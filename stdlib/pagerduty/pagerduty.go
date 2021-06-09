package pagerduty

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const DedupKeyKind = "dedupKey"
const dedupKeyColName = "_pagerdutyDedupKey" // if you change this, make sure to change it in the pagerduty.flux too!

type DedupKeyOpSpec struct {
	Exclude []string
}

func (s *DedupKeyOpSpec) Kind() flux.OperationKind {
	return DedupKeyKind
}

func init() {
	dedupKeySignature := runtime.MustLookupBuiltinType("pagerduty", "dedupKey")
	runtime.RegisterPackageValue("pagerduty", "dedupKey", flux.MustValue(flux.FunctionValue(DedupKeyKind, createDedupKeyOpSpec, dedupKeySignature)))
	plan.RegisterProcedureSpec(DedupKeyKind, newDedupKeyProcedure, DedupKeyKind)
	execute.RegisterTransformation(DedupKeyKind, createDedupKeyTransformation)
}

func createDedupKeyOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	exclude, ok, err := args.GetArrayAllowEmpty("exclude", semantic.String)
	if err != nil {
		return nil, err
	} else if !ok {
		exclude = values.NewArrayWithBacking(
			semantic.NewArrayType(semantic.BasicString),
			[]values.Value{
				values.NewString(execute.DefaultStartColLabel),
				values.NewString(execute.DefaultStopColLabel),
				values.NewString("_level"),
			},
		)
	}

	spec := &DedupKeyOpSpec{
		Exclude: make([]string, exclude.Len()),
	}
	exclude.Range(func(i int, v values.Value) {
		spec.Exclude[i] = v.Str()
	})
	return spec, nil
}

type DedupProcedureSpec struct {
	plan.DefaultCost
	Exclude []string
}

func (s *DedupProcedureSpec) Kind() plan.ProcedureKind {
	return DedupKeyKind
}

func (s *DedupProcedureSpec) Copy() plan.ProcedureSpec {
	ns := *s
	ns.Exclude = make([]string, len(s.Exclude))
	copy(ns.Exclude, s.Exclude)
	return &ns
}

func newDedupKeyProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*DedupKeyOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}
	return &DedupProcedureSpec{
		Exclude: spec.Exclude,
	}, nil
}

type DedupKeyTransformation struct {
	execute.ExecutionNode
	d       execute.Dataset
	cache   execute.TableBuilderCache
	exclude []string
}

func createDedupKeyTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	cache := execute.NewTableBuilderCache(a.Allocator())
	dataset := execute.NewDataset(id, mode, cache)
	s, ok := spec.(*DedupProcedureSpec)
	if !ok {
		return nil, nil, fmt.Errorf("invalid spec type %T", spec)
	}

	transform := NewDedupKeyTransformation(dataset, s, cache)
	return transform, dataset, nil
}

func NewDedupKeyTransformation(d execute.Dataset, spec *DedupProcedureSpec, cache execute.TableBuilderCache) *DedupKeyTransformation {
	return &DedupKeyTransformation{
		d:       d,
		cache:   cache,
		exclude: spec.Exclude,
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
	keys := make([]kvs, 0, len(groupCols))
	for i, col := range groupCols {
		if execute.ContainsStr(t.exclude, col.Label) {
			continue
		}

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
			return errors.Newf(codes.Invalid, "cannot convert %v to string", v.Type())
		}
		keys = append(keys, kvs{
			k: col.Label,
			v: str,
		})

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
