// Package control controls which resources a query may consume.
//
// The Controller manages the resources available to each query and ensures
// an optimal use of those resources to execute queries in a timely manner.
// The controller also maintains the state of a query as it goes through the
// various stages of execution and is responsible for killing currently
// executing queries when requested by the user.
//
// The Controller manages when a query is executed. This can be based on
// anything within the query's requested resources. For example, a basic
// implementation of the Controller may decide to execute anything with a high
// priority before anything with a low priority.  The implementation of the
// Controller will vary and change over time and this package may provide
// multiple implementations for different controller algorithms.
//
// During execution, the Controller manages the resources used by the query and
// provides observabiility into what resources are being used and by which
// queries. The Controller also imposes limitations so a query that uses more
// than its allocated resources or more resources than available on the system
// will be aborted.
package control

import (
	"context"
	"fmt"
	"math"
	"runtime/debug"
	"sync"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Controller provides a central location to manage all incoming queries.
// The controller is responsible for queueing, planning, and executing queries.
type Controller struct {
	newQueries    chan *Query
	queriesMu     sync.RWMutex
	queries       map[QueryID]*Query
	queryDone     chan *Query
	cancelRequest chan QueryID

	shutdownCtx context.Context
	shutdown    func()
	done        chan struct{}

	metrics   *controllerMetrics
	labelKeys []string

	planner  plan.Planner
	executor execute.Executor
	logger   *zap.Logger

	maxConcurrency       int
	availableConcurrency int
	availableMemory      int64
}

type Config struct {
	ConcurrencyQuota     int
	MemoryBytesQuota     int64
	ExecutorDependencies execute.Dependencies
	PPlannerOptions      []plan.PhysicalOption
	LPlannerOptions      []plan.LogicalOption
	Logger               *zap.Logger
	// MetricLabelKeys is a list of labels to add to the metrics produced by the controller.
	// The value for a given key will be read off the context.
	// The context value must be a string or an implementation of the Stringer interface.
	MetricLabelKeys []string
}

type QueryID uint64

func New(c Config) *Controller {
	logger := c.Logger
	if logger == nil {
		logger = zap.NewNop()
	}
	pb := plan.PlannerBuilder{}
	pb.AddLogicalOptions(c.LPlannerOptions...)
	pb.AddPhysicalOption(c.PPlannerOptions...)
	ctrl := &Controller{
		newQueries:           make(chan *Query),
		queries:              make(map[QueryID]*Query),
		queryDone:            make(chan *Query),
		cancelRequest:        make(chan QueryID),
		done:                 make(chan struct{}),
		maxConcurrency:       c.ConcurrencyQuota,
		availableConcurrency: c.ConcurrencyQuota,
		availableMemory:      c.MemoryBytesQuota,
		planner:              pb.Build(),
		executor:             execute.NewExecutor(c.ExecutorDependencies, logger),
		logger:               logger,
		metrics:              newControllerMetrics(c.MetricLabelKeys),
		labelKeys:            c.MetricLabelKeys,
	}
	ctrl.shutdownCtx, ctrl.shutdown = context.WithCancel(context.Background())
	go ctrl.run()
	return ctrl
}

// Query submits a query for execution returning immediately.
// Done must be called on any returned Query objects.
func (c *Controller) Query(ctx context.Context, compiler flux.Compiler) (flux.Query, error) {
	prog, err := compiler.Compile(ctx)
	if err != nil {
		return nil, err
	}

	a := new(memory.Allocator)
	return prog.Start(ctx, a)
}

type Stringer interface {
	String() string
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
	// Initiate the shutdown procedure by signaling to the run thread.
	c.shutdown()

	// Wait for the run loop to exit.
	select {
	case <-c.done:
		return nil
	case <-ctx.Done():
		c.CancelAll()
		return ctx.Err()
	}
}

// CancelAll cancels all executing queries.
func (c *Controller) CancelAll() {
	c.queriesMu.RLock()
	for _, q := range c.queries {
		q.Cancel()
	}
	c.queriesMu.RUnlock()
}

func (c *Controller) run() {
	defer close(c.done)

	pq := newPriorityQueue()
	for {
		select {
		// Wait for resources to free
		case q := <-c.queryDone:
			c.free(q)
			c.queriesMu.Lock()
			delete(c.queries, q.id)
			c.queriesMu.Unlock()
		// Wait for new queries
		case q := <-c.newQueries:
			pq.Push(q)
			c.queriesMu.Lock()
			c.queries[q.id] = q
			c.queriesMu.Unlock()
		// Wait for cancel query requests
		case id := <-c.cancelRequest:
			c.queriesMu.RLock()
			q := c.queries[id]
			c.queriesMu.RUnlock()
			q.Cancel()
		// Check if we have been signaled to shutdown.
		case <-c.shutdownCtx.Done():
			// We have been signaled to shutdown so drain the queues
			// and exit the for loop.
			c.drain(pq)
			return
		}

		// Peek at head of priority queue
		q := pq.Peek()
		if q != nil {
			pop, err := c.processQuery(q)
			if pop {
				pq.Pop()
			}
			if err != nil {
				q.setErr(err)
			}
		}
	}
}

// drain will continue processing queries from the priority queue and
// processing done queries.
func (c *Controller) drain(pq *PriorityQueue) {
	for {
		c.queriesMu.RLock()
		if len(c.queries) == 0 {
			c.queriesMu.RUnlock()
			return
		}
		c.queriesMu.RUnlock()

		// Wait for resources to free
		q := <-c.queryDone
		c.free(q)
		c.queriesMu.Lock()
		delete(c.queries, q.id)
		c.queriesMu.Unlock()

		// Peek at head of priority queue
		q = pq.Peek()
		if q != nil {
			pop, err := c.processQuery(q)
			if pop {
				pq.Pop()
			}
			if err != nil {
				go q.setErr(err)
			}
		}
	}
}

// processQuery move the query through the state machine and returns and errors and if the query should be popped.
func (c *Controller) processQuery(q *Query) (pop bool, err error) {
	defer func() {
		if e := recover(); e != nil {
			// If a query panicked, always pop it from the queue so we don't
			// try to reprocess it.
			pop = true

			// Update the error with information about the query if this is an
			// error type and create an error if it isn't.
			switch e := e.(type) {
			case error:
				err = errors.Wrap(e, "panic")
			default:
				err = fmt.Errorf("panic: %s", e)
			}
			if entry := c.logger.Check(zapcore.InfoLevel, "Controller panic"); entry != nil {
				entry.Stack = string(debug.Stack())
				entry.Write(zap.Error(err))
			}
		}
	}()

	// Check if we have enough resources
	if c.check(q) {
		// Update resource gauges
		c.consume(q)

		// Remove the query from the queue
		pop = true

		// Execute query
		if !q.tryExec() {
			return true, errors.New("failed to transition query into executing state")
		}
		q.alloc = new(memory.Allocator)
		// TODO: pass the plan to the executor here
		r, md, err := c.executor.Execute(q.currentCtx, q.plan, q.alloc)
		if err != nil {
			return true, errors.Wrap(err, "failed to execute query")
		}
		q.setResults(r, md)
	} else {
		// update state to queueing
		if !q.tryRequeue() {
			return true, errors.New("failed to transition query into requeueing state")
		}
	}
	return pop, nil
}

func (c *Controller) check(q *Query) bool {
	return c.availableConcurrency >= q.concurrency && (q.memory == math.MaxInt64 || c.availableMemory >= q.memory)
}
func (c *Controller) consume(q *Query) {
	c.availableConcurrency -= q.concurrency

	if q.memory != math.MaxInt64 {
		c.availableMemory -= q.memory
	}
}

func (c *Controller) free(q *Query) {
	c.availableConcurrency += q.concurrency

	if q.memory != math.MaxInt64 {
		c.availableMemory += q.memory
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

	spec flux.Spec

	ready  chan flux.Result
	metaCh <-chan flux.Metadata

	// query state. The stateMu protects access for the group below.
	stateMu sync.RWMutex
	state   State
	err     error
	cancel  func()

	parentCtx, currentCtx   context.Context
	parentSpan, currentSpan *span
	stats                   flux.Statistics

	plan *plan.Spec

	done        sync.Once
	concurrency int
	memory      int64

	alloc *memory.Allocator
}

// ID reports an ephemeral unique ID for the query.
func (q *Query) ID() QueryID {
	return q.id
}

func (q *Query) Spec() *flux.Spec {
	return &q.spec
}

// Concurrency reports the number of goroutines allowed to process the request.
func (q *Query) Concurrency() int {
	return q.concurrency
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
	return q.ready
}

// Done signals to the Controller that this query is no longer
// being used and resources related to the query may be freed.
//
// The Results method must have returned a result before calling
// this method either by the query executing, being canceled, or
// an error occurring.
func (q *Query) Done() {
	// We are not considered to be in the run loop anymore once
	// this is called.
	q.done.Do(func() {
		q.stateMu.Lock()
		q.transitionTo(Finished)

		// Read from the metadata channel. We only do this in this location so no locking
		// is needed.
		meta := make(flux.Metadata)
		for md := range q.metaCh {
			meta.AddAll(md)
		}
		q.stats.Metadata = meta
		q.stateMu.Unlock()

		q.c.queryDone <- q
	})
}

// Statistics reports the statistics for the query.
//
// This method must be called after Done. It will block until
// the query has been finalized unless a context is given.
func (q *Query) Statistics() flux.Statistics {
	q.stateMu.Lock()
	defer q.stateMu.Unlock()

	stats := q.stats
	stats.Concurrency = q.concurrency
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
func (q *Query) transitionTo(newState State, currentState ...State) bool {
	// If we are transitioning to a non-finished state, the query
	// may have been canceled. If the query was canceled, then
	// we need to transition to the canceled state
	if !isFinishedState(newState) {
		select {
		case <-q.parentCtx.Done():
			// Transition to the canceled state and report that
			// we failed to transition to the desired state.
			_ = q.transitionTo(Canceled)
			return false
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
		return false
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
		case Planning:
			q.stats.PlanDuration += q.currentSpan.Duration
		case Requeueing:
			q.stats.RequeueDuration += q.currentSpan.Duration
		case Executing:
			q.stats.ExecuteDuration += q.currentSpan.Duration
		}
	}
	q.currentSpan, q.currentCtx = nil, nil

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
	case Planning:
		dur, gauge = q.c.metrics.planningDur, q.c.metrics.planning
	case Requeueing:
		dur, gauge = q.c.metrics.requeueingDur, q.c.metrics.requeueing
	case Executing:
		dur, gauge = q.c.metrics.executingDur, q.c.metrics.executing
	default:
		// This state is not tracked so do not create a new span or context for it.
		return true
	}
	q.currentSpan, q.currentCtx = StartSpanFromContext(
		q.parentCtx,
		newState.String(),
		dur.WithLabelValues(labelValues...),
		gauge.WithLabelValues(labelValues...),
	)
	return true
}

func (q *Query) isOK() bool {
	q.stateMu.RLock()
	ok := q.state != Canceled && q.state != Errored && q.state != Finished
	q.stateMu.RUnlock()
	return ok
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
	close(q.ready)
}

// setResults will set the results and send them over the ready channel.
func (q *Query) setResults(_ map[string]flux.Result, md <-chan flux.Metadata) {
	q.stateMu.Lock()
	q.metaCh = md
	q.stateMu.Unlock()

	//q.ready <- r
	close(q.ready)
}

// tryRequeue attempts to transition the query into the Requeueing state.
func (q *Query) tryRequeue() bool {
	q.stateMu.Lock()
	defer q.stateMu.Unlock()

	if q.state == Requeueing {
		// Already in the correct state.
		return true
	}
	return q.transitionTo(Requeueing, Queueing)
}

// tryExec attempts to transition the query into the Executing state.
func (q *Query) tryExec() bool {
	q.stateMu.Lock()
	defer q.stateMu.Unlock()

	return q.transitionTo(Executing, Requeueing, Queueing)
}

// State is the query state.
type State int

const (
	// Created indicates the query has been created.
	Created State = iota

	// Compiling indicates that the query is in the process
	// of executing the compiler associated with the query.
	Compiling

	// Planning indicates that a query spec has been created
	// from the compiler and the query planner is executing.
	Planning

	// Queueing indicates the query is waiting inside of the
	// scheduler to be executed.
	Queueing

	// Requeueing indicates that the query scheduler wanted
	// to run the query, but not enough resources were available
	// so it is in the process of waiting again.
	Requeueing

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
	case Planning:
		return "planning"
	case Requeueing:
		return "requeueing"
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

var noMetadata = make(chan flux.Metadata)

func init() {
	close(noMetadata)
}
