package plan

import (
	"context"
	"fmt"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
)

// LogicalPlanner translates a flux.Spec into a plan.Spec and applies any
// registered logical rules to the plan.
//
// Logical planning should transform the plan in ways that are independent of
// actual physical algorithms used to implement operations, and independent of
// the actual data being processed.
type LogicalPlanner interface {
	CreateInitialPlan(spec *flux.Spec) (*Spec, error)
	Plan(context.Context, *Spec) (*Spec, error)
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

	thePlanner.addRules(rules...)

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
	disableIntegrityChecks bool
}

// OnlyLogicalRules produces a logical plan option that forces only a set of particular rules to be
// applied.
func OnlyLogicalRules(rules ...Rule) LogicalOption {
	return logicalOption(func(lp *logicalPlanner) {
		lp.clearRules()
		lp.addRules(rules...)
	})
}

func AddLogicalRules(rules ...Rule) LogicalOption {
	return logicalOption(func(lp *logicalPlanner) {
		lp.addRules(rules...)
	})
}

func RemoveLogicalRules(rules ...string) LogicalOption {
	return logicalOption(func(lp *logicalPlanner) {
		lp.removeRules(rules...)
	})
}

// DisableIntegrityChecks disables integrity checks in the logical planner.
func DisableIntegrityChecks() LogicalOption {
	return logicalOption(func(lp *logicalPlanner) {
		lp.disableIntegrityChecks = true
	})
}

// CreateInitialPlan translates the flux.Spec into an unoptimized, naive plan.
func (l *logicalPlanner) CreateInitialPlan(spec *flux.Spec) (*Spec, error) {
	return createLogicalPlan(spec)
}

// Plan transforms the given naive plan by applying rules.
func (l *logicalPlanner) Plan(ctx context.Context, logicalPlan *Spec) (*Spec, error) {
	newLogicalPlan, err := l.heuristicPlanner.Plan(ctx, logicalPlan)
	if err != nil {
		return nil, err
	}

	// check integrity after planning is complete
	if !l.disableIntegrityChecks {
		err := newLogicalPlan.CheckIntegrity()
		if err != nil {
			return nil, err
		}
	}

	return newLogicalPlan, nil
}

type administration struct {
	now time.Time
}

func (a administration) Now() time.Time {
	return a.now
}

// LogicalNode consists of the input and output edges and a procedure spec
// that describes what the node does.
type LogicalNode struct {
	edges
	bounds
	id     NodeID
	Spec   ProcedureSpec
	Source []interpreter.StackEntry
}

// ID returns a human-readable identifier unique to this plan.
func (lpn *LogicalNode) ID() NodeID {
	return lpn.id
}

// Kind returns the kind of procedure performed by this plan node.
func (lpn *LogicalNode) Kind() ProcedureKind {
	return lpn.Spec.Kind()
}

// CallStack returns the call stack that created this LogicalNode.
func (lpn *LogicalNode) CallStack() []interpreter.StackEntry {
	return lpn.Source
}

// ProcedureSpec returns the procedure spec for this plan node.
func (lpn *LogicalNode) ProcedureSpec() ProcedureSpec {
	return lpn.Spec
}

func (lpn *LogicalNode) ReplaceSpec(newSpec ProcedureSpec) error {
	lpn.Spec = newSpec
	return nil
}

func (lpn *LogicalNode) ShallowCopy() Node {
	newNode := new(LogicalNode)
	newNode.edges = lpn.edges.shallowCopy()
	newNode.id = lpn.id + "_copy"
	newNode.Spec = lpn.Spec.Copy()
	return newNode
}

// createLogicalPlan creates a logical query plan from a flux spec
func createLogicalPlan(spec *flux.Spec) (*Spec, error) {
	nodes := make(map[flux.OperationID]Node, len(spec.Operations))
	admin := administration{now: spec.Now}

	plan := NewPlanSpec()
	plan.Resources = spec.Resources
	plan.Now = spec.Now
	plan.NextNodeId = spec.NextNodeId

	v := &fluxSpecVisitor{
		a:          admin,
		spec:       spec,
		plan:       plan,
		nodes:      nodes,
		yieldNames: make(map[string]struct{}),
	}

	if err := spec.Walk(v.visitOperation); err != nil {
		return nil, err
	}

	return v.plan, nil
}

// fluxSpecVisitor visits a flux spec and constructs from it a logical plan DAG
type fluxSpecVisitor struct {
	a          Administration
	spec       *flux.Spec
	plan       *Spec
	nodes      map[flux.OperationID]Node
	yieldNames map[string]struct{}
}

func (v *fluxSpecVisitor) addYieldName(pn Node) error {
	yieldSpec := pn.ProcedureSpec().(YieldProcedureSpec)
	name := yieldSpec.YieldName()
	_, isDup := v.yieldNames[name]
	if isDup {
		return errors.Newf(codes.Invalid, "found more than one call to yield() with the name %q", name)
	}

	v.yieldNames[name] = struct{}{}
	return nil
}

func generateYieldNode(pred Node) Node {
	yieldSpec := &GeneratedYieldProcedureSpec{Name: DefaultYieldName}
	yieldNode := CreateLogicalNode(NodeID("generated_yield"), yieldSpec)
	pred.AddSuccessors(yieldNode)
	yieldNode.AddPredecessors(pred)
	return yieldNode
}

// visitOperation takes a flux spec operation, converts it to its equivalent
// logical procedure spec, and adds it to the current logical plan DAG.
func (v *fluxSpecVisitor) visitOperation(o *flux.Operation) error {
	// Retrieve the create function for this query operation
	createFns, ok := createProcedureFnsFromKind(o.Spec.Kind())

	if !ok {
		return fmt.Errorf("no ProcedureSpec available for %s", o.Spec.Kind())
	}

	// TODO: differentiate between logical and physical procedures.
	// There should be just one logical procedure for each operation, but could be
	// several physical procedures.
	create := createFns[0]

	// Create a ProcedureSpec from the query operation procedureSpec
	procedureSpec, err := create(o.Spec, v.a)
	if err != nil {
		return err
	}

	// Create a LogicalNode using the ProcedureSpec
	logicalNode := CreateLogicalNode(NodeID(o.ID), procedureSpec)
	logicalNode.Source = o.Source.Stack

	v.nodes[o.ID] = logicalNode

	// Add this node to the logical plan by connecting predecessors and successors
	for _, parent := range v.spec.Parents(o.ID) {
		logicalParent := v.nodes[parent.ID]
		logicalNode.AddPredecessors(logicalParent)
		logicalParent.AddSuccessors(logicalNode)
	}

	_, isYield := procedureSpec.(YieldProcedureSpec)
	if isYield {
		err = v.addYieldName(logicalNode)
		if err != nil {
			return err
		}
	}

	// no children => no successors => root node
	if len(v.spec.Children(o.ID)) == 0 {
		if isYield || HasSideEffect(procedureSpec) {
			v.plan.Roots[logicalNode] = struct{}{}
		} else {
			// Generate a yield node
			generateYieldNode := generateYieldNode(logicalNode)
			err = v.addYieldName(generateYieldNode)
			if err != nil {
				return err
			}
			v.plan.Roots[generateYieldNode] = struct{}{}

		}
	}

	return nil
}

// CreateLogicalNode creates a single logical plan node from a procedure spec.
// The newly created logical node has no incoming or outgoing edges.
func CreateLogicalNode(id NodeID, spec ProcedureSpec) *LogicalNode {
	return &LogicalNode{
		id:   id,
		Spec: spec,
	}
}
