package rules

import (
	"github.com/influxdata/flux/functions/inputs"
	"github.com/influxdata/flux/functions/transformations"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
)

// FromRangeTransformationRule rewrites the sequence from() |> range()
type FromRangeTransformationRule struct{}

// Pattern returns a pattern of the form:
//     range
//       |
//     from
func (rule FromRangeTransformationRule) Pattern() plan.Pattern {
	return plan.Pat(transformations.RangeKind, plan.Pat(inputs.FromKind))
}

// Rewrite performs the logical-to-physical transformation:
//     range
//       |     ->     FromRange()
//     from
func (rule FromRangeTransformationRule) Rewrite(node plan.PlanNode) (plan.PlanNode, bool) {
	rangeNode := node.ProcedureSpec().(*transformations.RangeProcedureSpec)
	fromNode := node.Predecessors()[0].ProcedureSpec().(*inputs.FromProcedureSpec)

	return &plan.PhysicalPlanNode{
		Spec: &plan.FromRangeProcedureSpec{
			Bucket:   fromNode.Bucket,
			BucketID: fromNode.BucketID,
			Bounds:   rangeNode.Bounds,
			TimeCol:  rangeNode.TimeCol,
			StartCol: rangeNode.StartCol,
			StopCol:  rangeNode.StopCol,
		},
	}, true
}

// FromTagFilterTransformationRule rewrites the sequence from() |> range() |> filter()
type FromTagFilterTransformationRule struct {
	visitor *PredicateVisitor
	catalog *Catalog
}

// Mock catalog for testing purposes
type Catalog struct {
	isTag map[string]bool
}

func (c *Catalog) IsTag(name string) bool {
	return c.isTag[name]
}

// PredicateVisitor is used to visit a filter's predicate, assert that it
// is a conjuctive predicate, and store all referenced (tag, tag value) pairs.
type PredicateVisitor struct {
	conjuctive        bool
	tagFilters        map[string]string
	predicateFunction *semantic.FunctionExpression
}

func (v *PredicateVisitor) Visit(node semantic.Node) semantic.Visitor {
	panic("Not Implemented")
}

func (v *PredicateVisitor) Done() {}

// Pattern returns a pattern of the form:
//     filter
//       |
//     from
func (rule FromTagFilterTransformationRule) Pattern() plan.Pattern {
	return plan.Pat(transformations.FilterKind, plan.Pat(inputs.FromKind))
}

// Rewrite performs the logical-to-physical transformation:
//                 filter
// ( tag0 == "something" AND <conjunctive predicate> )    ->    FromTagFilter()
//                   |
//                 from
func (rule FromTagFilterTransformationRule) Rewrite(node plan.PlanNode) (plan.PlanNode, bool) {
	filterNode := node.ProcedureSpec().(*transformations.FilterProcedureSpec)
	fromNode := node.Predecessors()[0].ProcedureSpec().(*inputs.FromProcedureSpec)

	predicate := filterNode.Fn
	semantic.Walk(rule.visitor, predicate)

	if !rule.visitor.conjuctive || len(rule.visitor.tagFilters) == 0 {
		return nil, false
	}

	tagFilters := make(map[string]string, len(rule.visitor.tagFilters))
	for k, v := range rule.visitor.tagFilters {
		tagFilters[k] = v
	}

	return &plan.PhysicalPlanNode{
		Spec: &plan.FromTagFilterProcedureSpec{
			Bucket:   fromNode.Bucket,
			BucketID: fromNode.BucketID,

			TagEqualityFilters: tagFilters,
			PredicateFunction:  rule.visitor.predicateFunction,
		},
	}, true
}
