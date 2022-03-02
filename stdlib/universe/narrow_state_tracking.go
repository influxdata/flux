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

func NewNarrowStateTrackingTransformation(t *stateTrackingTransformation, id execute.DatasetID, mem memory.Allocator) (execute.Transformation, execute.Dataset, error) {
	a := &narrowStateTrackingTransformationAdapter{t: t}
	nt, d, err := execute.NewNarrowStateTransformation(id, a, mem)
	if err != nil {
		return nil, nil, err
	}
	return nt, d, nil
}

type narrowStateTrackingTransformationAdapter struct {
	t *stateTrackingTransformation
}

type trackedState struct {
	start,
	prevTime values.Time

	countInState,
	durationInState bool

	count,
	duration int64
}

func (a *narrowStateTrackingTransformationAdapter) Process(chunk table.Chunk, state interface{}, d *execute.TransportDataset, mem memory.Allocator) (interface{}, bool, error) {
	// Track whether or not the state has been modified
	mod := false

	if chunk.Len() == 0 {
		return state, mod, nil
	}

	if state == nil {
		state = trackedState{
			count:    -1,
			duration: -1,
		}
		mod = true
	}
	s := state.(trackedState)

	cols := chunk.Cols()
	fn, err := a.t.fn.Prepare(cols)
	if err != nil {
		return s, mod, err
	}

	// Get the index for the time column
	timeIdx := execute.ColIdx(a.t.timeCol, cols)
	if timeIdx < 0 {
		// Do we need to retain and process the unmodified chunk
		// before an early return?
		//
		// chunk.Retain()
		// d.Process(chunk)
		return s, mod, errors.Newf(codes.FailedPrecondition, "column %q does not exist", a.t.timeCol)
	}

	stateCounts := array.NewIntBuilder(mem)
	stateDurations := array.NewIntBuilder(mem)

	nrows := chunk.Len()

	stateCounts.Resize(nrows)
	stateDurations.Resize(nrows)

	for i := 0; i < nrows; i++ {
		// Evaluate the predicate for the current row
		match, err := fn.EvalRow(a.t.ctx, i, chunk)
		if err != nil {
			return s, mod, err
		}

		// Update state duration
		if a.t.durationColumn != "" {
			times := chunk.Times(timeIdx)
			if times.IsNull(i) {
				return s, mod, errors.New(codes.FailedPrecondition, "got a null timestamp")
			}
			ts := values.Time(times.Value(i))
			if s.prevTime > ts {
				return s, mod, errors.New(codes.FailedPrecondition, "got an out-of-order timestamp")
			}
			s.prevTime = ts
			mod = true

			if match {
				if !s.durationInState {
					s.durationInState = true
					s.start = ts
					s.duration = 0
				} else {
					s.duration = int64(ts - s.start)
					if a.t.durationUnit > 0 {
						s.duration = s.duration / a.t.durationUnit
					}
				}
			} else {
				s.durationInState = false
				s.duration = -1
			}
			stateDurations.Append(s.duration)
		}

		// Update state count
		if a.t.countColumn != "" {
			if match {
				if !s.countInState {
					s.countInState = true
					s.count = 1
				} else {
					s.count++
				}
			} else {
				s.countInState = false
				s.count = -1
			}
			mod = true
			stateCounts.Append(s.count)
		}
	}

	ncols := chunk.NCols()
	newCols := make([]flux.ColMeta, 0, ncols+2)
	newCols = append(newCols, cols...)

	vs := make([]array.Interface, 0, ncols+2)
	for i := 0; i < ncols; i++ {
		vs = append(vs, chunk.Values(i))
	}

	if a.t.countColumn != "" {
		newCols = append(newCols, flux.ColMeta{Label: a.t.countColumn, Type: flux.TInt})
		vs = append(vs, stateCounts.NewArray())
	} else {
		stateCounts.Release()
	}

	if a.t.durationColumn != "" {
		newCols = append(newCols, flux.ColMeta{Label: a.t.durationColumn, Type: flux.TInt})
		vs = append(vs, stateDurations.NewArray())
	} else {
		stateDurations.Release()
	}

	buffer := arrow.TableBuffer{
		GroupKey: chunk.Key(),
		Columns:  newCols,
		Values:   vs,
	}
	c := table.ChunkFromBuffer(buffer)
	c.Retain()
	err = d.Process(c)
	return s, mod, err
}

func (a *narrowStateTrackingTransformationAdapter) Close() error { return nil }
