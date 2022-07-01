package interpolate

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
)

const LinearInterpolateKind = "linearInterpolateKind"

type LinearInterpolateOpSpec struct {
	Every flux.Duration `json:"every"`
}

func init() {
	runtime.RegisterPackageValue("interpolate", "linear",
		flux.MustValue(flux.FunctionValue("linear",
			createInterpolateOpSpec,
			runtime.MustLookupBuiltinType("interpolate", "linear"),
		)),
	)
	flux.RegisterOpSpec(LinearInterpolateKind,
		func() flux.OperationSpec {
			return new(LinearInterpolateOpSpec)
		},
	)
	plan.RegisterProcedureSpec(
		LinearInterpolateKind,
		newInterpolateProcedure,
		LinearInterpolateKind,
	)
	execute.RegisterTransformation(
		LinearInterpolateKind,
		createInterpolateTransformation,
	)
}

func createInterpolateOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	every, err := args.GetRequiredDuration("every")
	if err != nil {
		return nil, err
	}

	return &LinearInterpolateOpSpec{
		Every: every,
	}, nil
}

func (s *LinearInterpolateOpSpec) Kind() flux.OperationKind {
	return LinearInterpolateKind
}

type LinearInterpolateProcedureSpec struct {
	plan.DefaultCost
	Every flux.Duration `json:"every"`
}

func newInterpolateProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*LinearInterpolateOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	return &LinearInterpolateProcedureSpec{
		Every: spec.Every,
	}, nil
}

func (s *LinearInterpolateProcedureSpec) Kind() plan.ProcedureKind {
	return LinearInterpolateKind
}
func (s *LinearInterpolateProcedureSpec) Copy() plan.ProcedureSpec {
	return &LinearInterpolateProcedureSpec{
		Every: s.Every,
	}
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *LinearInterpolateProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

func createInterpolateTransformation(
	id execute.DatasetID,
	mode execute.AccumulationMode,
	spec plan.ProcedureSpec,
	a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*LinearInterpolateProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewInterpolateTransformation(d, cache, s)
	return t, d, nil
}

type interpolateTransformation struct {
	execute.ExecutionNode
	d      execute.Dataset
	cache  execute.TableBuilderCache
	spec   LinearInterpolateProcedureSpec
	window execute.Window
}

func NewInterpolateTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *LinearInterpolateProcedureSpec) *interpolateTransformation {
	return &interpolateTransformation{
		d:     d,
		cache: cache,
		spec:  *spec,
		window: execute.Window{
			Every:  spec.Every,
			Period: spec.Every,
		},
	}
}

func (t *interpolateTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *interpolateTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	key, columns, firstPoint := tbl.Key(), tbl.Cols(), true

	for _, c := range columns {
		if key.HasCol(c.Label) {
			continue
		}
		if c.Label == execute.DefaultTimeColLabel {
			continue
		}
		if c.Label == execute.DefaultValueColLabel {
			continue
		}
		return errors.Newf(codes.FailedPrecondition,
			"interpolate.linear requires column %q to be in group key", c.Label,
		)
	}

	b, created := t.cache.TableBuilder(tbl.Key())
	if !created {
		return errors.Newf(codes.FailedPrecondition,
			"duplicate table with key: %v", tbl.Key(),
		)
	}

	if err := execute.AddTableCols(tbl, b); err != nil {
		return err
	}

	ti := execute.ColIdx("_time", columns)
	if ti < 0 {
		return errors.New(codes.FailedPrecondition,
			"_time column does not exist",
		)
	}

	vi := execute.ColIdx("_value", columns)
	if vi < 0 {
		return errors.New(codes.FailedPrecondition,
			"_value column does not exist",
		)
	}

	if ty := columns[vi].Type; ty != flux.TFloat {
		return errors.Newf(codes.FailedPrecondition,
			"cannot interpolate %v values; expected float values", ty,
		)
	}

	fn := appendFn(b, ti, vi)

	var x0 int64
	var y0 float64

	if err := tbl.Do(func(cr flux.ColReader) error {

		if err := execute.AppendKeyValuesN(key, b, cr.Len()); err != nil {
			return err
		}

		tc := cr.Times(ti)
		vc := cr.Floats(vi)

		i := 0

		if firstPoint && cr.Len() > 0 {
			if tc.IsNull(i) {
				return errors.New(codes.FailedPrecondition,
					"null _time found during linear interpolation",
				)
			}
			x0 = tc.Value(i)

			if vc.IsNull(i) {
				return errors.New(codes.FailedPrecondition,
					"null _value found during linear interpolation",
				)
			}
			y0 = vc.Value(i)

			i++

			if err := fn(x0, y0); err != nil {
				return err
			}
		}
		for ; i < cr.Len(); i++ {

			if tc.IsNull(i) {
				return errors.New(codes.FailedPrecondition,
					"null _time found during linear interpolation",
				)
			}
			xn := tc.Value(i)

			if vc.IsNull(i) {
				return errors.New(codes.FailedPrecondition,
					"null _value found during linear interpolation",
				)
			}
			yn := vc.Value(i)

			xi := int64(t.window.GetEarliestBounds(values.Time(x0)).Stop)

			m := (yn - y0) / float64(xn-x0)

			for xi < xn {
				yi := float64(y0) + m*float64(xi-x0)

				if err := fn(xi, yi); err != nil {
					return err
				}

				if err := execute.AppendKeyValues(key, b); err != nil {
					return err
				}

				xi = int64(execute.Time(xi).Add(t.window.Every))
			}

			if err := fn(xn, yn); err != nil {
				return err
			}

			x0, y0, firstPoint = xn, yn, false
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func appendFn(b execute.TableBuilder, timeIdx, valueIdx int) func(int64, float64) error {
	return func(t int64, v float64) error {
		if err := b.AppendTime(timeIdx, execute.Time(t)); err != nil {
			return err
		}
		if err := b.AppendFloat(valueIdx, v); err != nil {
			return err
		}
		return nil
	}
}

func (t *interpolateTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *interpolateTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *interpolateTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
