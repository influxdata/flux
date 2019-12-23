package universe

import (
	"math"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
)

const CovarianceKind = "covariance"

type CovarianceOpSpec struct {
	PearsonCorrelation bool     `json:"pearsonr"`
	ValueDst           string   `json:"valueDst"`
	Columns            []string `json:"column"`
}

func init() {
	var covarianceSignature = flux.LookupBuiltInType("universe", "covariance")
	flux.RegisterPackageValue("universe", CovarianceKind, flux.MustValue(flux.FunctionValue(CovarianceKind, createCovarianceOpSpec, covarianceSignature)))
	flux.RegisterOpSpec(CovarianceKind, newCovarianceOp)
	plan.RegisterProcedureSpec(CovarianceKind, newCovarianceProcedure, CovarianceKind)
	execute.RegisterTransformation(CovarianceKind, createCovarianceTransformation)
}

func createCovarianceOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(CovarianceOpSpec)
	pearsonr, ok, err := args.GetBool("pearsonr")
	if err != nil {
		return nil, err
	} else if ok {
		spec.PearsonCorrelation = pearsonr
	}

	label, ok, err := args.GetString("valueDst")
	if err != nil {
		return nil, err
	} else if ok {
		spec.ValueDst = label
	} else {
		spec.ValueDst = execute.DefaultValueColLabel
	}

	if cols, err := args.GetRequiredArray("columns", semantic.String); err != nil {
		return nil, err
	} else {
		columns, err := interpreter.ToStringArray(cols)
		if err != nil {
			return nil, err
		}
		spec.Columns = columns
	}

	if len(spec.Columns) != 2 {
		return nil, errors.New(codes.Invalid, "must provide exactly two columns")
	}
	return spec, nil
}

func newCovarianceOp() flux.OperationSpec {
	return new(CovarianceOpSpec)
}

func (s *CovarianceOpSpec) Kind() flux.OperationKind {
	return CovarianceKind
}

type CovarianceProcedureSpec struct {
	plan.DefaultCost
	PearsonCorrelation bool
	ValueLabel         string
	Columns            []string
}

func newCovarianceProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*CovarianceOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}
	cs := CovarianceProcedureSpec{
		PearsonCorrelation: spec.PearsonCorrelation,
		ValueLabel:         spec.ValueDst,
	}
	cs.Columns = make([]string, len(spec.Columns))
	copy(cs.Columns, spec.Columns)

	return &cs, nil
}

func (s *CovarianceProcedureSpec) Kind() plan.ProcedureKind {
	return CovarianceKind
}

func (s *CovarianceProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(CovarianceProcedureSpec)
	*ns = *s

	if s.Columns != nil {
		ns.Columns = make([]string, len(s.Columns))
		copy(ns.Columns, s.Columns)
	}

	return ns
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *CovarianceProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

type CovarianceTransformation struct {
	d     execute.Dataset
	cache execute.TableBuilderCache
	spec  CovarianceProcedureSpec

	n,
	xm1,
	ym1,
	xm2,
	ym2,
	xym2 float64
}

func createCovarianceTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*CovarianceProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewCovarianceTransformation(d, cache, s)
	return t, d, nil
}

func NewCovarianceTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *CovarianceProcedureSpec) *CovarianceTransformation {
	return &CovarianceTransformation{
		d:     d,
		cache: cache,
		spec:  *spec,
	}
}

func (t *CovarianceTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *CovarianceTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	cols := tbl.Cols()
	builder, created := t.cache.TableBuilder(tbl.Key())
	if !created {
		return errors.Newf(codes.FailedPrecondition, "covariance found duplicate table with key: %v", tbl.Key())
	}
	err := execute.AddTableKeyCols(tbl.Key(), builder)
	if err != nil {
		return err
	}
	valueIdx, err := builder.AddCol(flux.ColMeta{
		Label: t.spec.ValueLabel,
		Type:  flux.TFloat,
	})
	if err != nil {
		return err
	}
	xIdx := execute.ColIdx(t.spec.Columns[0], cols)
	if xIdx < 0 {
		return errors.Newf(codes.FailedPrecondition, "specified column does not exist in table: %v", t.spec.Columns[0])
	}
	yIdx := execute.ColIdx(t.spec.Columns[1], cols)
	if yIdx < 0 {
		return errors.Newf(codes.FailedPrecondition, "specified column does not exist in table: %v", t.spec.Columns[1])
	}

	if cols[xIdx].Type != cols[yIdx].Type {
		return errors.New(codes.FailedPrecondition, "cannot compute the covariance between different types")
	}

	t.reset()
	err = tbl.Do(func(cr flux.ColReader) error {
		switch typ := cols[xIdx].Type; typ {
		case flux.TFloat:
			t.DoFloat(cr.Floats(xIdx), cr.Floats(yIdx))
		default:
			return errors.Newf(codes.Invalid, "covariance does not support %v", typ)
		}
		return nil
	})
	if err != nil {
		return err
	}

	if err := execute.AppendKeyValues(tbl.Key(), builder); err != nil {
		return err
	}
	return builder.AppendFloat(valueIdx, t.value())
}

func (t *CovarianceTransformation) reset() {
	t.n = 0
	t.xm1 = 0
	t.ym1 = 0
	t.xm2 = 0
	t.ym2 = 0
	t.xym2 = 0
}
func (t *CovarianceTransformation) DoFloat(xs, ys *array.Float64) {
	var xdelta, ydelta, xdelta2, ydelta2 float64
	for i := 0; i < xs.Len(); i++ {
		if xs.IsNull(i) || ys.IsNull(i) {
			continue
		}
		x, y := xs.Value(i), ys.Value(i)

		t.n++

		// Update means
		xdelta = x - t.xm1
		ydelta = y - t.ym1
		t.xm1 += xdelta / t.n
		t.ym1 += ydelta / t.n

		// Update variance sums
		xdelta2 = x - t.xm1
		ydelta2 = y - t.ym1
		t.xm2 += xdelta * xdelta2
		t.ym2 += ydelta * ydelta2

		// Update covariance sum
		// Covariance is symetric so we do not need to compute the yxm2 value.
		t.xym2 += xdelta * ydelta2
	}
}
func (t *CovarianceTransformation) value() float64 {
	if t.n < 2 {
		return math.NaN()
	}
	if t.spec.PearsonCorrelation {
		return (t.xym2) / math.Sqrt(t.xm2*t.ym2)
	}
	return t.xym2 / (t.n - 1)
}

func (t *CovarianceTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}

func (t *CovarianceTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}

func (t *CovarianceTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
