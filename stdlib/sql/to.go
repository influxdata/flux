package sql

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

const (
	ToSQLKind = "toSQL"
	BatchSize = 10000
)

type ToSQLOpSpec struct {
	DriverName     string `json:"driverName,omitempty"`
	DataSourceName string `json:"dataSourcename,omitempty"`
	Table          string `json:"table,omitempty"`
}

func init() {
	toSQLSignature := flux.FunctionSignature(
		map[string]semantic.PolyType{
			"driverName":     semantic.String,
			"dataSourceName": semantic.String,
			"table":          semantic.String,
		},
		[]string{"driverName", "dataSourceName", "table"},
	)
	flux.RegisterPackageValue("sql", "to", flux.FunctionValueWithSideEffect(ToSQLKind, createToSQLOpSpec, toSQLSignature))
	flux.RegisterOpSpec(ToSQLKind, func() flux.OperationSpec { return &ToSQLOpSpec{} })
	plan.RegisterProcedureSpecWithSideEffect(ToSQLKind, newToSQLProcedure, ToSQLKind)
	execute.RegisterTransformation(ToSQLKind, createToSQLTransformation)
}

func (o *ToSQLOpSpec) ReadArgs(args flux.Arguments) error {
	var err error

	o.DriverName, err = args.GetRequiredString("driverName")
	if err != nil {
		return err
	}
	if len(o.DriverName) == 0 {
		return errors.New(codes.Invalid, "invalid driver name")
	}

	o.DataSourceName, err = args.GetRequiredString("dataSourceName")
	if err != nil {
		return err
	}
	if len(o.DataSourceName) == 0 {
		return errors.New(codes.Invalid, "invalid data source name")
	}

	o.Table, err = args.GetRequiredString("table")
	if err != nil {
		return err
	}
	if len(o.Table) == 0 {
		return errors.New(codes.Invalid, "invalid table name")
	}

	return err
}

func createToSQLOpSpec(args flux.Arguments, a *flux.Administration) (flux.OperationSpec, error) {
	if err := a.AddParentFromArgs(args); err != nil {
		return nil, err
	}
	s := new(ToSQLOpSpec)
	if err := s.ReadArgs(args); err != nil {
		return nil, err
	}
	return s, nil
}

func (ToSQLOpSpec) Kind() flux.OperationKind {
	return ToSQLKind
}

type ToSQLProcedureSpec struct {
	plan.DefaultCost
	Spec *ToSQLOpSpec
}

func (o *ToSQLProcedureSpec) Kind() plan.ProcedureKind {
	return ToSQLKind
}

func (o *ToSQLProcedureSpec) Copy() plan.ProcedureSpec {
	s := o.Spec
	res := &ToSQLProcedureSpec{
		Spec: &ToSQLOpSpec{
			DriverName:     s.DriverName,
			DataSourceName: s.DataSourceName,
			Table:          s.Table,
		},
	}
	return res
}

func newToSQLProcedure(qs flux.OperationSpec, a plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*ToSQLOpSpec)
	if !ok && spec != nil {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}
	return &ToSQLProcedureSpec{Spec: spec}, nil
}

func createToSQLTransformation(id execute.DatasetID, mode execute.AccumulationMode, spec plan.ProcedureSpec, a execute.Administration) (execute.Transformation, execute.Dataset, error) {
	s, ok := spec.(*ToSQLProcedureSpec)
	if !ok {
		return nil, nil, errors.Newf(codes.Internal, "invalid spec type %T", spec)
	}
	cache := execute.NewTableBuilderCache(a.Allocator())
	d := execute.NewDataset(id, mode, cache)
	t := NewToSQLTransformation(d, cache, s)
	return t, d, nil
}

type ToSQLTransformation struct {
	d     execute.Dataset
	cache execute.TableBuilderCache
	spec  *ToSQLProcedureSpec
	db    *sql.DB
	tx    *sql.Tx
}

func (t *ToSQLTransformation) RetractTable(id execute.DatasetID, key flux.GroupKey) error {
	return t.d.RetractTable(key)
}

func NewToSQLTransformation(d execute.Dataset, cache execute.TableBuilderCache, spec *ToSQLProcedureSpec) *ToSQLTransformation {
	db, err := sql.Open(spec.Spec.DriverName, spec.Spec.DataSourceName)
	if err != nil {
		panic(err)
	}
	var tx *sql.Tx
	if spec.Spec.DriverName != "sqlmock" {
		tx, err = db.Begin()
		if err != nil {
			panic(err)
		}
	}
	return &ToSQLTransformation{
		d:     d,
		cache: cache,
		spec:  spec,
		db:    db,
		tx:    tx,
	}
}

type idxType struct {
	Idx  int
	Type flux.ColType
}

func (t *ToSQLTransformation) Process(id execute.DatasetID, tbl flux.Table) (err error) {
	colNames, valStrings, valArgs, err := CreateInsertComponents(t, tbl)
	if err != nil {
		return err
	}
	for i := range valStrings {
		if err := ExecuteQueries(t.tx, t.spec.Spec, colNames, &valStrings[i], &valArgs[i]); err != nil {
			return nil
		}
	}
	return err
}

func (t *ToSQLTransformation) UpdateWatermark(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateWatermark(pt)
}

func (t *ToSQLTransformation) UpdateProcessingTime(id execute.DatasetID, pt execute.Time) error {
	return t.d.UpdateProcessingTime(pt)
}

func (t *ToSQLTransformation) Finish(id execute.DatasetID, err error) {
	if t.spec.Spec.DriverName != "sqlmock" {
		var txErr error
		if err == nil {
			txErr = t.tx.Commit()
		} else {
			txErr = t.tx.Rollback()
		}
		if txErr != nil {
			panic(txErr)
		}
	}
	t.d.Finish(err)
}

func CreateInsertComponents(t *ToSQLTransformation, tbl flux.Table) (colNames []string, valStringArray [][]string, valArgsArray [][]interface{}, err error) {
	cols := tbl.Cols()
	labels := make(map[string]idxType, len(cols))
	var questionMarks, newSQLTableCols []string
	for i, col := range cols {
		labels[col.Label] = idxType{Idx: i, Type: col.Type}
		questionMarks = append(questionMarks, "?")
		colNames = append(colNames, col.Label)

		switch col.Type {
		case flux.TFloat:
			newSQLTableCols = append(newSQLTableCols, fmt.Sprintf("%s FLOAT", col.Label))
		case flux.TInt:
			newSQLTableCols = append(newSQLTableCols, fmt.Sprintf("%s BIGINT", col.Label))
		case flux.TUInt:
			newSQLTableCols = append(newSQLTableCols, fmt.Sprintf("%s BIGINT", col.Label))
		case flux.TString:
			switch t.spec.Spec.DriverName {
			case "mysql":
				newSQLTableCols = append(newSQLTableCols, fmt.Sprintf("%s TEXT(16383)", col.Label))
			case "postgres":
				newSQLTableCols = append(newSQLTableCols, fmt.Sprintf("%s text", col.Label))
			}
		case flux.TTime:
			newSQLTableCols = append(newSQLTableCols, fmt.Sprintf("%s DATETIME", col.Label))
		case flux.TBool:
			newSQLTableCols = append(newSQLTableCols, fmt.Sprintf("%s BOOL", col.Label))
		default:
			return nil, nil, nil, errors.Newf(codes.Internal, "invalid type for column %s", col.Label)
		}
	}

	// Creates the placeholders for values in the query
	// eg: (?,?)
	valuePlaceHolders := fmt.Sprintf("(%s)", strings.Join(questionMarks, ","))

	builder, new := t.cache.TableBuilder(tbl.Key())
	if new {
		if err := execute.AddTableCols(tbl, builder); err != nil {
			return nil, nil, nil, err
		}
	}

	if err := tbl.Do(func(er flux.ColReader) error {
		l := er.Len()

		// valueStrings is an array of valuePlaceHolders, which will be joined later
		valueStrings := make([]string, 0, l)
		// valueArgs holds all the values to pass into the query
		valueArgs := make([]interface{}, 0, l*len(cols))

		if t.spec.Spec.DriverName != "sqlmock" {
			q := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s)", t.spec.Spec.Table, strings.Join(newSQLTableCols, ","))
			_, err = t.tx.Exec(q)
			if err != nil {
				return err
			}
		}

		for i := 0; i < l; i++ {
			valueStrings = append(valueStrings, valuePlaceHolders)
			for j, col := range er.Cols() {
				switch col.Type {
				case flux.TFloat:
					if er.Floats(j).IsNull(i) {
						valueArgs = append(valueArgs, nil)
						break
					}
					valueArgs = append(valueArgs, er.Floats(j).Value(i))
				case flux.TInt:
					if er.Ints(j).IsNull(i) {
						valueArgs = append(valueArgs, nil)
						break
					}
					valueArgs = append(valueArgs, er.Ints(j).Value(i))
				case flux.TUInt:
					if er.UInts(j).IsNull(i) {
						valueArgs = append(valueArgs, nil)
						break
					}
					valueArgs = append(valueArgs, er.UInts(j).Value(i))
				case flux.TString:
					if er.Strings(j).IsNull(i) {
						valueArgs = append(valueArgs, nil)
						break
					}
					valueArgs = append(valueArgs, er.Strings(j).ValueString(i))
				case flux.TTime:
					if er.Times(j).IsNull(i) {
						valueArgs = append(valueArgs, nil)
						break
					}
					valueArgs = append(valueArgs, values.Time(er.Times(j).Value(i)).Time())
				case flux.TBool:
					if er.Bools(j).IsNull(i) {
						valueArgs = append(valueArgs, nil)
						break
					}
					valueArgs = append(valueArgs, er.Bools(j).Value(i))
				default:
					return fmt.Errorf("invalid type for column %s", col.Label)
				}
			}

			if err := execute.AppendRecord(i, er, builder); err != nil {
				return err
			}

			if i != 0 && i%BatchSize == 0 {
				valArgsArray = append(valArgsArray, valueArgs)
				valStringArray = append(valStringArray, valueStrings)
				valueArgs = make([]interface{}, 0)
				valueStrings = make([]string, 0)
			}
		}
		if len(valueArgs) > 0 && len(valueStrings) > 0 {
			valArgsArray = append(valArgsArray, valueArgs)
			valStringArray = append(valStringArray, valueStrings)
		}
		return nil
	}); err != nil {
		return nil, nil, nil, err
	}

	return colNames, valStringArray, valArgsArray, err
}

func ExecuteQueries(tx *sql.Tx, s *ToSQLOpSpec, colNames []string, valueStrings *[]string, valueArgs *[]interface{}) (err error) {
	concatValueStrings := strings.Join(*valueStrings, ",")

	// PostgreSQL uses $n instead of ? for placeholders
	if s.DriverName == "postgres" {
		for pqCounter := 1; strings.Contains(concatValueStrings, "?"); pqCounter++ {
			concatValueStrings = strings.Replace(concatValueStrings, "?", fmt.Sprintf("$%v", pqCounter), 1)
		}
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", s.Table, strings.Join(colNames, ","), concatValueStrings)
	if s.DriverName != "sqlmock" {
		_, err := tx.Exec(query, *valueArgs...)
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				return fmt.Errorf("transaction failed (%s) while recovering from %s", err, rbErr)
			}
			return err
		}
	}
	return err
}
