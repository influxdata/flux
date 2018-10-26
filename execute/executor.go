// Package execute contains the implementation of the execution phase in the query engine.
package execute

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/planner"
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
		results:   make(map[string]flux.Result, len(p.Results)),
		// TODO(nathanielc): Have the planner specify the dispatcher throughput
		dispatcher: newPoolDispatcher(10, e.logger),
	}
	nodes := make(map[plan.ProcedureID]Node, len(p.Procedures))
	for name, yield := range p.Results {
		ds, err := es.createNode(ctx, p.Procedures[yield.ID], nodes)
		if err != nil {
			return nil, err
		}
		r := newResult(name, yield)
		ds.AddTransformation(r)
		es.results[name] = r
	}
	return es, nil
}

// DefaultTriggerSpec defines the triggering that should be used for datasets
// whose parent transformation is not a windowing transformation.
var DefaultTriggerSpec = flux.AfterWatermarkTriggerSpec{}

type triggeringSpec interface {
	TriggerSpec() flux.TriggerSpec
}

func (es *executionState) createNode(ctx context.Context, pr *plan.Procedure, nodes map[plan.ProcedureID]Node) (Node, error) {
	// Check if we already created this node
	if n, ok := nodes[pr.ID]; ok {
		return n, nil
	}

	// Add explicit stream context if bounds are set on this node
	var streamContext streamContext
	if pr.Bounds != nil {
		streamContext.bounds = &Bounds{
			Start: pr.Bounds.Start,
			Stop:  pr.Bounds.Stop,
		}
	}

	// Build execution context
	ec := executionContext{
		ctx:           ctx,
		es:            es,
		streamContext: streamContext,
	}

	if len(pr.Parents) > 0 {
		ec.parents = make([]DatasetID, len(pr.Parents))
		for i, parentID := range pr.Parents {
			ec.parents[i] = DatasetID(parentID)
		}
	}

	// If source create source
	if createS, ok := procedureToSource[pr.Spec.Kind()]; ok {
		s, err := createS(pr.Spec, DatasetID(pr.ID), ec)
		if err != nil {
			return nil, err
		}
		es.sources = append(es.sources, s)
		nodes[pr.ID] = s
		return s, nil
	}

	createT, ok := procedureToTransformation[pr.Spec.Kind()]
	if !ok {
		return nil, fmt.Errorf("unsupported procedure %v", pr.Spec.Kind())
	}

	// Create the transformation
	t, ds, err := createT(DatasetID(pr.ID), AccumulatingMode, pr.Spec, ec)
	if err != nil {
		return nil, err
	}
	nodes[pr.ID] = ds

	// Setup triggering
	var ts flux.TriggerSpec = DefaultTriggerSpec
	if t, ok := pr.Spec.(triggeringSpec); ok {
		ts = t.TriggerSpec()
	}
	ds.SetTriggerSpec(ts)

	// Recurse creating parents
	for _, parentID := range pr.Parents {
		parent, err := es.createNode(ctx, es.p.Procedures[parentID], nodes)
		if err != nil {
			return nil, err
		}
		transport := newConescutiveTransport(es.dispatcher, t)
		es.transports = append(es.transports, transport)
		parent.AddTransformation(transport)
	}

	return ds, nil
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
func (ec executionContext) ConvertID(id planner.ProcedureID) DatasetID {
	return DatasetID(plan.ProcedureID(id))
}

func (ec executionContext) Dependencies() Dependencies {
	return ec.es.deps
}
