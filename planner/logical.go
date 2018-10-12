package planner

import (
	"errors"
	"fmt"

	"github.com/influxdata/flux"
)

// LogicalPlanner translates a flux.Spec into a PlanSpec and applies any
// registered logical rules to the plan.
//
// Logical planning should transform the plan in ways that are independent of
// actual physical algorithms used to implement operations, and independent of
// the actual data being processed.
type LogicalPlanner interface {
	Plan(spec *flux.Spec) (*PlanSpec, error)
}

// NewLogicalPlanner returns a new logical planner with the given options.
// The planner will be configured to apply any logical rules that have
// been registered.
func NewLogicalPlanner(options ...LogicalOption) LogicalPlanner {
	thePlanner := &logicalPlanner{
		heuristicPlanner: newHeuristicPlanner(),
	}

	// TODO: add any logical rules that have been registered:
	thePlanner.addRules([]Rule{})

	// Options may add or remove rules, so process them after we've
	// added registered rules.
	for _, opt := range options {
		opt.apply(thePlanner)
	}

	return thePlanner
}

// LogicalOption is an option to configure the behavior of the logical planner.
type LogicalOption interface {
	apply(*logicalPlanner)
}

type logicalOption func(*logicalPlanner)

func (opt logicalOption) apply(lp *logicalPlanner) {
	opt(lp)
}

type logicalPlanner struct {
	*heuristicPlanner
}

// WithRule produces a logical planner option that forces a particular rule to be
// applied.
func WithRule(rule Rule) LogicalOption {
	return logicalOption(func(lp *logicalPlanner) {
		lp.addRules([]Rule{rule})
	})
}

// Plan translates the given flux.Spec to a plan and transforms it by applying rules.
func (l *logicalPlanner) Plan(spec *flux.Spec) (*PlanSpec, error) {
	logicalPlan, err := createLogicalPlan(spec, l)
	if err != nil {
		return nil, err
	}

	return l.heuristicPlanner.Plan(logicalPlan)
}

func (logicalPlanner) ConvertID(oid flux.OperationID) ProcedureID {
	return ProcedureIDFromOperationID(oid)
}

// LogicalPlanNode consists of the input and output edges and a procedure spec
// that describes what the node does.
type LogicalPlanNode struct {
	edges
	id   NodeID
	Spec ProcedureSpec
}

// ID returns a human-readable identifier unique to this plan.
func (lpn *LogicalPlanNode) ID() NodeID {
	return lpn.id
}

// Kind returns the kind of procedure performed by this plan node.
func (lpn *LogicalPlanNode) Kind() ProcedureKind {
	return lpn.Spec.Kind()
}

// ProcedureSpec returns the procedure spec for this plan node.
func (lpn *LogicalPlanNode) ProcedureSpec() ProcedureSpec {
	return lpn.Spec
}

// createLogicalPlan creates a logical query plan from a flux spec
func createLogicalPlan(spec *flux.Spec, a Administration) (*PlanSpec, error) {
	nodes := make(map[flux.OperationID]PlanNode, len(spec.Operations))

	v := &fluxSpecVisitor{
		a:     a,
		spec:  spec,
		nodes: nodes,
	}

	if err := spec.Walk(v.visitOperation); err != nil {
		return nil, err
	}

	logicalPlan, err := validate(CreatePlanSpec(v.roots, spec.Resources, spec.Now))

	if err != nil {
		return nil, err
	}

	return logicalPlan, nil
}

// fluxSpecVisitor visits a flux spec and constructs from it a logical plan DAG
type fluxSpecVisitor struct {
	a     Administration
	spec  *flux.Spec
	roots []PlanNode
	nodes map[flux.OperationID]PlanNode
}

// visitOperation takes a flux spec operation, converts it to its equivalent
// logical procedure spec, and adds it to the current logical plan DAG.
func (v *fluxSpecVisitor) visitOperation(o *flux.Operation) error {
	// Retrieve the create function for this query operation
	createFns, ok := queryOpToProcedure[o.Spec.Kind()]

	if !ok {
		return fmt.Errorf("No ProcedureSpec available for %s", o.Spec.Kind())
	}

	// TODO: differentiate between logical and physical procedures.
	// There should be just one logical procedure for each operation, but could be
	// several physical procedures.
	create := createFns[0]

	// Create a ProcedureSpec from the query operation spec
	spec, err := create(o.Spec, v.a)

	if err != nil {
		return err
	}

	// Create a LogicalPlanNode using the ProcedureSpec
	logicalNode := CreateLogicalNode(NodeID(o.ID), spec)

	v.nodes[o.ID] = logicalNode

	// Add this node to the logical plan by connecting predecessors and successors
	for _, parent := range v.spec.Parents(o.ID) {
		logicalParent := v.nodes[parent.ID]
		logicalNode.AddPredecessors(logicalParent)
		logicalParent.AddSuccessors(logicalNode)
	}

	// no children => no successors => root node
	if len(v.spec.Children(o.ID)) == 0 {
		v.roots = append(v.roots, logicalNode)
	}

	return nil
}

func validate(plan *PlanSpec) (*PlanSpec, error) {
	if len(plan.Results()) > 1 {
		names := make(map[string]struct{}, len(plan.Results()))

		for _, root := range plan.Results() {
			spec, ok := root.ProcedureSpec().(YieldProcedureSpec)

			if !ok {
				return nil, errors.New("query must have explicit yields for multiple result")
			}

			if name, ok := names[spec.YieldName()]; ok {
				return nil, fmt.Errorf("found duplicate yield name %q", name)
			}

			if len(root.Predecessors()) != 1 {
				return nil, errors.New("yield procedures must have exactly one predecessor")
			}
		}
	}
	return plan, nil
}

// CreateLogicalNode creates a single logical plan node from a procedure spec.
// The newly created logical node has no incoming or outgoing edges.
func CreateLogicalNode(id NodeID, spec ProcedureSpec) *LogicalPlanNode {
	return &LogicalPlanNode{
		id:   id,
		Spec: spec,
	}
}
