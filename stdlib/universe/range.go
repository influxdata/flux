package universe

import (
	"fmt"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
	"github.com/pkg/errors"
)

const RangeKind = "range"

type RangeOpSpec struct {
	Start       flux.Time `json:"start"`
	Stop        flux.Time `json:"stop"`
	TimeColumn  string    `json:"timeColumn"`
	StartColumn string    `json:"startColumn"`
	StopColumn  string    `json:"stopColumn"`
}

func init() {
	rangeSignature := flux.FunctionSignature(
		map[string]semantic.PolyType{
			// TODO(nathanielc): Add polymorphic constants and a type class for time/durations
			"start":       semantic.Tvar(1),
			"stop":        semantic.Tvar(2),
			"timeColumn":  semantic.String,
			"startColumn": semantic.String,
			"stopColumn":  semantic.String,
		},
		[]string{"start"},
	)

	flux.RegisterPackageValue("universe", RangeKind, flux.FunctionValue(RangeKind, createRangeOpSpec, rangeSignature))
	flux.RegisterOpSpec(RangeKind, newRangeOp)
	plan.RegisterProcedureSpec(RangeKind, newRangeProcedure, RangeKind)
	// TODO register a range transformation. Currently range is only supported if it is pushed down into a select procedure.
	execute.RegisterTransformation(RangeKind, createRangeTransformation)
}

func createRangeOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}
	start, err := args.GetRequiredTime("start")
	if err != nil {
		return nil, err
	}
	spec := &RangeOpSpec{
		Start: start,
	}

	if stop, ok, err := args.GetTime("stop"); err != nil {
		return nil, err
	} else if ok {
		spec.Stop = stop
	} else {
		// Make stop time implicit "now"
		spec.Stop = flux.Now
	}

	if col, ok, err := args.GetString("timeColumn"); err != nil {
		return nil, err
	} else if ok {
		spec.TimeColumn = col
	} else {
		spec.TimeColumn = execute.DefaultTimeColLabel
	}

	if label, ok, err := args.GetString("startColumn"); err != nil {
		return nil, err
	} else if ok {
		spec.StartColumn = label
	} else {
		spec.StartColumn = execute.DefaultStartColLabel
	}

	if label, ok, err := args.GetString("stopColumn"); err != nil {
		return nil, err
	} else if ok {
		spec.StopColumn = label
	} else {
		spec.StopColumn = execute.DefaultStopColLabel
	}

	return spec, nil
}

func newRangeOp() flux.OperationSpec {
	return new(RangeOpSpec)
}

func (s *RangeOpSpec) Kind() flux.OperationKind {
	return RangeKind
}

type RangeProcedureSpec struct {
	plan.DefaultCost
	Bounds      flux.Bounds
	TimeColumn  string
	StartColumn string
	StopColumn  string
}

// TimeBounds implements plan.BoundsAwareProcedureSpec
func (s *RangeProcedureSpec) TimeBounds(predecessorBounds *plan.Bounds) *plan.Bounds {
	bounds := &plan.Bounds{
		Start: values.ConvertTime(s.Bounds.Start.Time(s.Bounds.Now)),
		Stop:  values.ConvertTime(s.Bounds.Stop.Time(s.Bounds.Now)),
	}
	if predecessorBounds != nil {
		bounds = bounds.Intersect(predecessorBounds)
	}
	return bounds
}

func newRangeProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*RangeOpSpec)

	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	if spec.TimeColumn == "" {
		spec.TimeColumn = execute.DefaultTimeColLabel
	}

	bounds := flux.Bounds{
		Start: spec.Start,
		Stop:  spec.Stop,
		Now:   pa.Now(),
	}

	if bounds.Start.IsZero() {
		return nil, errors.New(`must specify the start time in 'range'`)
	} else if bounds.Stop.IsZero() {
		return nil, errors.New(`must specify the stop time in 'range'`)
	} else if bounds.IsEmpty() {
		return nil, errors.New("cannot query an empty range")
	}

	return &RangeProcedureSpec{
		Bounds:      bounds,
		TimeColumn:  spec.TimeColumn,
		StartColumn: spec.StartColumn,
		StopColumn:  spec.StopColumn,
	}, nil
}

func (s *RangeProcedureSpec) Kind() plan.ProcedureKind {
	return RangeKind
}
func (s *RangeProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(RangeProcedureSpec)
	ns.Bounds = s.Bounds
	ns.TimeColumn = s.TimeColumn
	ns.StartColumn = s.StartColumn
	ns.StopColumn = s.StopColumn
	return ns
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *RangeProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

func createRangeTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*RangeProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)

	bounds := a.StreamContext().Bounds()

	t, err := NewRangeTransformation(d, cache, s, *bounds)
	if err != nil {
		return nil, nil, err
	}
	return t, d, nil
}

type rangeTransformation struct {
	d        execute.Dataset
	cache    execute.TableBuilderCache
	bounds   execute.Bounds
	timeCol  string
	startCol string
	stopCol  string
}

func NewRangeTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *RangeProcedureSpec, absolute execute.Bounds) (*rangeTransformation, error) {
	return &rangeTransformation{
		d:        d,
		cache:    cache,
		bounds:   absolute,
		timeCol:  spec.TimeColumn,
		startCol: spec.StartColumn,
		stopCol:  spec.StopColumn,
	}, nil
}

func (t *rangeTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *rangeTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	timeIdx := execute.ColIdx(t.timeCol, tbl.Cols())
	if timeIdx < 0 {
		return fmt.Errorf("range error: supplied time column %s doesn't exist", t.timeCol)
	}

	if tbl.Cols()[timeIdx].Type != flux.TTime {
		return fmt.Errorf("range error: provided time column %s is not of type time", t.timeCol)
	}

	// Determine index of start and stop columns in table
	startColIdx := execute.ColIdx(t.startCol, tbl.Cols())
	if startColIdx >= 0 && tbl.Cols()[startColIdx].Type != flux.TTime {
		return fmt.Errorf("range error: provided start column %s is not of type time", t.timeCol)
	}

	stopColIdx := execute.ColIdx(t.stopCol, tbl.Cols())
	if stopColIdx >= 0 && tbl.Cols()[stopColIdx].Type != flux.TTime {
		return fmt.Errorf("range error: provided stop column %s is not of type time", t.timeCol)
	}

	// Determine index of start and stop columns in group key
	startKeyColIdx := execute.ColIdx(t.startCol, tbl.Key().Cols())
	stopKeyColIdx := execute.ColIdx(t.stopCol, tbl.Key().Cols())

	// Compute the group key for the output table
	outKey := t.createRangeGroupKey(tbl.Key(), startKeyColIdx, stopKeyColIdx)

	builder, created := t.cache.TableBuilder(outKey)
	if !created {
		return fmt.Errorf("range found duplicate table with key: %v", tbl.Key())
	}

	// If the start and/or stop columns don't exist,
	// They must be added to the table
	if startColIdx < 0 {
		c := flux.ColMeta{
			Label: t.startCol,
			Type:  flux.TTime,
		}
		if _, err := builder.AddCol(c); err != nil {
			return err
		}
	}

	if stopColIdx < 0 {
		c := flux.ColMeta{
			Label: t.stopCol,
			Type:  flux.TTime,
		}
		if _, err := builder.AddCol(c); err != nil {
			return err
		}
	}

	if err := execute.AddTableCols(tbl, builder); err != nil {
		return err
	}

	// map from builder column to input table column
	tCols := len(tbl.Cols())
	bCols := builder.NCols()
	offset := bCols - tCols

	colMap := make([]int, bCols)
	for i := 0; i < bCols; i++ {
		if i < offset {
			colMap[i] = -1
		} else {
			colMap[i] = i - offset
		}
	}

	startTime := outKey.Value(execute.ColIdx(t.startCol, outKey.Cols()))
	stopTime := outKey.Value(execute.ColIdx(t.stopCol, outKey.Cols()))
	return tbl.Do(func(cr flux.ColReader) error {
		l := cr.Len()
		for i := 0; i < l; i++ {
			ts := cr.Times(timeIdx)
			if ts.IsNull(i) {
				continue
			}
			tVal := values.Time(ts.Value(i))
			if !t.bounds.Contains(tVal) {
				continue
			}
			for j, c := range builder.Cols() {
				switch c.Label {
				case t.startCol:
					if err := builder.AppendValue(j, startTime); err != nil {
						return err
					}
				case t.stopCol:
					if err := builder.AppendValue(j, stopTime); err != nil {
						return err
					}
				default:
					if err := builder.AppendValue(j, execute.ValueForRow(cr, i, colMap[j])); err != nil {
						return err
					}
				}
			}
		}
		return nil
	})
}

func (t *rangeTransformation) createRangeGroupKey(inKey flux.GroupKey, startKeyColIdx, stopKeyColIdx int) flux.GroupKey {
	var outKeyCols []flux.ColMeta
	var outKeyValues []values.Value
	if startKeyColIdx >= 0 && stopKeyColIdx >= 0 {
		// Both start and stop columns exist and they are in the group key.
		// If start/stop intersects with the range bounds, use the intersection
		// as start/stop in the group key, otherwise use the range bounds.
		outKeyCols = inKey.Cols()
		outKeyValues = make([]values.Value, len(inKey.Values()))
		copy(outKeyValues, inKey.Values())

		useBounds := t.bounds
		if !inKey.Value(startKeyColIdx).IsNull() && !inKey.Value(stopKeyColIdx).IsNull() {
			inBounds := execute.Bounds{
				Start: inKey.ValueTime(startKeyColIdx),
				Stop:  inKey.ValueTime(stopKeyColIdx),
			}
			if t.bounds.Overlaps(inBounds) {
				useBounds = t.bounds.Intersect(inBounds)
			}
		}
		outKeyValues[startKeyColIdx] = values.NewTime(useBounds.Start)
		outKeyValues[stopKeyColIdx] = values.NewTime(useBounds.Stop)
	} else {
		// One or both of the start/stop columns is missing or is not in the group key.
		// Put them both in the group key and set start/stop to the range bounds.
		outKeyCols = make([]flux.ColMeta, 0, len(inKey.Cols())+2)
		outKeyValues = make([]values.Value, 0, len(inKey.Cols())+2)

		if startKeyColIdx < 0 {
			outKeyCols = append(outKeyCols, flux.ColMeta{Label: t.startCol, Type: flux.TTime})
			outKeyValues = append(outKeyValues, values.New(t.bounds.Start))
		}
		if stopKeyColIdx < 0 {
			outKeyCols = append(outKeyCols, flux.ColMeta{Label: t.stopCol, Type: flux.TTime})
			outKeyValues = append(outKeyValues, values.New(t.bounds.Stop))
		}

		for j := 0; j < len(inKey.Cols()); j++ {
			outKeyCols = append(outKeyCols, inKey.Cols()[j])
			outKeyValues = append(outKeyValues, inKey.Value(j))
		}
	}

	return execute.NewGroupKey(outKeyCols, outKeyValues)
}

func (t *rangeTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *rangeTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *rangeTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
