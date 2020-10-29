package events

import (
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

	stopIdx := execute.ColIdx(t.stopColumn, cols)
	if stopIdx < 0 && !t.isStop {
		return errors.Newf(codes.FailedPrecondition, "column %q does not exist", t.stopColumn)
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

	return tbl.Do(func(cr flux.ColReader) error {
		l := cr.Len()
		if l != 0 {
			// If no stop timestamp is provided, get last value in stopColumn
			if !t.isStop {
				for j, c := range cols {
					if c.Type == flux.TTime && c.Label == t.stopColumn {
						stopColumn := cr.Times(j)
						t.stop = execute.Time(stopColumn.Value(l - 1))
					}
				}
			}
			for j, c := range cols {
				if c.Type == flux.TTime && c.Label == t.timeColumn {
					ts := cr.Times(j)
					for i := 0; i < l-1; i++ {
						if err := execute.AppendMappedRecordExplicit(i, cr, builder, colMap); err != nil {
							return err
						}

						// Calculate difference between this record and next
						cTime := execute.Time(ts.Value(i))
						nTime := execute.Time(ts.Value(i + 1))
						currentTime := float64(cTime)
						nextTime := float64(nTime)
						if err := builder.AppendInt(numCol, int64((nextTime-currentTime)/t.unit)); err != nil {
							return err
						}
					}

					if err := execute.AppendMappedRecordExplicit(l-1, cr, builder, colMap); err != nil {
						return err
					}

					// Calculate difference between last record and stop
					cTime := execute.Time(ts.Value(l - 1))
					sTime := t.stop
					currentTime := float64(cTime)
					stopTime := float64(sTime)
					if err := builder.AppendInt(numCol, int64((stopTime-currentTime)/t.unit)); err != nil {
						return err
					}
				}
			}
		}

		return nil
	})
}
