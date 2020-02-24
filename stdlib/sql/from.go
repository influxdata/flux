package sql

import (
	"context"
	"database/sql"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/runtime"
	"github.com/influxdata/flux/semantic"
	_ "github.com/lib/pq"
)

const FromSQLKind = "fromSQL"

// For SQL DATETIME parsing
const layout = "2006-01-02 15:04:05.999999999"

type FromSQLOpSpec struct {
	DriverName     string `json:"driverName,omitempty"`
	DataSourceName string `json:"dataSourceName,omitempty"`
	Query          string `json:"query,omitempty"`
}

func init() {
	fromSQLSignature := semantic.MustLookupBuiltinType("sql", "from")
	runtime.RegisterPackageValue("sql", "from", flux.MustValue(flux.FunctionValue(FromSQLKind, createFromSQLOpSpec, fromSQLSignature)))
	flux.RegisterOpSpec(FromSQLKind, newFromSQLOp)
	plan.RegisterProcedureSpec(FromSQLKind, newFromSQLProcedure, FromSQLKind)
	execute.RegisterSource(FromSQLKind, createFromSQLSource)
}

func createFromSQLOpSpec(args flux.Arguments, administration *flux.Administration) (flux.OperationSpec, error) {
	spec := new(FromSQLOpSpec)

	if driverName, err := args.GetRequiredString("driverName"); err != nil {
		return nil, err
	} else {
		spec.DriverName = driverName
	}
	if dataSourceName, err := args.GetRequiredString("dataSourceName"); err != nil {
		return nil, err
	} else {
		spec.DataSourceName = dataSourceName
	}
	if query, err := args.GetRequiredString("query"); err != nil {
		return nil, err
	} else {
		spec.Query = query
	}
	return spec, nil
}

func newFromSQLOp() flux.OperationSpec {
	return new(FromSQLOpSpec)
}

func (s *FromSQLOpSpec) Kind() flux.OperationKind {
	return FromSQLKind
}

type FromSQLProcedureSpec struct {
	plan.DefaultCost
	DriverName     string
	DataSourceName string
	Query          string
}

func newFromSQLProcedure(qs flux.OperationSpec, pa plan.Administration) (plan.ProcedureSpec, error) {
	spec, ok := qs.(*FromSQLOpSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", qs)
	}

	return &FromSQLProcedureSpec{
		DriverName:     spec.DriverName,
		DataSourceName: spec.DataSourceName,
		Query:          spec.Query,
	}, nil
}

func (s *FromSQLProcedureSpec) Kind() plan.ProcedureKind {
	return FromSQLKind
}

func (s *FromSQLProcedureSpec) Copy() plan.ProcedureSpec {
	ns := new(FromSQLProcedureSpec)
	ns.DriverName = s.DriverName
	ns.DataSourceName = s.DataSourceName
	ns.Query = s.Query
	return ns
}

func createFromSQLSource(prSpec plan.ProcedureSpec, dsid execute.DatasetID, a execute.Administration) (execute.Source, error) {
	spec, ok := prSpec.(*FromSQLProcedureSpec)
	if !ok {
		return nil, errors.Newf(codes.Internal, "invalid spec type %T", prSpec)
	}

	// validate the data driver name and source name.
	deps := flux.GetDependencies(a.Context())
	validator, err := deps.URLValidator()
	if err != nil {
		return nil, err
	}
	if err := validateDataSource(validator, spec.DriverName, spec.DataSourceName); err != nil {
		return nil, err
	}

	// Retrieve the row reader implementation for the driver.
	var newRowReader func(rows *sql.Rows) (execute.RowReader, error)
	switch spec.DriverName {
	case "mysql":
		newRowReader = NewMySQLRowReader
	case "sqlite3":
		newRowReader = NewSqliteRowReader
	case "postgres", "sqlmock":
		newRowReader = NewPostgresRowReader
	default:
		return nil, errors.Newf(codes.Invalid, "sql driver %s not supported", spec.DriverName)
	}

	readFn := func(ctx context.Context, rows *sql.Rows) (flux.Table, error) {
		reader, err := newRowReader(rows)
		if err != nil {
			_ = rows.Close()
			return nil, err
		}
		return read(ctx, reader, a.Allocator())
	}
	iterator := &sqlIterator{spec: spec, id: dsid, read: readFn}
	return execute.CreateSourceFromIterator(iterator, dsid)
}

var _ execute.SourceIterator = (*sqlIterator)(nil)

type sqlIterator struct {
	spec *FromSQLProcedureSpec
	id   execute.DatasetID
	read func(ctx context.Context, rows *sql.Rows) (flux.Table, error)
}

func (c *sqlIterator) connect(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open(c.spec.DriverName, c.spec.DataSourceName)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func (c *sqlIterator) Do(ctx context.Context, f func(flux.Table) error) error {
	// Connect to the database so we can execute the query.
	db, err := c.connect(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	rows, err := db.QueryContext(ctx, c.spec.Query)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()

	table, err := c.read(ctx, rows)
	if err != nil {
		return err
	}
	return f(table)
}

// read will use the RowReader to construct a flux.Table.
func read(ctx context.Context, reader execute.RowReader, alloc *memory.Allocator) (flux.Table, error) {
	// Ensure that the reader is always freed so the underlying
	// cursor can be returned.
	defer func() { _ = reader.Close() }()

	groupKey := execute.NewGroupKey(nil, nil)
	builder := execute.NewColListTableBuilder(groupKey, alloc)
	for i, dataType := range reader.ColumnTypes() {
		if _, err := builder.AddCol(flux.ColMeta{Label: reader.ColumnNames()[i], Type: dataType}); err != nil {
			return nil, err
		}
	}
	for reader.Next() {
		row, err := reader.GetNextRow()
		if err != nil {
			return nil, err
		}

		for i, col := range row {
			if err := builder.AppendValue(i, col); err != nil {
				return nil, err
			}
		}
	}

	// An error may have been encountered while reading.
	// This will get reported when we go to close the reader.
	if err := reader.Close(); err != nil {
		return nil, err
	}
	return builder.Table()
}
