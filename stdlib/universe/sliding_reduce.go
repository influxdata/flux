//package universe
//
//import (
//	"fmt"
//	"github.com/influxdata/flux"
//	"github.com/influxdata/flux/execute"
//	"github.com/influxdata/flux/interpreter"
//	"github.com/influxdata/flux/plan"
//	"github.com/influxdata/flux/semantic"
//)
//
//const SlidingReduceKind = "slidingReduce"
//
//type SlidingReduceOpSpec struct {
//	N  int64                        `json:"n"`
//	Fn *semantic.FunctionExpression `json:"fn"`
//}
//
//func init() {
//	slidingReduceSignature := flux.FunctionSignature(
//		map[string]semantic.PolyType{
//			"n": semantic.Int,
//			"fn": semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
//				Parameters: map[string]semantic.PolyType{
//					"rows": semantic.Tvar(1),
//				},
//				Required: semantic.LabelSet{"r"},
//				Return: semantic.Tvar(2),
//			}),
//		},
//		[]string{"n", "fn"},
//	)
//
//}
//
//func createSlidingReduceOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
//	if err := a.AddParentFromArgs(args); err != nil {
//		return nil, err
//	}
//
//	spec := new(SlidingReduceOpSpec)
//
//	if n, err := args.GetRequiredInt("n"); err != nil {
//		return nil, err
//	} else {
//		spec.N = n
//	}
//
//	if f, err := args.GetRequiredFunction("fn"); err != nil {
//		return nil, err
//	} else {
//		fn, err := interpreter.ResolveFunction(f)
//		if err != nil {
//			return nil, err
//		}
//		spec.Fn = fn
//	}
//
//	return spec, nil
//}
//
//func newSlidingReduceOp() flux.OperationSpec {
//	return new(SlidingReduceOpSpec)
//}
//
//func (s *SlidingReduceOpSpec) Kind() flux.OperationKind {
//	return SlidingReduceKind
//}
//
//type SlidingReduceProcedureSpec struct {
//	plan.DefaultCost
//	N  int64
//	Fn *semantic.FunctionExpression
//}
//
//func newSlidingReduceProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
//	spec, ok := qs.(*SlidingReduceOpSpec)
//	if !ok {
//		return nil, fmt.Errorf("invalid spec type %T", qs)
//	}
//
//	return &SlidingReduceProcedureSpec{
//		N: spec.N,
//		Fn: spec.Fn,
//	}, nil
//}
//
//func (s *SlidingReduceProcedureSpec) Kind() plan.ProcedureKind {
//	return SlidingReduceKind
//}
//
//func (s *SlidingReduceProcedureSpec) Copy() plan.ProcedureSpec {
//	ns := new (SlidingReduceProcedureSpec)
//	*ns = *s
//	ns.Fn = s.Fn.Copy().(*semantic.FunctionExpression)
//	return ns
//}
//
//func createSlidingReduceTransformation() {
//
//}
//
//type slidingReduceTransformation struct {
//	d execute.Dataset
//	cache execute.TableBuilderCache
//
//	n int64
//}