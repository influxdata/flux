package promql

import (
	"fmt"
	"regexp"

	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/plan"
)

const (
	// LabelReplaceKind is the Kind for the LabelReplace Flux function
	LabelReplaceKind = "labelReplace"
)

type LabelReplaceOpSpec struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Regex       string `json:"regex"`
	Replacement string `json:"replacement"`
}

func init() {
	labelReplaceSignature := runtime.MustLookupBuiltinType("internal/promql", "labelReplace")
	runtime.RegisterPackageValue("internal/promql", "labelReplace", flux.MustValue(flux.FunctionValue(LabelReplaceKind, createLabelReplaceOpSpec, labelReplaceSignature)))
	flux.RegisterOpSpec(LabelReplaceKind, func() flux.OperationSpec { return &LabelReplaceOpSpec{} })
	plan.RegisterProcedureSpec(LabelReplaceKind, newLabelReplaceProcedure, LabelReplaceKind)
	execute.RegisterTransformation(LabelReplaceKind, createLabelReplaceTransformation)
}

func createLabelReplaceOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(LabelReplaceOpSpec)

	if src, err := args.GetRequiredString("source"); err != nil {
		return nil, err
	} else {
		spec.Source = src
	}

	if dst, err := args.GetRequiredString("destination"); err != nil {
		return nil, err
	} else {
		spec.Destination = dst
	}

	if re, err := args.GetRequiredString("regex"); err != nil {
		return nil, err
	} else {
		spec.Regex = re
	}

	if repl, err := args.GetRequiredString("replacement"); err != nil {
		return nil, err
	} else {
		spec.Replacement = repl
	}

	return spec, nil
}

func (s *LabelReplaceOpSpec) Kind() flux.OperationKind {
	return LabelReplaceKind
}

func newLabelReplaceProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	s, ok := qs.(*LabelReplaceOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &LabelReplaceProcedureSpec{
		Source:      s.Source,
		Destination: s.Destination,
		Regex:       s.Regex,
		Replacement: s.Replacement,
	}, nil
}

type LabelReplaceProcedureSpec struct {
	plan.DefaultCost
	Source      string
	Destination string
	Regex       string
	Replacement string
}

func (s *LabelReplaceProcedureSpec) Kind() plan.ProcedureKind {
	return LabelReplaceKind
}

func (s *LabelReplaceProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(LabelReplaceProcedureSpec)
	*ns = *s
	return ns
}

func createLabelReplaceTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*LabelReplaceProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewLabelReplaceTransformation(d, cache, s)
	return t, d, nil
}

type labelReplaceTransformation struct {
	execute.ExecutionNode
	d     execute.Dataset
	cache execute.TableBuilderCache

	source      string
	destination string
	regex       string
	replacement string
}

func NewLabelReplaceTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *LabelReplaceProcedureSpec) *labelReplaceTransformation {
	return &labelReplaceTransformation{
		d:     d,
		cache: cache,

		source:      spec.Source,
		destination: spec.Destination,
		regex:       spec.Regex,
		replacement: spec.Replacement,
	}
}

func (t *labelReplaceTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *labelReplaceTransformation) Process(id execute.DatasetID, tbl flux.Table) (err error) {
	re, err := regexp.Compile("^(?:" + t.regex + ")$")
	if err != nil {
		return fmt.Errorf("invalid regular expression in label_replace(): %s", t.regex)
	}

	key := tbl.Key()
	var srcVal string
	if key.HasCol(t.source) {
		srcVal = key.LabelValue(t.source).Str()
	}

	indexes := re.FindStringSubmatchIndex(srcVal)

	var outKey flux.GroupKey

	if indexes == nil {
		// If there is no match, no replacement should take place.
		outKey = key
	} else {
		res := re.ExpandString([]byte{}, t.replacement, srcVal, indexes)

		outCols := make([]flux.ColMeta, 0, len(key.Cols()))
		outVals := make([]values.Value, 0, len(key.Cols()))
		for i, col := range key.Cols() {
			if col.Label == t.destination {
				continue
			}

			outCols = append(outCols, col)
			outVals = append(outVals, key.Values()[i])
		}

		if len(res) > 0 {
			outCols = append(outCols, flux.ColMeta{Label: t.destination, Type: flux.TString})
			outVals = append(outVals, values.NewString(string(res)))
		}

		outKey = execute.NewGroupKey(outCols, outVals)
	}

	builder, created := t.cache.TableBuilder(outKey)
	if !created {
		return fmt.Errorf("labelReplace found duplicate table with key: %v", tbl.Key())
	}
	if err := execute.AddTableKeyCols(outKey, builder); err != nil {
		return err
	}

	cols := tbl.Cols()
	valIdx := execute.ColIdx(execute.DefaultValueColLabel, cols)
	if valIdx < 0 {
		return fmt.Errorf("value column not found: %s", execute.DefaultValueColLabel)
	}

	outValIdx, err := builder.AddCol(flux.ColMeta{Label: execute.DefaultValueColLabel, Type: flux.TFloat})
	if err != nil {
		return fmt.Errorf("error appending value column: %s", err)
	}

	return tbl.Do(func(cr flux.ColReader) error {
		for i := 0; i < cr.Len(); i++ {
			err := execute.AppendKeyValues(outKey, builder)
			if err != nil {
				return err
			}

			v := execute.ValueForRow(cr, i, valIdx)
			if err := builder.AppendValue(outValIdx, v); err != nil {
				return err
			}
		}

		return nil
	})
}

func (t *labelReplaceTransformation) UpdateWatermark(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateWatermark(pt)
}

func (t *labelReplaceTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}

func (t *labelReplaceTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
