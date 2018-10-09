package planner

import (
	"fmt"

	"github.com/influxdata/flux"
)

// LogicalProcedureSpec is just a ProcedureSpec.
type LogicalProcedureSpec interface {
	Kind() ProcedureKind
	Copy() ProcedureSpec
}

// LogicalPlanNode consists of the input and output edges and a procedure spec
// that describes what the node does.
type LogicalPlanNode struct {
	Edges
	id   NodeID
	Spec LogicalProcedureSpec
}

func (lpn *LogicalPlanNode) ID() NodeID {
	return lpn.id
}

func (lpn *LogicalPlanNode) Kind() ProcedureKind {
	return lpn.Spec.Kind()
}

func (lpn *LogicalPlanNode) ProcedureSpec() ProcedureSpec {
	return lpn.Spec
}

// CreateLogicalPlan creates a logical query plan from a flux spec
func CreateLogicalPlan(spec *flux.Spec, ops map[flux.OperationKind]CreateLogicalProcedureSpec) (*QueryPlan, error) {
	nodes := make(map[flux.OperationID]PlanNode, len(spec.Operations))

	v := &fluxSpecVisitor{
		ops:   ops,
		spec:  spec,
		nodes: nodes,
	}

	if err := spec.Walk(v.VisitOperation); err != nil {
		return nil, err
	}

	return NewQueryPlan(v.roots), nil
}

// CreateLogicalProcedureSpec takes a flux spec operation and returns the equivalent LogicalProcedureSpec
type CreateLogicalProcedureSpec func(flux.OperationSpec) (LogicalProcedureSpec, error)

// fluxSpecVisitor visits a flux spec and constructs from it a logical plan DAG
type fluxSpecVisitor struct {
	ops   map[flux.OperationKind]CreateLogicalProcedureSpec
	spec  *flux.Spec
	roots []PlanNode
	nodes map[flux.OperationID]PlanNode
}

// VisitOperation takes a flux spec operation, converts it to its equivalent
// logical procedure spec, and adds it to the current logical plan DAG.
func (v *fluxSpecVisitor) VisitOperation(o *flux.Operation) error {
	// Retrieve the create function for this query operation
	create, ok := v.ops[o.Spec.Kind()]

	if !ok {
		return fmt.Errorf("No LogicalProcedureSpec available for %s", o.Spec.Kind())
	}

	// Create a LogicalProcedureSpec from the query operation spec
	spec, err := create(o.Spec)

	if err != nil {
		return err
	}

	// Create a LogicalPlanNode using the LogicalProcedureSpec
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

// CreateLogicalNode creates a single logical plan node from a procedure spec.
// The newly created logical node has no incoming or outgoing edges.
func CreateLogicalNode(id NodeID, spec ProcedureSpec) *LogicalPlanNode {
	return &LogicalPlanNode{
		id:   id,
		Spec: spec,
	}
}
