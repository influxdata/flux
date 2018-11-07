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
	"log"
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
	newQueries chan *Query
	lastID     QueryID
	queriesMu  sync.RWMutex
	queries    map[QueryID]*Query
	queryDone  chan *Query

	metrics   *controllerMetrics
	labelKeys []string

	verbose bool

	lplanner plan.LogicalPlanner
	pplanner plan.PhysicalPlanner
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
	Verbose              bool
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
	ctrl := &Controller{
		newQueries:           make(chan *Query),
		queries:              make(map[QueryID]*Query),
		queryDone:            make(chan *Query),
		maxConcurrency:       c.ConcurrencyQuota,
		availableConcurrency: c.ConcurrencyQuota,
		availableMemory:      c.MemoryBytesQuota,
		lplanner:             plan.NewLogicalPlanner(c.LPlannerOptions...),
		pplanner:             plan.NewPhysicalPlanner(c.PPlannerOptions...),
		executor:             execute.NewExecutor(c.ExecutorDependencies, logger),
		logger:               logger,
		verbose:              c.Verbose,
		metrics:              newControllerMetrics(c.MetricLabelKeys),
		labelKeys:            c.MetricLabelKeys,
	}
	go ctrl.run()
	return ctrl
}

// Query submits a query for execution returning immediately.
// Done must be called on any returned Query objects.
func (c *Controller) Query(ctx context.Context, compiler flux.Compiler) (flux.Query, error) {
	q := c.createQuery(ctx, compiler.CompilerType())
	if err := c.compileQuery(q, compiler); err != nil {
		q.transitionTo(Errored)
		return nil, err
	}
	if err := c.enqueueQuery(q); err != nil {
		q.transitionTo(Errored)
		return nil, err
	}
	return q, nil
}

type Stringer interface {
	String() string
}

func (c *Controller) createQuery(ctx context.Context, ct flux.CompilerType) *Query {
	id := c.nextID()
	labelValues := make([]string, len(c.labelKeys))
	compileLabelValues := make([]string, len(c.labelKeys)+1)
	for i, k := range c.labelKeys {
		value := ctx.Value(k)
		var str string
		switch v := value.(type) {
		case string:
			str = v
		case Stringer:
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
	ready := make(chan map[string]flux.Result, 1)
	return &Query{
		id:                 id,
		labelValues:        labelValues,
		compileLabelValues: compileLabelValues,
		state:              Created,
		c:                  c,
		now:                time.Now().UTC(),
		ready:              ready,
		parentCtx:          parentCtx,
		parentSpan:         parentSpan,
		cancel:             cancel,
	}
}

func (c *Controller) compileQuery(q *Query, compiler flux.Compiler) error {
	if !q.tryCompile() {
		return errors.New("failed to transition query to compiling state")
	}
	spec, err := compiler.Compile(q.currentCtx)
	if err != nil {
		return errors.Wrap(err, "failed to compile query")
	}

	// Incoming query spec may have been produced by an entity other than the
	// Flux interpreter, so we must set the default Now time if not already set.
	if spec.Now.IsZero() {
		spec.Now = q.now
	}

	q.spec = *spec
	return nil
}

func (c *Controller) enqueueQuery(q *Query) error {
	if c.verbose {
		log.Println("query", flux.Formatted(&q.spec, flux.FmtJSON))
	}
	if !q.tryQueue() {
		return errors.New("failed to transition query to queueing state")
	}
	if err := q.spec.Validate(); err != nil {
		return errors.Wrap(err, "invalid query")
	}
	// Add query to the queue
	select {
	case c.newQueries <- q:
		return nil
	case <-q.parentCtx.Done():
		return q.parentCtx.Err()
	}
}

func (c *Controller) nextID() QueryID {
	c.queriesMu.Lock()
	defer c.queriesMu.Unlock()
	ok := true
	for ok {
		c.lastID++
		_, ok = c.queries[c.lastID]
	}
	return c.lastID
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

func (c *Controller) run() {
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
		}

		// Peek at head of priority queue
		q := pq.Peek()
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

	if q.tryPlan() {
		// Plan query to determine needed resources
		lp, err := c.lplanner.Plan(&q.spec)
		if err != nil {
			return true, errors.Wrap(err, "failed to create logical plan")
		}
		if c.verbose {
			log.Println("logical plan", plan.Formatted(lp))
		}

		p, err := c.pplanner.Plan(lp)
		if err != nil {
			return true, errors.Wrap(err, "failed to create physical plan")
		}
		q.plan = p
		q.concurrency = p.Resources.ConcurrencyQuota
		if q.concurrency > c.maxConcurrency {
			q.concurrency = c.maxConcurrency
		}
		q.memory = p.Resources.MemoryBytesQuota
		if c.verbose {
			log.Println("physical plan", plan.Formatted(q.plan))
		}
	}

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
		r, err := c.executor.Execute(q.currentCtx, q.plan, q.alloc)
		if err != nil {
			return true, errors.Wrap(err, "failed to execute query")
		}
		q.setResults(r)
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
	now  time.Time

	err error

	ready chan map[string]flux.Result

	mu     sync.Mutex
	state  State
	cancel func()

	parentCtx, currentCtx   context.Context
	parentSpan, currentSpan *span
	stats                   flux.Statistics

	plan *plan.PlanSpec

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
	q.mu.Lock()
	defer q.mu.Unlock()

	// call cancel func
	q.cancel()

	if q.state != Errored && q.state != Canceled {
		q.transitionTo(Canceled)
	}
}

// Ready returns a channel that will deliver the query results.
// Its possible that the channel is closed before any results arrive, in which case the query should be
// inspected for an error using Err().
func (q *Query) Ready() <-chan map[string]flux.Result {
	return q.ready
}

// Done must always be called to free resources.
func (q *Query) Done() {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.state != Errored {
		q.transitionTo(Finished)
	}
}

// Statistics reports the statisitcs for the query.
// The statisitcs are not complete until the query is finished.
func (q *Query) Statistics() flux.Statistics {
	q.mu.Lock()
	defer q.mu.Unlock()

	stats := q.stats
	stats.TotalDuration = q.parentSpan.Duration
	stats.Concurrency = q.concurrency
	if q.alloc != nil {
		stats.MaxAllocated = q.alloc.MaxAllocated()
	}
	return stats
}

// State reports the current state of the query.
func (q *Query) State() State {
	q.mu.Lock()
	s := q.state
	q.mu.Unlock()
	return s
}

// transitionTo will transition from one state to another. If a list of current states
// is given, then the query must be in one of those states for the transition to succeed.
// This method must be called with a lock and it must not transition to a final state
// from within the run loop.
func (q *Query) transitionTo(newState State, currentState ...State) bool {
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

	// If we are transitioning to a finished state from a non-finished state, finish the parent span.
	switch newState {
	case Errored, Canceled, Finished:
		if q.parentSpan != nil {
			q.parentSpan.Finish()
			q.parentSpan = nil

			// Close the ready channel on the first time we move to one of these states.
			// It should signal any queries waiting on the results that no results will come.
			// Signal to the main loop that this query has completed.
			// TODO(jsternberg): The existence of this channel here is incorrect and a result of not implementing
			// the interface correctly. This channel send should be in the Done method, but because it wasn't
			// placed there originally, many of the implementors don't call Done when they have called Cancel.
			// This channel send is in the wrong place and can't be done when there is a lock on the query because
			// it can cause a deadlock with the controller. Fixing that is more involved though so, temporarily,
			// we spawn a goroutine to send the query to queryDone to avoid a potential deadlock.
			go func() { q.c.queryDone <- q }()
			close(q.ready)
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
	q.mu.Lock()
	ok := q.state != Canceled && q.state != Errored && q.state != Finished
	q.mu.Unlock()
	return ok
}

// Err reports any error the query may have encountered.
func (q *Query) Err() error {
	q.mu.Lock()
	err := q.err
	q.mu.Unlock()
	return err
}
func (q *Query) setErr(err error) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.err = err

	q.transitionTo(Errored)
}

func (q *Query) setResults(r map[string]flux.Result) {
	q.mu.Lock()
	if q.state == Executing {
		q.ready <- r
	}
	q.mu.Unlock()
}

// tryCompile attempts to transition the query into the Compiling state.
func (q *Query) tryCompile() bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.transitionTo(Compiling, Created)
}

// tryQueue attempts to transition the query into the Queueing state.
func (q *Query) tryQueue() bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.transitionTo(Queueing, Compiling)
}

// tryRequeue attempts to transition the query into the Requeueing state.
func (q *Query) tryRequeue() bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.state == Requeueing {
		// Already in the correct state.
		return true
	}
	return q.transitionTo(Requeueing, Planning)
}

// tryPlan attempts to transition the query into the Planning state.
func (q *Query) tryPlan() bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.transitionTo(Planning, Queueing)
}

// tryExec attempts to transition the query into the Executing state.
func (q *Query) tryExec() bool {
	q.mu.Lock()
	defer q.mu.Unlock()

	return q.transitionTo(Executing, Requeueing, Planning)
}

// State is the query state.
type State int

const (
	Created State = iota
	Compiling
	Queueing
	Planning
	Requeueing
	Executing
	Errored
	Finished
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
