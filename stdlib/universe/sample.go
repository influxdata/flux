package universe

import (
	"math/rand"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
)

const SampleKind = "sample"

type SampleOpSpec struct {
	N   int64 `json:"n"`
	Pos int64 `json:"pos"`
	execute.SelectorConfig
}

func init() {
	sampleSignature := runtime.MustLookupBuiltinType("universe", "sample")

	runtime.RegisterPackageValue("universe", SampleKind, flux.MustValue(flux.FunctionValue(SampleKind, createSampleOpSpec, sampleSignature)))
	flux.RegisterOpSpec(SampleKind, newSampleOp)
	plan.RegisterProcedureSpec(SampleKind, newSampleProcedure, SampleKind)
	execute.RegisterTransformation(SampleKind, createSampleTransformation)
}

func createSampleOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(SampleOpSpec)

	n, err := args.GetRequiredInt("n")
	if err != nil {
		return nil, err
	} else if n <= 0 {
		return nil, errors.Newf(codes.Invalid, "n must be a positive integer, but was %d", n)
	}
	spec.N = n

	if pos, ok, err := args.GetInt("pos"); err != nil {
		return nil, err
	} else if ok {
		if pos >= spec.N {
			return nil, errors.Newf(codes.Invalid, "pos must be less than n, but %d >= %d", pos, spec.N)
		}
		spec.Pos = pos
	} else {
		spec.Pos = -1
	}

	if err := spec.SelectorConfig.ReadArgs(args); err != nil {
		return nil, err
	}

	return spec, nil
}

func newSampleOp() flux.OperationSpec {
	return new(SampleOpSpec)
}

func (s *SampleOpSpec) Kind() flux.OperationKind {
	return SampleKind
}

type SampleProcedureSpec struct {
	N   int64
	Pos int64
	execute.SelectorConfig
}

func newSampleProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*SampleOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}
	return &SampleProcedureSpec{
		N:              spec.N,
		Pos:            spec.Pos,
		SelectorConfig: spec.SelectorConfig,
	}, nil
}

func (s *SampleProcedureSpec) Kind() plan.ProcedureKind {
	return SampleKind
}
func (s *SampleProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(SampleProcedureSpec)
	ns.N = s.N
	ns.Pos = s.Pos
	ns.SelectorConfig = s.SelectorConfig
	return ns
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *SampleProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

type SampleSelector struct {
	N   int
	Pos int

	offset   int
	selected []int
}

func createSampleTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	ps, ok := spec.(*SampleProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", ps)
	}

	ss := &SampleSelector{
		N:   int(ps.N),
		Pos: int(ps.Pos),
	}
	t, d := execute.NewIndexSelectorTransformationAndDataset(id, mode, ss, ps.SelectorConfig, a.Allocator())
	return t, d, nil
}

func (s *SampleSelector) reset() {
	pos := s.Pos
	if pos < 0 {
		pos = rand.Intn(s.N)
	}
	s.offset = pos
}

func (s *SampleSelector) NewTimeSelector() execute.DoTimeIndexSelector {
	s.reset()
	return s
}

func (s *SampleSelector) NewBoolSelector() execute.DoBoolIndexSelector {
	s.reset()
	return s
}

func (s *SampleSelector) NewIntSelector() execute.DoIntIndexSelector {
	s.reset()
	return s
}

func (s *SampleSelector) NewUIntSelector() execute.DoUIntIndexSelector {
	s.reset()
	return s
}

func (s *SampleSelector) NewFloatSelector() execute.DoFloatIndexSelector {
	s.reset()
	return s
}

func (s *SampleSelector) NewStringSelector() execute.DoStringIndexSelector {
	s.reset()
	return s
}

func (s *SampleSelector) selectSample(l int) []int {
	var i int
	s.selected = s.selected[0:0]
	for i = s.offset; i < l; i += s.N {
		s.selected = append(s.selected, i)
	}
	s.offset = i - l
	return s.selected
}

func (s *SampleSelector) DoTime(vs *array.Int64) []int {
	return s.selectSample(vs.Len())
}
func (s *SampleSelector) DoBool(vs *array.Boolean) []int {
	return s.selectSample(vs.Len())
}
func (s *SampleSelector) DoInt(vs *array.Int64) []int {
	return s.selectSample(vs.Len())
}
func (s *SampleSelector) DoUInt(vs *array.Uint64) []int {
	return s.selectSample(vs.Len())
}
func (s *SampleSelector) DoFloat(vs *array.Float64) []int {
	return s.selectSample(vs.Len())
}
func (s *SampleSelector) DoString(vs *array.Binary) []int {
	return s.selectSample(vs.Len())
}
