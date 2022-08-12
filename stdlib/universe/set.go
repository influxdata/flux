package universe

import (
	"github.com/apache/arrow/go/v7/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/feature"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
)

const SetKind = "set"

type SetOpSpec struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func init() {
	setSignature := runtime.MustLookupBuiltinType("universe", "set")

	runtime.RegisterPackageValue("universe", SetKind, flux.MustValue(flux.FunctionValue(SetKind, createSetOpSpec, setSignature)))
	plan.RegisterProcedureSpec(SetKind, newSetProcedure, SetKind)
	execute.RegisterTransformation(SetKind, createSetTransformation)
}

func createSetOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(SetOpSpec)
	key, err := args.GetRequiredString("key")
	if err != nil {
		return nil, err
	}
	spec.Key = key

	value, err := args.GetRequiredString("value")
	if err != nil {
		return nil, err
	}
	spec.Value = value

	return spec, nil
}

func (s *SetOpSpec) Kind() flux.OperationKind {
	return SetKind
}

type SetProcedureSpec struct {
	plan.DefaultCost
	Key, Value string
}

func newSetProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	s, ok := qs.(*SetOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}
	p := &SetProcedureSpec{
		Key:   s.Key,
		Value: s.Value,
	}
	return p, nil
}

func (s *SetProcedureSpec) Kind() plan.ProcedureKind {
	return SetKind
}
func (s *SetProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(SetProcedureSpec)
	ns.Key = s.Key
	ns.Value = s.Value
	return ns
}

func createSetTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*SetProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}

	if feature.OptimizeSetTransformation().Enabled(a.Context()) {
		tr := &setTransformation2{
			key:   s.Key,
			value: s.Value,
		}
		return execute.NewGroupTransformation(id, tr, a.Allocator())
	}

	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewSetTransformation(d, cache, s)
	return t, d, nil
}

type setTransformation struct {
	execute.ExecutionNode
	d     execute.Dataset
	cache execute.TableBuilderCache

	key, value string
}

func NewSetTransformation(
	d execute.Dataset,
	cache execute.TableBuilderCache,
	spec *SetProcedureSpec,
) execute.Transformation {
	return &setTransformation{
		d:     d,
		cache: cache,
		key:   spec.Key,
		value: spec.Value,
	}
}

func (t *setTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	// TODO
	return nil
}

func (t *setTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	key := tbl.Key()
	if idx := execute.ColIdx(t.key, key.Cols()); idx >= 0 {
		// Update key
		cols := make([]flux.ColMeta, len(key.Cols()))
		vs := make([]values.Value, len(key.Cols()))
		for j, c := range key.Cols() {
			cols[j] = c
			if j == idx {
				vs[j] = values.NewString(t.value)
			} else {
				vs[j] = key.Value(j)
			}
		}
		key = execute.NewGroupKey(cols, vs)
	}
	builder, created := t.cache.TableBuilder(key)
	if created {
		err := execute.AddTableCols(tbl, builder)
		if err != nil {
			return err
		}
		if !execute.HasCol(t.key, builder.Cols()) {
			if _, err = builder.AddCol(flux.ColMeta{
				Label: t.key,
				Type:  flux.TString,
			}); err != nil {
				return err
			}
		}
	}
	idx := execute.ColIdx(t.key, builder.Cols())
	return tbl.Do(func(cr flux.ColReader) error {
		for j := range cr.Cols() {
			if j == idx {
				continue
			}
			if err := execute.AppendCol(j, j, cr, builder); err != nil {
				return err
			}
		}
		// Set new value
		l := cr.Len()
		for i := 0; i < l; i++ {
			if err := builder.AppendString(idx, t.value); err != nil {
				return err
			}
		}
		return nil
	})
}

func (t *setTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *setTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *setTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}

type setTransformation2 struct {
	key, value string
}

func (s *setTransformation2) Process(chunk table.Chunk, d *execute.TransportDataset, mem memory.Allocator) error {
	if chunk.Key().HasCol(s.key) {
		// Need to create a new group key and change that.
		return s.processKey(chunk, d, mem)
	}
	return s.processChunk(chunk.Key(), chunk, d, mem)
}

func (s *setTransformation2) processKey(chunk table.Chunk, d *execute.TransportDataset, mem memory.Allocator) error {
	keyCols := chunk.Key().Cols()
	keyValues := make([]values.Value, len(keyCols))
	for i, col := range keyCols {
		if col.Label != s.key {
			keyValues[i] = chunk.Key().Value(i)
			continue
		}
		keyValues[i] = values.NewString(s.value)
	}
	key := execute.NewGroupKey(keyCols, keyValues)
	return s.processChunk(key, chunk, d, mem)
}

func (s *setTransformation2) processChunk(key flux.GroupKey, chunk table.Chunk, d *execute.TransportDataset, mem memory.Allocator) error {
	keyIdx := chunk.Index(s.key)
	overwrite := keyIdx >= 0 && chunk.Col(keyIdx).Type == flux.TString

	// Set the schema for the table.
	cols := chunk.Cols()
	if !overwrite {
		// If we are not overwriting an existing column or appending a new one,
		// we skip this to reduce the memory allocations.
		newCols := make([]flux.ColMeta, len(cols), len(cols)+1)
		copy(newCols, cols)
		if keyIdx >= 0 {
			newCols[keyIdx].Type = flux.TString
		} else {
			keyIdx = len(newCols)
			newCols = append(newCols, flux.ColMeta{
				Label: s.key,
				Type:  flux.TString,
			})
		}
		cols = newCols
	}

	// Copy over the arrays from the existing chunk using retain
	// and construct the string column when we find it.
	buffer := arrow.TableBuffer{
		GroupKey: key,
		Columns:  cols,
		Values:   make([]array.Array, len(cols)),
	}
	for i := range chunk.Cols() {
		if i == keyIdx {
			continue
		}
		arr := chunk.Values(i)
		arr.Retain()
		buffer.Values[i] = arr
	}
	buffer.Values[keyIdx] = array.StringRepeat(s.value, chunk.Len(), mem)

	out := table.ChunkFromBuffer(buffer)
	return d.Process(out)
}

func (s *setTransformation2) Close() error {
	return nil
}
