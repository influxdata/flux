package events

import (
	"fmt"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
)

const pkgPath = "contrib/tomhollingworth/events"

const DurationKind = "duration"

type DurationOpSpec struct {
	Unit       flux.Duration `json:"unit"`
	TimeColumn string        `json:"timeColumn"`
	ColumnName string        `json:"columnName"`
	StopColumn string        `json:"stopColumn"`
	Stop       flux.Time     `json:"stop"`
	IsStop     bool
}

func init() {
	durationSignature := runtime.MustLookupBuiltinType(pkgPath, DurationKind)
	runtime.RegisterPackageValue(pkgPath, DurationKind, flux.MustValue(flux.FunctionValue(DurationKind, createDurationOpSpec, durationSignature)))
	flux.RegisterOpSpec(DurationKind, newDurationOp)
	plan.RegisterProcedureSpec(DurationKind, newDurationProcedure, DurationKind)
	execute.RegisterTransformation(DurationKind, createDurationTransformation)
}

func createDurationOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(DurationOpSpec)

	if unit, ok, err := args.GetDuration("unit"); err != nil {
		return nil, err
	} else if ok {
		spec.Unit = unit
	} else {
		spec.Unit = flux.ConvertDuration(time.Second)
	}

	if timeCol, ok, err := args.GetString("timeColumn"); err != nil {
		return nil, err
	} else if ok {
		spec.TimeColumn = timeCol
	} else {
		spec.TimeColumn = execute.DefaultTimeColLabel
	}

	if name, ok, err := args.GetString("columnName"); err != nil {
		return nil, err
	} else if ok {
		spec.ColumnName = name
	} else {
		spec.ColumnName = "duration"
	}

	if stopCol, ok, err := args.GetString("stopColumn"); err != nil {
		return nil, err
	} else if ok {
		spec.StopColumn = stopCol
	} else {
		spec.StopColumn = execute.DefaultStopColLabel
	}

	spec.IsStop = false
	if stop, ok, err := args.GetTime("stop"); err != nil {
		return nil, err
	} else if ok {
		spec.IsStop = true
		spec.Stop = stop
	} else {
		spec.Stop = flux.Now
	}

	return spec, nil
}

func newDurationOp() flux.OperationSpec {
	return new(DurationOpSpec)
}

func (s *DurationOpSpec) Kind() flux.OperationKind {
	return DurationKind
}

type DurationProcedureSpec struct {
	plan.DefaultCost
	Unit       flux.Duration `json:"unit"`
	TimeColumn string        `json:"timeColumn"`
	ColumnName string        `json:"columnName"`
	StopColumn string        `json:"stopColumn"`
	Stop       flux.Time     `json:"stop"`
	IsStop     bool
}

func newDurationProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*DurationOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	return &DurationProcedureSpec{
		Unit:       spec.Unit,
		TimeColumn: spec.TimeColumn,
		ColumnName: spec.ColumnName,
		StopColumn: spec.StopColumn,
		Stop:       spec.Stop,
		IsStop:     spec.IsStop,
	}, nil
}

func (s *DurationProcedureSpec) Kind() plan.ProcedureKind {
	return DurationKind
}

func (s *DurationProcedureSpec) Copy() plan.ProcedureSpec {
	return &DurationProcedureSpec{
		Unit:       s.Unit,
		TimeColumn: s.TimeColumn,
		ColumnName: s.ColumnName,
		StopColumn: s.StopColumn,
		Stop:       s.Stop,
		IsStop:     s.IsStop,
	}
}

func createDurationTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*DurationProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewDurationTransformation(d, cache, s)
	return t, d, nil
}

type durationTransformation struct {
	execute.ExecutionNode
	d     execute.Dataset
	cache execute.TableBuilderCache

	unit       float64
	timeColumn string
	columnName string
	stopColumn string
	stop       values.Time
	isStop     bool
}

func NewDurationTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *DurationProcedureSpec) *durationTransformation {
	return &durationTransformation{
		d:     d,
		cache: cache,

		unit:       float64(values.Duration(spec.Unit).Duration()),
		timeColumn: spec.TimeColumn,
		columnName: spec.ColumnName,
		stopColumn: spec.StopColumn,
		stop:       values.ConvertTime(spec.Stop.Absolute),
		isStop:     spec.IsStop,
	}
}

func (t *durationTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *durationTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}

func (t *durationTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}

func (t *durationTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}

func (t *durationTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	builder, created := t.cache.TableBuilder(tbl.Key())
	if !created {
		return errors.Newf(codes.FailedPrecondition, "found duplicate table with key: %v", tbl.Key())
	}
	cols := tbl.Cols()
	numCol := 0

	err := execute.AddTableCols(tbl, builder)
	if err != nil {
		return err
	}

	timeIdx := execute.ColIdx(t.timeColumn, cols)
	if timeIdx < 0 {
		return errors.Newf(codes.FailedPrecondition, "column %q does not exist", t.timeColumn)
	}

	valueIdx := execute.ColIdx("_value", cols)
	if valueIdx < 0 {
		return errors.Newf(codes.FailedPrecondition, "column %q does not exist", "_value")
	}

	var stopIdx int
	if !t.isStop {
		stopIdx = execute.ColIdx(t.stopColumn, cols)
		if stopIdx < 0 {
			return errors.Newf(codes.FailedPrecondition, "column %q does not exist", t.stopColumn)
		} else if c := tbl.Cols()[stopIdx]; c.Type != flux.TTime {
			return errors.Newf(codes.FailedPrecondition, "stop column %q must be of type %s, got %s", c.Label, flux.TTime, c.Type)
		}
	}

	timeCol := cols[timeIdx]
	if timeCol.Type == flux.TTime {
		if numCol, err = builder.AddCol(flux.ColMeta{
			Label: t.columnName,
			Type:  flux.TInt,
		}); err != nil {
			return err
		}
	}

	colMap := execute.ColMap([]int{0}, builder, tbl.Cols())

	var (
		cTime      int64
		cTimeValid bool
		sTime      int64
	)

	// If we have specified a stop value, record it here.
	if t.isStop {
		sTime = int64(t.stop)
	}

	if err := tbl.Do(func(cr flux.ColReader) error {
		l := cr.Len()
		prevIndex := 0
		lastStatusIndex := 0
		//tempIndex := -1
		//tempDuration := float64(0)

		ts := cr.Times(timeIdx)

		values := cr.Ints(valueIdx)
		if !t.isStop {
			stopTimes := cr.Times(stopIdx)
			sTime = stopTimes.Value(l - 1)
		}

		for currIndex := 1; currIndex <= l; currIndex++ {
			// Read the current time value. If we have a current time to compare
			// it to, then append the difference between them.
			//
			// This section will always append the previous row. During the first
			// invocation of this section, it is skipped.
			//fmt.Println(currIndex)
			//fmt.Println(prevIndex)
			fmt.Println(lastStatusIndex)

			prevValue := values.Value(prevIndex)
			var currValue int64
			if currIndex < l {
				currValue = values.Value(currIndex)
			} else {
				currValue = -1
			}

			if prevValue != currValue {
				// if current status is 0 and check if change from green to red or red to green and duration < 8, then omit status 0
				// if d[current_index]['status'] == '0' and ((0 < int(d[prev_index]['status']) <=15 and 15 < int(d[current_index+1]['status']) <=240) or (15 < int(d[prev_index]['status']) <=240 and 0 < int(d[current_index+1]['status']) <=15)) and (d[current_index+1]['time'] - d[current_index]['time']).total_seconds() < 8:
				if currValue == 0 {
					var nextValue int64
					if currIndex < l-1 {
						nextValue = values.Value(currIndex + 1)
					} else {
						nextValue = -1
					}
					if ((0 < prevValue && prevValue <= 15) && (15 < nextValue && nextValue <= 240)) || ((15 < prevValue && prevValue <= 240) && (0 < nextValue && nextValue <= 15)) {
						duration := float64(ts.Value(currIndex+1)) - float64(ts.Value(currIndex))
						if duration < 8000000000 {
							//prevIndex = lastStatusIndex
						} else {
							duration := float64(ts.Value(currIndex)) - float64(ts.Value(prevIndex))
							if err := builder.AppendInt(numCol, int64((duration)/t.unit)); err != nil {
								return err
							}
							if err := execute.AppendMappedRecordExplicit(prevIndex, cr, builder, colMap); err != nil {
								return err
							}
							lastStatusIndex = prevIndex
							prevIndex = currIndex
						}
					} else {
						var duration float64
						if currIndex < l {
							duration = float64(ts.Value(currIndex)) - float64(ts.Value(prevIndex))
						} else {
							duration = float64(sTime) - float64(ts.Value(prevIndex))
						}
						if err := builder.AppendInt(numCol, int64((duration)/t.unit)); err != nil {
							return err
						}
						if err := execute.AppendMappedRecordExplicit(prevIndex, cr, builder, colMap); err != nil {
							return err
						}
						lastStatusIndex = prevIndex
						prevIndex = currIndex
					}
				} else {
					var duration float64
					if currIndex < l {
						duration = float64(ts.Value(currIndex)) - float64(ts.Value(prevIndex))
					} else {
						duration = float64(sTime) - float64(ts.Value(prevIndex))
					}
					if err := builder.AppendInt(numCol, int64((duration)/t.unit)); err != nil {
						return err
					}
					if err := execute.AppendMappedRecordExplicit(prevIndex, cr, builder, colMap); err != nil {
						return err
					}
					lastStatusIndex = prevIndex
					prevIndex = currIndex
				}
			} else {
				// go next
			}

			//
			//nTime := ts.Value(i)
			//if cTimeValid {
			//	currentTime := float64(cTime)
			//	nextTime := float64(nTime)
			//	if err := builder.AppendInt(numCol, int64((nextTime-currentTime+float64(values.Value(i)))/t.unit)); err != nil {
			//		return err
			//	}
			//}
			//cTime, cTimeValid = nTime, true

			// Append the existing columns. We always append the currently
			// processed row except for the duration between the two.
			// The reason is we need to copy over these values, but
			// we don't know the duration comparison until we read the next row
			// which may exist in a separate buffer.

		}

		//if true || tempIndex != -1 {
		//if err := builder.AppendInt(numCol, int64((tempDuration)/t.unit)); err != nil {
		//	return err
		//}
		//if err := execute.AppendMappedRecordExplicit(tempIndex, cr, builder, colMap); err != nil {
		//	return err
		//}
		//}

		// If no stop timestamp is provided, get last value in stopColumn.
		// We just record this as the actual append happens outside this loop.
		// We do not know if this is the final buffer until we have already
		// finished reading the buffers so we just record this in case it is the
		// proper value.

		return nil
	}); err != nil {
		return err
	}

	// If there was at least one valid time, append the difference between
	// the last time and the stop time.
	if cTimeValid {
		currentTime := float64(cTime)
		nextTime := float64(sTime)
		if err := builder.AppendInt(numCol, int64((nextTime-currentTime)/t.unit)); err != nil {
			return err
		}
	}
	return nil
}
