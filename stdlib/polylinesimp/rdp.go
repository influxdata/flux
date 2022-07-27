package polylinesimp

import (
	"fmt"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/array"
	fluxarrow "github.com/influxdata/flux/arrow"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/stdlib/polylinesimp/rdp"
)

const RdpKind = "rdp"

type RdpOpSpec struct {
	Column     string  `json:"column"`
	TimeColumn string  `json:"timecolumn"`
	Epsilon    float64 `json:"epsilon"`
	Retention  float64 `json:"retentionpercent"`
}

func init() {
	rdpSignature := runtime.MustLookupBuiltinType("polylinesimp", "rdp")
	runtime.RegisterPackageValue("polylinesimp", RdpKind, flux.MustValue(flux.FunctionValue(RdpKind, createRdpOpSpec, rdpSignature)))
	flux.RegisterOpSpec(RdpKind, newRdpOp)
	plan.RegisterProcedureSpec(RdpKind, newRdpProcedure, RdpKind)
	execute.RegisterTransformation(RdpKind, createRdpTransformation)
}

// Creating the operational spec for the RDP function

func createRdpOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}
	spec := new(RdpOpSpec)
	if col, ok, err := args.GetString("column"); err != nil {
		return nil, err
	} else if ok {
		spec.Column = col
	} else {
		spec.Column = execute.DefaultValueColLabel
	}
	if col, ok, err := args.GetString("timeColumn"); err != nil {
		return nil, err
	} else if ok {
		spec.TimeColumn = col
	} else {
		spec.TimeColumn = execute.DefaultTimeColLabel
	}
	if s, ok, err := args.GetFloat("epsilon"); err != nil {
		return nil, err
	} else if ok {
		if s < 0.0 {
			panic("Epsilon values should be greater than 0.0")
		}
		spec.Epsilon = s
	}
	if sp, ok, err := args.GetFloat("retention"); err != nil {
		return nil, err
	} else if ok {
		if sp < 0.0 || sp > 100.0 {
			panic("The retention rate should be between 0.0 and 100.0")
		}
		spec.Retention = sp
	}
	return spec, nil
}

func newRdpOp() flux.OperationSpec {
	return new(RdpOpSpec)
}

func (s *RdpOpSpec) Kind() flux.OperationKind {
	return RdpKind
}

type RdpProcedureSpec struct {
	plan.DefaultCost
	Column     string
	TimeColumn string
	Epsilon    float64
	Retention  float64
}

func newRdpProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*RdpOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}
	return &RdpProcedureSpec{
		Column:     spec.Column,
		TimeColumn: spec.TimeColumn,
		Epsilon:    spec.Epsilon,
		Retention:  spec.Retention,
	}, nil
}

func (s *RdpProcedureSpec) Kind() plan.ProcedureKind {
	return RdpKind
}
func (s *RdpProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(RdpProcedureSpec)
	*ns = *s
	return ns
}

func (s *RdpProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

// For RDP we create a narrow transformation.
func createRdpTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*RdpProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewRdpTransformation(d, cache, a.Allocator(), s)
	return t, d, nil
}

type RdpTransformation struct {
	execute.ExecutionNode
	d                execute.Dataset
	cache            execute.TableBuilderCache
	alloc            memory.Allocator
	column           string
	timeColumn       string
	epsilon          float64
	retentionpercent float64
}

func NewRdpTransformation(d execute.Dataset, cache execute.TableBuilderCache, alloc memory.Allocator, spec *RdpProcedureSpec) *RdpTransformation {
	return &RdpTransformation{
		d:                d,
		cache:            cache,
		alloc:            alloc,
		column:           spec.Column,
		timeColumn:       spec.TimeColumn,
		epsilon:          spec.Epsilon,
		retentionpercent: spec.Retention,
	}
}

// Execution of the transformation
func (rdpt *RdpTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	// Sanity checks.
	builder, created := rdpt.cache.TableBuilder(tbl.Key())
	if !created {
		return errors.Newf(codes.FailedPrecondition, "Rdp found duplicate table with key: %v", tbl.Key())
	}
	cols := tbl.Cols()
	timeIdx := execute.ColIdx(rdpt.timeColumn, cols)
	if timeIdx < 0 {
		return errors.Newf(codes.FailedPrecondition, "cannot find time column %s", rdpt.timeColumn)
	}
	colIdx := execute.ColIdx(rdpt.column, cols)
	if colIdx < 0 {
		return errors.Newf(codes.FailedPrecondition, "cannot find column %s", rdpt.column)
	}
	typ := cols[colIdx].Type
	if typ != flux.TInt &&
		typ != flux.TUInt &&
		typ != flux.TFloat {
		return errors.Newf(codes.FailedPrecondition, "rdp can work only on numerical types, got %s", typ.String())
	}

	// Building schema.
	if err := execute.AddTableKeyCols(tbl.Key(), builder); err != nil {
		return err
	}
	newTimeIdx, err := builder.AddCol(flux.ColMeta{
		Label: execute.DefaultTimeColLabel,
		Type:  flux.TTime,
	})
	if err != nil {
		return err
	}
	newValueIdx, err := builder.AddCol(flux.ColMeta{
		Label: execute.DefaultValueColLabel,
		Type:  flux.TFloat,
	})
	if err != nil {
		return err
	}

	// Cleaning data for RDP input.
	vs, ts, err := rdpt.getCleanData(tbl, colIdx, timeIdx)
	if err != nil {
		return err
	}

	// Ensuring that the algorithm is not passed with irrelevant inputs
	if rdpt.retentionpercent < 0.0 || rdpt.retentionpercent >= 100.0 {
		panic("Retention percent has to be between 0 and 100")
	}
	if rdpt.epsilon < 0.0 {
		panic("Epsilon value cannot be negative")
	}

	// Passing the cleaned input data to the main RDP function

	rdp_obj := rdp.New(rdpt.timeColumn, rdpt.column, rdpt.epsilon, rdpt.retentionpercent, fluxarrow.NewAllocator(rdpt.alloc))
	newTs, newVs := rdp_obj.Do(vs, ts)
	// don't need vs and ts anymore
	vs.Release()
	ts.Release()

	defer func() {
		newVs.Release()
		newTs.Release()
	}()

	// Appending columns.
	if err := builder.AppendTimes(newTimeIdx, newTs); err != nil {
		return err
	}
	if err := builder.AppendFloats(newValueIdx, newVs); err != nil {
		return err
	}
	if err := execute.AppendKeyValuesN(tbl.Key(), builder, newVs.Len()); err != nil {
		return err
	}
	return nil
}

// getCleanData handles NULL values effectively, drops invalid timestamps and returns two arrow arrays containing X values and Y values.

func (rdpt *RdpTransformation) getCleanData(tbl flux.Table, colIdx, timeIdx int) (*array.Float, *array.Float, error) {
	vs := array.NewFloatBuilder(fluxarrow.NewAllocator(rdpt.alloc))
	ts := array.NewFloatBuilder(fluxarrow.NewAllocator(rdpt.alloc))
	appendV := func(cr flux.ColReader, i int) {
		switch typ := tbl.Cols()[colIdx].Type; typ {
		case flux.TInt:
			c := cr.Ints(colIdx)
			if c.IsNull(i) {
				vs.AppendNull()
			} else {
				vs.Append(float64(c.Value(i)))
			}
		case flux.TUInt:
			c := cr.UInts(colIdx)
			if c.IsNull(i) {
				vs.AppendNull()
			} else {
				vs.Append(float64(c.Value(i)))
			}
		case flux.TFloat:
			c := cr.Floats(colIdx)
			if c.IsNull(i) {
				vs.AppendNull()
			} else {
				vs.Append(float64(c.Value(i)))
			}
		default:
			panic(fmt.Sprintf("cannot append non-numerical type %s", typ.String()))
		}

	}
	if err := tbl.Do(func(cr flux.ColReader) error {
		// we work row-wise
		for i := 0; i < cr.Len(); i++ {
			// drop values with invalid timestamp
			if cts := cr.Times(timeIdx); cts.IsValid(i) {
				trueT := cts.Value(i)
				ts.Append(float64(trueT))
				appendV(cr, i)

			}
		}
		return nil
	}); err != nil {
		return nil, nil, err
	}
	return vs.NewFloatArray(), ts.NewFloatArray(), nil
}

func (rdpt *RdpTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return rdpt.d.RetractTable(key)
}

func (rdpt *RdpTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return rdpt.d.UpdateWatermark(mark)
}
func (rdpt *RdpTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return rdpt.d.UpdateProcessingTime(pt)
}
func (rdpt *RdpTransformation) Finish(id execute.DatasetID, err error) {
	rdpt.d.Finish(err)
}
