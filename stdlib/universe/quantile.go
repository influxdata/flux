package universe

import (
	"fmt"
	"math"
	"sort"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
	"github.com/influxdata/tdigest"
	"github.com/pkg/errors"
)

const QuantileKind = "quantile"
const ExactQuantileAggKind = "exact-quantile-aggregate"
const ExactQuantileSelectKind = "exact-quantile-selector"

const (
	methodEstimateTdigest = "estimate_tdigest"
	methodExactMean       = "exact_mean"
	methodExactSelector   = "exact_selector"
)

type QuantileOpSpec struct {
	Quantile    float64 `json:"quantile"`
	Compression float64 `json:"compression"`
	Method      string  `json:"method"`
	// quantile is either an aggregate, or a selector based on the options
	execute.AggregateConfig
	execute.SelectorConfig
}

func init() {
	quantileSignature := flux.FunctionSignature(
		map[string]semantic.PolyType{
			"column":      semantic.String,                            // selector
			"columns":     semantic.NewArrayPolyType(semantic.String), // aggregate
			"q":           semantic.Float,
			"compression": semantic.Float,
			"method":      semantic.String,
		},
		[]string{"q"},
	)

	flux.RegisterPackageValue("universe", QuantileKind, flux.FunctionValue(QuantileKind, createQuantileOpSpec, quantileSignature))

	flux.RegisterOpSpec(QuantileKind, newQuantileOp)
	plan.RegisterProcedureSpec(QuantileKind, newQuantileProcedure, QuantileKind)
	execute.RegisterTransformation(QuantileKind, createQuantileTransformation)
	execute.RegisterTransformation(ExactQuantileAggKind, createExactQuantileAggTransformation)
	execute.RegisterTransformation(ExactQuantileSelectKind, createExactQuantileSelectTransformation)
}

func createQuantileOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(QuantileOpSpec)
	p, err := args.GetRequiredFloat("q")
	if err != nil {
		return nil, err
	}
	spec.Quantile = p

	if spec.Quantile < 0 || spec.Quantile > 1 {
		return nil, errors.New("quantile must be between 0 and 1.")
	}

	if m, ok, err := args.GetString("method"); err != nil {
		return nil, err
	} else if ok {
		spec.Method = m
	}

	if c, ok, err := args.GetFloat("compression"); err != nil {
		return nil, err
	} else if ok {
		spec.Compression = c
	}

	if spec.Compression > 0 && spec.Method != methodEstimateTdigest {
		return nil, errors.New("compression parameter is only valid for method estimate_tdigest.")
	}

	// Set default Compression if not exact
	if spec.Method == methodEstimateTdigest && spec.Compression == 0 {
		spec.Compression = 1000
	}

	switch spec.Method {
	case methodExactSelector:
		if err := spec.SelectorConfig.ReadArgs(args); err != nil {
			return nil, err
		}
	case methodEstimateTdigest, methodExactMean:
		if err := spec.AggregateConfig.ReadArgs(args); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unknown method %s", spec.Method)
	}

	return spec, nil
}

func newQuantileOp() flux.OperationSpec {
	return new(QuantileOpSpec)
}

func (s *QuantileOpSpec) Kind() flux.OperationKind {
	return QuantileKind
}

type TDigestQuantileProcedureSpec struct {
	Quantile    float64 `json:"quantile"`
	Compression float64 `json:"compression"`
	execute.AggregateConfig
}

func (s *TDigestQuantileProcedureSpec) Kind() plan.ProcedureKind {
	return QuantileKind
}
func (s *TDigestQuantileProcedureSpec) Copy() plan.ProcedureSpec {
	return &TDigestQuantileProcedureSpec{
		Quantile:        s.Quantile,
		Compression:     s.Compression,
		AggregateConfig: s.AggregateConfig,
	}
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *TDigestQuantileProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

type ExactQuantileAggProcedureSpec struct {
	Quantile float64 `json:"quantile"`
	execute.AggregateConfig
}

func (s *ExactQuantileAggProcedureSpec) Kind() plan.ProcedureKind {
	return ExactQuantileAggKind
}
func (s *ExactQuantileAggProcedureSpec) Copy() plan.ProcedureSpec {
	return &ExactQuantileAggProcedureSpec{Quantile: s.Quantile, AggregateConfig: s.AggregateConfig}
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *ExactQuantileAggProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

type ExactQuantileSelectProcedureSpec struct {
	Quantile float64 `json:"quantile"`
	execute.SelectorConfig
}

func (s *ExactQuantileSelectProcedureSpec) Kind() plan.ProcedureKind {
	return ExactQuantileSelectKind
}
func (s *ExactQuantileSelectProcedureSpec) Copy() plan.ProcedureSpec {
	return &ExactQuantileSelectProcedureSpec{Quantile: s.Quantile}
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *ExactQuantileSelectProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

func newQuantileProcedure(qs flux.OperationSpec, a plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*QuantileOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	switch spec.Method {
	case methodExactMean:
		return &ExactQuantileAggProcedureSpec{
			Quantile:        spec.Quantile,
			AggregateConfig: spec.AggregateConfig,
		}, nil
	case methodExactSelector:
		return &ExactQuantileSelectProcedureSpec{
			Quantile: spec.Quantile,
		}, nil
	case methodEstimateTdigest:
		fallthrough
	default:
		// default to estimated quantile
		return &TDigestQuantileProcedureSpec{
			Quantile:        spec.Quantile,
			Compression:     spec.Compression,
			AggregateConfig: spec.AggregateConfig,
		}, nil
	}
}

type QuantileAgg struct {
	Quantile,
	Compression float64

	digest *tdigest.TDigest
	ok     bool
}

func createQuantileTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	ps, ok := spec.(*TDigestQuantileProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", ps)
	}
	agg := &QuantileAgg{
		Quantile:    ps.Quantile,
		Compression: ps.Compression,
	}
	t, d := execute.NewAggregateTransformationAndDataset(id, mode, agg, ps.AggregateConfig, a.Allocator())
	return t, d, nil
}
func (a *QuantileAgg) Copy() *QuantileAgg {
	na := new(QuantileAgg)
	*na = *a
	na.digest = tdigest.NewWithCompression(na.Compression)
	return na
}

func (a *QuantileAgg) NewBoolAgg() execute.DoBoolAgg {
	return nil
}

func (a *QuantileAgg) NewIntAgg() execute.DoIntAgg {
	return nil
}

func (a *QuantileAgg) NewUIntAgg() execute.DoUIntAgg {
	return nil
}

func (a *QuantileAgg) NewFloatAgg() execute.DoFloatAgg {
	return a.Copy()
}

func (a *QuantileAgg) NewStringAgg() execute.DoStringAgg {
	return nil
}

func (a *QuantileAgg) DoFloat(vs *array.Float64) {
	for i := 0; i < vs.Len(); i++ {
		if vs.IsValid(i) {
			a.digest.Add(vs.Value(i), 1)
			a.ok = true
		}
	}
}

func (a *QuantileAgg) Type() flux.ColType {
	return flux.TFloat
}

func (a *QuantileAgg) ValueFloat() float64 {
	return a.digest.Quantile(a.Quantile)
}

func (a *QuantileAgg) IsNull() bool {
	return !a.ok
}

type ExactQuantileAgg struct {
	Quantile float64
	data     []float64
}

func createExactQuantileAggTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	ps, ok := spec.(*ExactQuantileAggProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", ps)
	}
	agg := &ExactQuantileAgg{
		Quantile: ps.Quantile,
	}
	t, d := execute.NewAggregateTransformationAndDataset(id, mode, agg, ps.AggregateConfig, a.Allocator())
	return t, d, nil
}

func (a *ExactQuantileAgg) Copy() *ExactQuantileAgg {
	na := new(ExactQuantileAgg)
	*na = *a
	na.data = nil
	return na
}
func (a *ExactQuantileAgg) NewBoolAgg() execute.DoBoolAgg {
	return nil
}

func (a *ExactQuantileAgg) NewIntAgg() execute.DoIntAgg {
	return nil
}

func (a *ExactQuantileAgg) NewUIntAgg() execute.DoUIntAgg {
	return nil
}

func (a *ExactQuantileAgg) NewFloatAgg() execute.DoFloatAgg {
	return a.Copy()
}

func (a *ExactQuantileAgg) NewStringAgg() execute.DoStringAgg {
	return nil
}

func (a *ExactQuantileAgg) DoFloat(vs *array.Float64) {
	if vs.NullN() == 0 {
		a.data = append(a.data, vs.Float64Values()...)
		return
	}

	// Check if we have enough space for the floats
	// inside of the array.
	l := vs.Len() - vs.NullN()
	if len(a.data)+l > cap(a.data) {
		// We do not. Create an array with the needed size and
		// copy over the existing data.
		data := make([]float64, len(a.data), len(a.data)+l)
		copy(data, a.data)
		a.data = data
	}

	for i := 0; i < vs.Len(); i++ {
		if vs.IsValid(i) {
			a.data = append(a.data, vs.Value(i))
		}
	}
}

func (a *ExactQuantileAgg) Type() flux.ColType {
	return flux.TFloat
}

func (a *ExactQuantileAgg) ValueFloat() float64 {
	sort.Float64s(a.data)

	x := a.Quantile * float64(len(a.data)-1)
	x0 := math.Floor(x)
	x1 := math.Ceil(x)

	if x0 == x1 {
		return a.data[int(x0)]
	}

	// Linear interpolate
	y0 := a.data[int(x0)]
	y1 := a.data[int(x1)]
	y := y0*(x1-x) + y1*(x-x0)

	return y
}

func (a *ExactQuantileAgg) IsNull() bool {
	return len(a.data) == 0
}

func createExactQuantileSelectTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	ps, ok := spec.(*ExactQuantileSelectProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", ps)
	}

	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewExactQuantileSelectorTransformation(d, cache, ps, a.Allocator())

	return t, d, nil
}

type ExactQuantileSelectorTransformation struct {
	d     execute.Dataset
	cache execute.TableBuilderCache
	spec  ExactQuantileSelectProcedureSpec
	a     *memory.Allocator
}

func NewExactQuantileSelectorTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *ExactQuantileSelectProcedureSpec, a *memory.Allocator) *ExactQuantileSelectorTransformation {
	if spec.SelectorConfig.Column == "" {
		spec.SelectorConfig.Column = execute.DefaultValueColLabel
	}

	sel := &ExactQuantileSelectorTransformation{
		d:     d,
		cache: cache,
		spec:  *spec,
		a:     a,
	}
	return sel
}

func (t *ExactQuantileSelectorTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	valueIdx := execute.ColIdx(t.spec.Column, tbl.Cols())
	if valueIdx < 0 {
		return fmt.Errorf("no column %q exists", t.spec.Column)
	}

	var row execute.Row
	switch typ := tbl.Cols()[valueIdx].Type; typ {
	case flux.TFloat:
		type floatValue struct {
			value float64
			row   execute.Row
		}

		var rows []floatValue
		if err := tbl.Do(func(cr flux.ColReader) error {
			vs := cr.Floats(valueIdx)
			for i := 0; i < vs.Len(); i++ {
				if vs.IsValid(i) {
					rows = append(rows, floatValue{
						value: vs.Value(i),
						row:   execute.ReadRow(i, cr),
					})
				}
			}
			return nil
		}); err != nil {
			return err
		}

		if len(rows) > 0 {
			sort.SliceStable(rows, func(i, j int) bool {
				return rows[i].value < rows[j].value
			})
			index := getQuantileIndex(t.spec.Quantile, len(rows))
			row = rows[index].row
		}
	case flux.TInt:
		type intValue struct {
			value int64
			row   execute.Row
		}

		var rows []intValue
		if err := tbl.Do(func(cr flux.ColReader) error {
			vs := cr.Ints(valueIdx)
			for i := 0; i < vs.Len(); i++ {
				if vs.IsValid(i) {
					rows = append(rows, intValue{
						value: vs.Value(i),
						row:   execute.ReadRow(i, cr),
					})
				}
			}
			return nil
		}); err != nil {
			return err
		}

		if len(rows) > 0 {
			sort.SliceStable(rows, func(i, j int) bool {
				return rows[i].value < rows[j].value
			})
			index := getQuantileIndex(t.spec.Quantile, len(rows))
			row = rows[index].row
		}
	case flux.TUInt:
		type uintValue struct {
			value uint64
			row   execute.Row
		}

		var rows []uintValue
		if err := tbl.Do(func(cr flux.ColReader) error {
			vs := cr.UInts(valueIdx)
			for i := 0; i < vs.Len(); i++ {
				if vs.IsValid(i) {
					rows = append(rows, uintValue{
						value: vs.Value(i),
						row:   execute.ReadRow(i, cr),
					})
				}
			}
			return nil
		}); err != nil {
			return err
		}

		if len(rows) > 0 {
			sort.SliceStable(rows, func(i, j int) bool {
				return rows[i].value < rows[j].value
			})
			index := getQuantileIndex(t.spec.Quantile, len(rows))
			row = rows[index].row
		}
	case flux.TString:
		type stringValue struct {
			value string
			row   execute.Row
		}

		var rows []stringValue
		if err := tbl.Do(func(cr flux.ColReader) error {
			vs := cr.Strings(valueIdx)
			for i := 0; i < vs.Len(); i++ {
				if vs.IsValid(i) {
					rows = append(rows, stringValue{
						value: vs.ValueString(i),
						row:   execute.ReadRow(i, cr),
					})
				}
			}
			return nil
		}); err != nil {
			return err
		}

		if len(rows) > 0 {
			sort.SliceStable(rows, func(i, j int) bool {
				return rows[i].value < rows[j].value
			})
			index := getQuantileIndex(t.spec.Quantile, len(rows))
			row = rows[index].row
		}
	case flux.TTime:
		type timeValue struct {
			value values.Time
			row   execute.Row
		}

		var rows []timeValue
		if err := tbl.Do(func(cr flux.ColReader) error {
			vs := cr.Times(valueIdx)
			for i := 0; i < vs.Len(); i++ {
				if vs.IsValid(i) {
					rows = append(rows, timeValue{
						value: values.Time(vs.Value(i)),
						row:   execute.ReadRow(i, cr),
					})
				}
			}
			return nil
		}); err != nil {
			return err
		}

		if len(rows) > 0 {
			sort.SliceStable(rows, func(i, j int) bool {
				return rows[i].value < rows[j].value
			})
			index := getQuantileIndex(t.spec.Quantile, len(rows))
			row = rows[index].row
		}
	case flux.TBool:
		type boolValue struct {
			value bool
			row   execute.Row
		}

		var rows []boolValue
		if err := tbl.Do(func(cr flux.ColReader) error {
			vs := cr.Bools(valueIdx)
			for i := 0; i < vs.Len(); i++ {
				if vs.IsValid(i) {
					rows = append(rows, boolValue{
						value: vs.Value(i),
						row:   execute.ReadRow(i, cr),
					})
				}
			}
			return nil
		}); err != nil {
			return err
		}

		if len(rows) > 0 {
			sort.SliceStable(rows, func(i, j int) bool {
				if rows[i].value == rows[j].value {
					return false
				}
				return rows[j].value
			})
			index := getQuantileIndex(t.spec.Quantile, len(rows))
			row = rows[index].row
		}
	default:
		execute.PanicUnknownType(typ)
	}

	builder, created := t.cache.TableBuilder(tbl.Key())
	if !created {
		return fmt.Errorf("found duplicate table with key: %v", tbl.Key())
	}
	if err := execute.AddTableCols(tbl, builder); err != nil {
		return err
	}

	for j, col := range builder.Cols() {
		if row.Values == nil {
			if idx := execute.ColIdx(col.Label, tbl.Key().Cols()); idx != -1 {
				v := tbl.Key().Value(idx)
				if err := builder.AppendValue(j, v); err != nil {
					return err
				}
			} else {
				if err := builder.AppendNil(j); err != nil {
					return err
				}
			}
			continue
		}

		v := values.New(row.Values[j])
		if err := builder.AppendValue(j, v); err != nil {
			return err
		}
	}

	return nil
}

func getQuantileIndex(quantile float64, len int) int {
	x := quantile * float64(len)
	index := int(math.Ceil(x))
	if index > 0 {
		index--
	}
	return index
}

func (t *ExactQuantileSelectorTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *ExactQuantileSelectorTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}

func (t *ExactQuantileSelectorTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}

func (t *ExactQuantileSelectorTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
