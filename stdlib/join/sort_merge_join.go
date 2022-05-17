package join

import (
	"context"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/stdlib/universe"
)

const SortMergeJoinKind = "sortmergejoin"

func init() {
	plan.RegisterPhysicalRules(SortMergeJoinPredicateRule{})
	execute.RegisterTransformation(SortMergeJoinKind, createSortMergeJoinTransformation)
}

func createSortMergeJoinTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*SortMergeJoinProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}

	return newSortMergeJoinTransformation(id, s, a.Allocator())
}

func newSortMergeJoinTransformation(id execute.DatasetID, s *SortMergeJoinProcedureSpec, mem memory.Allocator) (execute.Transformation, execute.Dataset, error) {
	return nil, nil, errors.Newf(codes.Internal, "sort merge join transformation is not implemented")
}

type SortMergeJoinProcedureSpec struct {
	As     interpreter.ResolvedFunction
	Left   *flux.TableObject
	Right  *flux.TableObject
	Method string
}

func (p *SortMergeJoinProcedureSpec) Kind() plan.ProcedureKind {
	return plan.ProcedureKind(SortMergeJoinKind)
}

func (p *SortMergeJoinProcedureSpec) Copy() plan.ProcedureSpec {
	return &SortMergeJoinProcedureSpec{
		As:     p.As,
		Left:   p.Left,
		Right:  p.Right,
		Method: p.Method,
	}
}

func (p *SortMergeJoinProcedureSpec) Cost(inStats []plan.Statistics) (cost plan.Cost, outStats plan.Statistics) {
	return plan.Cost{}, plan.Statistics{}
}

func newSortMergeJoin(spec *JoinProcedureSpec) *SortMergeJoinProcedureSpec {
	return &SortMergeJoinProcedureSpec{
		As:     spec.As,
		Left:   spec.Left,
		Right:  spec.Right,
		Method: spec.Method,
	}
}

type SortMergeJoinPredicateRule struct{}

func (SortMergeJoinPredicateRule) Name() string {
	return "sortMergeJoinPredicate"
}

func (SortMergeJoinPredicateRule) Pattern() plan.Pattern {
	return plan.Pat(Join2Kind, plan.Any(), plan.Any())
}

func (SortMergeJoinPredicateRule) Rewrite(ctx context.Context, n plan.Node) (plan.Node, bool, error) {
	s := n.ProcedureSpec()
	spec, ok := s.(*JoinProcedureSpec)
	if !ok {
		return nil, false, errors.New(codes.Internal, "invalid spec type on join node")
	}

	for _, parentNode := range n.Predecessors() {
		sortProc := universe.SortProcedureSpec{}
		sortNode := plan.CreateUniquePhysicalNode(ctx, "sortMergeJoin", &sortProc)
		for i, successorNode := range n.Successors() {
			if successorNode.ID() == n.ID() {
				n.Successors()[i] = sortNode
			}
		}
		sortNode.AddPredecessors(parentNode)
		sortNode.AddSuccessors(n)
	}

	n.ReplaceSpec(newSortMergeJoin(spec))

	return n, true, nil
}
