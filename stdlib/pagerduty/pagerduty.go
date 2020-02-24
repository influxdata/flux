package pagerduty

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const DedupKeyKind = "dedupKey"
const dedupKeyColName = "_pagerdutyDedupKey" // if you change this, make sure to change it in the pagerduty.flux too!

type DedupKeyOpSpec struct{}

func (s *DedupKeyOpSpec) Kind() flux.OperationKind {
	return DedupKeyKind
}

func init() {
	dedupKeySignature := semantic.MustLookupBuiltinType("pagerduty", "dedupKey")
	runtime.RegisterPackageValue("pagerduty", "dedupKey", flux.MustValue(flux.FunctionValue(DedupKeyKind, createDedupKeyOpSpec, dedupKeySignature)))
	flux.RegisterOpSpec(DedupKeyKind, newDedupKeyOp)
	plan.RegisterProcedureSpec(DedupKeyKind, newDedupKeyProcedure, DedupKeyKind)
	execute.RegisterTransformation(DedupKeyKind, createDedupKeyTransformation)
}

func createDedupKeyOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}
	return &DedupKeyOpSpec{}, nil
}

func newDedupKeyOp() flux.OperationSpec {
	return new(DedupKeyOpSpec)
}

type DedupProcedureSpec struct {
	plan.DefaultCost
}

func (s *DedupProcedureSpec) Kind() plan.ProcedureKind {
	return DedupKeyKind
}

func (s *DedupProcedureSpec) Copy() plan.ProcedureSpec {
	ns := *s
	return &ns
}

func newDedupKeyProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	if _, ok := qs.(*DedupKeyOpSpec); !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}
	return &DedupProcedureSpec{}, nil
}

type DedupKeyTransformation struct {
	d     execute.Dataset
	cache execute.TableBuilderCache
}

func createDedupKeyTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	cache := execute.NewTableBuilderCache(a.Allocator())
	dataset := execute.NewDataset(id, mode, cache)
	if _, ok := spec.(*DedupProcedureSpec); !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}

	transform := NewDedupKeyTransformation(dataset, cache)
	return transform, dataset, nil
}

func NewDedupKeyTransformation(d execute.Dataset, cache execute.TableBuilderCache) *DedupKeyTransformation {
	return &DedupKeyTransformation{
		d:     d,
		cache: cache,
	}
}

func (t *DedupKeyTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

type kvs struct {
	k string
	v string
}

func (t *DedupKeyTransformation) Process(id execute.DatasetID, tbl flux.Table) error {
	groupCols := tbl.Key().Cols()
	groupVals := tbl.Key().Values()
	keys := make([]kvs, len(groupCols))
	for i := range groupCols {
		keys[i].k = groupCols[i].Label
		v := groupVals[i]
		var str string
		switch v.Type().Nature() {
		case semantic.String:
			str = v.Str()
		case semantic.Int:
			str = strconv.FormatInt(v.Int(), 10)
		case semantic.UInt:
			str = strconv.FormatUint(v.UInt(), 10)
		case semantic.Float:
			str = strconv.FormatFloat(v.Float(), 'f', -1, 64)
		case semantic.Bool:
			str = strconv.FormatBool(v.Bool())
		case semantic.Time:
			str = v.Time().String()
		case semantic.Duration:
			str = v.Duration().String()
		default:
			return fmt.Errorf("cannot convert %v to string", v.Type())
		}

		keys[i].v = str

	}
	sort.Slice(keys, func(i, j int) bool {
		return keys[i].k < keys[j].k
	})

	sb := strings.Builder{}
	for i := range keys {
		sb.WriteString(keys[i].k)
		sb.WriteRune('\n')
		sb.WriteString(keys[i].v)
		sb.WriteRune('\n')
	}

	dedupKey := sb.String()

	builder, isNew := t.cache.TableBuilder(tbl.Key())

	if isNew {
		if err := execute.AddTableCols(tbl, builder); err != nil { // adds the other columns to builder
			return err
		}
	}

	colIDX, err := builder.AddCol(flux.ColMeta{
		Label: dedupKeyColName,
		Type:  flux.TString,
	})
	if err != nil {
		return err
	}

	// because pagerduty restricts the dedupKey size we are hashing to reduce chance of collisions, by distributing the key more evenly
	dedupKeyHash := sha256.Sum256([]byte(dedupKey))
	dedupKeyHashHex := hex.EncodeToString(dedupKeyHash[:])

	err = tbl.Do(func(cr flux.ColReader) error {
		l := cr.Len()
		for j := range builder.Cols() {
			if j == colIDX {
				for i := 0; i < l; i++ {
					if err := builder.AppendValue(j, values.NewString(dedupKeyHashHex)); err != nil {
						return err
					}
				}
			} else {
				if err := execute.AppendCol(j, j, cr, builder); err != nil {
					return err
				}
			}
		}
		return nil
	})
	return err
}

func (t *DedupKeyTransformation) UpdateWatermark(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateWatermark(mark)
}

func (t *DedupKeyTransformation) UpdateProcessingTime(id execute.DatasetID, mark execute.Time) error {
	return t.d.UpdateProcessingTime(mark)
}

func (t *DedupKeyTransformation) Finish(id execute.DatasetID, err error) {
	t.d.Finish(err)
}
