package universe

import (
	"context"
	"strconv"

	arrowmem "github.com/apache/arrow/go/v7/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/execute/table"
	"github.com/influxdata/flux/internal/feature"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

//go:generate -command tmpl ../../gotool.sh github.com/benbjohnson/tmpl
//go:generate tmpl -data=@../../internal/types.tmpldata -o fill.gen.go fill.gen.go.tmpl

const FillKind = "fill"

type FillOpSpec struct {
	Column      string `json:"column"`
	Type        string `json:"type"`
	Value       string `json:"value"`
	UsePrevious bool   `json:"use_previous"`
}

func init() {
	fillSignature := runtime.MustLookupBuiltinType("universe", "fill")

	runtime.RegisterPackageValue("universe", FillKind, flux.MustValue(flux.FunctionValue(FillKind, CreateFillOpSpec, fillSignature)))
	flux.RegisterOpSpec(FillKind, newFillOp)
	plan.RegisterProcedureSpec(FillKind, newFillProcedure, FillKind)
	execute.RegisterTransformation(FillKind, createFillTransformation)
}

func CreateFillOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(FillOpSpec)

	if col, ok, err := args.GetString("column"); err != nil {
		return nil, err
	} else if ok {
		spec.Column = col
	} else {
		spec.Column = execute.DefaultValueColLabel
	}

	val, valOk := args.Get("value")
	if valOk {
		typ := val.Type()
		spec.Type = typ.Nature().String()
		switch typ.Nature() {
		case semantic.Bool:
			spec.Value = strconv.FormatBool(val.Bool())
		case semantic.Int:
			spec.Value = strconv.FormatInt(val.Int(), 10)
		case semantic.UInt:
			spec.Value = strconv.FormatUint(val.UInt(), 10)
		case semantic.Float:
			spec.Value = strconv.FormatFloat(val.Float(), 'f', -1, 64)
		case semantic.String:
			spec.Value = val.Str()
		case semantic.Time:
			spec.Value = val.Time().String()
		default:
			return nil, errors.New(codes.Invalid, "value type for fill must be a valid primitive type (bool, int, uint, float, string, time)")
		}

	}

	usePrevious, prevOk, err := args.GetBool("usePrevious")
	if err != nil {
		return nil, err
	}
	if prevOk == valOk {
		return nil, errors.New(codes.Invalid, "fill requires exactly one of value or usePrevious")
	}

	if prevOk {
		spec.UsePrevious = usePrevious
	}

	return spec, nil
}

func newFillOp() flux.OperationSpec {
	return new(FillOpSpec)
}

func (s *FillOpSpec) Kind() flux.OperationKind {
	return FillKind
}

type FillProcedureSpec struct {
	plan.DefaultCost
	Column      string
	Value       values.Value
	UsePrevious bool
}

func newFillProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*FillOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	pspec := &FillProcedureSpec{
		Column:      spec.Column,
		UsePrevious: spec.UsePrevious,
	}
	if !spec.UsePrevious {
		switch spec.Type {
		case "bool":
			v, err := strconv.ParseBool(spec.Value)
			if err != nil {
				return nil, err
			}
			pspec.Value = values.New(v)
		case "int":
			v, err := strconv.ParseInt(spec.Value, 10, 64)
			if err != nil {
				return nil, err
			}
			pspec.Value = values.New(v)
		case "uint":
			v, err := strconv.ParseUint(spec.Value, 10, 64)
			if err != nil {
				return nil, err
			}
			pspec.Value = values.New(v)
		case "float":
			v, err := strconv.ParseFloat(spec.Value, 64)
			if err != nil {
				return nil, err
			}
			pspec.Value = values.New(v)
		case "string":
			pspec.Value = values.New(spec.Value)
		case "time":
			v, err := values.ParseTime(spec.Value)
			if err != nil {
				return nil, err
			}
			pspec.Value = values.New(v)
		default:
			return nil, errors.New(codes.Internal, "unknown type in fill op-spec")
		}
	}

	return pspec, nil
}

func (s *FillProcedureSpec) Kind() plan.ProcedureKind {
	return FillKind
}
func (s *FillProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(FillProcedureSpec)

	*ns = *s

	return ns
}

func createFillTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*FillProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}

	if feature.NarrowTransformationFill().Enabled(a.Context()) {
		return NewNarrowFillTransformation(a.Context(), s, id, a.Allocator())
	}

	t, d := NewFillTransformation(a.Context(), s, id, a.Allocator())
	return t, d, nil
}

type fillTransformation struct {
	execute.ExecutionNode
	d     *execute.PassthroughDataset
	ctx   context.Context
	spec  *FillProcedureSpec
	alloc memory.Allocator
}

func NewFillTransformation(ctx context.Context, spec *FillProcedureSpec, id execute.DatasetID, alloc memory.Allocator) (execute.Transformation, execute.Dataset) {
	t := &fillTransformation{
		d:     execute.NewPassthroughDataset(id),
		ctx:   ctx,
		spec:  spec,
		alloc: alloc,
	}
	return t, t.d
}

func (t *fillTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *fillTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	colIdx := execute.ColIdx(t.spec.Column, tbl.Cols())
	if colIdx < 0 && t.spec.UsePrevious {
		// usePrevious was used on a column that doesn't exist. In this case, just
		// act as a passthrough. This functionality says "I was provided a non-existent
		// value, so the new value also doesn't exist.
		return t.d.Process(tbl)
	}
	key := tbl.Key()
	if idx := execute.ColIdx(t.spec.Column, key.Cols()); idx >= 0 {
		if key.IsNull(idx) {
			var err error
			gkb := execute.NewGroupKeyBuilder(key)
			gkb.SetKeyValue(t.spec.Column, t.spec.Value)
			key, err = gkb.Build()
			if err != nil {
				return err
			}
		} else {
			return t.d.Process(tbl)
		}
	}

	var fillValue interface{}
	if !t.spec.UsePrevious {
		if colIdx > -1 && tbl.Cols()[colIdx].Type != flux.ColumnType(t.spec.Value.Type()) {
			return errors.Newf(codes.FailedPrecondition, "fill column type mismatch: %s/%s", tbl.Cols()[colIdx].Type.String(), flux.ColumnType(t.spec.Value.Type()).String())
		}
		fillValue = values.Unwrap(t.spec.Value)
	}

	// In case of missing fill column, add it to the existing columns
	tableCols := tbl.Cols()
	if colIdx < 0 {
		newCols := make([]flux.ColMeta, len(tableCols), len(tableCols)+1)
		copy(newCols, tableCols)
		c := flux.ColMeta{
			Label: t.spec.Column,
			Type:  flux.ColumnType(t.spec.Value.Type()),
		}
		tableCols = append(newCols, c)
		colIdx = len(tableCols) - 1
	}

	table, err := table.StreamWithContext(t.ctx, key, tableCols, func(ctx context.Context, w *table.StreamWriter) error {
		return tbl.Do(func(cr flux.ColReader) error {
			return t.fillTable(w, cr, colIdx, &fillValue)
		})
	})
	if err != nil {
		return err
	}
	return t.d.Process(table)
}

func (t *fillTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *fillTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *fillTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}

type fillTransformationAdapter struct {
	fillTransformation fillTransformation
}

func NewNarrowFillTransformation(ctx context.Context, spec *FillProcedureSpec, id execute.DatasetID, alloc memory.Allocator) (execute.Transformation, execute.Dataset, error) {
	fillTransformation := fillTransformation{
		ctx:  ctx,
		spec: spec,
	}
	t := &fillTransformationAdapter{
		fillTransformation,
	}
	return execute.NewNarrowStateTransformation[*fillState](id, t, alloc)
}

type fillState struct {
	fillValue interface{}
}

func (t *fillTransformationAdapter) Process(chunk table.Chunk, state *fillState, d *execute.TransportDataset, mem arrowmem.Allocator) (*fillState, bool, error) {
	return t.fillTransformation.adaptedProcess(chunk, state, d, mem)
}

func (t *fillTransformation) adaptedProcess(chunk table.Chunk, state *fillState, d *execute.TransportDataset, mem arrowmem.Allocator) (*fillState, bool, error) {
	if state == nil {
		// fill value
		var fillValue interface{}
		if !t.spec.UsePrevious {
			fillValue = values.Unwrap(t.spec.Value)
		}

		// populate state
		state = &fillState{
			fillValue: fillValue,
		}
	}

	colIdx := execute.ColIdx(t.spec.Column, chunk.Cols())
	if !t.spec.UsePrevious {
		if colIdx > -1 && chunk.Cols()[colIdx].Type != flux.ColumnType(t.spec.Value.Type()) {
			return nil, false, errors.Newf(codes.FailedPrecondition, "fill column type mismatch: %s/%s", chunk.Cols()[colIdx].Type.String(), flux.ColumnType(t.spec.Value.Type()).String())
		}
	}

	if colIdx < 0 && t.spec.UsePrevious {
		// usePrevious was used on a column that doesn't exist. In this case, just
		// act as a passthrough. This functionality says "I was provided a non-existent
		// value, so the new value also doesn't exist.
		chunk.Retain()
		err := d.Process(chunk)
		return nil, false, err
	}

	// key
	key := chunk.Key()
	if idx := execute.ColIdx(t.spec.Column, key.Cols()); idx >= 0 {
		if key.IsNull(idx) {
			var err error
			gkb := execute.NewGroupKeyBuilder(key)
			gkb.SetKeyValue(t.spec.Column, t.spec.Value)
			key, err = gkb.Build()
			if err != nil {
				return nil, false, err
			}
		} else {
			chunk.Retain()
			err := d.Process(chunk)
			return nil, false, err
		}
	}

	// output columns
	var outputColumns []flux.ColMeta
	// In case of missing fill column, add it to the existing columns
	if colIdx < 0 {
		colsLen := len(chunk.Cols())
		newCols := make([]flux.ColMeta, colsLen, colsLen+1)
		copy(newCols, chunk.Cols())
		c := flux.ColMeta{
			Label: t.spec.Column,
			Type:  flux.ColumnType(t.spec.Value.Type()),
		}
		outputColumns = append(newCols, c)
		colIdx = len(outputColumns) - 1
	} else {
		outputColumns = chunk.Cols()
	}

	buffer := arrow.TableBuffer{
		GroupKey: key,
		Columns:  outputColumns,
		Values:   make([]array.Array, len(chunk.Cols())),
	}

	if err := t.fillChunk(&buffer, chunk, colIdx, &state.fillValue, mem); err != nil {
		return nil, false, err
	}

	out := table.ChunkFromBuffer(buffer)
	if err := d.Process(out); err != nil {
		return nil, false, err
	}
	return state, true, nil
}

func (t *fillTransformationAdapter) Close() error { return nil }

func (t *fillTransformation) fillChunk(buffer *arrow.TableBuffer, chunk table.Chunk, colIdx int, fillValue *interface{}, mem arrowmem.Allocator) error {
	l := chunk.Len()
	vs := make([]array.Array, len(buffer.Cols()))

	// Iterate over the existing columns and if column already exist(colIdx matches with i) call fillColumn on it
	for i, col := range chunk.Cols() {
		if l == 0 {
			vs[i] = arrow.Empty(col.Type)
		} else {
			arr := chunk.Values(i)
			if i != colIdx {
				vs[i] = arr
				vs[i].Retain()
				continue
			}
			vs[i] = t.fillColumn(col.Type, arr, fillValue, mem)
		}
	}

	// If the fill column is new, create a completely null column and call fillColumn on it
	if vs[colIdx] == nil {
		colType := flux.ColumnType(t.spec.Value.Type())
		arr := t.addNullColumn(colType, chunk.Len(), mem)
		defer arr.Release()
		vs[colIdx] = t.fillColumn(colType, arr, fillValue, mem)
	}
	buffer.Values = vs
	return nil
}

func (t *fillTransformation) fillTable(w *table.StreamWriter, cr flux.ColReader, colIdx int, fillValue *interface{}) error {
	crLen := cr.Len()
	if crLen == 0 {
		return nil
	}
	vs := make([]array.Array, len(w.Cols()))

	// Iterate over the existing columns and if column already exist(colIdx matches with i) call fillColumn on it
	for i, col := range cr.Cols() {
		arr := table.Values(cr, i)
		if i != colIdx {
			vs[i] = arr
			vs[i].Retain()
			continue
		}
		vs[i] = t.fillColumn(col.Type, arr, fillValue, t.alloc)
	}

	// If the fill column is new, create a completely null column and call fillColumn on it
	if vs[colIdx] == nil {
		colType := flux.ColumnType(t.spec.Value.Type())
		arr := t.addNullColumn(colType, crLen, t.alloc)
		defer arr.Release()
		vs[colIdx] = t.fillColumn(colType, arr, fillValue, t.alloc)
	}
	return w.Write(vs)
}

func (t *fillTransformation) addNullColumn(typ flux.ColType, l int, mem arrowmem.Allocator) array.Array {
	builder := arrow.NewBuilder(typ, mem)
	builder.Resize(l)
	for i := 0; i < l; i++ {
		builder.AppendNull()
	}
	return builder.NewArray()
}
