package universe

import (
	"context"

	"github.com/apache/arrow/go/v7/arrow/memory"
	"github.com/mvn-trinhnguyen2-dn/flux"
	"github.com/mvn-trinhnguyen2-dn/flux/array"
	"github.com/mvn-trinhnguyen2-dn/flux/arrow"
	"github.com/mvn-trinhnguyen2-dn/flux/codes"
	"github.com/mvn-trinhnguyen2-dn/flux/compiler"
	"github.com/mvn-trinhnguyen2-dn/flux/execute"
	"github.com/mvn-trinhnguyen2-dn/flux/execute/table"
	"github.com/mvn-trinhnguyen2-dn/flux/internal/errors"
	"github.com/mvn-trinhnguyen2-dn/flux/values"
)

func NewNarrowStateTrackingTransformation(ctx context.Context, spec *StateTrackingProcedureSpec, id execute.DatasetID, mem memory.Allocator) (execute.Transformation, execute.Dataset, error) {
	fn := execute.NewRowPredicateFn(spec.Fn.Fn, compiler.ToScope(spec.Fn.Scope))
	t := &narrowStateTrackingTransformation{
		ctx:      ctx,
		fn:       fn,
		timeCol:  spec.TimeCol,
		countCol: spec.CountColumn,
		durCol:   spec.DurationColumn,
		unit:     int64(spec.DurationUnit.Duration()),
	}
	nt, d, err := execute.NewNarrowStateTransformation(id, t, mem)
	if err != nil {
		return nil, nil, err
	}
	return nt, d, nil
}

type narrowStateTrackingTransformation struct {
	ctx context.Context
	fn  *execute.RowPredicateFn

	timeCol,
	countCol,
	durCol string

	unit int64
}

type trackedState struct {
	start,
	prevTime values.Time

	countInState,
	durationInState bool

	count,
	duration int64
}

func (n *narrowStateTrackingTransformation) Process(chunk table.Chunk, state interface{}, d *execute.TransportDataset, mem memory.Allocator) (interface{}, bool, error) {
	// Track whether or not the state has been modified
	mod := false

	// Initialize state
	if state == nil {
		state = trackedState{
			count:    -1,
			duration: -1,
		}
		mod = true
	}
	s := state.(trackedState)
	mod, err := n.processChunk(chunk, &s, d, mem, mod)
	return s, mod, err
}

func (n *narrowStateTrackingTransformation) Close() error { return nil }

// Updates the state object for each row in the chunk, creates a new chunk with
// columns tracking counts and/or durations, and passes that chunk to the next
// transport node.
func (n *narrowStateTrackingTransformation) processChunk(chunk table.Chunk, state *trackedState, d *execute.TransportDataset, mem memory.Allocator, mod bool) (bool, error) {
	fn, err := n.fn.Prepare(chunk.Cols())
	if err != nil {
		return mod, err
	}

	timeIdx := chunk.Index(n.timeCol)
	if timeIdx < 0 {
		return mod, errors.Newf(codes.FailedPrecondition, "column %q does not exist", n.timeCol)
	}

	buf := chunk.Buffer()
	times := buf.Times(timeIdx)
	if n.durCol != "" {
		if times.NullN() > 0 {
			return mod, errors.New(codes.FailedPrecondition, "got a null timestamp")
		}
	}

	// Create the new columns
	counts := array.NewIntBuilder(mem)
	counts.Resize(chunk.Len())

	durations := array.NewIntBuilder(mem)
	durations.Resize(chunk.Len())

	for i := 0; i < chunk.Len(); i++ {
		// Evaluate the predicate for the current row
		match, err := fn.EvalRow(n.ctx, i, &buf)
		if err != nil {
			return mod, err
		}

		if mod, err = n.updateState(state, times, match, i, mod); err != nil {
			return mod, err
		}

		counts.Append(state.count)
		durations.Append(state.duration)
	}

	return mod, d.Process(n.createChunk(chunk, counts, durations))
}

// Updates the state and returns `true` if the state has been modfied.
func (n *narrowStateTrackingTransformation) updateState(state *trackedState, times *array.Int, match bool, i int, mod bool) (bool, error) {
	if n.durCol != "" {
		ts := values.Time(times.Value(i))
		if state.prevTime > ts {
			return mod, errors.New(codes.FailedPrecondition, "got an out-of-order timestamp")
		}
		state.prevTime = ts
		mod = true

		if match {
			if !state.durationInState {
				state.durationInState = true
				state.start = ts
				state.duration = 0
			} else {
				state.duration = int64(ts - state.start)
				if n.unit > 0 {
					state.duration = state.duration / n.unit
				}
			}
		} else {
			state.durationInState = false
			state.duration = -1
		}
		mod = true
	}

	if n.countCol != "" {
		if match {
			if !state.countInState {
				state.countInState = true
				state.count = 1
			} else {
				state.count++
			}
		} else {
			state.countInState = false
			state.count = -1
		}
		mod = true
	}
	return mod, nil
}

// Returns a copy of an existing chunk with the new columns appended to it.
// `counts` is released if there isn't a column name for it; ditto for `durations`.
func (n *narrowStateTrackingTransformation) createChunk(chunk table.Chunk, counts, durations *array.IntBuilder) table.Chunk {
	ncols := chunk.NCols()
	newCols := append(make([]flux.ColMeta, 0, ncols+2), chunk.Cols()...)

	vs := make([]array.Array, 0, ncols+2)
	for i := 0; i < ncols; i++ {
		col := chunk.Values(i)
		col.Retain()
		vs = append(vs, col)
	}

	if n.countCol != "" {
		newCols = append(newCols, flux.ColMeta{Label: n.countCol, Type: flux.TInt})
		vs = append(vs, counts.NewArray())
	} else {
		counts.Release()
		counts = nil
	}

	if n.durCol != "" {
		newCols = append(newCols, flux.ColMeta{Label: n.durCol, Type: flux.TInt})
		vs = append(vs, durations.NewArray())
	} else {
		durations.Release()
		durations = nil
	}

	buffer := arrow.TableBuffer{
		GroupKey: chunk.Key(),
		Columns:  newCols,
		Values:   vs,
	}
	return table.ChunkFromBuffer(buffer)
}
