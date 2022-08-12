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
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/values"
)

const (
	ToSQLKind        = "toSQL"
	DefaultBatchSize = 10000 // TODO: decide if this should be kept low enough for the lowest (SQLite), or not.
)

type ToSQLOpSpec struct {
	DriverName     string `json:"driverName,omitempty"`
	DataSourceName string `json:"dataSourcename,omitempty"`
	Table          string `json:"table,omitempty"`
	BatchSize      int    `json:"batchSize,omitempty"`
}

func init() {
	toSQLSignature := runtime.MustLookupBuiltinType("sql", "to")
	runtime.RegisterPackageValue("sql", "to", flux.MustValue(flux.FunctionValueWithSideEffect(ToSQLKind, createToSQLOpSpec, toSQLSignature)))
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
	execute.ExecutionNode
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
	db, err := getOpenFunc(spec.Spec.DriverName, spec.Spec.DataSourceName)()
	if err != nil {
		return nil, err
	}
	var tx *sql.Tx
	if supportsTx(spec.Spec.DriverName) {
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
	if supportsTx(t.spec.Spec.DriverName) {
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

// quoteIdentFunc is used to quote identifiers like table and column names for a
// given SQL dialect.
type quoteIdentFunc func(name string) string

// doubleQuote wraps the input in double quotes and escapes any interior quotes.
// If the input contains an interior nul byte, it will be truncated.
// Useful for quoting _table or column identifiers_ for many database engines.
func doubleQuote(s string) string {
	end := strings.IndexRune(s, 0)
	if end > -1 {
		s = s[:end]
	}
	return fmt.Sprintf(`"%s"`, strings.ReplaceAll(s, `"`, `""`))
}

// singleQuote wraps the input in single quotes and escapes any interior quotes.
// If the input contains an interior nul byte, it will be truncated.
// Useful for producing _string literals_ for many database engines.
func singleQuote(s string) string {
	end := strings.IndexRune(s, 0)
	if end > -1 {
		s = s[:end]
	}
	return fmt.Sprintf("'%s'", strings.ReplaceAll(s, "'", "''"))
}

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
	case "vertica", "vertigo":
		return VerticaColumnTranslateFunc, nil
	case "mysql":
		return MysqlColumnTranslateFunc, nil
	case "snowflake":
		return SnowflakeColumnTranslateFunc, nil
	case "mssql", "sqlserver":
		return MssqlColumnTranslateFunc, nil
	case "awsathena": // read-only support for AWS Athena (see awsathena.go)
		return nil, errors.Newf(codes.Invalid, "writing is not supported for %s", driverName)
	case "bigquery":
		return BigQueryColumnTranslateFunc, nil
	case "hdb":
		return HdbColumnTranslateFunc, nil
	default:
		return nil, errors.Newf(codes.Internal, "invalid driverName: %s", driverName)
	}
}

func getQuoteIdentFunc(driverName string) (quoteIdentFunc, error) {
	switch driverName {
	case "sqlite3":
		return doubleQuote, nil
	case "postgres", "sqlmock":
		return postgresQuoteIdent, nil
	case "vertica", "vertigo":
		return doubleQuote, nil
	case "mysql":
		return mysqlQuoteIdent, nil
	case "snowflake":
		// n.b. snowflake automatically UPPERCASES identifiers when they are
		// specified as bare words in queries (and DDL). Case is preserved when
		// identifiers are quoted.
		// Therefore, snowflake users may see breakage with this quoting behavior
		// since they will now need to explicitly match the case of the
		// identifiers already defined in schema.
		return doubleQuote, nil
	case "mssql", "sqlserver":
		return doubleQuote, nil
	case "awsathena": // read-only support for AWS Athena (see awsathena.go)
		return nil, errors.Newf(codes.Invalid, "writing is not supported for %s", driverName)
	case "bigquery":
		// BigQuery offers 2 dialects, "legacy" and "standard."
		// For the "legacy" dialect, it seems to use MS-style quoting with
		// square brackets.
		// The "standard" dialect (which is the default) seems to use backticks
		// in the style of MySQL (which makes sense for Google from a product
		// standpoint since their "Cloud SQL" product also speaks this dialect).
		return mysqlQuoteIdent, nil
	case "hdb":
		return func(name string) string { return hdbEscapeName(name, true) }, nil
	default:
		return nil, errors.Newf(codes.Internal, "invalid driverName: %s", driverName)
	}
}

func supportsTx(driverName string) bool {
	return driverName != "sqlmock" && driverName != "awsathena"
}

func CreateInsertComponents(t *ToSQLTransformation, tbl flux.Table) (colNames []string, valStringArray [][]string, valArgsArray [][]interface{}, err error) {
	cols := tbl.Cols()
	batchSize := correctBatchSize(t.spec.Spec.BatchSize, len(cols))
	driverName := t.spec.Spec.DriverName

	quoteIdent, err := getQuoteIdentFunc(driverName)
	if err != nil {
		return nil, nil, nil, err
	}
	// the following allows driver-specific type errors (of which there can be MANY) to be returned, rather than the default of invalid type
	translateColumn, err := getTranslationFunc(driverName)
	if err != nil {
		return nil, nil, nil, err
	}

	labels := make(map[string]idxType, len(cols))
	var questionMarks, newSQLTableCols []string
	for i, col := range cols {
		labels[col.Label] = idxType{Idx: i, Type: col.Type}
		questionMarks = append(questionMarks, "?")
		colNames = append(colNames, col.Label)

		switch col.Type {
		case flux.TFloat, flux.TInt, flux.TUInt, flux.TString, flux.TBool, flux.TTime:
			// Each type is handled within the function - precise mapping is
			// handled within each driver's implementation.
			// The expectation is the identifiers in these values are
			// quoted/escaped making them safe for formatting into SQL below.
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
			var q string
			if isMssqlDriver(t.spec.Spec.DriverName) { // SQL Server does not support IF NOT EXIST
				q = fmt.Sprintf("IF OBJECT_ID(%s, 'U') IS NULL BEGIN CREATE TABLE %s (%s) END",
					singleQuote(t.spec.Spec.Table),
					quoteIdent(t.spec.Spec.Table),
					// XXX: Items in `newSQLTableCols` should include _quoted column identifiers_, ref: influxdata/idpe#8689
					strings.Join(newSQLTableCols, ","),
				)
			} else if t.spec.Spec.DriverName == "hdb" { // SAP HANA does not support IF NOT EXIST
				// wrap CREATE TABLE statement with HDB-specific "if not exists" SQLScript check
				q = fmt.Sprintf(
					"CREATE TABLE %s (%s)",
					hdbEscapeName(t.spec.Spec.Table, true),
					// XXX: Items in `newSQLTableCols` should include _quoted column identifiers_, ref: influxdata/idpe#8689
					strings.Join(newSQLTableCols, ","),
				)
				// The table name we pass to `hdbAddIfNotExist` cannot be escaped
				// using `hdbEscapeName` here since it needs to appear as both a
				// string literal and a quoted identifier in the SQL generated within.
				q = hdbAddIfNotExist(t.spec.Spec.Table, q)
				// SAP HANA does not support INSERT/UPDATE batching via a single SQL command
				batchSize = 1
			} else {
				q = fmt.Sprintf(
					"CREATE TABLE IF NOT EXISTS %s (%s)",
					quoteIdent(t.spec.Spec.Table),
					// XXX: Items in `newSQLTableCols` should include _quoted column identifiers_, ref: influxdata/idpe#8689
					strings.Join(newSQLTableCols, ","),
				)
			}
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
					valueArgs = append(valueArgs, er.Strings(j).Value(i))
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

			if (i != 0 && i%batchSize == 0) || (batchSize == 1) {
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

// ExecuteQueries runs the SQL statements required to insert the new rows.
func ExecuteQueries(tx *sql.Tx, s *ToSQLOpSpec, colNames []string, valueStrings *[]string, valueArgs *[]interface{}) (err error) {
	concatValueStrings := strings.Join(*valueStrings, ",")

	quoteIdent, err := getQuoteIdentFunc(s.DriverName)
	if err != nil {
		return err
	}

	// PostgreSQL uses $n instead of ? for placeholders
	if s.DriverName == "postgres" {
		for pqCounter := 1; strings.Contains(concatValueStrings, "?"); pqCounter++ {
			concatValueStrings = strings.Replace(concatValueStrings, "?", fmt.Sprintf("$%v", pqCounter), 1)
		}
	}
	// SQLServer uses @p instead of ? for placeholders
	if isMssqlDriver(s.DriverName) {
		for pqCounter := 1; strings.Contains(concatValueStrings, "?"); pqCounter++ {
			concatValueStrings = strings.Replace(concatValueStrings, "?", fmt.Sprintf("@p%v", pqCounter), 1)
		}
	}

	// N.B. identifiers that will be string formatted into SQL statements must be
	// quoted/escaped, ref: influxdata/idpe#8689
	quotedTable := quoteIdent(s.Table)
	quotedColNames := make([]string, len(colNames))
	for idx, name := range colNames {
		quotedColNames[idx] = quoteIdent(name)
	}
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", quotedTable, strings.Join(quotedColNames, ","), concatValueStrings)

	if isMssqlDriver(s.DriverName) && mssqlCheckParameter(s.DataSourceName, mssqlIdentityInsertEnabled) {
		// XXX: identifiers that will be string formatted into SQL statements must be quoted, ref: influxdata/idpe#8689
		prologue := fmt.Sprintf(
			"SET QUOTED_IDENTIFIER ON; DECLARE @tableHasIdentity INT = OBJECTPROPERTY(OBJECT_ID(%s), 'TableHasIdentity'); IF @tableHasIdentity = 1 BEGIN SET IDENTITY_INSERT %s ON END",
			singleQuote(s.Table),
			quotedTable,
		)
		epilogue := fmt.Sprintf("IF @tableHasIdentity = 1 BEGIN SET IDENTITY_INSERT %s OFF END", quotedTable)
		query = strings.Join([]string{prologue, query, epilogue}, "; ")
	}
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
