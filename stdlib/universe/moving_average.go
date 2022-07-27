package universe

import (
	stdarrow "github.com/apache/arrow/go/v7/arrow"
	"github.com/apache/arrow/go/v7/arrow/bitutil"
	"github.com/apache/arrow/go/v7/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
)

const MovingAverageKind = "movingAverage"

type MovingAverageOpSpec struct {
	N int64 `json:"n"`
}

func init() {
	movingAverageSignature := runtime.MustLookupBuiltinType("universe", "movingAverage")

	runtime.RegisterPackageValue("universe", MovingAverageKind, flux.MustValue(flux.FunctionValue(MovingAverageKind, createMovingAverageOpSpec, movingAverageSignature)))
	flux.RegisterOpSpec(MovingAverageKind, newMovingAverageOp)
	plan.RegisterProcedureSpec(MovingAverageKind, newMovingAverageProcedure, MovingAverageKind)
	execute.RegisterTransformation(MovingAverageKind, createMovingAverageTransformation)
}

func createMovingAverageOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(MovingAverageOpSpec)

	if n, err := args.GetRequiredInt("n"); err != nil {
		return nil, err
	} else if n <= 0 {
		return nil, errors.Newf(codes.Invalid, "cannot take moving average with a period of %v (must be greater than 0)", n)
	} else {
		spec.N = n
	}

	return spec, nil
}

func newMovingAverageOp() flux.OperationSpec {
	return new(MovingAverageOpSpec)
}

func (s *MovingAverageOpSpec) Kind() flux.OperationKind {
	return MovingAverageKind
}

type MovingAverageProcedureSpec struct {
	plan.DefaultCost
	N int64 `json:"n"`
}

func newMovingAverageProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*MovingAverageOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	return &MovingAverageProcedureSpec{
		N: spec.N,
	}, nil
}

func (s *MovingAverageProcedureSpec) Kind() plan.ProcedureKind {
	return MovingAverageKind
}

func (s *MovingAverageProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(MovingAverageProcedureSpec)
	*ns = *s
	return ns
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *MovingAverageProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

func createMovingAverageTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*MovingAverageProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	return NewMovingAverageTransformation(id, s, a.Allocator())
}

type movingAverageTransformation struct {
	n int64
}

func NewMovingAverageTransformation(id execute.DatasetID, spec *MovingAverageProcedureSpec, mem memory.Allocator) (execute.Transformation, execute.Dataset, error) {
	tr := &movingAverageTransformation{
		n: spec.N,
	}
	return execute.NewNarrowStateTransformation[*movingAverageState](id, tr, mem)
}

func (m *movingAverageTransformation) Process(chunk table.Chunk, state *movingAverageState, d *execute.TransportDataset, mem memory.Allocator) (*movingAverageState, bool, error) {
	newState, err := m.processChunk(chunk, state, d, mem)
	if err != nil {
		return nil, false, err
	}
	return newState, true, nil
}

func (m *movingAverageTransformation) processChunk(chunk table.Chunk, state *movingAverageState, d *execute.TransportDataset, mem memory.Allocator) (*movingAverageState, error) {
	idx := chunk.Index(execute.DefaultValueColLabel)
	if idx < 0 {
		return nil, errors.Newf(codes.FailedPrecondition, "cannot find _value column")
	}

	col := chunk.Col(idx)
	if col.Type != flux.TInt && col.Type != flux.TUInt && col.Type != flux.TFloat {
		return nil, errors.Newf(codes.FailedPrecondition, "cannot take moving average of column %s (type %s)", col.Label, col.Type.String())
	} else if m.n <= 0 {
		// Defensive code. This should already be forbidden by the argument parser.
		return nil, errors.Newf(codes.Invalid, "cannot take moving average with a period of %v (must be greater than 0)", m.n)
	}

	if state == nil {
		state = newMovingAverageState(m.n, col.Type, d, mem)
	} else {
		if state.inType != col.Type {
			return nil, errors.Newf(codes.FailedPrecondition, "schema collision detected: column \"%s\" is both of type %s and %s", col.Label, col.Type, state.inType)
		}
	}

	cols := chunk.Cols()
	if col.Type != flux.TFloat {
		// The schema is changing so we have to recreate the columns.
		newCols := make([]flux.ColMeta, len(cols))
		copy(newCols, cols)
		newCols[idx].Type = flux.TFloat
		cols = newCols
	}

	buffer := arrow.TableBuffer{
		GroupKey: chunk.Key(),
		Columns:  cols,
		Values:   make([]array.Array, len(cols)),
	}

	n := int64(chunk.Len()) - state.needed
	for i, col := range cols {
		if i == idx {
			continue
		} else if n < 0 {
			buffer.Values[i] = arrow.Empty(col.Type)
			continue
		}

		arr := chunk.Values(i)
		if n < int64(arr.Len()) {
			buffer.Values[i] = arrow.Slice(arr, state.needed, int64(arr.Len()))
		} else {
			arr.Retain()
			buffer.Values[i] = arr
		}
	}

	b := array.NewFloatBuilder(mem)
	b.Resize(int(n))
	switch arr := chunk.Values(idx).(type) {
	case *array.Float:
		state.ProcessFloats(b, arr)
	case *array.Int:
		state.ProcessInts(b, arr)
	case *array.Uint:
		state.ProcessUints(b, arr)
	}
	buffer.Values[idx] = b.NewArray()

	out := table.ChunkFromBuffer(buffer)
	if err := d.Process(out); err != nil {
		return nil, err
	}

	// We have not output a row.
	// Store the chunk in case we need it.
	if state.needed > 0 && chunk.Len() > 0 {
		if state.last != nil {
			state.last.Release()
		}
		chunk.Retain()
		state.last = &chunk
	} else if state.needed == 0 && state.last != nil {
		state.last.Release()
		state.last = nil
	}
	return state, nil
}

func (m *movingAverageTransformation) Close() error {
	return nil
}

type movingAverageState struct {
	data   []float64
	mask   []byte
	index  int
	needed int64
	inType flux.ColType
	last   *table.Chunk
	d      *execute.TransportDataset
	mem    memory.Allocator
}

func newMovingAverageState(n int64, inType flux.ColType, d *execute.TransportDataset, mem memory.Allocator) *movingAverageState {
	data := mem.Allocate(stdarrow.Float64Traits.BytesRequired(int(n)))
	mask := mem.Allocate(int(bitutil.BytesForBits(n)))
	return &movingAverageState{
		data:   stdarrow.Float64Traits.CastFromBytes(data),
		mask:   mask,
		needed: n - 1,
		inType: inType,
		d:      d,
		mem:    mem,
	}
}

func (m *movingAverageState) ProcessFloats(b *array.FloatBuilder, arr *array.Float) {
	for i, n := 0, arr.Len(); i < n; i++ {
		if arr.IsNull(i) {
			bitutil.ClearBit(m.mask, m.index)
		} else {
			m.data[m.index] = arr.Value(i)
			bitutil.SetBit(m.mask, m.index)
		}

		m.index++
		if m.index >= len(m.data) {
			m.index = 0
		}
		if m.needed == 0 {
			if v, ok := m.Compute(); ok {
				b.Append(v)
			} else {
				b.AppendNull()
			}
		} else {
			m.needed--
		}
	}
}

func (m *movingAverageState) ProcessInts(b *array.FloatBuilder, arr *array.Int) {
	for i, n := 0, arr.Len(); i < n; i++ {
		if arr.IsNull(i) {
			bitutil.ClearBit(m.mask, m.index)
		} else {
			m.data[m.index] = float64(arr.Value(i))
			bitutil.SetBit(m.mask, m.index)
		}

		m.index++
		if m.index >= len(m.data) {
			m.index = 0
		}
		if m.needed == 0 {
			if v, ok := m.Compute(); ok {
				b.Append(v)
			} else {
				b.AppendNull()
			}
		} else {
			m.needed--
		}
	}
}

func (m *movingAverageState) ProcessUints(b *array.FloatBuilder, arr *array.Uint) {
	for i, n := 0, arr.Len(); i < n; i++ {
		if arr.IsNull(i) {
			bitutil.ClearBit(m.mask, m.index)
		} else {
			m.data[m.index] = float64(arr.Value(i))
			bitutil.SetBit(m.mask, m.index)
		}

		m.index++
		if m.index >= len(m.data) {
			m.index = 0
		}
		if m.needed == 0 {
			if v, ok := m.Compute(); ok {
				b.Append(v)
			} else {
				b.AppendNull()
			}
		} else {
			m.needed--
		}
	}
}

func (m *movingAverageState) Compute() (float64, bool) {
	var (
		sum float64
		n   int64
	)
	for i, f := range m.data {
		if bitutil.BitIsSet(m.mask, i) {
			sum += f
			n++
		}
	}
	if n == 0 {
		return 0, false
	}
	return sum / float64(n), true
}

func (m *movingAverageState) forceValue() error {
	if m.last == nil {
		// No points at all so nothing to force.
		return nil
	}

	chunk := *m.last
	defer chunk.Release()

	idx := chunk.Index(execute.DefaultValueColLabel)
	col := chunk.Col(idx)

	cols := chunk.Cols()
	if col.Type != flux.TFloat {
		// The schema is changing so we have to recreate the columns.
		newCols := make([]flux.ColMeta, len(cols))
		copy(newCols, cols)
		newCols[idx].Type = flux.TFloat
		cols = newCols
	}

	buffer := arrow.TableBuffer{
		GroupKey: chunk.Key(),
		Columns:  cols,
		Values:   make([]array.Array, len(cols)),
	}
	for i, col := range cols {
		if i == idx {
			b := array.NewFloatBuilder(m.mem)
			b.Resize(1)
			if v, ok := m.Compute(); ok {
				b.Append(v)
			} else {
				b.AppendNull()
			}
			buffer.Values[i] = b.NewArray()
			continue
		}

		b := arrow.NewBuilder(col.Type, m.mem)
		b.Resize(1)
		arr := chunk.Values(i)
		if arr.IsNull(arr.Len() - 1) {
			b.AppendNull()
		} else {
			switch b := b.(type) {
			case *array.IntBuilder:
				arr := arr.(*array.Int)
				b.Append(arr.Value(arr.Len() - 1))
			case *array.UintBuilder:
				arr := arr.(*array.Uint)
				b.Append(arr.Value(arr.Len() - 1))
			case *array.FloatBuilder:
				arr := arr.(*array.Float)
				b.Append(arr.Value(arr.Len() - 1))
			case *array.StringBuilder:
				arr := arr.(*array.String)
				b.Append(arr.Value(arr.Len() - 1))
			case *array.BooleanBuilder:
				arr := arr.(*array.Boolean)
				b.Append(arr.Value(arr.Len() - 1))
			default:
				return errors.Newf(codes.Internal, "unknown builder type: %T", b)
			}
		}
		buffer.Values[i] = b.NewArray()
	}

	out := table.ChunkFromBuffer(buffer)
	return m.d.Process(out)
}

func (m *movingAverageState) Close() (err error) {
	if m.needed > 0 {
		err = m.forceValue()
	}

	if m.data != nil {
		buf := stdarrow.Float64Traits.CastToBytes(m.data)
		m.mem.Free(buf)
		m.data = nil
	}
	if m.mask != nil {
		m.mem.Free(m.mask)
		m.mask = nil
	}
	return err
}
