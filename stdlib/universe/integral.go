package universe

import (
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
)

const IntegralKind = "integral"

type IntegralOpSpec struct {
	Unit        flux.Duration `json:"unit"`
	TimeColumn  string        `json:"timeColumn"`
	Interpolate string        `json:"interpolate"`
	execute.AggregateConfig
}

func init() {
	integralSignature := runtime.MustLookupBuiltinType("universe", "integral")

	runtime.RegisterPackageValue("universe", IntegralKind, flux.MustValue(flux.FunctionValue(IntegralKind, createIntegralOpSpec, integralSignature)))
	flux.RegisterOpSpec(IntegralKind, newIntegralOp)
	plan.RegisterProcedureSpec(IntegralKind, newIntegralProcedure, IntegralKind)
	execute.RegisterTransformation(IntegralKind, createIntegralTransformation)
}

func createIntegralOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(IntegralOpSpec)

	if unit, ok, err := args.GetDuration("unit"); err != nil {
		return nil, err
	} else if ok {
		spec.Unit = unit
	} else {
		// Default is 1s
		spec.Unit = flux.ConvertDuration(time.Second)
	}

	if timeValue, ok, err := args.GetString("timeColumn"); err != nil {
		return nil, err
	} else if ok {
		spec.TimeColumn = timeValue
	} else {
		spec.TimeColumn = execute.DefaultTimeColLabel
	}

	if interpolate, ok, err := args.GetString("interpolate"); err != nil {
		return nil, err
	} else if ok {
		spec.Interpolate = interpolate
	} else {
		spec.Interpolate = ""
	}

	if err := spec.AggregateConfig.ReadArgs(args); err != nil {
		return nil, err
	}
	return spec, nil
}

func newIntegralOp() flux.OperationSpec {
	return new(IntegralOpSpec)
}

func (s *IntegralOpSpec) Kind() flux.OperationKind {
	return IntegralKind
}

type IntegralProcedureSpec struct {
	Unit        flux.Duration `json:"unit"`
	TimeColumn  string        `json:"timeColumn"`
	Interpolate bool          `json:"interpolate"`
	execute.AggregateConfig
}

func newIntegralProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*IntegralOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	return &IntegralProcedureSpec{
		Unit:            spec.Unit,
		TimeColumn:      spec.TimeColumn,
		Interpolate:     spec.Interpolate == "linear",
		AggregateConfig: spec.AggregateConfig,
	}, nil
}

func (s *IntegralProcedureSpec) Kind() plan.ProcedureKind {
	return IntegralKind
}
func (s *IntegralProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(IntegralProcedureSpec)
	*ns = *s

	ns.AggregateConfig = s.AggregateConfig.Copy()

	return ns
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *IntegralProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

func createIntegralTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*IntegralProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewIntegralTransformation(d, cache, s)
	return t, d, nil
}

type integralTransformation struct {
	execute.ExecutionNode
	d     execute.Dataset
	cache execute.TableBuilderCache

	spec IntegralProcedureSpec
}

func NewIntegralTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *IntegralProcedureSpec) *integralTransformation {
	return &integralTransformation{
		d:     d,
		cache: cache,
		spec:  *spec,
	}
}

func (t *integralTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *integralTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	if !tbl.Key().HasCol("_start") {
		return errors.New(codes.Invalid, "integral function needs _start column to be part of group key")
	}
	if !tbl.Key().HasCol("_stop") {
		return errors.New(codes.Invalid, "integral function needs _stop column to be part of group key")
	}

	var start execute.Time
	var stop execute.Time

	for j, col := range tbl.Key().Cols() {
		if col.Label == "_start" {
			start = tbl.Key().ValueTime(j)
		}
		if col.Label == "_stop" {
			stop = tbl.Key().ValueTime(j)
		}
	}

	builder, created := t.cache.TableBuilder(tbl.Key())
	if !created {
		return errors.Newf(codes.FailedPrecondition, "integral found duplicate table with key: %v", tbl.Key())
	}

	if err := execute.AddTableKeyCols(tbl.Key(), builder); err != nil {
		return err
	}

	cols := tbl.Cols()
	integrals := make([]*integral, len(cols))
	colMap := make([]int, len(cols))

	for _, c := range t.spec.Columns {
		idx := execute.ColIdx(c, cols)
		if idx < 0 {
			return errors.Newf(codes.FailedPrecondition, "column %q does not exist", c)
		}

		if tbl.Key().HasCol(c) {
			return errors.Newf(codes.FailedPrecondition, "cannot aggregate columns that are part of the group key")
		}

		if typ := cols[idx].Type; typ != flux.TFloat &&
			typ != flux.TInt &&
			typ != flux.TUInt {
			return errors.Newf(codes.FailedPrecondition, "cannot perform integral over %v", typ)
		}

		integrals[idx] = newIntegral(values.Duration(t.spec.Unit).Duration(), start, stop, t.spec.Interpolate)
		newIdx, err := builder.AddCol(flux.ColMeta{
			Label: c,
			Type:  flux.TFloat,
		})
		if err != nil {
			return err
		}
		colMap[idx] = newIdx
	}

	timeIdx := execute.ColIdx(t.spec.TimeColumn, cols)
	if timeIdx < 0 {
		return errors.Newf(codes.FailedPrecondition, "no column %q exists", t.spec.TimeColumn)
	}
	if err := tbl.Do(func(cr flux.ColReader) error {
		if cr.Times(timeIdx).NullN() > 0 {
			return errors.Newf(codes.FailedPrecondition, "integral found null time in time column")
		}

		for j, in := range integrals {
			if in == nil {
				continue
			}

			var prevTime values.Time
			l := cr.Len()
			for i := 0; i < l; i++ {
				tm := execute.Time(cr.Times(timeIdx).Value(i))
				if prevTime > tm {
					return errors.Newf(codes.FailedPrecondition, "integral found out-of-order times in time column")
				} else if prevTime == tm && i != 0 {
					// skip repeated times as in IFQL https://github.com/influxdata/influxdb/blob/1.8/query/functions.go
					continue
				}
				prevTime = tm

				switch tbl.Cols()[j].Type {
				case flux.TInt:
					if vs := cr.Ints(j); vs.IsValid(i) {
						in.updateFloat(tm, float64(vs.Value(i)))
					}
				case flux.TUInt:
					if vs := cr.UInts(j); vs.IsValid(i) {
						in.updateFloat(tm, float64(vs.Value(i)))
					}
				case flux.TFloat:
					if vs := cr.Floats(j); vs.IsValid(i) {
						in.updateFloat(tm, vs.Value(i))
					}
				}
			}
		}
		return nil
	}); err != nil {
		return err
	}

	if err := execute.AppendKeyValues(tbl.Key(), builder); err != nil {
		return err
	}
	for j, in := range integrals {
		if in == nil {
			continue
		}
		switch in.points {
		case 0:
			if err := builder.AppendFloat(colMap[j], 0.0); err != nil {
				return err
			}
		case 1:
			v := in.vs[0] * float64(in.bounds[1]-in.bounds[0])
			if !in.interpolate {
				v = 0
			}
			if err := builder.AppendFloat(colMap[j], v); err != nil {
				return err
			}
		default:
			in.interpolateStop()
			if err := builder.AppendFloat(colMap[j], in.value()); err != nil {
				return err
			}
		}
	}

	return nil
}

func (t *integralTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *integralTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *integralTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}

func newIntegral(unit time.Duration, start, stop execute.Time, interpolate bool) *integral {
	return &integral{
		interpolate: interpolate,
		bounds:      [2]execute.Time{start, stop},
		unit:        float64(unit),
	}
}

type integral struct {
	interpolate bool

	ts [2]execute.Time
	vs [2]float64

	bounds [2]execute.Time
	points uint8

	unit float64
	sum  float64
}

func (in *integral) value() float64 {
	return in.sum
}

func (in *integral) updateFloat(t execute.Time, v float64) {
	switch in.points {
	case 0:
		in.ts[0], in.vs[0] = t, v
		in.points++
	case 1:
		in.sum += 0.5 * (v + in.vs[0]) * float64(t-in.ts[0]) / in.unit
		in.ts[1], in.vs[1] = t, v
		in.points++
		in.interpolateStart()
	default:
		in.sum += 0.5 * (v + in.vs[1]) * float64(t-in.ts[1]) / in.unit
		in.ts[0], in.ts[1] = in.ts[1], t
		in.vs[0], in.vs[1] = in.vs[1], v
	}
}
func (in *integral) interpolateStart() {
	if in.interpolate && in.bounds[0] < in.ts[0] {
		m := (in.vs[1] - in.vs[0]) / float64(in.ts[1]-in.ts[0])
		y := in.vs[0] - m*float64(in.ts[0]-in.bounds[0])
		in.sum += 0.5 * (y + in.vs[0]) * float64(in.ts[0]-in.bounds[0]) / in.unit
	}
}
func (in *integral) interpolateStop() {
	if in.interpolate {
		m := (in.vs[1] - in.vs[0]) / float64(in.ts[1]-in.ts[0])
		y := in.vs[1] + m*float64(in.bounds[1]-in.ts[1])
		in.sum += 0.5 * (y + in.vs[1]) * float64(in.bounds[1]-in.ts[1]) / in.unit
	}
}
