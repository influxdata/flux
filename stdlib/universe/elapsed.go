package universe

import (
	"fmt"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
)

const ElapsedKind = "elapsed"

type ElapsedOpSpec struct {
	Unit       flux.Duration `json:"unit"`
	TimeColumn string        `json:"timeColumn"`
	ColumnName string        `json:"columnName"`
}

func init() {
	elapsedSignature := flux.FunctionSignature(
		map[string]semantic.PolyType{
			"unit":       semantic.Duration,
			"timeColumn": semantic.String,
			"columnName": semantic.String,
		},
		nil,
	)

	flux.RegisterPackageValue("universe", ElapsedKind, flux.FunctionValue(ElapsedKind, createElapsedOpSpec, elapsedSignature))
	flux.RegisterOpSpec(ElapsedKind, newElapsedOp)
	plan.RegisterProcedureSpec(ElapsedKind, newElapsedProcedure, ElapsedKind)
	execute.RegisterTransformation(ElapsedKind, createElapsedTransformation)
}

func createElapsedOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(ElapsedOpSpec)

	if unit, ok, err := args.GetDuration("unit"); err != nil {
		return nil, err
	} else if ok {
		spec.Unit = unit
	} else {
		spec.Unit = flux.Duration(time.Second)
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
		spec.ColumnName = "elapsed"
	}

	return spec, nil
}

func newElapsedOp() flux.OperationSpec {
	return new(ElapsedOpSpec)
}

func (s *ElapsedOpSpec) Kind() flux.OperationKind {
	return ElapsedKind
}

type ElapsedProcedureSpec struct {
	plan.DefaultCost
	Unit       flux.Duration `json:"unit"`
	TimeColumn string        `json:"timeColumn"`
	ColumnName string        `json:"columnName"`
}

func newElapsedProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*ElapsedOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &ElapsedProcedureSpec{
		Unit:       spec.Unit,
		TimeColumn: spec.TimeColumn,
		ColumnName: spec.ColumnName,
	}, nil
}

func (s *ElapsedProcedureSpec) Kind() plan.ProcedureKind {
	return ElapsedKind
}

func (s *ElapsedProcedureSpec) Copy() plan.ProcedureSpec {
	return &ElapsedProcedureSpec{
		Unit:       s.Unit,
		TimeColumn: s.TimeColumn,
		ColumnName: s.ColumnName,
	}
}

func createElapsedTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*ElapsedProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewElapsedTransformation(d, cache, s)
	return t, d, nil
}

type elapsedTransformation struct {
	d     execute.Dataset
	cache execute.TableBuilderCache

	unit       float64
	timeColumn string
	columnName string
}

func NewElapsedTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *ElapsedProcedureSpec) *elapsedTransformation {
	return &elapsedTransformation{
		d:     d,
		cache: cache,

		unit:       float64(spec.Unit),
		timeColumn: spec.TimeColumn,
		columnName: spec.ColumnName,
	}
}

func (t *elapsedTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *elapsedTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}

func (t *elapsedTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}

func (t *elapsedTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}

func (t *elapsedTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	builder, created := t.cache.TableBuilder(tbl.Key())
	if !created {
		return fmt.Errorf("found duplicate table with key: %v", tbl.Key())
	}
	cols := tbl.Cols()
	numCol := 0

	err := execute.AddTableCols(tbl, builder)
	if err != nil {
		return err
	}

	timeIdx := execute.ColIdx(t.timeColumn, cols)
	if timeIdx < 0 {
		return fmt.Errorf("column %q does not exist", t.timeColumn)
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

	prevTime := float64(0)

	colMap := execute.ColMap([]int{0}, builder, tbl.Cols())

	return tbl.Do(func(cr flux.ColReader) error {
		l := cr.Len()
		if l != 0 {
			for j, c := range cols {

				if c.Type == flux.TTime && c.Label == t.timeColumn {
					ts := cr.Times(j)
					prevTime = float64(execute.Time(ts.Value(0)))
					currTime := float64(0)
					for i := 1; i < l; i++ {

						if err := execute.AppendMappedRecordExplicit(i, cr, builder, colMap); err != nil {
							return err
						}

						pTime := execute.Time(ts.Value(i))
						currTime = float64(pTime)
						if err := builder.AppendInt(numCol, int64((currTime-prevTime)/t.unit)); err != nil {
							return err
						}

						prevTime = currTime
					}
				}
			}
		}

		return nil
	})
}
