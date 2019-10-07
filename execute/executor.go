// Package execute contains the implementation of the execution phase in the query engine.
package execute

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Executor interface {
	// Execute will begin execution of the plan.Spec using the memory allocator.
	// This returns a mapping of names to the query results.
	// This will also return a channel for the Metadata from the query. The channel
	// may return zero or more values. The returned channel must not require itself to
	// be read so the executor must allocate enough space in the channel so if the channel
	// is unread that it will not block.
	Execute(ctx context.Context, p *plan.Spec, a *memory.Allocator) (map[string]flux.Result, <-chan flux.Metadata, error)
}

type executor struct {
	logger *zap.Logger
}

func NewExecutor(logger *zap.Logger) Executor {
	if logger == nil {
		logger = zap.NewNop()
	}
	e := &executor{
		logger: logger,
	}
	return e
}

type streamContext struct {
	bounds *Bounds
}

func (ctx streamContext) Bounds() *Bounds {
	return ctx.bounds
}

type executionState struct {
	p *plan.Spec

	alloc *memory.Allocator

	resources flux.ResourceManagement

	results map[string]flux.Result
	sources []Source
	metaCh  chan flux.Metadata

	transports []Transport

	dispatcher *poolDispatcher
	logger     *zap.Logger
}

func (e *executor) Execute(ctx context.Context, p *plan.Spec, a *memory.Allocator) (map[string]flux.Result, <-chan flux.Metadata, error) {
	es, err := e.createExecutionState(ctx, p, a)
	if err != nil {
		return nil, nil, errors.Wrap(err, codes.Inherit, "failed to initialize execute state")
	}
	es.logger = e.logger
	es.do(ctx)
	return es.results, es.metaCh, nil
}

func validatePlan(p *plan.Spec) error {
	if p.Resources.ConcurrencyQuota == 0 {
		return errors.New(codes.Invalid, "plan must have a non-zero concurrency quota")
	}
	return nil
}

func (e *executor) createExecutionState(ctx context.Context, p *plan.Spec, a *memory.Allocator) (*executionState, error) {
	if err := validatePlan(p); err != nil {
		return nil, errors.Wrap(err, codes.Invalid, "invalid plan")
	}
	es := &executionState{
		p:         p,
		alloc:     a,
		resources: p.Resources,
		results:   make(map[string]flux.Result),
		// TODO(nathanielc): Have the planner specify the dispatcher throughput
		dispatcher: newPoolDispatcher(10, e.logger),
	}
	v := &createExecutionNodeVisitor{
		ctx:   ctx,
		es:    es,
		nodes: make(map[plan.Node]Node),
	}

	if err := p.BottomUpWalk(v.Visit); err != nil {
		return nil, err
	}

	// Only sources can be a MetadataNode at the moment so allocate enough
	// space for all of them to report metadata. Not all of them will necessarily
	// report metadata.
	es.metaCh = make(chan flux.Metadata, len(es.sources))

	return v.es, nil
}

// createExecutionNodeVisitor visits each node in a physical query plan
// and creates a node responsible for executing that physical operation.
type createExecutionNodeVisitor struct {
	ctx   context.Context
	es    *executionState
	nodes map[plan.Node]Node
}

func skipYields(pn plan.Node) plan.Node {
	isYield := func(pn plan.Node) bool {
		_, ok := pn.ProcedureSpec().(plan.YieldProcedureSpec)
		return ok
	}

	for isYield(pn) {
		pn = pn.Predecessors()[0]
	}

	return pn
}

func nonYieldPredecessors(pn plan.Node) []plan.Node {
	nodes := make([]plan.Node, len(pn.Predecessors()))
	for i, pred := range pn.Predecessors() {
		nodes[i] = skipYields(pred)
	}

	return nodes
}

// Visit creates the node that will execute a particular plan node
func (v *createExecutionNodeVisitor) Visit(node plan.Node) error {
	ppn, ok := node.(*plan.PhysicalPlanNode)
	if !ok {
		return fmt.Errorf("cannot execute plan node of type %T", node)
	}
	spec := node.ProcedureSpec()
	kind := spec.Kind()
	id := DatasetIDFromNodeID(node.ID())

	if yieldSpec, ok := spec.(plan.YieldProcedureSpec); ok {
		r := newResult(yieldSpec.YieldName())
		v.es.results[yieldSpec.YieldName()] = r
		v.nodes[skipYields(node)].AddTransformation(r)
		return nil
	}

	// Add explicit stream context if bounds are set on this node
	var streamContext streamContext
	if node.Bounds() != nil {
		streamContext.bounds = &Bounds{
			Start: node.Bounds().Start,
			Stop:  node.Bounds().Stop,
		}
	}

	// Build execution context
	ec := executionContext{
		ctx:           v.ctx,
		es:            v.es,
		parents:       make([]DatasetID, len(node.Predecessors())),
		streamContext: streamContext,
	}

	for i, pred := range nonYieldPredecessors(node) {
		ec.parents[i] = DatasetIDFromNodeID(pred.ID())
	}

	// If node is a leaf, create a source
	if len(node.Predecessors()) == 0 {
		createSourceFn, ok := procedureToSource[kind]

		if !ok {
			return fmt.Errorf("unsupported source kind %v", kind)
		}

		source, err := createSourceFn(spec, id, ec)

		if err != nil {
			return err
		}

		v.es.sources = append(v.es.sources, source)
		v.nodes[node] = source
	} else {

		// If node is internal, create a transformation.
		// For each predecessor, add a transport for sending data upstream.
		createTransformationFn, ok := procedureToTransformation[kind]

		if !ok {
			return fmt.Errorf("unsupported procedure %v", kind)
		}

		tr, ds, err := createTransformationFn(id, DiscardingMode, spec, ec)

		if err != nil {
			return err
		}

		if ppn.TriggerSpec == nil {
			ppn.TriggerSpec = plan.DefaultTriggerSpec
		}
		ds.SetTriggerSpec(ppn.TriggerSpec)
		v.nodes[node] = ds

		for _, p := range nonYieldPredecessors(node) {
			executionNode := v.nodes[p]
			transport := newConsecutiveTransport(v.es.dispatcher, tr)
			v.es.transports = append(v.es.transports, transport)
			executionNode.AddTransformation(transport)
		}

		if plan.HasSideEffect(spec) && len(node.Successors()) == 0 {
			name := string(node.ID())
			r := newResult(name)
			v.es.results[name] = r
			v.nodes[skipYields(node)].AddTransformation(r)
		}
	}

	return nil
}

func (es *executionState) abort(err error) {
	for _, r := range es.results {
		r.(*result).abort(err)
	}
}

func (es *executionState) do(ctx context.Context) {
	var wg sync.WaitGroup
	for _, src := range es.sources {
		wg.Add(1)
		go func(src Source) {
			defer wg.Done()

			// Setup panic handling on the source goroutines
			defer func() {
				if e := recover(); e != nil {
					// We had a panic, abort the entire execution.
					err, ok := e.(error)
					if !ok {
						err = fmt.Errorf("%v", e)
					}

					if errors.Code(err) == codes.ResourceExhausted {
						es.abort(err)
						return
					}

					err = errors.Wrap(err, codes.Internal, "panic")
					es.abort(err)
					if entry := es.logger.Check(zapcore.InfoLevel, "Execute source panic"); entry != nil {
						entry.Stack = string(debug.Stack())
						entry.Write(zap.Error(err))
					}
				}
			}()
			src.Run(ctx)

			if mdn, ok := src.(MetadataNode); ok {
				es.metaCh <- mdn.Metadata()
			}
		}(src)
	}

	go func() {
		defer close(es.metaCh)
		wg.Wait()
	}()

	es.dispatcher.Start(es.resources.ConcurrencyQuota, ctx)
	go func() {
		// Wait for all transports to finish
		for _, t := range es.transports {
			select {
			case <-t.Finished():
			case <-ctx.Done():
				es.abort(ctx.Err())
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

func (ec executionContext) Allocator() *memory.Allocator {
	return ec.es.alloc
}

func (ec executionContext) Parents() []DatasetID {
	return ec.parents
}
