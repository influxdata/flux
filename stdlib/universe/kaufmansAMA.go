package universe

import (
	"fmt"
	"github.com/influxdata/flux/interpreter"
	"math"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
)

const kamaKind = "kaufmansAMA"

type KamaOpSpec struct {
	N       int64    `json:"n"`
	Columns []string `json:"columns"`
}

func init() {
	kamaSignature := flux.FunctionSignature(
		map[string]semantic.PolyType{
			"n":       semantic.Int,
			"columns": semantic.NewArrayPolyType(semantic.String),
		},
		[]string{"n"},
	)

	flux.RegisterPackageValue("universe", kamaKind, flux.FunctionValue(kamaKind, createkamaOpSpec, kamaSignature))
	flux.RegisterOpSpec(kamaKind, newkamaOp)
	plan.RegisterProcedureSpec(kamaKind, newkamaProcedure, kamaKind)
	execute.RegisterTransformation(kamaKind, createkamaTransformation)
}

func createkamaOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	spec := new(KamaOpSpec)

	if n, err := args.GetRequiredInt("n"); err != nil {
		return nil, err
	} else {
		spec.N = n
	}

	if cols, ok, err := args.GetArray("columns", semantic.String); err != nil {
		return nil, err
	} else if ok {
		columns, err := interpreter.ToStringArray(cols)
		if err != nil {
			return nil, err
		}
		spec.Columns = columns
	} else {
		spec.Columns = []string{execute.DefaultValueColLabel}
	}

	return spec, nil
}

func newkamaOp() flux.OperationSpec {
	return new(KamaOpSpec)
}

func (s *KamaOpSpec) Kind() flux.OperationKind {
	return kamaKind
}

type KamaProcedureSpec struct {
	plan.DefaultCost
	N       int64    `json:"n"`
	Columns []string `json:"columns"`
}

func newkamaProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*KamaOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &KamaProcedureSpec{
		N:       spec.N,
		Columns: spec.Columns,
	}, nil
}

func (s *KamaProcedureSpec) Kind() plan.ProcedureKind {
	return kamaKind
}

func (s *KamaProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(KamaProcedureSpec)
	*ns = *s
	return ns
}

// TriggerSpec implements plan.TriggerAwareProcedureSpec
func (s *KamaProcedureSpec) TriggerSpec() plan.TriggerSpec {
	return plan.NarrowTransformationTriggerSpec{}
}

func createkamaTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*KamaProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewkamaTransformation(d, cache, s)
	return t, d, nil
}

type kamaTransformation struct {
	d     execute.Dataset
	cache execute.TableBuilderCache

	n       int64
	columns []string
}

func NewkamaTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *KamaProcedureSpec) *kamaTransformation {
	return &kamaTransformation{
		d:     d,
		cache: cache,

		n:       spec.N,
		columns: spec.Columns,
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
	cols := tbl.Cols()
	doKAMA := make([]bool, len(cols))
	for j, c := range cols {
		found := false
		for _, label := range t.columns {
			if c.Label == label {
				if c.Type != flux.TInt && c.Type != flux.TUInt && c.Type != flux.TFloat {
					return fmt.Errorf("cannot take KAMA of column %s (type %s)", c.Label, c.Type.String())
				}
				found = true
				break
			}
		}

		if found {
			mac := c
			mac.Type = flux.TFloat
			_, err := builder.AddCol(mac)
			if err != nil {
				return err
			}
			doKAMA[j] = true
		} else {
			_, err := builder.AddCol(c)
			if err != nil {
				return err
			}
		}
	}

	return tbl.Do(func(cr flux.ColReader) error {
		if cr.Len() == 0 || len(cr.Cols()) == 0 {
			return nil
		}

		for j, c := range cr.Cols() {
			if !doKAMA[j] {
				for i := int(t.n); i < cr.Len(); i++ {
					if err := builder.AppendValue(j, execute.ValueForRow(cr, i, j)); err != nil {
						return err
					}
				}
			} else {
				var values []float64
				switch c.Type {
				case flux.TInt:
					arr := cr.Ints(j)
					for i := 0; i < arr.Len(); i++ {
						values = append(values, float64(arr.Value(i)))
					}
				case flux.TUInt:
					arr := cr.UInts(j)
					for i := 0; i < arr.Len(); i++ {
						values = append(values, float64(arr.Value(i)))
					}
				case flux.TFloat:
					arr := cr.Floats(j)
					values = arr.Float64Values()
				}
				kers := t.makeKERS(values, int(t.n))
				var kama float64
				sc := 0.0
				kama = values[t.n-1]
				for i := int(t.n); i < cr.Len(); i++ {
					sc = math.Pow(kers[i]*(2.0/(2.0+1.0)-2.0/(30.0+1.0))+2.0/(30.0+1.0), 2)
					kama = kama + sc*(values[i]-kama)
					if err := builder.AppendFloat(j, kama); err != nil {
						return err
					}
				}
			}
		}

		return nil
	})
}

func (t *kamaTransformation) makeKERS(arr []float64, n int) []float64 {
	var sumUp float64
	var sumDown float64
	sumUp = 0
	sumDown = 0

	var kers []float64

	prev := arr[0]
	curr := 0.0
	for i := 0; i < len(arr); i++ {
		curr = arr[i]
		diff := curr - prev
		if i >= n {
			diffNAgo := arr[i-n+1] - arr[i-n]
			val, su, sd := nextCMO(sumUp, sumDown, diff, diffNAgo)
			sumUp = su
			sumDown = sd
			kers = append(kers, math.Abs(val)/100)
		} else {
			_, sumUp, sumDown = nextCMO(sumUp, sumDown, diff, 0)
			kers = append(kers, -999)
		}
		prev = curr
	}

	return kers
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
