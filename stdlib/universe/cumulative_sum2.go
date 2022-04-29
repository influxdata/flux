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
)

type cumulativeSumTransformation2 struct {
	columns []string
}

func newCumulativeSumTransformation2(id execute.DatasetID, spec *CumulativeSumProcedureSpec, mem memory.Allocator) (execute.Transformation, execute.Dataset, error) {
	tr := &cumulativeSumTransformation2{
		columns: spec.Columns,
	}
	return execute.NewNarrowStateTransformation(id, tr, mem)
}

func (c *cumulativeSumTransformation2) Process(chunk table.Chunk, state interface{}, d *execute.TransportDataset, mem memory.Allocator) (interface{}, bool, error) {
	s, _ := state.(map[string]*cumulativeSumState)
	if s == nil {
		s = make(map[string]*cumulativeSumState)
	}

	if err := c.processChunk(chunk, s, d, mem); err != nil {
		return nil, false, err
	}
	return s, true, nil
}

func (c *cumulativeSumTransformation2) processChunk(chunk table.Chunk, state map[string]*cumulativeSumState, d *execute.TransportDataset, mem memory.Allocator) error {
	buffer := arrow.TableBuffer{
		GroupKey: chunk.Key(),
		Columns:  chunk.Cols(),
		Values:   make([]array.Array, chunk.NCols()),
	}

	for j, col := range chunk.Cols() {
		arr := chunk.Values(j)
		if !execute.ContainsStr(c.columns, col.Label) {
			arr.Retain()
			buffer.Values[j] = arr
			continue
		}

		sumer, ok := state[col.Label]
		if !ok {
			sumer = newCumulativeSumState(col.Type)
			if sumer == nil {
				arr.Retain()
				buffer.Values[j] = arr
				continue
			}
			state[col.Label] = sumer
		} else if sumer.inType != col.Type {
			return errors.Newf(codes.FailedPrecondition, "schema collision detected: column \"%s\" is both of type %s and %s", col.Label, col.Type, sumer.inType)
		}

		buffer.Values[j] = sumer.Sum(arr, mem)
		if execute.ContainsStr(c.columns, col.Label) {
			if _, ok := state[col.Label]; !ok {
				state[col.Label] = &cumulativeSumState{}
			}
		}
	}

	out := table.ChunkFromBuffer(buffer)
	return d.Process(out)
}

func (c *cumulativeSumTransformation2) Close() error {
	return nil
}

type cumulativeSum2 interface {
	Sum(arr array.Array, mem memory.Allocator) array.Array
}

type cumulativeSumState struct {
	inType flux.ColType
	cumulativeSum2
}

func newCumulativeSumState(inType flux.ColType) *cumulativeSumState {
	state := &cumulativeSumState{inType: inType}
	switch inType {
	case flux.TFloat:
		state.cumulativeSum2 = &cumulativeSumFloat{}
	case flux.TInt:
		state.cumulativeSum2 = &cumulativeSumInt{}
	case flux.TUInt:
		state.cumulativeSum2 = &cumulativeSumUint{}
	default:
		return nil
	}
	return state
}

type cumulativeSumFloat struct {
	sum float64
}

func (c *cumulativeSumFloat) Sum(arr array.Array, mem memory.Allocator) array.Array {
	b := array.NewFloatBuilder(mem)
	b.Resize(arr.Len())

	vs := arr.(*array.Float)
	for i, n := 0, vs.Len(); i < n; i++ {
		if vs.IsValid(i) {
			c.sum += vs.Value(i)
		}
		b.Append(c.sum)
	}
	return b.NewArray()
}

type cumulativeSumInt struct {
	sum int64
}

func (c *cumulativeSumInt) Sum(arr array.Array, mem memory.Allocator) array.Array {
	b := array.NewIntBuilder(mem)
	b.Resize(arr.Len())

	vs := arr.(*array.Int)
	for i, n := 0, vs.Len(); i < n; i++ {
		if vs.IsValid(i) {
			c.sum += vs.Value(i)
		}
		b.Append(c.sum)
	}
	return b.NewArray()
}

type cumulativeSumUint struct {
	sum uint64
}

func (c *cumulativeSumUint) Sum(arr array.Array, mem memory.Allocator) array.Array {
	b := array.NewUintBuilder(mem)
	b.Resize(arr.Len())

	vs := arr.(*array.Uint)
	for i, n := 0, vs.Len(); i < n; i++ {
		if vs.IsValid(i) {
			c.sum += vs.Value(i)
		}
		b.Append(c.sum)
	}
	return b.NewArray()
}
