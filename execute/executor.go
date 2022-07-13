// Package execute contains the implementation of the execution phase in the query engine.
package execute

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"sync"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/metadata"
	"github.com/influxdata/flux/plan"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
)

type Executor interface {
	// Execute will begin execution of the plan.Spec using the memory allocator.
	// This returns a mapping of names to the query results.
	// This will also return a channel for the Metadata from the query. The channel
	// may return zero or more values. The returned channel must not require itself to
	// be read so the executor must allocate enough space in the channel so if the channel
	// is unread that it will not block.
	Execute(ctx context.Context, p *plan.Spec, a memory.Allocator) (map[string]flux.Result, <-chan flux.Statistics, error)
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

	ctx    context.Context
	cancel func()
	alloc  memory.Allocator

	resources flux.ResourceManagement

	results map[string]flux.Result
	sources []Source
	statsCh chan flux.Statistics

	transports []AsyncTransport

	dispatcher *poolDispatcher
	logger     *zap.Logger
}

func (e *executor) Execute(ctx context.Context, p *plan.Spec, a memory.Allocator) (map[string]flux.Result, <-chan flux.Statistics, error) {
	es, err := e.createExecutionState(ctx, p, a)
	if err != nil {
		return nil, nil, errors.Wrap(err, codes.Inherit, "failed to initialize execute state")
	}
	es.do()
	return es.results, es.statsCh, nil
}

func (e *executor) createExecutionState(ctx context.Context, p *plan.Spec, a memory.Allocator) (*executionState, error) {
	ctx, cancel := context.WithCancel(ctx)
	es := &executionState{
		p:         p,
		ctx:       ctx,
		cancel:    cancel,
		alloc:     a,
		resources: p.Resources,
		results:   make(map[string]flux.Result),
		// TODO(nathanielc): Have the planner specify the dispatcher throughput
		dispatcher: newPoolDispatcher(10, e.logger),
		logger:     e.logger,
	}
	v := &createExecutionNodeVisitor{
		es:    es,
		nodes: make(map[plan.Node][]Node),
	}

	if err := p.BottomUpWalk(v.Visit); err != nil {
		return nil, err
	}

	// Only one statistics struct will be sent. Allocate space for it
	// so we don't block on its creation.
	es.statsCh = make(chan flux.Statistics, 1)

	// Choose some default resource limits based on execution options, if necessary.
	es.chooseDefaultResources(ctx, p)

	if err := es.validate(); err != nil {
		return nil, errors.Wrap(err, codes.Invalid, "execution state")
	}

	return v.es, nil
}

// createExecutionNodeVisitor visits each node in a physical query plan
// and creates a node responsible for executing that physical operation.
type createExecutionNodeVisitor struct {
	es    *executionState
	nodes map[plan.Node][]Node
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

	// N.b. yields become results here, but other terminal nodes are handled
	// further below.
	if yieldSpec, ok := spec.(plan.YieldProcedureSpec); ok {
		if err := v.generateResult(yieldSpec.YieldName(), node, 0); err != nil {
			return err
		}
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

	// There are three types of instantiations we need to support here, and for
	// each one of these, there are source nodes and non-source nodes.
	//
	// 1. Standard instantiation. These are non-parallel, non-merge nodes.
	//    There is a 1:1 relationship between planner node and execution graph
	//    node. Only a single copy of the node is made and it references the
	//    single copy of the predecessor nodes.
	//
	// 2. Parallel instantiation. There are multiple copies of the node and of
	//    all predecessors. This applies to both source and non-source nodes.
	//
	// 3. Merge instantiation. There is a single copy of the node, but multiple copies of the
	//    predecessors. These copies merge into the node.

	copies := 1
	if attr := plan.GetOutputAttribute(ppn, plan.ParallelRunKey); attr != nil {
		copies = attr.(plan.ParallelRunAttribute).Factor
	}

	isParallelMerge := false
	predCopies := 1
	if attr := plan.GetOutputAttribute(ppn, plan.ParallelMergeKey); attr != nil {
		isParallelMerge = true
		predCopies = attr.(plan.ParallelMergeAttribute).Factor
	}

	// Build execution context for each copy.
	ec := make([]executionContext, copies)
	for i := 0; i < copies; i++ {
		ec[i] = executionContext{
			es:            v.es,
			parents:       make([]DatasetID, len(node.Predecessors())*predCopies),
			streamContext: streamContext,
			parallelOpts:  ParallelOpts{Group: i, Factor: copies},
		}

		for pi, pred := range nonYieldPredecessors(node) {
			for j := 0; j < predCopies; j++ {
				ec[i].parents[pi*predCopies+j] = datasetIDFromNodeID(pred.ID(), i+j)
			}
		}
	}

	v.nodes[node] = make([]Node, copies)

	// If node is a leaf, create a source
	if len(node.Predecessors()) == 0 {
		createSourceFn, ok := procedureToSource[kind]
		if !ok {
			return fmt.Errorf("unsupported source kind %v", kind)
		}

		for i := 0; i < copies; i++ {
			id := datasetIDFromNodeID(node.ID(), i)

			source, err := createSourceFn(spec, id, ec[i])

			if err != nil {
				return err
			}

			source.SetLabel(string(node.ID()))
			v.es.sources = append(v.es.sources, source)
			v.nodes[node][i] = source
		}
	} else {
		// If node is internal, create a transformation. For each
		// predecessor, add a transport for sending data upstream.
		createTransformationFn, ok := procedureToTransformation[kind]

		if !ok {
			return fmt.Errorf("unsupported procedure %v", kind)
		}

		for i := 0; i < copies; i++ {
			id := datasetIDFromNodeID(node.ID(), i)

			tr, ds, err := createTransformationFn(id, DiscardingMode, spec, ec[i])

			if err != nil {
				return err
			}

			if ds, ok := ds.(DatasetContext); ok {
				ds.WithContext(v.es.ctx)
			}

			if ppn.TriggerSpec == nil {
				ppn.TriggerSpec = plan.DefaultTriggerSpec
			}
			ds.SetTriggerSpec(ppn.TriggerSpec)
			v.nodes[node][i] = ds

			for _, p := range nonYieldPredecessors(node) {
				// In case (1) above, both copies and predCopies are 1. We link
				// forward from the only copy of the predecessor node.
				//   i == 0 AND j == 0
				//
				// In case (2) above, both copies is > 1 and predCopies is 1.
				// We link forward from corresponding copy for the node.
				//   ( iterating i ) AND j == 0
				//
				// In case (3) above, both copies is 1 and predCopies is > 1.
				// We link forward from all copies for the node to achieve the
				// fan-in.
				//   i == 0 AND ( iterating j )
				for j := 0; j < predCopies; j++ {
					// Either i == 0 && j == 0: we are either iterating i, or we are iterating j.
					executionNode := v.nodes[p][i+j]
					transport := newConsecutiveTransport(v.es.ctx, v.es.dispatcher, tr, node, v.es.logger, v.es.alloc)
					v.es.transports = append(v.es.transports, transport)
					executionNode.AddTransformation(transport)
				}
			}
		}
	}
	// Results should be generated for terminal nodes.
	//
	// If user entered `from |> range |> filter` but forgot to add the yield, we
	// generate a result for them. In this case, the terminal node has no
	// side-effects and the result name uses the default of `_result` (per the
	// body of `getResultName()`).
	//
	// This is in contrast to other specific cases, like terminal nodes
	// _with side-effects_ which generate results named for the node itself.
	//
	// All queries must have an associated result transformation, otherwise
	// the query context will be cancelled immediately and execution
	// will be cancelled before any work for the query has been done.
	// TODO: understand if this (preventing cancellation) could be addressed by another means.
	if len(node.Successors()) == 0 {
		resultName, err := getResultName(node, spec, isParallelMerge)
		if err != nil {
			return err
		}
		if err := v.generateResult(resultName, node, 0); err != nil {
			return err
		}
	}
	return nil
}

// generateResult will attach a result to the query for the specified node.
func (v *createExecutionNodeVisitor) generateResult(resultName string, node plan.Node, idx int) error {
	// if the result name is already present in the result set, that's an error.
	if _, ok := v.es.results[resultName]; ok {
		// XXX: we produce an error like this in the planner for duplicate yield
		// names, but since we're generating results that aren't necessarily from
		// yields, we need a similar check here.
		return errors.Newf(codes.Invalid, "tried to produce more than one result with the name %q", resultName)
	}
	r := newResult(resultName)
	v.es.results[resultName] = r
	v.nodes[skipYields(node)][idx].AddTransformation(r)
	return nil
}

// getResultName will offer a "best guess" name for a given node's result.
//
// For nodes that have side-effects, the result will be based on the node ID.
// If the node is a parallel merge node, then it was added by the planner; in
// this case check the side-effect status of the node's predecessor. For other
// cases, the default yield name of `_result` is used.
func getResultName(node plan.Node, spec plan.ProcedureSpec, isParallelMerge bool) (string, error) {
	name := plan.DefaultYieldName

	// Parallel merge nodes are added by the planner and can mask the presence
	// of a side-effect result; skip them.
	if isParallelMerge {
		if len(node.Predecessors()) != 1 {
			return "", errors.New(codes.Internal, "parallel merge must have a single predecessor")
		}

		node := node.Predecessors()[0]
		spec = node.ProcedureSpec()
	}

	if plan.HasSideEffect(spec) {
		name = string(node.ID())
	}
	return name, nil
}

func (es *executionState) validate() error {
	if es.resources.ConcurrencyQuota == 0 {
		return errors.New(codes.Invalid, "execution state must have a non-zero concurrency quota")
	}
	return nil
}

// getResourceExecOptions returns the DefaultMemoryLimit and ConcurrencyLimit
// from exec options, if present.
func getResourceLimits(ctx context.Context) (int64, int) {
	// Initialize resources from the execution dependencies and/or properties of the plan.
	if !HaveExecutionDependencies(ctx) {
		return 0, math.MaxInt
	}

	execOptions := GetExecutionDependencies(ctx).ExecutionOptions
	return execOptions.DefaultMemoryLimit, execOptions.ConcurrencyLimit
}

func (es *executionState) chooseDefaultResources(ctx context.Context, p *plan.Spec) {
	defaultMemoryLimit, concurrencyLimit := getResourceLimits(ctx)

	// Update memory quota
	if es.resources.MemoryBytesQuota == 0 {
		es.resources.MemoryBytesQuota = defaultMemoryLimit
	}

	// Update concurrency quota
	if es.resources.ConcurrencyQuota == 0 {
		es.resources.ConcurrencyQuota = len(p.Roots)

		// If the query concurrency limit is greater than zero,
		// we will use the new behavior that sets the concurrency
		// quota equal to the number of transformations and limits
		// it to the value specified.
		if concurrencyLimit > 0 {
			concurrencyQuota := 0
			_ = p.TopDownWalk(func(node plan.Node) error {
				// Do not include source nodes in the node list as
				// they do not use the dispatcher.
				if len(node.Predecessors()) > 0 {
					addend := 1
					ppn := node.(*plan.PhysicalPlanNode)
					if attr := plan.GetOutputAttribute(ppn, plan.ParallelRunKey); attr != nil {
						addend = attr.(plan.ParallelRunAttribute).Factor
					}
					concurrencyQuota += addend
				}
				return nil
			})

			if concurrencyQuota > int(concurrencyLimit) {
				concurrencyQuota = int(concurrencyLimit)
			} else if concurrencyQuota == 0 {
				concurrencyQuota = 1
			}
			es.resources.ConcurrencyQuota = concurrencyQuota
		}
	}
}

func (es *executionState) abort(err error) {
	for _, r := range es.results {
		r.(*result).abort(err)
	}
	es.cancel()
}

func (es *executionState) do() {
	var (
		wg      sync.WaitGroup
		stats   flux.Statistics
		statsMu sync.Mutex
	)

	updateStats := func(fn func(stats *flux.Statistics)) {
		statsMu.Lock()
		defer statsMu.Unlock()
		fn(&stats)
	}

	stats.Metadata = make(metadata.Metadata)
	for _, src := range es.sources {
		wg.Add(1)
		go func(src Source) {
			ctx := es.ctx
			opName := reflect.TypeOf(src).String()

			// If operator profiling is enabled for this execution, begin profiling
			profile := flux.TransportProfile{
				NodeType: opName,
				Label:    src.Label(),
			}
			profileSpan := profile.StartSpan()

			if span, spanCtx := opentracing.StartSpanFromContext(ctx, opName, opentracing.Tag{Key: "label", Value: src.Label()}); span != nil {
				ctx = spanCtx
				defer span.Finish()
			}

			defer wg.Done()

			// Setup panic handling on the source goroutines
			defer es.recover()
			src.Run(ctx)
			profileSpan.Finish()

			updateStats(func(stats *flux.Statistics) {
				stats.Profiles = append(stats.Profiles, profile)
				if mdn, ok := src.(MetadataNode); ok {
					stats.Metadata.AddAll(mdn.Metadata())
				}
			})
		}(src)
	}

	wg.Add(1)
	es.dispatcher.Start(es.resources.ConcurrencyQuota, es.ctx)

	// Keep the transport profiles in a separate array from the source profiles.
	// This ensures that sources are before transformations.
	profiles := make([]flux.TransportProfile, 0, len(es.transports))
	go func() {
		defer wg.Done()

		// Wait for all transports to finish
		for _, t := range es.transports {
			select {
			case <-t.Finished():
				tp := t.TransportProfile()
				profiles = append(profiles, tp)
			case <-es.ctx.Done():
				es.abort(es.ctx.Err())
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

	go func() {
		defer close(es.statsCh)
		wg.Wait()

		// Merge the transport profiles in with the ones already filled
		// by the sources.
		stats.Profiles = append(stats.Profiles, profiles...)

		es.statsCh <- stats
	}()
}

type ParallelOpts struct {
	Group  int
	Factor int
}

// Need a unique stream context per execution context
type executionContext struct {
	es            *executionState
	parents       []DatasetID
	streamContext streamContext
	parallelOpts  ParallelOpts
}

func resolveTime(qt flux.Time, now time.Time) Time {
	return Time(qt.Time(now).UnixNano())
}

func (ec executionContext) Context() context.Context {
	return ec.es.ctx
}

func (ec executionContext) ResolveTime(qt flux.Time) Time {
	return resolveTime(qt, ec.es.p.Now)
}

func (ec executionContext) StreamContext() StreamContext {
	return ec.streamContext
}

func (ec executionContext) Allocator() memory.Allocator {
	return ec.es.alloc
}

func (ec executionContext) Parents() []DatasetID {
	return ec.parents
}

func (ec executionContext) ParallelOpts() ParallelOpts {
	return ec.parallelOpts
}
