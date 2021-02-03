package universe

import (
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interval"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/values"
)

const WindowIntervalKind = "windowInterval"

type WindowIntervalOpSpec struct {
	Every       flux.Duration `json:"every"`
	Period      flux.Duration `json:"period"`
	Offset      flux.Duration `json:"offset"`
	TimeColumn  string        `json:"timeColumn"`
	StopColumn  string        `json:"stopColumn"`
	StartColumn string        `json:"startColumn"`
	CreateEmpty bool          `json:"createEmpty"`
}

func init() {
	execute.RegisterTransformation(WindowIntervalKind, createWindowIntervalTransformation)
}

func (s *WindowIntervalOpSpec) Kind() flux.OperationKind {
	return WindowIntervalKind
}

type WindowIntervalProcedureSpec struct {
	plan.DefaultCost
	Window plan.WindowSpec
	TimeColumn,
	StartColumn,
	StopColumn string
	CreateEmpty bool
}

func (s *WindowIntervalProcedureSpec) Kind() plan.ProcedureKind {
	return WindowIntervalKind
}
func (s *WindowIntervalProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(WindowIntervalProcedureSpec)
	ns.Window = s.Window
	ns.TimeColumn = s.TimeColumn
	ns.StartColumn = s.StartColumn
	ns.StopColumn = s.StopColumn
	ns.CreateEmpty = s.CreateEmpty
	return ns
}

func createWindowIntervalTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*WindowIntervalProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)

	bounds := a.StreamContext().Bounds()
	if bounds == nil {
		const docURL = "https://v2.docs.influxdata.com/v2.0/reference/flux/stdlib/built-in/transformations/window/#nil-bounds-passed-to-window"
		return nil, nil, errors.New(codes.Invalid, "nil bounds passed to window; use range to set the window range").
			WithDocURL(docURL)
	}

	newBounds := interval.NewBounds(bounds.Start, bounds.Stop)

	w, err := interval.NewWindow(
		s.Window.Every,
		s.Window.Period,
		s.Window.Offset,
	)

	if err != nil {
		return nil, nil, err
	}
	t := NewIntervalFixedWindowTransformation(
		d,
		cache,
		newBounds,
		w,
		s.TimeColumn,
		s.StartColumn,
		s.StopColumn,
		s.CreateEmpty,
	)
	return t, d, nil
}

type intervalFixedWindowTransformation struct {
	execute.ExecutionNode
	d         execute.Dataset
	cache     execute.TableBuilderCache
	w         interval.Window
	bounds    interval.Bounds
	allBounds []interval.Bounds

	timeCol,
	startCol,
	stopCol string
	createEmpty bool
}

func NewIntervalFixedWindowTransformation(
	d execute.Dataset,
	cache execute.TableBuilderCache,
	bounds interval.Bounds,
	w interval.Window,
	timeCol,
	startCol,
	stopCol string,
	createEmpty bool,
) execute.Transformation {
	t := &intervalFixedWindowTransformation{
		d:           d,
		cache:       cache,
		w:           w,
		bounds:      bounds,
		timeCol:     timeCol,
		startCol:    startCol,
		stopCol:     stopCol,
		createEmpty: createEmpty,
	}

	if createEmpty {
		t.generateWindowsWithinBounds()
	}

	return t
}

func (t *intervalFixedWindowTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) (err error) {
	panic("not implemented")
}

func (t *intervalFixedWindowTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	timeIdx := execute.ColIdx(t.timeCol, tbl.Cols())
	if timeIdx < 0 {
		const docURL = "https://v2.docs.influxdata.com/v2.0/reference/flux/stdlib/built-in/transformations/window/#missing-time-column"
		return errors.Newf(codes.FailedPrecondition, "missing time column %q", t.timeCol).
			WithDocURL(docURL)
	}

	newCols := make([]flux.ColMeta, 0, len(tbl.Cols())+2)
	keyCols := make([]flux.ColMeta, 0, len(tbl.Cols())+2)
	keyColMap := make([]int, 0, len(tbl.Cols())+2)
	startColIdx := -1
	stopColIdx := -1
	for j, c := range tbl.Cols() {
		keyIdx := execute.ColIdx(c.Label, tbl.Key().Cols())
		keyed := keyIdx >= 0
		if c.Label == t.startCol {
			startColIdx = j
			keyed = true
		}
		if c.Label == t.stopCol {
			stopColIdx = j
			keyed = true
		}
		newCols = append(newCols, c)
		if keyed {
			keyCols = append(keyCols, c)
			keyColMap = append(keyColMap, keyIdx)
		}
	}
	if startColIdx == -1 {
		startColIdx = len(newCols)
		c := flux.ColMeta{
			Label: t.startCol,
			Type:  flux.TTime,
		}
		newCols = append(newCols, c)
		keyCols = append(keyCols, c)
		keyColMap = append(keyColMap, len(keyColMap))
	}
	if stopColIdx == -1 {
		stopColIdx = len(newCols)
		c := flux.ColMeta{
			Label: t.stopCol,
			Type:  flux.TTime,
		}
		newCols = append(newCols, c)
		keyCols = append(keyCols, c)
		keyColMap = append(keyColMap, len(keyColMap))
	}

	// Abort processing if no data will match bounds
	if t.bounds.IsEmpty() {
		return nil
	}

	for _, bnds := range t.allBounds {
		key := t.newWindowGroupKey(tbl, keyCols, bnds, keyColMap)
		builder, created := t.cache.TableBuilder(key)
		if created {
			for _, c := range newCols {
				_, err := builder.AddCol(c)
				if err != nil {
					return err
				}
			}
		}
	}

	return tbl.Do(func(cr flux.ColReader) error {
		l := cr.Len()
		for i := 0; i < l; i++ {
			tm := values.Time(cr.Times(timeIdx).Value(i))
			bounds := t.getWindowBounds(tm)

			for _, bnds := range bounds {
				key := t.newWindowGroupKey(tbl, keyCols, bnds, keyColMap)
				builder, created := t.cache.TableBuilder(key)
				if created {
					for _, c := range newCols {
						_, err := builder.AddCol(c)
						if err != nil {
							return err
						}
					}
				}

				for j, c := range builder.Cols() {
					switch c.Label {
					case t.startCol:
						if err := builder.AppendTime(startColIdx, bnds.Start()); err != nil {
							return err
						}
					case t.stopCol:
						if err := builder.AppendTime(stopColIdx, bnds.Stop()); err != nil {
							return err
						}
					default:
						if err := builder.AppendValue(j, execute.ValueForRow(cr, i, j)); err != nil {
							return err
						}
					}
				}
			}
		}
		return nil
	})
}

func (t *intervalFixedWindowTransformation) newWindowGroupKey(tbl flux.Table, keyCols []flux.ColMeta, bnds interval.Bounds, keyColMap []int) flux.GroupKey {
	cols := make([]flux.ColMeta, len(keyCols))
	vs := make([]values.Value, len(keyCols))
	for j, c := range keyCols {
		cols[j] = c
		switch c.Label {
		case t.startCol:
			vs[j] = values.NewTime(bnds.Start())
		case t.stopCol:
			vs[j] = values.NewTime(bnds.Stop())
		default:
			vs[j] = tbl.Key().Value(keyColMap[j])
		}
	}
	return execute.NewGroupKey(cols, vs)
}

func (t *intervalFixedWindowTransformation) clipBounds(bs []interval.Bounds) {
	for i := range bs {
		bs[i] = t.bounds.Intersect(bs[i])
	}
}

func (t *intervalFixedWindowTransformation) getWindowBounds(tm execute.Time) []interval.Bounds {
	if t.w.Every() == infinityVar.Duration() {
		return []interval.Bounds{t.bounds}
	}
	bs := t.w.GetOverlappingBounds(tm, tm+1)
	t.clipBounds(bs)
	return bs
}

func (t *intervalFixedWindowTransformation) generateWindowsWithinBounds() {
	if t.w.Every() == infinityVar.Duration() {
		bounds := interval.NewBounds(interval.MinTime, interval.MaxTime)
		t.allBounds = []interval.Bounds{bounds}
		return
	}
	bs := t.w.GetOverlappingBounds(t.bounds.Start(), t.bounds.Stop())
	t.clipBounds(bs)
	t.allBounds = bs
}

func (t *intervalFixedWindowTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *intervalFixedWindowTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *intervalFixedWindowTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
