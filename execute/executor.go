// Package execute contains the implementation of the execution phase in the query engine.
package execute

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/influxdata/flux"
	plan "github.com/influxdata/flux/planner"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Executor interface {
	Execute(ctx context.Context, p *plan.PlanSpec, a *Allocator) (map[string]flux.Result, error)
}

type executor struct {
	deps   Dependencies
	logger *zap.Logger
}

func NewExecutor(deps Dependencies, logger *zap.Logger) Executor {
	if logger == nil {
		logger = zap.NewNop()
	}
	e := &executor{
		deps:   deps,
		logger: logger,
	}
	return e
}

type streamContext struct {
	bounds *Bounds
}

func newStreamContext(b *Bounds) streamContext {
	return streamContext{
		bounds: b,
	}
}

func (ctx streamContext) Bounds() *Bounds {
	return ctx.bounds
}

type executionState struct {
	p    *plan.PlanSpec
	deps Dependencies

	alloc *Allocator

	resources flux.ResourceManagement

	results map[string]flux.Result
	sources []Source

	transports []Transport

	dispatcher *poolDispatcher
	logger     *zap.Logger
}

func (e *executor) Execute(ctx context.Context, p *plan.PlanSpec, a *Allocator) (map[string]flux.Result, error) {
	es, err := e.createExecutionState(ctx, p, a)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize execute state")
	}
	es.logger = e.logger
	es.do(ctx)
	return es.results, nil
}

func validatePlan(p *plan.PlanSpec) error {
	if p.Resources.ConcurrencyQuota == 0 {
		return errors.New("plan must have a non-zero concurrency quota")
	}
	return nil
}

func (e *executor) createExecutionState(ctx context.Context, p *plan.PlanSpec, a *Allocator) (*executionState, error) {
	if err := validatePlan(p); err != nil {
		return nil, errors.Wrap(err, "invalid plan")
	}
	// Set allocation limit
	a.Limit = p.Resources.MemoryBytesQuota
	es := &executionState{
		p:         p,
		deps:      e.deps,
		alloc:     a,
		resources: p.Resources,
		results:   make(map[string]flux.Result, len(p.Roots())),
		// TODO(nathanielc): Have the planner specify the dispatcher throughput
		dispatcher: newPoolDispatcher(10, e.logger),
	}
	v := &ExecutionNodeVisitor{
		ctx: ctx,
		es:  es,
		en:  make(map[plan.PlanNode]Node),
		ec:  make(map[plan.PlanNode]executionContext),
		tr:  make(map[plan.PlanNode]Transformation),
	}

	if err := plan.Walk(p, v.CreateExecutionContext); err != nil {
		return nil, err
	}
	if err := plan.Walk(p, v.CreateExecutionNode); err != nil {
		return nil, err
	}
	if err := plan.Walk(p, v.CreateTransports); err != nil {
		return nil, err
	}
	return v.es, nil

	/*
		v := &CreateExecutionNodeVisitor{
			ctx:   ctx,
			state: es,
			trans: make(map[plan.PlanNode]Transformation),
		}
		if err := plan.Walk(p, v.Visit); err != nil {
			return nil, err
		}
		return v.state, nil
	*/
}

// DefaultTriggerSpec defines the triggering that should be used for datasets
// whose parent transformation is not a windowing transformation.
var DefaultTriggerSpec = flux.AfterWatermarkTriggerSpec{}

type triggeringSpec interface {
	TriggerSpec() flux.TriggerSpec
}

// ExecutionNodeVisitor has several `visit` methods for
// constructing an execution node for a physical procedure.
type ExecutionNodeVisitor struct {
	ctx context.Context
	es  *executionState

	// map plan node to its execution context
	ec map[plan.PlanNode]executionContext
	// map plan node to its execution node
	en map[plan.PlanNode]Node
	// map plan node to its transformation
	tr map[plan.PlanNode]Transformation
}

// CreateExecutionContext creates an execution context for each node in the plan
func (v *ExecutionNodeVisitor) CreateExecutionContext(node plan.PlanNode) error {
	ec := executionContext{
		ctx:     v.ctx,
		es:      v.es,
		parents: make([]DatasetID, len(node.Predecessors())),
	}
	for i, pred := range node.Predecessors() {
		id := plan.ProcedureIDFromNodeID(pred.ID())
		ec.parents[i] = DatasetID(id)
	}
	v.ec[node] = ec
	return nil
}

// CreateExecutionNode creates the node that will execute a particular plan node
func (v *ExecutionNodeVisitor) CreateExecutionNode(node plan.PlanNode) error {
	spec := node.ProcedureSpec()
	kind := spec.Kind()
	id := plan.ProcedureIDFromNodeID(node.ID())

	// Leaf node <=> source node
	if len(node.Predecessors()) == 0 {

		createSourceFn, ok := procedureToSource[kind]

		if !ok {
			return fmt.Errorf("unsupported source kind %v", kind)
		}

		source, err := createSourceFn(spec, DatasetID(id), v.ec[node])

		if err != nil {
			return err
		}

		v.en[node] = source
		v.es.sources = append(v.es.sources, source)
	}

	// Non-leaf node <=> transformation node
	createTransformationFn, ok := procedureToTransformation[kind]

	if !ok {
		return fmt.Errorf("unsupported procedure %v", kind)
	}

	transformation, dataset, err := createTransformationFn(DatasetID(id), AccumulatingMode, spec, v.ec[node])

	if err != nil {
		return err
	}

	// Setup triggering
	var ts flux.TriggerSpec = DefaultTriggerSpec
	if t, ok := spec.(triggeringSpec); ok {
		ts = t.TriggerSpec()
	}
	dataset.SetTriggerSpec(ts)

	v.tr[node] = transformation
	v.en[node] = dataset
	return nil
}

// CreateTransports creates the transport transformations for sending data downstream
func (v *ExecutionNodeVisitor) CreateTransports(node plan.PlanNode) error {
	if len(node.Successors()) == 0 {
		result := newResult(string(node.ID()))
		v.en[node].AddTransformation(result)
		v.es.results[string(node.ID())] = result
	}

	for _, successor := range node.Successors() {
		transformation := v.tr[successor]
		transport := newConescutiveTransport(v.es.dispatcher, transformation)
		v.es.transports = append(v.es.transports, transport)
		v.en[node].AddTransformation(transport)
	}
	return nil
}

func (es *executionState) abort(err error) {
	for _, r := range es.results {
		r.(*result).abort(err)
	}
}

func (es *executionState) do(ctx context.Context) {
	for _, src := range es.sources {
		go func(src Source) {
			// Setup panic handling on the source goroutines
			defer func() {
				if e := recover(); e != nil {
					// We had a panic, abort the entire execution.
					var err error
					switch e := e.(type) {
					case error:
						err = e
					default:
						err = fmt.Errorf("%v", e)
					}
					es.abort(fmt.Errorf("panic: %v\n%s", err, debug.Stack()))
					if entry := es.logger.Check(zapcore.InfoLevel, "Execute source panic"); entry != nil {
						entry.Stack = string(debug.Stack())
						entry.Write(zap.Error(err))
					}
				}
			}()
			src.Run(ctx)
		}(src)
	}
	es.dispatcher.Start(es.resources.ConcurrencyQuota, ctx)
	go func() {
		// Wait for all transports to finish
		for _, t := range es.transports {
			select {
			case <-t.Finished():
			case <-ctx.Done():
				es.abort(errors.New("context done"))
			case err := <-es.dispatcher.Err():
				if err != nil {
					es.abort(err)
				}
			}
		}
		// Check for any errors on the dispatcher
		err := es.dispatcher.Stop()
		if err != nil {
			es.abort(err)
		}
	}()
}

// Need a unique stream context per execution context
type executionContext struct {
	ctx           context.Context
	es            *executionState
	parents       []DatasetID
	streamContext streamContext
}

func resolveTime(qt flux.Time, now time.Time) Time {
	return Time(qt.Time(now).UnixNano())
}

func (ec executionContext) Context() context.Context {
	return ec.ctx
}

func (ec executionContext) ResolveTime(qt flux.Time) Time {
	return resolveTime(qt, ec.es.p.Now)
}

func (ec executionContext) StreamContext() StreamContext {
	return ec.streamContext
}

func (ec executionContext) Allocator() *Allocator {
	return ec.es.alloc
}

func (ec executionContext) Parents() []DatasetID {
	return ec.parents
}
func (ec executionContext) ConvertID(id plan.ProcedureID) DatasetID {
	return DatasetID(id)
}

func (ec executionContext) Dependencies() Dependencies {
	return ec.es.deps
}
