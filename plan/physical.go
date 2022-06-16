package plan

import (
	"context"
	"fmt"
	"math"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/interpreter"
)

// PhysicalPlanner performs transforms a logical plan to a physical plan,
// by applying any registered physical rules.
type PhysicalPlanner interface {
	Plan(ctx context.Context, lplan *Spec) (*Spec, error)
}

// NewPhysicalPlanner creates a new physical plan with the specified options.
// The new plan will be configured to apply any physical rules that have been registered.
func NewPhysicalPlanner(options ...PhysicalOption) PhysicalPlanner {
	pp := &physicalPlanner{
		heuristicPlannerPhysical: newHeuristicPlanner(),
		heuristicPlannerParallel: newHeuristicPlanner(),
		defaultMemoryLimit:       math.MaxInt64,
	}

	rulesPhysical := make([]Rule, len(ruleNameToPhysicalRule))
	i := 0
	for _, v := range ruleNameToPhysicalRule {
		rulesPhysical[i] = v
		i++
	}

	rulesParallel := make([]Rule, len(ruleNameToParallelizeRules))
	i = 0
	for _, v := range ruleNameToParallelizeRules {
		rulesParallel[i] = v
		i++
	}

	pp.heuristicPlannerPhysical.addRules(rulesPhysical...)

	pp.heuristicPlannerPhysical.addRules(physicalConverterRule{})

	pp.heuristicPlannerParallel.addRules(rulesParallel...)

	// Options may add or remove rules, so process them after we've
	// added registered rules.
	for _, opt := range options {
		opt.apply(pp)
	}

	return pp
}

func (pp *physicalPlanner) Plan(ctx context.Context, spec *Spec) (*Spec, error) {
	intermediateSpec, err := pp.heuristicPlannerPhysical.Plan(ctx, spec)
	if err != nil {
		return nil, err
	}

	transformedSpec, err := pp.heuristicPlannerParallel.Plan(ctx, intermediateSpec)
	if err != nil {
		return nil, err
	}

	// Compute time bounds for nodes in the plan
	if err := transformedSpec.BottomUpWalk(ComputeBounds); err != nil {
		return nil, err
	}

	// Set all default and/or registered trigger specs
	if err := transformedSpec.TopDownWalk(SetTriggerSpec); err != nil {
		return nil, err
	}

	// Ensure that the plan is valid
	if !pp.disableValidation {
		err := transformedSpec.CheckIntegrity()
		if err != nil {
			return nil, err
		}

		err = ValidatePhysicalPlan(transformedSpec)
		if err != nil {
			return nil, err
		}
	}

	return transformedSpec, nil
}

func ValidatePhysicalPlan(plan *Spec) error {
	err := plan.BottomUpWalk(func(pn Node) error {
		if validator, ok := pn.ProcedureSpec().(PostPhysicalValidator); ok {
			if err := validator.PostPhysicalValidate(pn.ID()); err != nil {
				return err
			}
		}

		ppn, ok := pn.(*PhysicalPlanNode)
		if !ok {
			return &flux.Error{
				Code: codes.Internal,
				Msg:  fmt.Sprintf("invalid physical query plan; found logical operation \"%v\"", pn.ID()),
			}
		}
		if ppn.TriggerSpec == nil {
			return &flux.Error{
				Code: codes.Internal,
				Msg:  fmt.Sprintf("invalid physical query plan; trigger spec not set on \"%v\"", ppn.id),
			}
		}

		// Check if required attributes are present in the output of
		// predecessors.

		// If there are any required attributes for this node, there should be one set of
		// required attributes for each input.
		reqAttrs := ppn.RequiredAttrs()
		if lra, lpred := len(reqAttrs), len(ppn.Predecessors()); lra > 0 && lra != lpred {
			return &flux.Error{
				Code: codes.Internal,
				Msg:  fmt.Sprintf("node has %d predecessors but has %d sets of required attributes", lpred, lra),
			}
		}
		for i, attrs := range reqAttrs {
			for key, attr := range attrs {
				ppred := pn.Predecessors()[i].(*PhysicalPlanNode)
				if !NodeSatisfiesRequiredAttribute(ppred, attr) {
					return &flux.Error{
						Code: codes.Internal,
						Msg: fmt.Sprintf("invalid physical query plan; attribute \"%v\" required by "+
							"\"%v\" is not satisfied by predecessor \"%v\"", key, ppn.id, ppred.id),
					}
				}
			}
		}

		// Check if attributes that must be required in successors are indeed
		// required there.
		outputAttrs := ppn.OutputAttrs()
		for key, attr := range outputAttrs {
			if attr.SuccessorsMustRequire() {
				if len(ppn.Successors()) == 0 {
					return &flux.Error{
						Code: codes.Internal,
						Msg: fmt.Sprintf("node %v provides attribute %v that must be required but has no "+
							"successors to require it", ppn.ID(), key),
					}
				}

				for _, succ := range ppn.Successors() {
					psucc := succ.(*PhysicalPlanNode)
					whichPred := indexOfNode(pn, psucc.Predecessors())
					if whichPred == -1 {
						return &flux.Error{
							Code: codes.Internal,
							Msg: fmt.Sprintf("plan integrity violated: %v has %v as successor, "+
								"but %v does not have %v as a predecessor", pn.ID(), psucc.ID(), psucc.ID(), pn.ID()),
						}
					}
					if attr := psucc.RequiredAttrs()[whichPred][key]; attr == nil {
						return &flux.Error{
							Code: codes.Internal,
							Msg: fmt.Sprintf("invalid physical query plan; attribute \"%v\" on "+
								"\"%v\" must be required by all successors, but isn't on \"%v\"",
								key, ppn.id, psucc.id),
						}
					}
				}
			}
		}

		return nil
	})
	return err
}

type physicalPlanner struct {
	heuristicPlannerPhysical *heuristicPlanner
	heuristicPlannerParallel *heuristicPlanner
	defaultMemoryLimit       int64
	disableValidation        bool
}

// PhysicalOption is an option to configure the behavior of the physical plan.
type PhysicalOption interface {
	apply(*physicalPlanner)
}

type physicalOption func(*physicalPlanner)

func (opt physicalOption) apply(p *physicalPlanner) {
	opt(p)
}

// WithDefaultMemoryLimit sets the default memory limit for plans generated by the plan.
// If the query spec explicitly sets a memory limit, that limit is used instead of the default.
func WithDefaultMemoryLimit(memBytes int64) PhysicalOption {
	return physicalOption(func(p *physicalPlanner) {
		p.defaultMemoryLimit = memBytes
	})
}

// OnlyPhysicalRules produces a physical plan option that forces only a particular set of rules to be applied.
func OnlyPhysicalRules(rules ...Rule) PhysicalOption {
	return physicalOption(func(pp *physicalPlanner) {
		pp.heuristicPlannerPhysical.clearRules()
		pp.heuristicPlannerParallel.clearRules()
		// Always add physicalConverterRule. It doesn't change the plan but only convert nodes to physical.
		// This is required for some pieces to work on the physical plan (e.g. SetTriggerSpec).
		pp.heuristicPlannerPhysical.addRules(physicalConverterRule{})
		pp.heuristicPlannerPhysical.addRules(rules...)
	})
}

func AddParallelRules(rules ...Rule) PhysicalOption {
	return physicalOption(func(pp *physicalPlanner) {
		pp.heuristicPlannerParallel.addRules(rules...)
	})
}

func RemovePhysicalRules(rules ...string) PhysicalOption {
	return physicalOption(func(pp *physicalPlanner) {
		pp.heuristicPlannerPhysical.removeRules(rules...)
		pp.heuristicPlannerParallel.removeRules(rules...)
	})
}

// DisableValidation disables validation in the physical planner.
func DisableValidation() PhysicalOption {
	return physicalOption(func(p *physicalPlanner) {
		p.disableValidation = true
	})
}

// physicalConverterRule rewrites logical nodes that have a ProcedureSpec that implements
// PhysicalProcedureSpec as a physical node.  For operations that have a 1:1 relationship
// between their physical and logical operations, this is the default behavior.
type physicalConverterRule struct {
}

func (physicalConverterRule) Name() string {
	return "physicalConverterRule"
}

func (physicalConverterRule) Pattern() Pattern {
	return Any()
}

func (physicalConverterRule) Rewrite(ctx context.Context, pn Node) (Node, bool, error) {
	if _, ok := pn.(*PhysicalPlanNode); ok {
		// Already converted
		return pn, false, nil
	}

	ln := pn.(*LogicalNode)
	pspec, ok := ln.Spec.(PhysicalProcedureSpec)
	if !ok {
		// A different rule will do the conversion
		return pn, false, nil
	}

	newNode := PhysicalPlanNode{
		bounds: ln.bounds,
		id:     ln.id,
		Spec:   pspec,
		Source: ln.Source,
	}

	ReplaceNode(pn, &newNode)

	return &newNode, true, nil
}

// PhysicalProcedureSpec is similar to its logical counterpart but must provide a method to determine cost.
type PhysicalProcedureSpec interface {
	Kind() ProcedureKind
	Copy() ProcedureSpec
	Cost(inStats []Statistics) (cost Cost, outStats Statistics)
}

// PhysicalPlanNode represents a physical operation in a plan.
type PhysicalPlanNode struct {
	edges
	bounds
	id     NodeID
	Spec   PhysicalProcedureSpec
	Source []interpreter.StackEntry

	// The trigger spec defines how and when a transformation
	// sends its tables to downstream operators
	TriggerSpec TriggerSpec
}

// ID returns a human-readable id for this plan node.
func (ppn *PhysicalPlanNode) ID() NodeID {
	return ppn.id
}

// ProcedureSpec returns the procedure spec for this plan node.
func (ppn *PhysicalPlanNode) ProcedureSpec() ProcedureSpec {
	return ppn.Spec
}

func (ppn *PhysicalPlanNode) ReplaceSpec(newSpec ProcedureSpec) error {
	physSpec, ok := newSpec.(PhysicalProcedureSpec)
	if !ok {
		return &flux.Error{
			Code: codes.Internal,
			Msg:  fmt.Sprintf("couldn't replace ProcedureSpec for physical plan node \"%v\"", ppn.ID()),
		}
	}

	ppn.Spec = physSpec
	return nil
}

// Kind returns the procedure kind for this plan node.
func (ppn *PhysicalPlanNode) Kind() ProcedureKind {
	return ppn.Spec.Kind()
}

func (ppn *PhysicalPlanNode) CallStack() []interpreter.StackEntry {
	return ppn.Source
}

func (ppn *PhysicalPlanNode) ShallowCopy() Node {
	newNode := new(PhysicalPlanNode)
	newNode.edges = ppn.edges.shallowCopy()
	newNode.id = ppn.id + "_copy"
	// TODO: the type assertion below... is it needed?
	newNode.Spec = ppn.Spec.Copy().(PhysicalProcedureSpec)
	return newNode
}

// Cost provides the self-cost (i.e., does not include the cost of its predecessors) for
// this plan node.  Caller must provide statistics of predecessors to this node.
func (ppn *PhysicalPlanNode) Cost(inStats []Statistics) (cost Cost, outStats Statistics) {
	return ppn.Spec.Cost(inStats)
}

func (ppn *PhysicalPlanNode) OutputAttrs() PhysicalAttributes {
	if oa, ok := ppn.Spec.(OutputAttributer); ok {
		return oa.OutputAttributes()
	}
	return nil
}

func (ppn *PhysicalPlanNode) RequiredAttrs() []PhysicalAttributes {
	if ra, ok := ppn.Spec.(RequiredAttributer); ok {
		return ra.RequiredAttributes()
	}
	return nil
}

// CreatePhysicalNode creates a single physical plan node from a procedure spec.
// The newly created physical node has no incoming or outgoing edges.
func CreatePhysicalNode(id NodeID, spec PhysicalProcedureSpec) *PhysicalPlanNode {
	return &PhysicalPlanNode{
		id:   id,
		Spec: spec,
	}
}

type nodeIDKey string

const NextPlanNodeIDKey nodeIDKey = "NextPlanNodeID"

func CreateUniquePhysicalNode(ctx context.Context, prefix string, spec PhysicalProcedureSpec) *PhysicalPlanNode {
	if value := ctx.Value(NextPlanNodeIDKey); value != nil {
		nextNodeID := value.(*int)
		id := NodeID(fmt.Sprintf("%s%d", prefix, *nextNodeID))
		*nextNodeID++
		return CreatePhysicalNode(id, spec)
	}
	return CreatePhysicalNode(NodeID(prefix), spec)
}

// PostPhysicalValidator provides an interface that can be implemented by PhysicalProcedureSpecs for any
// validation checks to be performed post-physical planning.
type PostPhysicalValidator interface {
	PostPhysicalValidate(id NodeID) error
}
