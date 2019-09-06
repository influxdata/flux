package bigtable

import (
	"cloud.google.com/go/bigtable"
	"context"
	"fmt"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
	"google.golang.org/api/option"
	"strconv"
	"strings"
)

const ToBigtableKind = "toBigtable"

type ToBigtableOpSpec struct {
	Token     string `json:"token,omitempty"`
	Project   string `json:"project,omitempty"`
	Instance  string `json:"instance,omitempty"`
	Table     string `json:"table,omitempty"`
	RowkeyCol string `json:"RowkeyCol,omitempty"`
}

func init() {
	toBigtableSignature := flux.FunctionSignature(
		map[string]semantic.PolyType{
			"token":     semantic.String,
			"project":   semantic.String,
			"instance":  semantic.String,
			"table":     semantic.String,
			"rowkeyCol": semantic.String,
		},
		[]string{"token", "project", "instance", "table"},
	)
	flux.RegisterPackageValue("experimental/bigtable", "to", flux.FunctionValueWithSideEffect(ToBigtableKind, createToBigtableOpSpec, toBigtableSignature))
	flux.RegisterOpSpec(ToBigtableKind, func() flux.OperationSpec { return &ToBigtableOpSpec{} })
	plan.RegisterProcedureSpecWithSideEffect(ToBigtableKind, newToBigtableProcedure, ToBigtableKind)
	execute.RegisterTransformation(ToBigtableKind, createToBigtableTransformation)
}

func (o *ToBigtableOpSpec) ReadArgs(args flux.Arguments) error {
	var err error

	o.Token, err = args.GetRequiredString("token")
	if err != nil {
		return err
	}

	o.Instance, err = args.GetRequiredString("instance")
	if err != nil {
		return err
	}

	o.Project, err = args.GetRequiredString("project")
	if err != nil {
		return err
	}

	o.Table, err = args.GetRequiredString("table")
	if err != nil {
		return err
	}

	var ok bool
	o.RowkeyCol, ok, err = args.GetString("rowkeyCol")
	if err != nil {
		return err
	}
	if !ok {
		o.RowkeyCol = "_time"
	}

	return err
}

func createToBigtableOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}

	b := new(ToBigtableOpSpec)
	if err := b.ReadArgs(args); err != nil {
		return nil, err
	}

	return b, nil
}

func (ToBigtableOpSpec) Kind() flux.OperationKind {
	return ToBigtableKind
}

func (o *ToBigtableProcedureSpec) Copy() plan.ProcedureSpec {
	s := o.Spec
	res := &ToBigtableProcedureSpec{
		Spec: &ToBigtableOpSpec{
			Token:     s.Token,
			Project:   s.Project,
			Instance:  s.Instance,
			Table:     s.Table,
			RowkeyCol: s.RowkeyCol,
		},
	}
	return res
}

type ToBigtableProcedureSpec struct {
	plan.DefaultCost
	Spec *ToBigtableOpSpec
}

func (o *ToBigtableProcedureSpec) Kind() plan.ProcedureKind {
	return ToBigtableKind
}

func newToBigtableProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*ToBigtableOpSpec)
	if !ok {
		return nil, fmt.Errorf("invalid spec type %T", qs)
	}

	return &ToBigtableProcedureSpec{Spec: spec}, nil
}

func createToBigtableTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*ToBigtableProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t, err := NewToBigtableTransformation(d, cache, s)
	if err != nil {
		return nil, nil, err
	}
	return t, d, nil
}

type ToBigtableTransformation struct {
	d     execute.Dataset
	cache execute.TableBuilderCache
	spec  *ToBigtableProcedureSpec

	client  *bigtable.AdminClient
	tblInfo *bigtable.TableInfo
}

func (t *ToBigtableTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func NewToBigtableTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *ToBigtableProcedureSpec) (*ToBigtableTransformation, error) {
	client, err := bigtable.NewAdminClient(context.Background(), spec.Spec.Project, spec.Spec.Instance, option.WithCredentialsJSON([]byte(spec.Spec.Token)))
	if err != nil {
		return nil, err
	}
	if err := client.CreateTable(context.Background(), spec.Spec.Table); err != nil {
		return nil, err
	}
	tblInfo, err := client.TableInfo(context.Background(), spec.Spec.Table)
	if err != nil {
		return nil, err
	}
	return &ToBigtableTransformation{
		d:       d,
		cache:   cache,
		spec:    spec,
		client:  client,
		tblInfo: tblInfo,
	}, nil
}

func isKeyCol(list []flux.ColMeta, target flux.ColMeta) bool {
	for _, s := range list {
		if s == target {
			return true
		}
	}
	return false
}

func (t *ToBigtableTransformation) Process(id execute.DatasetID, tbl flux.Table) (err error) {
	return tbl.Do(func(cr flux.ColReader) error {
		cols := tbl.Cols()
		keyCols := tbl.Key()

		keyValueNames := tbl.Key().Values() // store the values in the key columns

		var keyValueNamesStr []string
		// get the values in the key columns as strings so we can concatenate them

		for i := 0; i < len(keyValueNames); i++ {
			keyValueNamesStr = append(keyValueNamesStr, keyValueNames[i].Str())
		}

		rows := cr.Len()
		familyName := strings.Join(keyValueNamesStr, "-")

		newCli, err := bigtable.NewClient(context.Background(), t.spec.Spec.Project, t.spec.Spec.Instance, option.WithCredentialsJSON([]byte(t.spec.Spec.Token)))
		if err != nil {
			return err
		}

		newTbl := newCli.Open(t.spec.Spec.Table)
		muts := make([]*bigtable.Mutation, rows)
		rowKeys := make([]string, rows)

		rowKeyColIdx := execute.ColIdx(t.spec.Spec.RowkeyCol, tbl.Cols())
		rowKeyType := tbl.Cols()[rowKeyColIdx].Type

		for i := 0; i < rows; i++ {
			for j := 0; j < len(cols); j++ {
				// if this is a key column, we need to skip it
				if isKeyCol(keyCols.Cols(), cols[j]) {
					continue
				}

				muts[i] = bigtable.NewMutation()

				v := execute.ValueForRow(cr, i, j)
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

				muts[i].Set(familyName, cols[j].Label, bigtable.Now(), []byte(str))
			}

			// set the rowkey value for the current row
			switch rowKeyType {
			case flux.TBool:
				rowKeys[i] = fmt.Sprintf("%v", cr.Bools(rowKeyColIdx).Value(i))
			case flux.TFloat:
				rowKeys[i] = fmt.Sprintf("%v", cr.Floats(rowKeyColIdx).Value(i))
			case flux.TInt:
				rowKeys[i] = fmt.Sprintf("%v", cr.Ints(rowKeyColIdx).Value(i))
			case flux.TString:
				rowKeys[i] = fmt.Sprintf("%v", cr.Strings(rowKeyColIdx).Value(i))
			case flux.TUInt:
				rowKeys[i] = fmt.Sprintf("%v", cr.UInts(rowKeyColIdx).Value(i))
			case flux.TTime:
				rowKeys[i] = fmt.Sprintf("%v", cr.Times(rowKeyColIdx).Value(i))
			}
		}

		rowErrs, err := newTbl.ApplyBulk(context.Background(), rowKeys, muts)
		if err != nil {
			return err
		}
		for _, rowErr := range rowErrs {
			return fmt.Errorf("error writing row: %v", rowErr)
		}

		return nil
	})
}

func (t *ToBigtableTransformation) UpdateWatermark(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateWatermark(pt)
}

func (t *ToBigtableTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}

func (t *ToBigtableTransformation) Finish(id execute.DatasetID, err error) {
	if cliErr := t.client.Close(); cliErr != nil {
		err = errors.Wrap(err, codes.Inherit, cliErr)
	}
	t.d.Finish(err)
}
