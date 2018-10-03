package planner

import (
	"github.com/influxdata/flux/semantic"
)

// FromRangeTransformationRule rewrites the sequence from() |> range()
type FromRangeTransformationRule struct{}

// Pattern returns a pattern of the form:
//     range
//       |
//     from
func (rule FromRangeTransformationRule) Pattern() Pattern {
	return &TreePattern{
		RootType: ProcedureKind("RangeKind"),
		Predecessors: []Pattern{
			&LeafPattern{
				RootType: ProcedureKind("FromKind"),
			},
		},
	}
}

// Rewrite performs the logical-to-physical transformation:
//     range
//       |     ->     FromRange()
//     from
func (rule FromRangeTransformationRule) Rewrite(node PlanNode) (PlanNode, bool) {
	rangeNode := node.ProcedureSpec().(*RangeProcedureSpec)
	fromNode := node.Predecessors()[0].ProcedureSpec().(*FromProcedureSpec)
	return &PhysicalPlanNode{
		Spec: &FromRangeProcedureSpec{
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
func (rule FromTagFilterTransformationRule) Pattern() Pattern {
	return &TreePattern{
		RootType: ProcedureKind("FilterKind"),
		Predecessors: []Pattern{
			&TreePattern{
				RootType: ProcedureKind("RangeKind"),
				Predecessors: []Pattern{
					&LeafPattern{
						RootType: ProcedureKind("FromKind"),
					},
				},
			},
		},
	}
}

// Rewrite performs the logical-to-physical transformation:
//                 filter
// ( tag0 == "something" AND <conjunctive predicate> )    ->    FromRangeTagFilter()
//                   |
//                 from
func (rule FromTagFilterTransformationRule) Rewrite(node PlanNode) (PlanNode, bool) {
	filterNode := node.ProcedureSpec().(*FilterProcedureSpec)
	fromNode := node.Predecessors()[0].ProcedureSpec().(*FromProcedureSpec)

	predicate := filterNode.Fn
	semantic.Walk(rule.visitor, predicate)

	if !rule.visitor.conjuctive || len(rule.visitor.tagFilters) == 0 {
		return nil, false
	}

	tagFilters := make(map[string]string, len(rule.visitor.tagFilters))
	for k, v := range rule.visitor.tagFilters {
		tagFilters[k] = v
	}

	return &PhysicalPlanNode{
		Spec: &FromTagFilterProcedureSpec{
			Bucket:   fromNode.Bucket,
			BucketID: fromNode.BucketID,

			TagEqualityFilters: tagFilters,
			PredicateFunction:  rule.visitor.predicateFunction,
		},
	}, true
}
