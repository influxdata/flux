package universe

import (
	"github.com/apache/arrow/go/arrow/memory"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	"github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/values"
)

type shiftTransformation2 struct {
	columns []string
	shift   execute.Duration
}

func newShiftTransformation2(id execute.DatasetID, spec *ShiftProcedureSpec, mem memory.Allocator) (execute.Transformation, execute.Dataset, error) {
	tr := &shiftTransformation2{
		columns: spec.Columns,
		shift:   spec.Shift,
	}
	return execute.NewNarrowTransformation(id, tr, mem)
}

func (s *shiftTransformation2) Process(chunk table.Chunk, d *execute.TransportDataset, mem memory.Allocator) error {
	key := chunk.Key()
	for _, c := range key.Cols() {
		if execute.ContainsStr(s.columns, c.Label) {
			k, err := s.regenerateKey(key)
			if err != nil {
				return err
			}
			key = k
			break
		}
	}

	buffer := arrow.TableBuffer{
		GroupKey: key,
		Columns:  chunk.Cols(),
		Values:   make([]array.Interface, chunk.NCols()),
	}
	for j, c := range chunk.Cols() {
		vs := chunk.Values(j)
		if execute.ContainsStr(s.columns, c.Label) {
			if c.Type != flux.TTime {
				return errors.Newf(codes.FailedPrecondition, "column %q is not of type time", c.Label)
			}
			buffer.Values[j] = s.shiftTimes(vs.(*array.Int), mem)
		} else {
			vs.Retain()
			buffer.Values[j] = vs
		}
	}

	out := table.ChunkFromBuffer(buffer)
	return d.Process(out)
}

func (s *shiftTransformation2) regenerateKey(key flux.GroupKey) (flux.GroupKey, error) {
	cols := key.Cols()
	vals := make([]values.Value, len(cols))
	for j, c := range cols {
		if execute.ContainsStr(s.columns, c.Label) {
			if c.Type != flux.TTime {
				return nil, errors.Newf(codes.FailedPrecondition, "column %q is not of type time", c.Label)
			}
			vals[j] = values.NewTime(key.ValueTime(j).Add(s.shift))
		} else {
			vals[j] = key.Value(j)
		}
	}
	return execute.NewGroupKey(cols, vals), nil
}

func (s *shiftTransformation2) shiftTimes(vs *array.Int, mem memory.Allocator) *array.Int {
	b := array.NewIntBuilder(mem)
	b.Resize(vs.Len())
	for i, n := 0, vs.Len(); i < n; i++ {
		if vs.IsNull(i) {
			b.AppendNull()
			continue
		}

		ts := execute.Time(vs.Value(i)).Add(s.shift)
		b.Append(int64(ts))
	}
	return b.NewIntArray()
}

func (s *shiftTransformation2) Dispose() {}
