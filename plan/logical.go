package plan

import (
	"fmt"
	"time"

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

// NewLogicalPlanner returns a new logical plan with the given options.
// The plan will be configured to apply any logical rules that have
// been registered.
func NewLogicalPlanner(options ...LogicalOption) LogicalPlanner {
	thePlanner := &logicalPlanner{
		heuristicPlanner: newHeuristicPlanner(),
	}

	rules := make([]Rule, len(ruleNameToLogicalRule))
	i := 0
	for _, v := range ruleNameToLogicalRule {
		rules[i] = v
		i++
	}

	thePlanner.addRules(rules)

	// Options may add or remove rules, so process them after we've
	// added registered rules.
	for _, opt := range options {
		opt.apply(thePlanner)
	}

	return thePlanner
}

// LogicalOption is an option to configure the behavior of the logical plan.
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

// WithLogicalRule produces a logical plan option that forces a particular rule to be
// applied.
func WithLogicalRule(rule Rule) LogicalOption {
	return logicalOption(func(lp *logicalPlanner) {
		lp.addRules([]Rule{rule})
	})
}

// Plan translates the given flux.Spec to a plan and transforms it by applying rules.
func (l *logicalPlanner) Plan(spec *flux.Spec) (*PlanSpec, error) {
	logicalPlan, err := createLogicalPlan(spec)
	if err != nil {
		return nil, err
	}

	newLogicalPlan, err := l.heuristicPlanner.Plan(logicalPlan)
	if err != nil {
		return nil, err
	}

	return newLogicalPlan, nil
}

// TODO: This is unnecessary if we no longer need ProcedureIDs
type administration struct {
	now time.Time
}

func (a administration) Now() time.Time {
	return a.now
}

func (a administration) ConvertID(oid flux.OperationID) ProcedureID {
	return ProcedureIDFromOperationID(oid)
}

// LogicalPlanNode consists of the input and output edges and a procedure spec
// that describes what the node does.
type LogicalPlanNode struct {
	edges
	bounds
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

func (lpn *LogicalPlanNode) ShallowCopy() PlanNode {
	newNode := new(LogicalPlanNode)
	newNode.edges = lpn.edges.shallowCopy()
	newNode.id = lpn.id + "_copy"
	newNode.Spec = lpn.Spec.Copy()
	return newNode
}

// createLogicalPlan creates a logical query plan from a flux spec
func createLogicalPlan(spec *flux.Spec) (*PlanSpec, error) {
	nodes := make(map[flux.OperationID]PlanNode, len(spec.Operations))
	admin := administration{now: spec.Now}

	plan := NewPlanSpec()
	plan.Resources = spec.Resources
	plan.Now = spec.Now

	v := &fluxSpecVisitor{
		a:     admin,
		spec:  spec,
		plan:  plan,
		nodes: nodes,
	}

	if err := spec.Walk(v.visitOperation); err != nil {
		return nil, err
	}

	return v.plan, nil
}

// fluxSpecVisitor visits a flux spec and constructs from it a logical plan DAG
type fluxSpecVisitor struct {
	a     Administration
	spec  *flux.Spec
	plan  *PlanSpec
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
		v.plan.Roots[logicalNode] = struct{}{}
	}

	return nil
}

// CreateLogicalNode creates a single logical plan node from a procedure spec.
// The newly created logical node has no incoming or outgoing edges.
func CreateLogicalNode(id NodeID, spec ProcedureSpec) *LogicalPlanNode {
	return &LogicalPlanNode{
		id:   id,
		Spec: spec,
	}
}
