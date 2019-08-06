package universe

import (
	"fmt"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
)

const KAMAKind = "KAMAHelper"

type KAMAOpSpec struct {
	N int64 `json:"n"`
}

func init() {
	kamaSignature := flux.FunctionSignature(
		map[string]semantic.PolyType{
			"n": semantic.Int,
		},
		[]string{"n"},
	)

	flux.RegisterPackageValue("universe", KAMAKind, flux.FunctionValue(KAMAKind, createKAMAOpSpec, kamaSignature))
	flux.RegisterOpSpec(KAMAKind, newKAMAOp)
	plan.RegisterProcedureSpec(KAMAKind, newKAMAProcedure, KAMAKind)
	execute.RegisterTransformation(KAMAKind, createKAMATransformation)
}

func createKAMAOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(KAMAOpSpec)

	if n, err := args.GetRequiredInt("n"); err != nil {
		return nil, err
	} else {
		spec.N = n
	}

	return spec, nil
}

func newKAMAOp() flux.OperationSpec {
	return new(KAMAOpSpec)
}

func (s *KAMAOpSpec) Kind() flux.OperationKind {
	return KAMAKind
}

type KAMAProcedureSpec struct {
	plan.DefaultCost
	N int64 `json:"n"`
}

func newKAMAProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*KAMAOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &KAMAProcedureSpec{
		N: spec.N,
	}, nil
}

func (s *KAMAProcedureSpec) Kind() plan.ProcedureKind {
	return KAMAKind
}

func (s *KAMAProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(KAMAProcedureSpec)
	*ns = *s
	return ns
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *KAMAProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

func createKAMATransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*KAMAProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewKAMATransformation(d, cache, s)
	return t, d, nil
}

type kamaTransformation struct {
	d     execute.Dataset
	cache execute.TableBuilderCache

	n int64
}

func NewKAMATransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *KAMAProcedureSpec) *kamaTransformation {
	return &kamaTransformation{
		d:     d,
		cache: cache,

		n: spec.N,
	}
}

func (t *kamaTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func (t *kamaTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	builder, created := t.cache.TableBuilder(tbl.Key())
	if !created {
		return fmt.Errorf("KAMA found duplicate table with key: %v", tbl.Key())
	}
	if t.n <= 0 {
		return fmt.Errorf("cannot take KAMA with a period of %v (must be greater than 0)", t.n)
	}
	cols := tbl.Cols()
	valueIdx := -1
	for j, c := range cols {
		if c.Label == execute.DefaultValueColLabel {
			if c.Type != flux.TInt && c.Type != flux.TUInt && c.Type != flux.TFloat {
				return fmt.Errorf("cannot take exponential moving average of column %s (type %s)", c.Label, c.Type.String())
			}
			valueIdx = j
			mac := c
			mac.Type = flux.TFloat
			_, err := builder.AddCol(mac)
			if err != nil {
				return err
			}
		} else {
			_, err := builder.AddCol(c)
			if err != nil {
				return err
			}
		}
	}
	if valueIdx == -1 {
		return fmt.Errorf("cannot find _value column")
	}

	return tbl.Do(func(cr flux.ColReader) error {
		//
		return nil
	})
}

func (t *kamaTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}

func (t *kamaTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}

func (t *kamaTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
