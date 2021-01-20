package execute

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"

	"github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/jaeger"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
)

// Transport is an interface for handling raw messages.
type Transport interface {
	// ProcessMessage will process a message in the Transport.
	ProcessMessage(m Message) error
}

// AsyncTransport is a Transport that performs its work in a separate goroutine.
type AsyncTransport interface {
	Transport
	// Finished reports when the AsyncTransport has completed and there is no more work to do.
	Finished() <-chan struct{}
}

var _ Transformation = (*consecutiveTransport)(nil)

// consecutiveTransport implements AsyncTransport by transporting data consecutively to the downstream Transformation.
type consecutiveTransport struct {
	ctx        context.Context
	dispatcher Dispatcher
	logger     *zap.Logger

	t         Transport
	messages  MessageQueue
	op, label string
	stack     []interpreter.StackEntry

	finished chan struct{}
	errMu    sync.Mutex
	errValue error

	schedulerState int32
	inflight       int32
}

func newConsecutiveTransport(ctx context.Context, dispatcher Dispatcher, t Transformation, n plan.Node, logger *zap.Logger, mem memory.Allocator) *consecutiveTransport {
	return &consecutiveTransport{
		ctx:        ctx,
		dispatcher: dispatcher,
		logger:     logger,
		t:          WrapTransformationInTransport(t, mem),
		// TODO(nathanielc): Have planner specify message queue initial buffer size.
		messages: newMessageQueue(64),
		op:       OperationType(t),
		label:    string(n.ID()),
		stack:    n.CallStack(),
		finished: make(chan struct{}),
	}
}

func (t *consecutiveTransport) sourceInfo() string {
	if len(t.stack) == 0 {
		return ""
	}

	// Learn the filename from the bottom of the stack.
	// We want the top most entry (deepest in the stack)
	// from the primary file. We can retrieve the filename
	// for the primary file by looking at the bottom of the
	// stack and then finding the top-most entry with that
	// filename.
	filename := t.stack[len(t.stack)-1].Location.File
	for i := 0; i < len(t.stack); i++ {
		entry := t.stack[i]
		if entry.Location.File == filename {
			return fmt.Sprintf("@%s: %s", entry.Location, entry.FunctionName)
		}
	}
	entry := t.stack[0]
	return fmt.Sprintf("@%s: %s", entry.Location, entry.FunctionName)
}
func (t *consecutiveTransport) setErr(err error) {
	t.errMu.Lock()
	msg := "runtime error"
	if srcInfo := t.sourceInfo(); srcInfo != "" {
		msg += " " + srcInfo
	}
	err = errors.Wrap(err, codes.Inherit, msg)
	t.errValue = err
	t.errMu.Unlock()
}
func (t *consecutiveTransport) err() error {
	t.errMu.Lock()
	err := t.errValue
	t.errMu.Unlock()
	return err
}

func (t *consecutiveTransport) Finished() <-chan struct{} {
	return t.finished
}

func (t *consecutiveTransport) RetractTable(id DatasetID, key flux.GroupKey) error {
	select {
	case <-t.finished:
		return t.err()
	default:
	}
	t.pushMsg(&retractTableMsg{
		srcMessage: srcMessage(id),
		key:        key,
	})
	return nil
}

func (t *consecutiveTransport) Process(id DatasetID, tbl flux.Table) error {
	select {
	case <-t.finished:
		return t.err()
	default:
	}
	t.pushMsg(&processMsg{
		srcMessage: srcMessage(id),
		table:      newConsecutiveTransportTable(t, tbl),
	})
	return nil
}

func (t *consecutiveTransport) UpdateWatermark(id DatasetID, time Time) error {
	select {
	case <-t.finished:
		return t.err()
	default:
	}
	t.pushMsg(&updateWatermarkMsg{
		srcMessage: srcMessage(id),
		time:       time,
	})
	return nil
}

func (t *consecutiveTransport) UpdateProcessingTime(id DatasetID, time Time) error {
	select {
	case <-t.finished:
		return t.err()
	default:
	}
	t.pushMsg(&updateProcessingTimeMsg{
		srcMessage: srcMessage(id),
		time:       time,
	})
	return nil
}

func (t *consecutiveTransport) Finish(id DatasetID, err error) {
	select {
	case <-t.finished:
		return
	default:
	}
	t.pushMsg(&finishMsg{
		srcMessage: srcMessage(id),
		err:        err,
	})
}

func (t *consecutiveTransport) pushMsg(m Message) {
	t.messages.Push(m)
	atomic.AddInt32(&t.inflight, 1)
	t.schedule()
}

func (t *consecutiveTransport) ProcessMessage(m Message) error {
	t.pushMsg(m)
	return nil
}

const (
	// consecutiveTransport schedule states
	idle int32 = iota
	running
	finished
)

// schedule indicates that there is work available to schedule.
func (t *consecutiveTransport) schedule() {
	if t.tryTransition(idle, running) {
		t.dispatcher.Schedule(t.processMessages)
	}
}

// tryTransition attempts to transition into the new state and returns true on success.
func (t *consecutiveTransport) tryTransition(old, new int32) bool {
	return atomic.CompareAndSwapInt32(&t.schedulerState, old, new)
}

// transition sets the new state.
func (t *consecutiveTransport) transition(new int32) {
	atomic.StoreInt32(&t.schedulerState, new)
}

func (t *consecutiveTransport) processMessages(ctx context.Context, throughput int) {
PROCESS:
	i := 0
	for m := t.messages.Pop(); m != nil; m = t.messages.Pop() {
		atomic.AddInt32(&t.inflight, -1)
		if f, err := t.processMessage(ctx, m); err != nil || f {
			// Set the error if there was any
			t.setErr(err)

			// Transition to the finished state.
			if t.tryTransition(running, finished) {
				// Call Finish if we have not already
				if !f {
					m := &finishMsg{
						srcMessage: srcMessage(m.SrcDatasetID()),
						err:        t.err(),
					}
					_ = t.t.ProcessMessage(m)
				}
				// We are finished
				close(t.finished)
				return
			}
		}
		i++
		if i >= throughput {
			// We have done enough work.
			// Transition to the idle state and reschedule for later.
			t.transition(idle)
			t.schedule()
			return
		}
	}

	t.transition(idle)
	// Check if more messages arrived after the above loop finished.
	// This check must happen in the idle state.
	if atomic.LoadInt32(&t.inflight) > 0 {
		if t.tryTransition(idle, running) {
			goto PROCESS
		} // else we have already been scheduled again, we can return
	}
}

// processMessage processes the message on t.
// The return value is true if the message was a FinishMsg.
func (t *consecutiveTransport) processMessage(ctx context.Context, m Message) (finished bool, err error) {
	if _, span := StartSpanFromContext(ctx, t.op, t.label); span != nil {
		setMessageTags(span, m)
		defer span.Finish()
	}
	if err := t.t.ProcessMessage(m); err != nil {
		return false, err
	}
	finished = isFinishMessage(m)
	return finished, nil
}

// Message is a message sent from one Dataset to another.
type Message interface {
	// Type returns the MessageType for this Message.
	Type() MessageType

	// SrcDatasetID is the DatasetID that produced this Message.
	SrcDatasetID() DatasetID

	// Ack is used to acknowledge that the Message was received
	// and terminated. A Message may be passed between various
	// Transport implementations. When the Ack is received,
	// this signals to the Message to release any memory it may
	// have retained.
	Ack()

	// Dup is used to duplicate the Message.
	// This is useful when the Message has to be sent to multiple
	// receivers from a single sender.
	Dup() Message

	// SetTags will set the tags on the span.
	SetTags(span opentracing.Span)
}

type MessageType int

const (
	// RetractTableType is sent when the previous table for
	// a given group key should be retracted.
	RetractTableType MessageType = iota

	// ProcessType is sent when there is an entire flux.Table
	// ready to be processed from the upstream Dataset.
	ProcessType

	// UpdateWatermarkType is sent when there will be no more
	// points older than the watermark for any key.
	UpdateWatermarkType

	// UpdateProcessingTimeType is sent to update the current time.
	UpdateProcessingTimeType

	// FinishType is sent when there are no more messages from
	// the upstream Dataset or an upstream error occurred that
	// caused the execution to abort.
	FinishType

	// ProcessViewType is sent when a new table.View is ready to
	// be processed from the upstream Dataset.
	ProcessViewType

	// FlushKeyType is sent when the upstream Dataset wishes
	// to flush the data associated with a key presently stored
	// in the Dataset.
	FlushKeyType

	// WatermarkKeyType is sent when the upstream Dataset will send
	// no more rows with a time older than the time in the watermark
	// for the given key.
	WatermarkKeyType
)

func (m MessageType) String() string {
	switch m {
	case RetractTableType:
		return "RetractTableType"
	case ProcessType:
		return "ProcessType"
	case UpdateWatermarkType:
		return "UpdateWatermarkType"
	case UpdateProcessingTimeType:
		return "UpdateProcessingTimeType"
	case FinishType:
		return "FinishType"
	case ProcessViewType:
		return "ProcessViewType"
	case FlushKeyType:
		return "FlushKeyType"
	case WatermarkKeyType:
		return "WatermarkKeyType"
	default:
		return "UnknownMessageType"
	}
}

// setMessageTags sets the tags from a Message on the opentracing.Span.
func setMessageTags(span opentracing.Span, m Message) {
	span.SetTag("messageType", m.Type().String())
	m.SetTags(span)
}

type srcMessage DatasetID

func (m srcMessage) SrcDatasetID() DatasetID {
	return DatasetID(m)
}
func (m srcMessage) Ack() {}
func (m srcMessage) SetTags(span opentracing.Span) {
	span.SetTag("dataset", DatasetID(m).String())
}

type RetractTableMsg interface {
	Message
	Key() flux.GroupKey
}

type retractTableMsg struct {
	srcMessage
	key flux.GroupKey
}

func (m *retractTableMsg) Type() MessageType {
	return RetractTableType
}
func (m *retractTableMsg) Key() flux.GroupKey {
	return m.key
}
func (m *retractTableMsg) Dup() Message {
	return m
}
func (m *retractTableMsg) SetTags(span opentracing.Span) {
	m.srcMessage.SetTags(span)
	span.SetTag("key", m.key.String())
}

type ProcessMsg interface {
	Message
	Table() flux.Table
}

type processMsg struct {
	srcMessage
	table flux.Table
}

func (m *processMsg) Type() MessageType {
	return ProcessType
}
func (m *processMsg) Table() flux.Table {
	return m.table
}
func (m *processMsg) Ack() {
	m.table.Done()
}
func (m *processMsg) Dup() Message {
	cpy, _ := table.Copy(m.table)
	m.table = cpy.Copy()

	dup := *m
	dup.table = cpy
	return &dup
}
func (m *processMsg) SetTags(span opentracing.Span) {
	m.srcMessage.SetTags(span)
	span.SetTag("key", m.table.Key().String())
}

type UpdateWatermarkMsg interface {
	Message
	WatermarkTime() Time
}

type updateWatermarkMsg struct {
	srcMessage
	time Time
}

func (m *updateWatermarkMsg) Type() MessageType {
	return UpdateWatermarkType
}
func (m *updateWatermarkMsg) WatermarkTime() Time {
	return m.time
}
func (m *updateWatermarkMsg) Dup() Message {
	return m
}
func (m *updateWatermarkMsg) SetTags(span opentracing.Span) {
	m.srcMessage.SetTags(span)
	span.SetTag("time", int64(m.time))
}

type UpdateProcessingTimeMsg interface {
	Message
	ProcessingTime() Time
}

type updateProcessingTimeMsg struct {
	srcMessage
	time Time
}

func (m *updateProcessingTimeMsg) Type() MessageType {
	return UpdateProcessingTimeType
}
func (m *updateProcessingTimeMsg) ProcessingTime() Time {
	return m.time
}
func (m *updateProcessingTimeMsg) Dup() Message {
	return m
}
func (m *updateProcessingTimeMsg) SetTags(span opentracing.Span) {
	m.srcMessage.SetTags(span)
	span.SetTag("time", int64(m.time))
}

type FinishMsg interface {
	Message
	Error() error
}

type finishMsg struct {
	srcMessage
	err error
}

func (m *finishMsg) Type() MessageType {
	return FinishType
}
func (m *finishMsg) Error() error {
	return m.err
}
func (m *finishMsg) Dup() Message {
	return m
}
func (m *finishMsg) SetTags(span opentracing.Span) {
	m.srcMessage.SetTags(span)
	if m.err != nil {
		span.SetTag("error", m.err.Error())
	}
}

type ProcessViewMsg interface {
	Message
	View() table.View
}

type FlushKeyMsg interface {
	Message
	Key() flux.GroupKey
}

type WatermarkKeyMsg interface {
	Message
	ColumnName() string
	Time() int64
	Key() flux.GroupKey
}

// transformationTransportAdapter will translate Message values sent to
// a Transport to an underlying Transformation.
type transformationTransportAdapter struct {
	t     Transformation
	cache table.BuilderCache
}

// WrapTransformationInTransport will wrap a Transformation into
// a Transport to be used for the execution engine.
func WrapTransformationInTransport(t Transformation, mem memory.Allocator) Transport {
	// If the Transformation implements the Transport interface,
	// then we can just use that directly.
	if tr, ok := t.(Transport); ok {
		return tr
	}
	return &transformationTransportAdapter{
		t: t,
		cache: table.BuilderCache{
			New: func(key flux.GroupKey) table.Builder {
				return table.NewBufferedBuilder(key, mem)
			},
		},
	}
}

func (t *transformationTransportAdapter) ProcessMessage(m Message) error {
	switch m.Type() {
	case RetractTableType:
		m := m.(RetractTableMsg)
		return t.t.RetractTable(m.SrcDatasetID(), m.Key())
	case ProcessType:
		m := m.(ProcessMsg)
		return t.t.Process(m.SrcDatasetID(), m.Table())
	case UpdateWatermarkType:
		m := m.(UpdateWatermarkMsg)
		return t.t.UpdateWatermark(m.SrcDatasetID(), m.WatermarkTime())
	case UpdateProcessingTimeType:
		m := m.(UpdateProcessingTimeMsg)
		return t.t.UpdateProcessingTime(m.SrcDatasetID(), m.ProcessingTime())
	case FinishType:
		m := m.(FinishMsg)

		// If there are pending buffers that were never flushed,
		// do that here.
		if err := t.cache.ForEach(func(key flux.GroupKey, builder table.Builder) error {
			table, err := builder.Table()
			if err != nil {
				return err
			}
			return t.t.Process(m.SrcDatasetID(), table)
		}); err != nil {
			return err
		}
		t.t.Finish(m.SrcDatasetID(), m.Error())
		return nil
	case ProcessViewType:
		defer m.Ack()
		m := m.(ProcessViewMsg)

		// Retrieve the buffered builder and append the
		// table view to it. The view is implemented using
		// arrow.TableBuffer which is compatible with
		// flux.ColReader so we can append it directly.
		b, _ := table.GetBufferedBuilder(m.View().Key(), &t.cache)
		buffer := m.View().Buffer()
		return b.AppendBuffer(&buffer)
	case FlushKeyType:
		defer m.Ack()
		m := m.(FlushKeyMsg)

		// Retrieve the buffered builder for the given key
		// and send the data to the next transformation.
		tbl, err := t.cache.Table(m.Key())
		if err != nil {
			return err
		}
		t.cache.ExpireTable(m.Key())
		return t.t.Process(m.SrcDatasetID(), tbl)
	default:
		// Message is not handled by older Transformation implementations.
		m.Ack()
		return nil
	}
}

func (t *transformationTransportAdapter) OperationType() string {
	return OperationType(t.t)
}

// isFinishMessage will return true if the Message is a FinishMsg.
func isFinishMessage(m Message) bool {
	_, ok := m.(FinishMsg)
	return ok
}

// OperationType returns a string representation of the transformation
// operation represented by the Transport.
func OperationType(t interface{}) string {
	if t, ok := t.(interface {
		OperationType() string
	}); ok {
		return t.OperationType()
	}
	return reflect.TypeOf(t).String()
}

// consecutiveTransportTable is a flux.Table that is being processed
// within a consecutiveTransport.
type consecutiveTransportTable struct {
	transport *consecutiveTransport
	tbl       flux.Table
}

func newConsecutiveTransportTable(t *consecutiveTransport, tbl flux.Table) flux.Table {
	return &consecutiveTransportTable{
		transport: t,
		tbl:       tbl,
	}
}

func (t *consecutiveTransportTable) Key() flux.GroupKey {
	return t.tbl.Key()
}

func (t *consecutiveTransportTable) Cols() []flux.ColMeta {
	return t.tbl.Cols()
}

func (t *consecutiveTransportTable) Do(f func(flux.ColReader) error) error {
	return t.tbl.Do(func(cr flux.ColReader) error {
		if err := t.validate(cr); err != nil {
			fields := []zap.Field{
				zap.String("source", t.transport.sourceInfo()),
				zap.Error(err),
			}

			ctx, logger := t.transport.ctx, t.transport.logger
			if span := opentracing.SpanFromContext(ctx); span != nil {
				if traceID, sampled, found := jaeger.InfoFromSpan(span); found {
					fields = append(fields,
						zap.String("tracing/id", traceID),
						zap.Bool("tracing/sampled", sampled),
					)
				}
			}
			logger.Info("Invalid column reader received from predecessor", fields...)
		}
		return f(cr)
	})
}

func (t *consecutiveTransportTable) Done() {
	t.tbl.Done()
}

func (t *consecutiveTransportTable) Empty() bool {
	return t.tbl.Empty()
}

func (t *consecutiveTransportTable) validate(cr flux.ColReader) error {
	if len(cr.Cols()) == 0 {
		return nil
	}

	sz := table.Values(cr, 0).Len()
	for i, n := 1, len(cr.Cols()); i < n; i++ {
		nsz := table.Values(cr, i).Len()
		if sz != nsz {
			// Mismatched column lengths.
			// Look at all column lengths so we can give a more complete
			// error message.
			// We avoid this in the usual case to avoid allocating an array
			// of lengths for every table when it might not be needed.
			lens := make(map[string]int, len(cr.Cols()))
			for i, col := range cr.Cols() {
				label := fmt.Sprintf("%s:%s", col.Label, col.Type)
				lens[label] = table.Values(cr, i).Len()
			}
			return errors.Newf(codes.Internal, "mismatched column lengths: %v", lens)
		}
	}
	return nil
}
