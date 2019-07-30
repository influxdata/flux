package universe

import (
	"fmt"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const SlidingReduceKind = "slidingReduce"

type SlidingReduceOpSpec struct {
	N  int64                        `json:"n"`
	Fn *semantic.FunctionExpression `json:"fn"`
}

func init() {
	slidingReduceSignature := flux.FunctionSignature(
		map[string]semantic.PolyType{
			"n": semantic.Int,
			"fn": semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
				Parameters: map[string]semantic.PolyType{
					"rows": semantic.Tvar(1),
				},
				Required: semantic.LabelSet{"r"},
				Return: semantic.Tvar(2),
			}),
		},
		[]string{"n", "fn"},
	)

}

func createSlidingReduceOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(SlidingReduceOpSpec)

	if n, err := args.GetRequiredInt("n"); err != nil {
		return nil, err
	} else {
		spec.N = n
	}

	if f, err := args.GetRequiredFunction("fn"); err != nil {
		return nil, err
	} else {
		fn, err := interpreter.ResolveFunction(f)
		if err != nil {
			return nil, err
		}
		spec.Fn = fn
	}

	return spec, nil
}

func newSlidingReduceOp() flux.OperationSpec {
	return new(SlidingReduceOpSpec)
}

func (s *SlidingReduceOpSpec) Kind() flux.OperationKind {
	return SlidingReduceKind
}

type SlidingReduceProcedureSpec struct {
	plan.DefaultCost
	N  int64
	Fn *semantic.FunctionExpression
}

func newSlidingReduceProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*SlidingReduceOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &SlidingReduceProcedureSpec{
		N: spec.N,
		Fn: spec.Fn,
	}, nil
}

func (s *SlidingReduceProcedureSpec) Kind() plan.ProcedureKind {
	return SlidingReduceKind
}

func (s *SlidingReduceProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new (SlidingReduceProcedureSpec)
	*ns = *s
	ns.Fn = s.Fn.Copy().(*semantic.FunctionExpression)
	return ns
}

func createSlidingReduceTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*SlidingReduceProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t, err := NewSlidingReduceTransformation(d, cache, s)
	if err != nil {
		return nil, nil, err
	}
	return t, d, nil
}

type slidingReduceTransformation struct {
	d execute.Dataset
	cache execute.TableBuilderCache

	n int64
	fn *execute.RowListReduceFn
}

func NewSlidingReduceTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *SlidingReduceProcedureSpec) (*slidingReduceTransformation, error) {
	fn, err := execute.NewRowListReduceFn(spec.Fn)
	if err != nil {
		return nil, err
	}

	return &slidingReduceTransformation{
		d: d,
		cache: cache,
		fn: fn,
		n: spec.N,
	}, nil
}

func (t *slidingReduceTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	cols := tbl.Cols()
	if err := t.fn.Prepare(cols); err != nil {
		return err
	}

	data := make([][]values.Value, 0)

	if err := tbl.Do(func(cr flux.ColReader) error {
		l := cr.Len()
		for i := 0; i < l; i++ {
			data = append(data, make([]values.Value, len(cols)))
			for j := 0; j < len(cols); j++ {
				data[i][j] = execute.ValueForRow(cr, i, j)
			}
		}
		return nil
	}); err != nil {
		return err
	}


}

func (t *slidingReduceTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *slidingReduceTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *slidingReduceTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *slidingReduceTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
