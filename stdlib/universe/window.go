package universe

import (
	"context"
	"math"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
)

const WindowKind = "window"

type WindowOpSpec struct {
	Every       flux.Duration `json:"every"`
	Period      flux.Duration `json:"period"`
	Offset      flux.Duration `json:"offset"`
	TimeColumn  string        `json:"timeColumn"`
	StopColumn  string        `json:"stopColumn"`
	StartColumn string        `json:"startColumn"`
	CreateEmpty bool          `json:"createEmpty"`
}

var infinityVar = values.NewDuration(values.ConvertDurationNsecs(math.MaxInt64))

func init() {
	windowSignature := runtime.MustLookupBuiltinType("universe", "window")

	runtime.RegisterPackageValue("universe", WindowKind, flux.MustValue(flux.FunctionValue(WindowKind, createWindowOpSpec, windowSignature)))
	flux.RegisterOpSpec(WindowKind, newWindowOp)
	runtime.RegisterPackageValue("universe", "inf", infinityVar)
	plan.RegisterProcedureSpec(WindowKind, newWindowProcedure, WindowKind)
	plan.RegisterPhysicalRules(WindowTriggerPhysicalRule{})
	execute.RegisterTransformation(WindowKind, createWindowTransformation)
}

func createWindowOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(WindowOpSpec)

	every, everySet, err := func() (flux.Duration, bool, error) {
		d, ok, err := args.GetDuration("every")
		if err != nil || !ok {
			return flux.Duration{}, ok, err
		}

		if d.IsNegative() {
			return flux.Duration{}, false, errors.New(codes.Invalid, `parameter "every" must be nonnegative`)
		} else if d.IsZero() {
			return flux.Duration{}, false, errors.New(codes.Invalid, `parameter "every" must be nonzero`)
		}
		return d, true, nil
	}()
	if err != nil {
		const docURL = "https://v2.docs.influxdata.com/v2.0/reference/flux/stdlib/built-in/transformations/window/#every"
		return nil, errors.WithDocURL(err, docURL)
	} else if everySet {
		spec.Every = every
	}

	period, periodSet, err := args.GetDuration("period")
	if err != nil {
		const docURL = "https://v2.docs.influxdata.com/v2.0/reference/flux/stdlib/built-in/transformations/window/#period"
		return nil, errors.WithDocURL(err, docURL)
	}
	if periodSet {
		spec.Period = period
	}
	if offset, ok, err := args.GetDuration("offset"); err != nil {
		return nil, err
	} else if ok {
		spec.Offset = offset
	}

	if !everySet && !periodSet {
		const docURL = "https://v2.docs.influxdata.com/v2.0/reference/flux/stdlib/built-in/transformations/window/"
		return nil, errors.New(codes.Invalid, `window function requires at least one of "every" or "period" to be set`).
			WithDocURL(docURL)
	}

	if label, ok, err := args.GetString("timeColumn"); err != nil {
		return nil, err
	} else if ok {
		spec.TimeColumn = label
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
	if createEmpty, ok, err := args.GetBool("createEmpty"); err != nil {
		return nil, err
	} else if ok {
		spec.CreateEmpty = createEmpty
	} else {
		spec.CreateEmpty = false
	}

	// Apply defaults
	if !everySet {
		spec.Every = spec.Period
	}
	if !periodSet {
		spec.Period = spec.Every
	}
	return spec, nil
}

func newWindowOp() flux.OperationSpec {
	return new(WindowOpSpec)
}

func (s *WindowOpSpec) Kind() flux.OperationKind {
	return WindowKind
}

type WindowProcedureSpec struct {
	plan.DefaultCost
	Window plan.WindowSpec
	TimeColumn,
	StartColumn,
	StopColumn string
	CreateEmpty bool
}

func newWindowProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	s, ok := qs.(*WindowOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}
	p := &WindowProcedureSpec{
		Window: plan.WindowSpec{
			Every:  s.Every,
			Period: s.Period,
			Offset: s.Offset,
		},
		TimeColumn:  s.TimeColumn,
		StartColumn: s.StartColumn,
		StopColumn:  s.StopColumn,
		CreateEmpty: s.CreateEmpty,
	}
	return p, nil
}

func (s *WindowProcedureSpec) Kind() plan.ProcedureKind {
	return WindowKind
}
func (s *WindowProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(WindowProcedureSpec)
	ns.Window = s.Window
	ns.TimeColumn = s.TimeColumn
	ns.StartColumn = s.StartColumn
	ns.StopColumn = s.StopColumn
	ns.CreateEmpty = s.CreateEmpty
	return ns
}

func createWindowTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*WindowProcedureSpec)
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

	w, err := execute.NewWindow(
		s.Window.Every,
		s.Window.Period,
		s.Window.Offset,
	)
	if err != nil {
		return nil, nil, err
	}
	t := NewFixedWindowTransformation(
		d,
		cache,
		*bounds,
		w,
		s.TimeColumn,
		s.StartColumn,
		s.StopColumn,
		s.CreateEmpty,
	)
	return t, d, nil
}

type fixedWindowTransformation struct {
	d         execute.Dataset
	cache     execute.TableBuilderCache
	w         execute.Window
	bounds    execute.Bounds
	allBounds []execute.Bounds

	timeCol,
	startCol,
	stopCol string
	createEmpty bool
}

func NewFixedWindowTransformation(
	d execute.Dataset,
	cache execute.TableBuilderCache,
	bounds execute.Bounds,
	w execute.Window,
	timeCol,
	startCol,
	stopCol string,
	createEmpty bool,
) execute.Transformation {
	t := &fixedWindowTransformation{
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

func (t *fixedWindowTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) (err error) {
	panic("not implemented")
}

func (t *fixedWindowTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
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
						if err := builder.AppendTime(startColIdx, bnds.Start); err != nil {
							return err
						}
					case t.stopCol:
						if err := builder.AppendTime(stopColIdx, bnds.Stop); err != nil {
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

func (t *fixedWindowTransformation) newWindowGroupKey(tbl flux.Table, keyCols []flux.ColMeta, bnds execute.Bounds, keyColMap []int) flux.GroupKey {
	cols := make([]flux.ColMeta, len(keyCols))
	vs := make([]values.Value, len(keyCols))
	for j, c := range keyCols {
		cols[j] = c
		switch c.Label {
		case t.startCol:
			vs[j] = values.NewTime(bnds.Start)
		case t.stopCol:
			vs[j] = values.NewTime(bnds.Stop)
		default:
			vs[j] = tbl.Key().Value(keyColMap[j])
		}
	}
	return execute.NewGroupKey(cols, vs)
}

func (t *fixedWindowTransformation) clipBounds(bs []execute.Bounds) {
	for i := range bs {
		bs[i] = t.bounds.Intersect(bs[i])
	}
}

func (t *fixedWindowTransformation) getWindowBounds(tm execute.Time) []execute.Bounds {
	if t.w.Every == infinityVar.Duration() {
		return []execute.Bounds{t.bounds}
	}
	bs := t.w.GetOverlappingBounds(execute.Bounds{Start: tm, Stop: tm + 1})
	t.clipBounds(bs)
	return bs
}

func (t *fixedWindowTransformation) generateWindowsWithinBounds() {
	if t.w.Every == infinityVar.Duration() {
		t.allBounds = []execute.Bounds{
			{Start: execute.MinTime, Stop: execute.MaxTime},
		}
		return
	}
	bs := t.w.GetOverlappingBounds(t.bounds)
	t.clipBounds(bs)
	t.allBounds = bs
}

func (t *fixedWindowTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}
func (t *fixedWindowTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}
func (t *fixedWindowTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}

// WindowTriggerPhysicalRule rewrites a physical window operation
// to use a narrow trigger if certain conditions are met.
type WindowTriggerPhysicalRule struct{}

func (WindowTriggerPhysicalRule) Name() string {
	return "WindowTriggerPhysicalRule"
}

// Pattern matches the physical operator pattern consisting of a window
// operator with a single predecessor of any kind.
func (WindowTriggerPhysicalRule) Pattern() plan.Pattern {
	return plan.PhysPat(WindowKind, plan.Any())
}

// Rewrite modifies a window's trigger spec so long as it doesn't have any
// window descendents that occur earlier in the plan and as long as none
// of its descendents merge multiple streams together like union and join.
func (WindowTriggerPhysicalRule) Rewrite(ctx context.Context, window plan.Node) (plan.Node, bool, error) {
	// This rule's pattern ensures us only one predecessor
	if !hasValidPredecessors(window.Predecessors()[0]) {
		return window, false, nil
	}
	// This rule's pattern ensures us a physical operator
	ppn := window.(*plan.PhysicalPlanNode)
	if ppn.TriggerSpec != nil {
		return ppn, false, nil
	}
	ppn.TriggerSpec = plan.NarrowTransformationTriggerSpec{}
	return ppn, true, nil
}

func hasValidPredecessors(node plan.Node) bool {
	pred := node.Predecessors()
	// Source nodes might not produce uniform time bounds for all
	// tables in which case we can't optimize window. However if a
	// source is a BoundsAwareProcedureSpec then it must produce
	// bounded data in which case we can perform the optimization.
	if len(pred) == 0 {
		s := node.ProcedureSpec()
		n, ok := s.(plan.BoundsAwareProcedureSpec)
		return ok && n.TimeBounds(nil) != nil
	}
	kind := node.Kind()
	switch kind {
	// Range gives the same static time bounds to the entire stream,
	// so no need to recurse.
	case RangeKind:
		return true
	case ColumnsKind,
		CumulativeSumKind,
		DerivativeKind,
		DifferenceKind,
		DistinctKind,
		FilterKind,
		FirstKind,
		GroupKind,
		KeyValuesKind,
		KeysKind,
		LastKind,
		LimitKind,
		MaxKind,
		MinKind,
		ExactQuantileSelectKind,
		SampleKind,
		DropKind,
		KeepKind,
		DuplicateKind,
		RenameKind,
		ShiftKind,
		SortKind,
		StateTrackingKind,
		UniqueKind:
	default:
		return false
	}
	if len(pred) == 1 {
		return hasValidPredecessors(pred[0])
	}
	return false
}
