package universe

import (
	"github.com/apache/arrow/go/v7/arrow/memory"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/arrowutil"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/internal/execute/table"
	"github.com/influxdata/flux/internal/feature"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
)

const DifferenceKind = "difference"

type DifferenceOpSpec struct {
	NonNegative bool     `json:"nonNegative"`
	Columns     []string `json:"columns"`
	KeepFirst   bool     `json:"keepFirst"`
	InitialZero bool     `json:"initialZero"`
}

func init() {
	differenceSignature := runtime.MustLookupBuiltinType("universe", "difference")

	runtime.RegisterPackageValue("universe", DifferenceKind, flux.MustValue(flux.FunctionValue(DifferenceKind, createDifferenceOpSpec, differenceSignature)))
	flux.RegisterOpSpec(DifferenceKind, newDifferenceOp)
	plan.RegisterProcedureSpec(DifferenceKind, newDifferenceProcedure, DifferenceKind)
	execute.RegisterTransformation(DifferenceKind, createDifferenceTransformation)
}

func createDifferenceOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(DifferenceOpSpec)

	if nn, ok, err := args.GetBool("nonNegative"); err != nil {
		return nil, err
	} else if ok {
		spec.NonNegative = nn
	}

	if cols, ok, err := args.GetArray("columns", semantic.String); err != nil {
		return nil, err
	} else if ok {
		columns, err := interpreter.ToStringArray(cols)
		if err != nil {
			return nil, err
		}
		spec.Columns = columns
	} else {
		spec.Columns = []string{execute.DefaultValueColLabel}
	}

	if keepFirst, ok, err := args.GetBool("keepFirst"); err != nil {
		return nil, err
	} else if ok {
		spec.KeepFirst = keepFirst
	} else {
		spec.KeepFirst = false
	}

	if initialZero, ok, err := args.GetBool("initialZero"); err != nil {
		return nil, err
	} else if ok {
		spec.InitialZero = initialZero
	} else {
		spec.InitialZero = false
	}

	return spec, nil
}

func newDifferenceOp() flux.OperationSpec {
	return new(DifferenceOpSpec)
}

func (s *DifferenceOpSpec) Kind() flux.OperationKind {
	return DifferenceKind
}

type DifferenceProcedureSpec struct {
	plan.DefaultCost
	NonNegative bool     `json:"non_negative"`
	Columns     []string `json:"columns"`
	KeepFirst   bool     `json:"keepFirst"`
	InitialZero bool     `json:"initialZero"`
}

func newDifferenceProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*DifferenceOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	return &DifferenceProcedureSpec{
		NonNegative: spec.NonNegative,
		Columns:     spec.Columns,
		KeepFirst:   spec.KeepFirst,
		InitialZero: spec.InitialZero,
	}, nil
}

func (s *DifferenceProcedureSpec) Kind() plan.ProcedureKind {
	return DifferenceKind
}
func (s *DifferenceProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(DifferenceProcedureSpec)
	*ns = *s
	if s.Columns != nil {
		ns.Columns = make([]string, len(s.Columns))
		copy(ns.Columns, s.Columns)
	}
	return ns
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *DifferenceProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

func createDifferenceTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*DifferenceProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}

	if feature.NarrowTransformationDifference().Enabled(a.Context()) {
		return NewNarrowDifferenceTransformation(s, id, a.Allocator())
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewDifferenceTransformation(d, cache, s)
	return t, d, nil
}

type differenceTransformation struct {
	execute.ExecutionNode
	d     execute.Dataset
	cache execute.TableBuilderCache

	nonNegative bool
	columns     []string
	keepFirst   bool
	initialZero bool
}

func NewDifferenceTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *DifferenceProcedureSpec) *differenceTransformation {
	return &differenceTransformation{
		d:           d,
		cache:       cache,
		nonNegative: spec.NonNegative,
		columns:     spec.Columns,
		keepFirst:   spec.KeepFirst,
		initialZero: spec.InitialZero,
	}
}

func (t *differenceTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *differenceTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	builder, created := t.cache.TableBuilder(tbl.Key())
	if !created {
		return errors.Newf(codes.FailedPrecondition, "difference found duplicate table with key: %v", tbl.Key())
	}
	cols := tbl.Cols()
	differences := make([]*difference, len(cols))
	for j, c := range cols {
		found := false
		for _, label := range t.columns {
			if c.Label == label {
				found = true
				break
			}
		}
		if !found {
			if _, err := builder.AddCol(c); err != nil {
				return err
			}
			continue
		}
		var typ flux.ColType
		switch c.Type {
		case flux.TInt, flux.TUInt:
			typ = flux.TInt
		case flux.TFloat:
			typ = flux.TFloat
		case flux.TTime:
			return errors.New(codes.FailedPrecondition, "difference does not support time columns. Try the elapsed function")
		default:
			return errors.Newf(codes.Invalid, `difference does not support column "%s" of type "%v"`, c.Label, c.Type)
		}
		if _, err := builder.AddCol(flux.ColMeta{
			Label: c.Label,
			Type:  typ,
		}); err != nil {
			return err
		}
		differences[j] = newDifference(t.nonNegative, t.keepFirst, t.initialZero)
	}

	// We need to drop the first row since its difference is undefined
	firstIdx := 1
	if t.keepFirst {
		// The user wants to keep the first row
		firstIdx = 0
	}
	return tbl.Do(func(cr flux.ColReader) error {
		l := cr.Len()
		if l == 0 {
			return nil
		}
		for j, c := range cols {
			d := differences[j]
			switch c.Type {
			case flux.TBool:
				s := arrow.BoolSlice(cr.Bools(j), firstIdx, l)
				if err := builder.AppendBools(j, s); err != nil {
					s.Release()
					return err
				}
				s.Release()
			case flux.TInt:
				values := cr.Ints(j)
				if d == nil {
					s := arrow.IntSlice(values, firstIdx, l)
					if err := builder.AppendInts(j, s); err != nil {
						s.Release()
						return err
					}
					s.Release()
					continue
				}
				for i := 0; i < l; i++ {
					v, ok := d.updateInt(values.Value(i), values.IsValid(i))
					if i < firstIdx {
						continue
					}
					if ok {
						if err := builder.AppendInt(j, v); err != nil {
							return err
						}
					} else {
						if err := builder.AppendNil(j); err != nil {
							return err
						}
					}
				}
			case flux.TUInt:
				values := cr.UInts(j)
				if d == nil {
					s := arrow.UintSlice(values, firstIdx, l)
					if err := builder.AppendUInts(j, s); err != nil {
						s.Release()
						return err
					}
					s.Release()
					continue
				}
				for i := 0; i < l; i++ {
					v, ok := d.updateUInt(values.Value(i), values.IsValid(i))
					if i < firstIdx {
						continue
					}
					if ok {
						if err := builder.AppendInt(j, v); err != nil {
							return err
						}
					} else {
						if err := builder.AppendNil(j); err != nil {
							return err
						}
					}
				}
			case flux.TFloat:
				values := cr.Floats(j)
				if d == nil {
					s := arrow.FloatSlice(values, firstIdx, l)
					if err := builder.AppendFloats(j, s); err != nil {
						s.Release()
						return err
					}
					s.Release()
					continue
				}
				for i := 0; i < l; i++ {
					v, ok := d.updateFloat(values.Value(i), values.IsValid(i))
					if i < firstIdx {
						continue
					}
					if ok {
						if err := builder.AppendFloat(j, v); err != nil {
							return err
						}
					} else {
						if err := builder.AppendNil(j); err != nil {
							return err
						}
					}
				}
			case flux.TString:
				s := arrow.StringSlice(cr.Strings(j), firstIdx, l)
				if err := builder.AppendStrings(j, s); err != nil {
					s.Release()
					return err
				}
				s.Release()
			case flux.TTime:
				s := arrow.IntSlice(cr.Times(j), firstIdx, l)
				if err := builder.AppendTimes(j, s); err != nil {
					s.Release()
					return err
				}
				s.Release()
			}
		}

		// Now that we skipped the first row, start at 0 for the rest of the batches
		firstIdx = 0
		return nil
	})
}

func (t *differenceTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *differenceTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *differenceTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}

type differenceTransformationAdapter struct {
	differenceTransformation differenceTransformation
}

func NewNarrowDifferenceTransformation(spec *DifferenceProcedureSpec, id execute.DatasetID, alloc memory.Allocator) (execute.Transformation, execute.Dataset, error) {
	differenceTransformation := differenceTransformation{
		nonNegative: spec.NonNegative,
		columns:     spec.Columns,
		keepFirst:   spec.KeepFirst,
		initialZero: spec.InitialZero,
	}
	t := &differenceTransformationAdapter{
		differenceTransformation,
	}
	return execute.NewNarrowStateTransformation[*differenceState](id, t, alloc)
}

func (t *differenceTransformation) createDifferences(cols []flux.ColMeta) []*difference {

	differences := make([]*difference, len(cols))
	for j, c := range cols {
		found := false
		for _, label := range t.columns {
			if c.Label == label {
				found = true
				break
			}
		}
		if !found {
			continue
		}
		differences[j] = newDifference(t.nonNegative, t.keepFirst, t.initialZero)
	}
	return differences
}

func (t *differenceTransformation) createOutputColumns(cols []flux.ColMeta) ([]flux.ColMeta, error) {
	newCols := make([]flux.ColMeta, len(cols))
	for j, c := range cols {
		found := false
		for _, label := range t.columns {
			if c.Label == label {
				found = true
				break
			}
		}
		if !found {
			newCols[j] = c
			continue
		}
		var typ flux.ColType
		switch c.Type {
		case flux.TInt, flux.TUInt:
			typ = flux.TInt
		case flux.TFloat:
			typ = flux.TFloat
		case flux.TTime:
			return nil, errors.New(codes.FailedPrecondition, "difference does not support time columns. Try the elapsed function")
		default:
			return nil, errors.Newf(codes.Invalid, `difference does not support column "%s" of type "%v"`, c.Label, c.Type)
		}
		newCols[j] = flux.ColMeta{
			Label: c.Label,
			Type:  typ,
		}
	}
	return newCols, nil
}

type differenceState struct {
	differences   []*difference
	firstIdx      int
	outputColumns []flux.ColMeta
}

func (t *differenceTransformationAdapter) Process(chunk table.Chunk, state *differenceState, d *execute.TransportDataset, mem memory.Allocator) (*differenceState, bool, error) {
	return t.differenceTransformation.adaptedProcess(chunk, state, d, mem)
}

func (t *differenceTransformation) adaptedProcess(chunk table.Chunk, state *differenceState, d *execute.TransportDataset, mem memory.Allocator) (*differenceState, bool, error) {
	if state == nil {
		// We need to drop the first row since its difference is undefined
		firstIdx := 1
		if t.keepFirst {
			// The user wants to keep the first row
			firstIdx = 0
		}
		outputColumns, err := t.createOutputColumns(chunk.Cols())
		if err != nil {
			return nil, false, err
		}
		state = &differenceState{
			differences:   t.createDifferences(chunk.Cols()),
			firstIdx:      firstIdx,
			outputColumns: outputColumns,
		}
	}

	buffer := arrow.TableBuffer{
		GroupKey: chunk.Key(),
		Columns:  state.outputColumns,
		Values:   make([]array.Array, len(state.outputColumns)),
	}

	if err := t.processChunk(state.differences, state.firstIdx, mem, &buffer, chunk); err != nil {
		return nil, false, err
	}

	// Now that we skipped the first row, start at 0 for the rest of the batches
	state.firstIdx = 0

	out := table.ChunkFromBuffer(buffer)
	if err := d.Process(out); err != nil {
		return nil, false, err
	}

	return state, true, nil
}

func (t *differenceTransformationAdapter) Close() error { return nil }

func (t *differenceTransformation) processChunk(differences []*difference, firstIdx int, mem memory.Allocator, buffer *arrow.TableBuffer, chunk table.Chunk) error {

	l := chunk.Len()
	for j, c := range chunk.Cols() {
		d := differences[j]
		var out array.Array
		if l == 0 {
			out = arrow.Empty(c.Type)
		} else {
			switch c.Type {
			case flux.TInt:
				values := chunk.Ints(j)
				out = processInts(d, l, values, firstIdx, mem)
			case flux.TUInt:
				values := chunk.Uints(j)
				out = processUints(d, l, values, firstIdx, mem)
			case flux.TFloat:
				values := chunk.Floats(j)
				out = processFloats(d, l, values, firstIdx, mem)
			case flux.TString:
				out = arrow.StringSlice(chunk.Strings(j), firstIdx, l)
			case flux.TTime:
				out = arrow.IntSlice(chunk.Ints(j), firstIdx, l)
			case flux.TBool:
				out = arrow.BoolSlice(chunk.Bools(j), firstIdx, l)
			}
		}
		buffer.Values[j] = out
	}
	return nil
}

func processInts(d *difference, l int, values *array.Int, firstIdx int, mem memory.Allocator) *array.Int {
	if d == nil {
		return arrow.IntSlice(values, firstIdx, l)
	}

	b := arrowutil.NewIntBuilder(mem)
	b.Resize(l)
	for i := 0; i < l; i++ {
		v, ok := d.updateInt(values.Value(i), values.IsValid(i))
		if i < firstIdx {
			continue
		}
		if ok {
			b.Append(v)
		} else {
			b.AppendNull()
		}
	}
	return b.NewIntArray()
}

func processUints(d *difference, l int, values *array.Uint, firstIdx int, mem memory.Allocator) array.Array {
	if d == nil {
		return arrow.UintSlice(values, firstIdx, l)
	}

	b := arrowutil.NewIntBuilder(mem)
	b.Resize(l)
	for i := 0; i < l; i++ {
		v, ok := d.updateUInt(values.Value(i), values.IsValid(i))
		if i < firstIdx {
			continue
		}
		if ok {
			b.Append(v)
		} else {
			b.AppendNull()
		}
	}
	return b.NewIntArray()
}

func processFloats(d *difference, l int, values *array.Float, firstIdx int, mem memory.Allocator) *array.Float {
	if d == nil {
		return arrow.FloatSlice(values, firstIdx, l)
	}

	b := arrowutil.NewFloatBuilder(mem)
	b.Resize(l)
	for i := 0; i < l; i++ {
		v, ok := d.updateFloat(values.Value(i), values.IsValid(i))
		if i < firstIdx {
			continue
		}
		if ok {
			b.Append(v)
		} else {
			b.AppendNull()
		}
	}
	return b.NewFloatArray()
}

func newDifference(nonNegative, keepFirst, initialZero bool) *difference {
	return &difference{
		nonNegative: nonNegative,
		keepFirst:   keepFirst,
		initialZero: initialZero,
	}
}

type difference struct {
	nonNegative bool
	keepFirst   bool
	initialZero bool

	valid       bool
	pIntValue   int64
	pUIntValue  uint64
	pFloatValue float64
}

func (d *difference) updateInt(v int64, valid bool) (int64, bool) {
	if !valid {
		return 0, false
	}
	prev := d.pIntValue
	d.pIntValue = v
	if !d.valid && d.keepFirst && d.initialZero {
		d.valid = true
		return 0, true
	}
	if !d.valid {
		d.valid = true
		return 0, false
	}
	diff := v - prev
	if diff >= 0 || !d.nonNegative {
		return diff, true
	} else if d.nonNegative && d.initialZero && v >= 0 {
		return v, true
	}
	return 0, false
}

func (d *difference) updateUInt(v uint64, valid bool) (int64, bool) {
	if !valid {
		return 0, false
	}
	prev := d.pUIntValue
	d.pUIntValue = v
	if !d.valid && d.keepFirst && d.initialZero {
		d.valid = true
		return 0, true
	}
	if !d.valid {
		d.valid = true
		return 0, false
	}
	// Note: the unsigned substraction works correctly even for negative differences
	// because of two's-complement arithmetic.
	diff := int64(v - prev)
	if diff >= 0 || !d.nonNegative {
		return diff, true
	} else if d.nonNegative && d.initialZero && int64(v) >= 0 {
		return int64(v), true
	}
	return 0, false
}

func (d *difference) updateFloat(v float64, valid bool) (float64, bool) {
	if !valid {
		return 0, false
	}
	prev := d.pFloatValue
	d.pFloatValue = v
	if !d.valid && d.keepFirst && d.initialZero {
		d.valid = true
		return 0, true
	}
	if !d.valid {
		d.valid = true
		return 0, false
	}
	diff := v - prev
	if diff >= 0 || !d.nonNegative {
		return diff, true
	} else if d.nonNegative && d.initialZero && v >= 0 {
		return v, true
	}
	return 0, false
}
