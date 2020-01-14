package sql

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/values"
)

const (
	ToSQLKind        = "toSQL"
	DefaultBatchSize = 10000 //TODO: decide if this should be kept low enough for the lowest (SQLite), or not.
)

type ToSQLOpSpec struct {
	DriverName     string `json:"driverName,omitempty"`
	DataSourceName string `json:"dataSourcename,omitempty"`
	Table          string `json:"table,omitempty"`
	BatchSize      int    `json:"batchSize,omitempty"`
}

func init() {
	toSQLSignature := semantic.LookupBuiltInType("sql", "ro")
	flux.RegisterPackageValue("sql", "to", flux.MustValue(flux.FunctionValueWithSideEffect(ToSQLKind, createToSQLOpSpec, toSQLSignature)))
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

	b, _, err := args.GetInt("batchSize")
	if err != nil {
		return err
	}
	if b <= 0 {
		// set default as argument we not supplied
		o.BatchSize = DefaultBatchSize
	} else {
		o.BatchSize = int(b)
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
			BatchSize:      s.BatchSize,
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
	deps := flux.GetDependencies(a.Context())
	t, err := NewToSQLTransformation(d, deps, cache, s)
	if err != nil {
		return nil, nil, err
	}
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

func NewToSQLTransformation(d execute.Dataset, deps flux.Dependencies, cache execute.TableBuilderCache, spec *ToSQLProcedureSpec) (*ToSQLTransformation, error) {
	validator, err := deps.URLValidator()
	if err != nil {
		return nil, err
	}
	if err := validateDataSource(validator, spec.Spec.DriverName, spec.Spec.DataSourceName); err != nil {
		return nil, err
	}

	// validate the data driver name and source name.
	db, err := sql.Open(spec.Spec.DriverName, spec.Spec.DataSourceName)
	if err != nil {
		return nil, err
	}
	var tx *sql.Tx
	if spec.Spec.DriverName != "sqlmock" {
		tx, err = db.Begin()
		if err != nil {
			return nil, err
		}
	}
	return &ToSQLTransformation{
		d:     d,
		cache: cache,
		spec:  spec,
		db:    db,
		tx:    tx,
	}, nil
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
			return err
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
			err = errors.Wrap(err, codes.Inherit, txErr)
		}
		if dbErr := t.db.Close(); dbErr != nil {
			err = errors.Wrap(err, codes.Inherit, dbErr)
		}
	}
	t.d.Finish(err)
}

type translationFunc func(f flux.ColType, colname string) (string, error)

func correctBatchSize(batchSize, numberCols int) int {
	/*
		BatchSize for the DB is the number of parameters that can be queued within each call to Exec.
		As each row you send has a parameter count equal to the number of columns (i.e. the number of "?" used in the insert statement), and some DBs
		have a default limit on the number of parameters which can be queued before calling Exec; SQLite, for example, has a default of 999 (can only be changed
		at compile time).

		So if the row width is 10 columns, the maximum Batchsize would be:

		(999 - row_width) / row_width = 98 rows. (with 0.9 of a row unused)

		and,

		(1000 - row_width) / row_width = 99 rows. (no remainder)

		NOTE: Given a statement like:

		INSERT INTO data_table (now,values,difference) VALUES(?,?,?)

		each iteration of EXEC() would add 3 new values (one for each of the '?' placeholders) - but the final "parameter count" includes the initial 3 column names.
		this is why the calculation subracts an initial "column width" from the supplied Batchsize.

		Sending more would result in the call to Exec returning a "too many SQL variables" error, and the transaction would be rolled-back / aborted
	*/

	if batchSize < numberCols {
		// if this is because the width of a single row is very large, pass to DB driver, and if this exceeds the number of allowed parameters
		// this will be fed back to the user to handle - possibly by reducing the row width
		return numberCols
	}
	return (batchSize - numberCols) / numberCols
}

func getTranslationFunc(driverName string) (func() translationFunc, error) {
	// simply return the translationFunc that corresponds to the driver type
	switch driverName {
	case "sqlite3":
		return SqliteColumnTranslateFunc, nil
	case "postgres", "sqlmock":
		return PostgresColumnTranslateFunc, nil
	case "mysql":
		return MysqlColumnTranslateFunc, nil
	default:
		return nil, errors.Newf(codes.Internal, "invalid driverName: %s", driverName)
	}

}

func CreateInsertComponents(t *ToSQLTransformation, tbl flux.Table) (colNames []string, valStringArray [][]string, valArgsArray [][]interface{}, err error) {
	cols := tbl.Cols()
	batchSize := correctBatchSize(t.spec.Spec.BatchSize, len(cols))

	labels := make(map[string]idxType, len(cols))
	var questionMarks, newSQLTableCols []string
	for i, col := range cols {
		labels[col.Label] = idxType{Idx: i, Type: col.Type}
		questionMarks = append(questionMarks, "?")
		colNames = append(colNames, col.Label)
		driverName := t.spec.Spec.DriverName
		// the following allows driver-specific type errors (of which there can be MANY) to be returned, rather than the default of invalid type
		translateColumn, err := getTranslationFunc(driverName)
		if err != nil {
			return nil, nil, nil, err
		}

		switch col.Type {
		case flux.TFloat, flux.TInt, flux.TUInt, flux.TString, flux.TBool, flux.TTime:
			// each type is handled within the function - precise mapping is handled within each driver's implementation
			v, err := translateColumn()(col.Type, col.Label)
			if err != nil {
				return nil, nil, nil, err
			}
			newSQLTableCols = append(newSQLTableCols, v)
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
					return errors.Newf(codes.FailedPrecondition, "invalid type for column %s", col.Label)
				}
			}

			if err := execute.AppendRecord(i, er, builder); err != nil {
				return err
			}

			if i != 0 && i%batchSize == 0 {
				// create "mini batches" of values - each one represents a single db.Exec to SQL
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
			// this err which is extremely helpful as it comes from the SQL driver should be
			// bubbled up further up the stack so user can see the issue
			if rbErr := tx.Rollback(); rbErr != nil {
				return errors.Newf(codes.Aborted, "transaction failed (%s) while recovering from %s", err, rbErr)
			}
			return err
		}
	}
	return err
}
