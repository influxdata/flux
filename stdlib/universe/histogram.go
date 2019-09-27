package universe

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"sort"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const HistogramKind = "histogram"

type HistogramOpSpec struct {
	Column           string    `json:"column"`
	UpperBoundColumn string    `json:"upperBoundColumn"`
	CountColumn      string    `json:"countColumn"`
	Bins             []float64 `json:"bins"`
	Normalize        bool      `json:"normalize"`
}

func init() {
	histogramSignature := execute.AggregateSignature(
		map[string]semantic.PolyType{
			"column":           semantic.String,
			"upperBoundColumn": semantic.String,
			"countColumn":      semantic.String,
			"bins":             semantic.NewArrayPolyType(semantic.Float),
			"normalize":        semantic.Bool,
		},
		[]string{"bins"},
	)

	flux.RegisterPackageValue("universe", HistogramKind, flux.FunctionValue(HistogramKind, createHistogramOpSpec, histogramSignature))
	flux.RegisterPackageValue("universe", "linearBins", linearBins{})
	flux.RegisterPackageValue("universe", "logarithmicBins", logarithmicBins{})
	flux.RegisterOpSpec(HistogramKind, newHistogramOp)
	plan.RegisterProcedureSpec(HistogramKind, newHistogramProcedure, HistogramKind)
	execute.RegisterTransformation(HistogramKind, createHistogramTransformation)
}

func createHistogramOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(HistogramOpSpec)

	if col, ok, err := args.GetString("column"); err != nil {
		return nil, err
	} else if ok {
		spec.Column = col
	} else {
		spec.Column = execute.DefaultValueColLabel
	}
	if col, ok, err := args.GetString("upperBoundColumn"); err != nil {
		return nil, err
	} else if ok {
		spec.UpperBoundColumn = col
	} else {
		spec.UpperBoundColumn = DefaultUpperBoundColumnLabel
	}
	if col, ok, err := args.GetString("countColumn"); err != nil {
		return nil, err
	} else if ok {
		spec.CountColumn = col
	} else {
		spec.CountColumn = execute.DefaultValueColLabel
	}
	binsArry, err := args.GetRequiredArray("bins", semantic.Float)
	if err != nil {
		return nil, err
	}
	spec.Bins, err = interpreter.ToFloatArray(binsArry)
	if err != nil {
		return nil, err
	}
	if normalize, ok, err := args.GetBool("normalize"); err != nil {
		return nil, err
	} else if ok {
		spec.Normalize = normalize
	}

	return spec, nil
}

func newHistogramOp() flux.OperationSpec {
	return new(HistogramOpSpec)
}

func (s *HistogramOpSpec) Kind() flux.OperationKind {
	return HistogramKind
}

type HistogramProcedureSpec struct {
	plan.DefaultCost
	HistogramOpSpec
}

func newHistogramProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*HistogramOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	return &HistogramProcedureSpec{
		HistogramOpSpec: *spec,
	}, nil
}

func (s *HistogramProcedureSpec) Kind() plan.ProcedureKind {
	return HistogramKind
}
func (s *HistogramProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(HistogramProcedureSpec)
	*ns = *s
	if len(s.Bins) > 0 {
		ns.Bins = make([]float64, len(s.Bins))
		copy(ns.Bins, s.Bins)
	}
	return ns
}

func createHistogramTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*HistogramProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewHistogramTransformation(d, cache, s)
	return t, d, nil
}

type histogramTransformation struct {
	d     execute.Dataset
	cache execute.TableBuilderCache

	spec HistogramProcedureSpec
}

func NewHistogramTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *HistogramProcedureSpec) *histogramTransformation {
	sort.Float64s(spec.Bins)
	return &histogramTransformation{
		d:     d,
		cache: cache,
		spec:  *spec,
	}
}

func (t *histogramTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *histogramTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	builder, created := t.cache.TableBuilder(tbl.Key())
	if !created {
		return errors.Newf(codes.FailedPrecondition, "histogram found duplicate table with key: %v", tbl.Key())
	}
	valueIdx := execute.ColIdx(t.spec.Column, tbl.Cols())
	if valueIdx < 0 {
		return errors.Newf(codes.FailedPrecondition, "column %q is missing", t.spec.Column)
	}
	if col := tbl.Cols()[valueIdx]; col.Type != flux.TFloat {
		return errors.Newf(codes.FailedPrecondition, "column %q must be a float got %v", t.spec.Column, col.Type)
	}

	err := execute.AddTableKeyCols(tbl.Key(), builder)
	if err != nil {
		return err
	}
	boundIdx, err := builder.AddCol(flux.ColMeta{
		Label: t.spec.UpperBoundColumn,
		Type:  flux.TFloat,
	})
	if err != nil {
		return err
	}
	countIdx, err := builder.AddCol(flux.ColMeta{
		Label: t.spec.CountColumn,
		Type:  flux.TFloat,
	})
	if err != nil {
		return err
	}
	totalRows := 0.0
	counts := make([]float64, len(t.spec.Bins))
	err = tbl.Do(func(cr flux.ColReader) error {
		vs := cr.Floats(valueIdx)
		totalRows += float64(vs.Len() - vs.NullN())
		for i := 0; i < vs.Len(); i++ {
			if vs.IsNull(i) {
				continue
			}

			v := vs.Value(i)
			idx := sort.Search(len(t.spec.Bins), func(i int) bool {
				return v <= t.spec.Bins[i]
			})
			if idx >= len(t.spec.Bins) {
				// Greater than highest bin, or not found
				return fmt.Errorf("found value greater than any bin, %d %d %f %f", idx, len(t.spec.Bins), v, t.spec.Bins[len(t.spec.Bins)-1])
			}
			// Increment counter
			counts[idx]++
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Add records making counts cumulative
	total := 0.0
	for i, v := range counts {
		if err := execute.AppendKeyValues(tbl.Key(), builder); err != nil {
			return err
		}
		count := v + total
		if t.spec.Normalize {
			count /= totalRows
		}
		if err := builder.AppendFloat(countIdx, count); err != nil {
			return err
		}
		if err := builder.AppendFloat(boundIdx, t.spec.Bins[i]); err != nil {
			return err
		}
		total += v
	}
	return nil
}

func (t *histogramTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *histogramTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *histogramTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}

// linearBins is a helper function for creating bins spaced linearly
type linearBins struct{}

var linearBinsType = semantic.NewFunctionType(semantic.FunctionSignature{
	Parameters: map[string]semantic.Type{
		"start":    semantic.Float,
		"width":    semantic.Float,
		"count":    semantic.Int,
		"infinity": semantic.Bool,
	},
	Required: semantic.LabelSet{"start", "width", "count"},
	Return:   semantic.NewArrayType(semantic.Float),
})
var linearBinsPolyType = linearBinsType.PolyType()

func (b linearBins) Type() semantic.Type {
	return linearBinsType
}
func (b linearBins) PolyType() semantic.PolyType {
	return linearBinsPolyType
}

func (b linearBins) IsNull() bool {
	return false
}
func (b linearBins) Str() string {
	panic(values.UnexpectedKind(semantic.String, semantic.Function))
}

func (b linearBins) Bytes() []byte {
	panic(values.UnexpectedKind(semantic.Function, semantic.Bytes))
}

func (b linearBins) Int() int64 {
	panic(values.UnexpectedKind(semantic.Int, semantic.Function))
}

func (b linearBins) UInt() uint64 {
	panic(values.UnexpectedKind(semantic.UInt, semantic.Function))
}

func (b linearBins) Float() float64 {
	panic(values.UnexpectedKind(semantic.Float, semantic.Function))
}

func (b linearBins) Bool() bool {
	panic(values.UnexpectedKind(semantic.Bool, semantic.Function))
}

func (b linearBins) Time() values.Time {
	panic(values.UnexpectedKind(semantic.Time, semantic.Function))
}

func (b linearBins) Duration() values.Duration {
	panic(values.UnexpectedKind(semantic.Duration, semantic.Function))
}

func (b linearBins) Regexp() *regexp.Regexp {
	panic(values.UnexpectedKind(semantic.Regexp, semantic.Function))
}

func (b linearBins) Array() values.Array {
	panic(values.UnexpectedKind(semantic.Array, semantic.Function))
}

func (b linearBins) Object() values.Object {
	panic(values.UnexpectedKind(semantic.Object, semantic.Function))
}

func (b linearBins) Function() values.Function {
	return b
}

func (b linearBins) Equal(rhs values.Value) bool {
	if b.Type() != rhs.Type() {
		return false
	}
	_, ok := rhs.(linearBins)
	return ok
}

func (b linearBins) HasSideEffect() bool {
	return false
}

func (b linearBins) Call(ctx context.Context, args values.Object) (values.Value, error) {
	startV, ok := args.Get("start")
	if !ok {
		return nil, errors.New(codes.Invalid, "start is required")
	}
	if startV.Type() != semantic.Float {
		return nil, errors.New(codes.Invalid, "start must be a float")
	}
	widthV, ok := args.Get("width")
	if !ok {
		return nil, errors.New(codes.Invalid, "width is required")
	}
	if widthV.Type() != semantic.Float {
		return nil, errors.New(codes.Invalid, "width must be a float")
	}
	countV, ok := args.Get("count")
	if !ok {
		return nil, errors.New(codes.Invalid, "count is required")
	}
	if countV.Type() != semantic.Int {
		return nil, errors.New(codes.Invalid, "count must be an int")
	}
	infV, ok := args.Get("infinity")
	if !ok {
		infV = values.NewBool(true)
	}
	if infV.Type() != semantic.Bool {
		return nil, errors.New(codes.Invalid, "infinity must be a bool")
	}
	start := startV.Float()
	width := widthV.Float()
	count := countV.Int()
	inf := infV.Bool()
	l := int(count)
	if inf {
		l++
	}
	elements := make([]values.Value, l)
	bound := start
	for i := 0; i < l; i++ {
		elements[i] = values.NewFloat(bound)
		bound += width
	}
	if inf {
		elements[l-1] = values.NewFloat(math.Inf(1))
	}
	counts := values.NewArrayWithBacking(semantic.Float, elements)
	return counts, nil
}

// logarithmicBins is a helper function for creating bins spaced by an logarithmic factor.
type logarithmicBins struct{}

var logarithmicBinsType = semantic.NewFunctionType(semantic.FunctionSignature{
	Parameters: map[string]semantic.Type{
		"start":    semantic.Float,
		"factor":   semantic.Float,
		"count":    semantic.Int,
		"infinity": semantic.Bool,
	},
	Required: semantic.LabelSet{"start", "factor", "count"},
	Return:   semantic.NewArrayType(semantic.Float),
})
var logarithmicBinsPolyType = logarithmicBinsType.PolyType()

func (b logarithmicBins) Type() semantic.Type {
	return logarithmicBinsType
}
func (b logarithmicBins) PolyType() semantic.PolyType {
	return logarithmicBinsPolyType
}

func (b logarithmicBins) IsNull() bool {
	return false
}
func (b logarithmicBins) Str() string {
	panic(values.UnexpectedKind(semantic.String, semantic.Function))
}

func (b logarithmicBins) Bytes() []byte {
	panic(values.UnexpectedKind(semantic.Function, semantic.Bytes))
}

func (b logarithmicBins) Int() int64 {
	panic(values.UnexpectedKind(semantic.Int, semantic.Function))
}

func (b logarithmicBins) UInt() uint64 {
	panic(values.UnexpectedKind(semantic.UInt, semantic.Function))
}

func (b logarithmicBins) Float() float64 {
	panic(values.UnexpectedKind(semantic.Float, semantic.Function))
}

func (b logarithmicBins) Bool() bool {
	panic(values.UnexpectedKind(semantic.Bool, semantic.Function))
}

func (b logarithmicBins) Time() values.Time {
	panic(values.UnexpectedKind(semantic.Time, semantic.Function))
}

func (b logarithmicBins) Duration() values.Duration {
	panic(values.UnexpectedKind(semantic.Duration, semantic.Function))
}

func (b logarithmicBins) Regexp() *regexp.Regexp {
	panic(values.UnexpectedKind(semantic.Regexp, semantic.Function))
}

func (b logarithmicBins) Array() values.Array {
	panic(values.UnexpectedKind(semantic.Array, semantic.Function))
}

func (b logarithmicBins) Object() values.Object {
	panic(values.UnexpectedKind(semantic.Object, semantic.Function))
}

func (b logarithmicBins) Function() values.Function {
	return b
}

func (b logarithmicBins) Equal(rhs values.Value) bool {
	if b.Type() != rhs.Type() {
		return false
	}
	_, ok := rhs.(logarithmicBins)
	return ok
}

func (b logarithmicBins) HasSideEffect() bool {
	return false
}

func (b logarithmicBins) Call(ctx context.Context, args values.Object) (values.Value, error) {
	startV, ok := args.Get("start")
	if !ok {
		return nil, errors.New(codes.Invalid, "start is required")
	}
	if startV.Type() != semantic.Float {
		return nil, errors.New(codes.Invalid, "start must be a float")
	}
	factorV, ok := args.Get("factor")
	if !ok {
		return nil, errors.New(codes.Invalid, "factor is required")
	}
	if factorV.Type() != semantic.Float {
		return nil, errors.New(codes.Invalid, "factor must be a float")
	}
	countV, ok := args.Get("count")
	if !ok {
		return nil, errors.New(codes.Invalid, "count is required")
	}
	if countV.Type() != semantic.Int {
		return nil, errors.New(codes.Invalid, "count must be an int")
	}
	infV, ok := args.Get("infinity")
	if !ok {
		infV = values.NewBool(true)
	}
	if infV.Type() != semantic.Bool {
		return nil, errors.New(codes.Invalid, "infinity must be a bool")
	}
	start := startV.Float()
	factor := factorV.Float()
	count := countV.Int()
	inf := infV.Bool()
	l := int(count)
	if inf {
		l++
	}
	elements := make([]values.Value, l)
	bound := start
	for i := 0; i < l; i++ {
		elements[i] = values.NewFloat(bound)
		bound *= factor
	}
	if inf {
		elements[l-1] = values.NewFloat(math.Inf(1))
	}
	counts := values.NewArrayWithBacking(semantic.Float, elements)
	return counts, nil
}
