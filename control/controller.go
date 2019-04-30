// Package control keeps track of resources and manages queries.
//
// The Controller manages the resources available to each query by
// managing the memory allocation and concurrency usage of each query.
// The Controller will compile a program by using the passed in language
// and it will start the program using the ResourceManager.
//
// It will guarantee that each program that is started has at least
// one goroutine that it can use with the dispatcher and it will
// ensure a minimum amount of memory is available before the program
// runs.
//
// Other goroutines and memory usage is at the will of the specific
// resource strategy that the Controller is using.
//
// The Controller also provides visibility into the lifetime of the query
// and its current resource usage.
package control

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/memory"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

// Controller provides a central location to manage all incoming queries.
// The controller is responsible for compiling, queueing, and executing queries.
type Controller struct {
	lastID     uint64
	queriesMu  sync.RWMutex
	queries    map[QueryID]*Query
	queryQueue chan *Query
	wg         sync.WaitGroup
	shutdown   bool
	done       chan struct{}
	abortOnce  sync.Once
	abort      chan struct{}

	memoryBytesQuotaPerQuery int64

	metrics   *controllerMetrics
	labelKeys []string

	logger *zap.Logger

	dependencies execute.Dependencies
}

type Config struct {
	// ConcurrencyQuota is the number of queries that are allowed to execute concurrently.
	ConcurrencyQuota int
	// MemoryBytesQuotaPerQuery is the maximum number of bytes (in table memory) a query is allowed to use at
	// any given time.
	//
	// The maximum amount of memory the controller is allowed to consume is
	//   ConcurrencyQuota * MemoryBytesQuotaPerQuery
	MemoryBytesQuotaPerQuery int64
	// QueueSize is the number of queries that are allowed to be awaiting execution before new queries are
	// rejected.
	QueueSize int
	Logger    *zap.Logger
	// MetricLabelKeys is a list of labels to add to the metrics produced by the controller.
	// The value for a given key will be read off the context.
	// The context value must be a string or an implementation of the Stringer interface.
	MetricLabelKeys []string

	ExecutorDependencies execute.Dependencies
}

func (c *Config) Validate() error {
	if c.ConcurrencyQuota <= 0 {
		return errors.New("ConcurrencyQuota must be positive")
	}
	if c.MemoryBytesQuotaPerQuery <= 0 {
		return errors.New("MemoryBytesQuotaPerQuery must be positive")
	}
	if c.QueueSize <= 0 {
		return errors.New("QueueSize must be positive")
	}
	return nil
}

type QueryID uint64

func New(c Config) (*Controller, error) {
	if err := c.Validate(); err != nil {
		return nil, errors.Wrap(err, "invalid controller config")
	}
	logger := c.Logger
	if logger == nil {
		logger = zap.NewNop()
	}
	logger.Info("Starting query controller",
		zap.Int("concurrency_quota", c.ConcurrencyQuota),
		zap.Int64("memory_bytes_quota_per_query", c.MemoryBytesQuotaPerQuery),
		zap.Int("queue_size", c.QueueSize))
	ctrl := &Controller{
		queries:                  make(map[QueryID]*Query),
		queryQueue:               make(chan *Query, c.QueueSize),
		done:                     make(chan struct{}),
		abort:                    make(chan struct{}),
		memoryBytesQuotaPerQuery: c.MemoryBytesQuotaPerQuery,
		logger:                   logger,
		metrics:                  newControllerMetrics(c.MetricLabelKeys),
		labelKeys:                c.MetricLabelKeys,
		dependencies:             c.ExecutorDependencies,
	}
	ctrl.wg.Add(c.ConcurrencyQuota)
	for i := 0; i < c.ConcurrencyQuota; i++ {
		go func() {
			defer ctrl.wg.Done()
			ctrl.processQueryQueue()
		}()
	}
	return ctrl, nil
}

// Query submits a query for execution returning immediately.
// Done must be called on any returned Query objects.
func (c *Controller) Query(ctx context.Context, compiler flux.Compiler) (flux.Query, error) {
	q, err := c.createQuery(ctx, compiler.CompilerType())
	if err != nil {
		return nil, err
	}

	if err := c.compileQuery(q, compiler); err != nil {
		q.setErr(err)
		c.finish(q)
		c.countQueryRequest(q, labelCompileError)
		return nil, q.Err()
	}
	if err := c.enqueueQuery(q); err != nil {
		q.setErr(err)
		c.finish(q)
		c.countQueryRequest(q, labelQueueError)
		return nil, q.Err()
	}
	c.countQueryRequest(q, labelSuccess)
	return q, nil
}

func (c *Controller) createQuery(ctx context.Context, ct flux.CompilerType) (*Query, error) {
	c.queriesMu.RLock()
	if c.shutdown {
		c.queriesMu.RUnlock()
		return nil, errors.New("query controller shutdown")
	}
	c.queriesMu.RUnlock()

	id := c.nextID()
	labelValues := make([]string, len(c.labelKeys))
	compileLabelValues := make([]string, len(c.labelKeys)+1)
	for i, k := range c.labelKeys {
		value := ctx.Value(k)
		var str string
		switch v := value.(type) {
		case string:
			str = v
		case fmt.Stringer:
			str = v.String()
		}
		labelValues[i] = str
		compileLabelValues[i] = str
	}
	compileLabelValues[len(compileLabelValues)-1] = string(ct)

	cctx, cancel := context.WithCancel(ctx)
	parentSpan, parentCtx := StartSpanFromContext(
		cctx,
		"all",
		c.metrics.allDur.WithLabelValues(labelValues...),
		c.metrics.all.WithLabelValues(labelValues...),
	)
	q := &Query{
		id:                 id,
		labelValues:        labelValues,
		compileLabelValues: compileLabelValues,
		state:              Created,
		c:                  c,
		results:            make(chan flux.Result),
		parentCtx:          parentCtx,
		parentSpan:         parentSpan,
		cancel:             cancel,
	}

	// Lock the queries mutex for the rest of this method.
	c.queriesMu.Lock()
	defer c.queriesMu.Unlock()

	if c.shutdown {
		// Query controller was shutdown between when we started
		// creating the query and ending it.
		err := errors.New("query controller shutdown")
		q.setErr(err)
		return nil, err
	}
	c.queries[id] = q
	return q, nil
}

func (c *Controller) nextID() QueryID {
	nextID := atomic.AddUint64(&c.lastID, 1)
	return QueryID(nextID)
}

func (c *Controller) countQueryRequest(q *Query, result requestsLabel) {
	l := len(q.labelValues)
	lvs := make([]string, l+1)
	copy(lvs, q.labelValues)
	lvs[l] = string(result)
	c.metrics.requests.WithLabelValues(lvs...).Inc()
}

func (c *Controller) compileQuery(q *Query, compiler flux.Compiler) error {
	ctx, ok := q.tryCompile()
	if !ok {
		return errors.New("failed to transition query to compiling state")
	}

	prog, err := compiler.Compile(ctx)
	if err != nil {
		return errors.Wrap(err, "compilation failed")
	}

	if p, ok := prog.(lang.DependenciesAwareProgram); ok {
		p.SetExecutorDependencies(c.dependencies)
		p.SetLogger(c.logger)
	}

	q.program = prog
	return nil
}

func (c *Controller) enqueueQuery(q *Query) error {
	if _, ok := q.tryQueue(); !ok {
		return errors.New("failed to transition query to queueing state")
	}

	select {
	case c.queryQueue <- q:
	default:
		return errors.New("queue length exceeded")
	}

	return nil
}

func (c *Controller) processQueryQueue() {
	for {
		select {
		case <-c.done:
			return
		case q := <-c.queryQueue:
			c.executeQuery(q)
		}
	}
}

func (c *Controller) executeQuery(q *Query) {
	ctx, ok := q.tryExec()
	if !ok {
		// This may happen if the query was cancelled (either because the
		// client cancelled it, or because the controller is shutting down)
		// In the case of cancellation, SetErr() should reset the error to an
		// appropriate message.
		q.setErr(errors.New("impossible state transition"))
		return
	}

	q.alloc = new(memory.Allocator)
	q.alloc.Limit = func(v int64) *int64 { return &v }(c.memoryBytesQuotaPerQuery)
	exec, err := q.program.Start(ctx, q.alloc)
	if err != nil {
		q.addRuntimeError(err)
		q.setErr(err)
		return
	}
	q.exec = exec
	q.pump(exec, ctx.Done())
}

func (c *Controller) finish(q *Query) {
	c.queriesMu.Lock()
	delete(c.queries, q.id)
	if len(c.queries) == 0 && c.shutdown {
		close(c.done)
	}
	c.queriesMu.Unlock()
}

// Queries reports the active queries.
func (c *Controller) Queries() []*Query {
	c.queriesMu.RLock()
	defer c.queriesMu.RUnlock()
	queries := make([]*Query, 0, len(c.queries))
	for _, q := range c.queries {
		queries = append(queries, q)
	}
	return queries
}

// Shutdown will signal to the Controller that it should not accept any
// new queries and that it should finish executing any existing queries.
// This will return once the Controller's run loop has been exited and all
// queries have been finished or until the Context has been canceled.
func (c *Controller) Shutdown(ctx context.Context) error {
	// Mark that the controller is shutdown so it does not
	// accept new queries.
	c.queriesMu.Lock()
	c.shutdown = true
	if len(c.queries) == 0 {
		c.queriesMu.Unlock()
		return nil
	}
	c.queriesMu.Unlock()

	// Cancel all of the currently active queries.
	c.queriesMu.RLock()
	for _, q := range c.queries {
		q.Cancel()
	}
	c.queriesMu.RUnlock()

	// Wait for query processing goroutines to finish.
	defer c.wg.Wait()

	// Wait for all of the queries to be cleaned up or until the
	// context is done.
	select {
	case <-c.done:
		return nil
	case <-ctx.Done():
		c.abortOnce.Do(func() {
			close(c.abort)
		})
		return ctx.Err()
	}
}

// PrometheusCollectors satisifies the prom.PrometheusCollector interface.
func (c *Controller) PrometheusCollectors() []prometheus.Collector {
	return c.metrics.PrometheusCollectors()
}

// Query represents a single request.
type Query struct {
	id QueryID

	labelValues        []string
	compileLabelValues []string

	c *Controller

	// query state. The stateMu protects access for the group below.
	stateMu     sync.RWMutex
	state       State
	err         error
	runtimeErrs []error
	cancel      func()

	parentCtx               context.Context
	parentSpan, currentSpan *span
	stats                   flux.Statistics

	done sync.Once

	program flux.Program
	exec    flux.Query
	results chan flux.Result
	alloc   *memory.Allocator
}

// ID reports an ephemeral unique ID for the query.
func (q *Query) ID() QueryID {
	return q.id
}

// Cancel will stop the query execution.
func (q *Query) Cancel() {
	// Call the cancel function to signal that execution should
	// be interrupted.
	q.cancel()
}

// Results returns a channel that will deliver the query results.
//
// It's possible that the channel is closed before any results arrive.
// In particular, if a query's context or the query itself is canceled,
// the query may close the results channel before any results are computed.
//
// The query may also have an error during execution so the Err()
// function should be used to check if an error happened.
func (q *Query) Results() <-chan flux.Result {
	return q.results
}

// Done signals to the Controller that this query is no longer
// being used and resources related to the query may be freed.
func (q *Query) Done() {
	// We are not considered to be in the run loop anymore once
	// this is called.
	q.done.Do(func() {
		if q.exec != nil {
			q.exec.Done()
			if q.err == nil {
				// TODO(jsternberg): The underlying program never returns
				// this so maybe their interface should change?
				q.err = q.exec.Err()
			}
			stats := q.exec.Statistics()
			q.stats.Metadata = stats.Metadata
		}

		errMsgs := make([]string, 0, len(q.runtimeErrs))
		for _, e := range q.runtimeErrs {
			errMsgs = append(errMsgs, e.Error())
		}
		q.stats.RuntimeErrors = errMsgs

		q.transitionTo(Finished)
		q.c.finish(q)
	})
}

// Statistics reports the statistics for the query.
//
// This method must be called after Done. It will block until
// the query has been finalized unless a context is given.
func (q *Query) Statistics() flux.Statistics {
	stats := q.stats
	if q.alloc != nil {
		stats.MaxAllocated = q.alloc.MaxAllocated()
	}
	return stats
}

// State reports the current state of the query.
func (q *Query) State() State {
	q.stateMu.RLock()
	state := q.state
	if !isFinishedState(state) {
		// If the query is a non-finished state, check the
		// context to see if we have been interrupted.
		select {
		case <-q.parentCtx.Done():
			// The query has been canceled so report to the
			// outside world that we have been canceled.
			// Do NOT attempt to change the internal state
			// variable here. It is a minefield. Leave the
			// normal query execution to figure that out.
			state = Canceled
		default:
			// The context has not been canceled.
		}
	}
	q.stateMu.RUnlock()
	return state
}

// transitionTo will transition from one state to another. If a list of current states
// is given, then the query must be in one of those states for the transition to succeed.
// This method must be called with a lock and it must be called from within the run loop.
func (q *Query) transitionTo(newState State, currentState ...State) (context.Context, bool) {
	// If we are transitioning to a non-finished state, the query
	// may have been canceled. If the query was canceled, then
	// we need to transition to the canceled state
	if !isFinishedState(newState) {
		select {
		case <-q.parentCtx.Done():
			// Transition to the canceled state and report that
			// we failed to transition to the desired state.
			_, _ = q.transitionTo(Canceled)
			return nil, false
		default:
		}
	}

	if len(currentState) > 0 {
		// Find the current state in the list of current states.
		for _, st := range currentState {
			if q.state == st {
				goto TRANSITION
			}
		}
		return nil, false
	}

TRANSITION:
	// We are transitioning to a new state. Close the current span (if it exists).
	if q.currentSpan != nil {
		q.currentSpan.Finish()
		switch q.state {
		case Compiling:
			q.stats.CompileDuration += q.currentSpan.Duration
		case Queueing:
			q.stats.QueueDuration += q.currentSpan.Duration
		case Executing:
			q.stats.ExecuteDuration += q.currentSpan.Duration
		}
	}
	q.currentSpan = nil

	if isFinishedState(newState) {
		// Invoke the cancel function to ensure that we have signaled that the query should be done.
		// The user is supposed to read the entirety of the tables returned before we end up in a finished
		// state, but user error may have caused this not to happen so there's no harm to canceling multiple
		// times.
		q.cancel()

		// If we are transitioning to a finished state from a non-finished state, finish the parent span.
		if q.parentSpan != nil {
			q.parentSpan.Finish()
			q.stats.TotalDuration = q.parentSpan.Duration
			q.parentSpan = nil
		}
	}

	// Transition to the new state.
	q.state = newState

	// Start a new span and set a new context.
	var (
		dur         *prometheus.HistogramVec
		gauge       *prometheus.GaugeVec
		labelValues = q.labelValues
	)
	switch newState {
	case Compiling:
		dur, gauge = q.c.metrics.compilingDur, q.c.metrics.compiling
		labelValues = q.compileLabelValues
	case Queueing:
		dur, gauge = q.c.metrics.queueingDur, q.c.metrics.queueing
	case Executing:
		dur, gauge = q.c.metrics.executingDur, q.c.metrics.executing
	default:
		// This state is not tracked so do not create a new span or context for it.
		// Use the parent context if one is needed.
		return q.parentCtx, true
	}
	var currentCtx context.Context
	q.currentSpan, currentCtx = StartSpanFromContext(
		q.parentCtx,
		newState.String(),
		dur.WithLabelValues(labelValues...),
		gauge.WithLabelValues(labelValues...),
	)
	return currentCtx, true
}

// Err reports any error the query may have encountered.
func (q *Query) Err() error {
	q.stateMu.Lock()
	err := q.err
	q.stateMu.Unlock()
	return err
}

// setErr marks this query with an error. If the query was
// canceled, then the error is ignored.
//
// This will mark the query as ready so setResults must not
// be called if this method is invoked.
func (q *Query) setErr(err error) {
	q.stateMu.Lock()
	defer q.stateMu.Unlock()

	// We may have this get called when the query is canceled.
	// If that is the case, transition to the canceled state
	// instead and record the error from that since the error
	// we received is probably wrong.
	select {
	case <-q.parentCtx.Done():
		q.transitionTo(Canceled)
		err = q.parentCtx.Err()
	default:
		q.transitionTo(Errored)
	}
	q.err = err

	// Close the ready channel to report that no results
	// will be sent.
	close(q.results)
}

func (q *Query) addRuntimeError(e error) {
	q.stateMu.Lock()
	defer q.stateMu.Unlock()

	q.runtimeErrs = append(q.runtimeErrs, e)
}

// pump will read from the executing query results and pump the
// results to our destination.
// When there are no more results, then this will close our own
// results channel.
func (q *Query) pump(exec flux.Query, done <-chan struct{}) {
	defer close(q.results)

	// When our context is canceled, we need to propagate that cancel
	// signal down to the executing program just in case it is waiting
	// for a cancel signal and is ignoring the passed in context.
	// We want this signal to only be sent once and we want to continue
	// draining the results until the underlying program has actually
	// been finished so we copy this to a new channel and set it to
	// nil when it has been closed.
	signalCh := done
	for {
		select {
		case res, ok := <-exec.Results():
			if !ok {
				return
			}

			// It is possible for the underlying query to misbehave.
			// We have to continue pumping results even if this is the
			// case, but if the query has been canceled or finished with
			// done, nobody is going to read these values so we need
			// to avoid blocking.
			ecr := &errorCollectingResult{
				Result: res,
				q:      q,
			}
			select {
			case <-done:
			case q.results <- ecr:
			}
		case <-signalCh:
			// Signal to the underlying executor that the query
			// has been canceled. Usually, the signal on the context
			// is likely enough, but this explicitly signals just in case.
			exec.Cancel()

			// Set the done channel to nil so we don't do this again
			// and we continue to drain the results.
			signalCh = nil
		case <-q.c.abort:
			// If we get here, then any running queries should have been cancelled
			// in controller.Shutdown().
			return
		}
	}
}

// tryCompile attempts to transition the query into the Compiling state.
func (q *Query) tryCompile() (context.Context, bool) {
	q.stateMu.Lock()
	defer q.stateMu.Unlock()

	return q.transitionTo(Compiling, Created)
}

// tryQueue attempts to transition the query into the Queueing state.
func (q *Query) tryQueue() (context.Context, bool) {
	q.stateMu.Lock()
	defer q.stateMu.Unlock()

	return q.transitionTo(Queueing, Compiling)
}

// tryExec attempts to transition the query into the Executing state.
func (q *Query) tryExec() (context.Context, bool) {
	q.stateMu.Lock()
	defer q.stateMu.Unlock()

	return q.transitionTo(Executing, Queueing)
}

type errorCollectingResult struct {
	flux.Result
	q *Query
}

func (r *errorCollectingResult) Tables() flux.TableIterator {
	return &errorCollectingTableIterator{
		TableIterator: r.Result.Tables(),
		q:             r.q,
	}
}

type errorCollectingTableIterator struct {
	flux.TableIterator
	q *Query
}

func (ti *errorCollectingTableIterator) Do(f func(t flux.Table) error) error {
	err := ti.TableIterator.Do(f)
	if err != nil {
		ti.q.addRuntimeError(err)
	}
	return err
}

// State is the query state.
type State int

const (
	// Created indicates the query has been created.
	Created State = iota

	// Compiling indicates that the query is in the process
	// of executing the compiler associated with the query.
	Compiling

	// Queueing indicates the query is waiting inside of the
	// scheduler to be executed.
	// TODO(jsternberg): This stage isn't used currently, but
	// it makes sense to readd this once we have a work queue again.
	Queueing

	// Executing indicates that the query is currently executing.
	Executing

	// Errored indicates that there was an error when attempting
	// to execute a query within any state inside of the controller.
	Errored

	// Finished indicates that the query has been marked as Done
	// and it is awaiting removal from the Controller or has already
	// been removed.
	Finished

	// Canceled indicates that the query was signaled to be
	// canceled. A canceled query must still be released with Done.
	Canceled
)

func (s State) String() string {
	switch s {
	case Created:
		return "created"
	case Compiling:
		return "compiling"
	case Queueing:
		return "queueing"
	case Executing:
		return "executing"
	case Errored:
		return "errored"
	case Finished:
		return "finished"
	case Canceled:
		return "canceled"
	default:
		return "unknown"
	}
}

func isFinishedState(state State) bool {
	switch state {
	case Canceled, Errored, Finished:
		return true
	default:
		return false
	}
}

// span is a simple wrapper around opentracing.Span in order to
// get access to the duration of the span for metrics reporting.
type span struct {
	s        opentracing.Span
	start    time.Time
	Duration time.Duration
	hist     prometheus.Observer
	gauge    prometheus.Gauge
}

func StartSpanFromContext(ctx context.Context, operationName string, hist prometheus.Observer, gauge prometheus.Gauge) (*span, context.Context) {
	start := time.Now()
	s, sctx := opentracing.StartSpanFromContext(ctx, operationName, opentracing.StartTime(start))
	gauge.Inc()
	return &span{
		s:     s,
		start: start,
		hist:  hist,
		gauge: gauge,
	}, sctx
}

func (s *span) Finish() {
	finish := time.Now()
	s.Duration = finish.Sub(s.start)
	s.s.FinishWithOptions(opentracing.FinishOptions{
		FinishTime: finish,
	})
	s.hist.Observe(s.Duration.Seconds())
	s.gauge.Dec()
}
