package universe

import (
	"fmt"
	"log"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
	"github.com/pkg/errors"
)

const StateTrackingKind = "stateTracking"

type StateTrackingOpSpec struct {
	Fn             *semantic.FunctionExpression `json:"fn"`
	CountColumn    string                       `json:"countColumn"`
	DurationColumn string                       `json:"durationColumn"`
	DurationUnit   flux.Duration                `json:"durationUnit"`
	TimeColumn     string                       `json:"timeColumn"`
}

func init() {
	stateTrackingSignature := flux.FunctionSignature(
		map[string]semantic.PolyType{
			"fn": semantic.NewFunctionPolyType(semantic.FunctionPolySignature{
				Parameters: map[string]semantic.PolyType{
					"r": semantic.Tvar(1),
				},
				Required: semantic.LabelSet{"r"},
				Return:   semantic.Bool,
			}),
			"countColumn":    semantic.String,
			"durationColumn": semantic.String,
			"durationUnit":   semantic.Duration,
			"timeColumn":     semantic.String,
		},
		[]string{"fn"},
	)

	flux.RegisterPackageValue("universe", StateTrackingKind, flux.FunctionValue(StateTrackingKind, createStateTrackingOpSpec, stateTrackingSignature))
	flux.RegisterOpSpec(StateTrackingKind, newStateTrackingOp)
	plan.RegisterProcedureSpec(StateTrackingKind, newStateTrackingProcedure, StateTrackingKind)
	execute.RegisterTransformation(StateTrackingKind, createStateTrackingTransformation)
}

func createStateTrackingOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	f, err := args.GetRequiredFunction("fn")
	if err != nil {
		return nil, err
	}

	fn, err := interpreter.ResolveFunction(f)
	if err != nil {
		return nil, err
	}

	spec := &StateTrackingOpSpec{
		Fn:           fn,
		DurationUnit: flux.Duration(time.Second),
	}

	if label, ok, err := args.GetString("countColumn"); err != nil {
		return nil, err
	} else if ok {
		spec.CountColumn = label
	}
	if label, ok, err := args.GetString("durationColumn"); err != nil {
		return nil, err
	} else if ok {
		spec.DurationColumn = label
	}
	if unit, ok, err := args.GetDuration("durationUnit"); err != nil {
		return nil, err
	} else if ok {
		spec.DurationUnit = unit
	}
	if label, ok, err := args.GetString("timeColumn"); err != nil {
		return nil, err
	} else if ok {
		spec.TimeColumn = label
	} else {
		spec.TimeColumn = execute.DefaultTimeColLabel
	}

	if spec.DurationColumn != "" && spec.DurationUnit <= 0 {
		return nil, errors.New("state tracking duration unit must be greater than zero")
	}
	return spec, nil
}

func newStateTrackingOp() flux.OperationSpec {
	return new(StateTrackingOpSpec)
}

func (s *StateTrackingOpSpec) Kind() flux.OperationKind {
	return StateTrackingKind
}

type StateTrackingProcedureSpec struct {
	plan.DefaultCost
	Fn *semantic.FunctionExpression
	CountColumn,
	DurationColumn string
	DurationUnit flux.Duration
	TimeCol      string
}

func newStateTrackingProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*StateTrackingOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &StateTrackingProcedureSpec{
		Fn:             spec.Fn,
		CountColumn:    spec.CountColumn,
		DurationColumn: spec.DurationColumn,
		DurationUnit:   spec.DurationUnit,
		TimeCol:        spec.TimeColumn,
	}, nil
}

func (s *StateTrackingProcedureSpec) Kind() plan.ProcedureKind {
	return StateTrackingKind
}
func (s *StateTrackingProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(StateTrackingProcedureSpec)
	*ns = *s

	ns.Fn = s.Fn.Copy().(*semantic.FunctionExpression)

	return ns
}

func createStateTrackingTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*StateTrackingProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t, err := NewStateTrackingTransformation(d, cache, s)
	if err != nil {
		return nil, nil, err
	}
	return t, d, nil
}

type stateTrackingTransformation struct {
	d     execute.Dataset
	cache execute.TableBuilderCache

	fn *execute.RowPredicateFn

	timeCol,
	countColumn,
	durationColumn string

	durationUnit int64
}

func NewStateTrackingTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *StateTrackingProcedureSpec) (*stateTrackingTransformation, error) {
	fn, err := execute.NewRowPredicateFn(spec.Fn)
	if err != nil {
		return nil, err
	}
	return &stateTrackingTransformation{
		d:              d,
		cache:          cache,
		fn:             fn,
		countColumn:    spec.CountColumn,
		durationColumn: spec.DurationColumn,
		durationUnit:   int64(spec.DurationUnit),
		timeCol:        spec.TimeCol,
	}, nil
}

func (t *stateTrackingTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *stateTrackingTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	builder, created := t.cache.TableBuilder(tbl.Key())
	if !created {
		return fmt.Errorf("found duplicate table with key: %v", tbl.Key())
	}
	err := execute.AddTableCols(tbl, builder)
	if err != nil {
		return err
	}

	// Prepare the functions for the column types.
	cols := tbl.Cols()
	err = t.fn.Prepare(cols)
	if err != nil {
		// TODO(nathanielc): Should we not fail the query for failed compilation?
		return err
	}

	var countCol, durationCol = -1, -1

	// Add new value columns
	if t.countColumn != "" {
		countCol, err = builder.AddCol(flux.ColMeta{
			Label: t.countColumn,
			Type:  flux.TInt,
		})
		if err != nil {
			return err
		}
	}
	if t.durationColumn != "" {
		durationCol, err = builder.AddCol(flux.ColMeta{
			Label: t.durationColumn,
			Type:  flux.TInt,
		})
		if err != nil {
			return err
		}
	}

	var (
		startTime       values.Time
		prevTime        values.Time
		count           int64
		duration        int64
		countInState    bool
		durationInState bool
	)

	timeIdx := execute.ColIdx(t.timeCol, tbl.Cols())
	if timeIdx < 0 {
		return fmt.Errorf("no column %q exists", t.timeCol)
	}
	// Append modified rows
	return tbl.Do(func(cr flux.ColReader) error {
		l := cr.Len()
		for i := 0; i < l; i++ {
			match, err := t.fn.Eval(i, cr)
			if err != nil {
				log.Printf("failed to evaluate state tracking expression: %v", err)
				continue
			}

			// Duration
			if durationCol > 0 {
				if ts := cr.Times(timeIdx); ts.IsNull(i) {
					return errors.New("got a null timestamp")
				}

				tValue := values.Time(cr.Times(timeIdx).Value(i))

				if prevTime > tValue {
					return errors.New("got an out-of-order timestamp")
				}
				prevTime = tValue

				if !match {
					duration = -1
					durationInState = false
				} else {
					if !durationInState {
						startTime = tValue
						duration = 0
						durationInState = true
					}

					if t.durationUnit > 0 {
						tm := tValue
						duration = int64(tm-startTime) / t.durationUnit
					}
				}
			}

			// Count
			if countCol > 0 {
				if !match {
					count = -1
					countInState = false
				} else {
					if !countInState {
						count = 0
						countInState = true
					}
					count++
				}
			}

			colMap := make([]int, len(cr.Cols()))
			colMap = execute.ColMap(colMap, builder, cr)
			err = execute.AppendMappedRecordExplicit(i, cr, builder, colMap)
			if err != nil {
				return err
			}
			if countCol > 0 {
				err = builder.AppendInt(countCol, count)
				if err != nil {
					return err
				}
			}
			if durationCol > 0 {
				err = builder.AppendInt(durationCol, duration)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (t *stateTrackingTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *stateTrackingTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *stateTrackingTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
