package sql

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/plan"
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
	fromSQLSignature := semantic.FunctionPolySignature{
		Parameters: map[string]semantic.PolyType{
			"driverName":     semantic.String,
			"dataSourceName": semantic.String,
			"query":          semantic.String,
		},
		Required: semantic.LabelSet{"driverName", "dataSourceName", "query"},
		Return:   flux.TableObjectType,
	}
	flux.RegisterPackageValue("sql", "from", flux.FunctionValue(FromSQLKind, createFromSQLOpSpec, fromSQLSignature))
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
		return nil, fmt.Errorf("invalid spec type %T", qs)
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
		return nil, fmt.Errorf("invalid spec type %T", prSpec)
	}

	// Allow for "sqlmock" for testing purposes in "sql_test.go"
	if spec.DriverName != "postgres" && spec.DriverName != "mysql" && spec.DriverName != "sqlmock" {
		return nil, fmt.Errorf("sql driver %s not supported", spec.DriverName)
	}

	SQLIterator := SQLIterator{id: dsid, spec: spec, administration: a}

	return execute.CreateSourceFromDecoder(&SQLIterator, dsid, a)
}

type SQLIterator struct {
	id             execute.DatasetID
	administration execute.Administration
	spec           *FromSQLProcedureSpec
	db             *sql.DB
	reader         *execute.RowReader
}

var _ execute.SourceDecoder = (*SQLIterator)(nil)

func (c *SQLIterator) Connect(ctx context.Context) error {
	db, err := sql.Open(c.spec.DriverName, c.spec.DataSourceName)

	if err != nil {
		return err
	}
	if err = db.Ping(); err != nil {
		return err
	}
	c.db = db

	return nil
}

func (c *SQLIterator) Fetch(ctx context.Context) (bool, error) {
	rows, err := c.db.Query(c.spec.Query)
	if err != nil {
		return false, err
	}

	var reader execute.RowReader
	switch c.spec.DriverName {
	case "mysql":
		reader, err = NewMySQLRowReader(rows)
	case "postgres", "sqlmock":
		reader, err = NewPostgresRowReader(rows)
	default:
		return false, fmt.Errorf("unsupported driver %s", c.spec.DriverName)
	}

	if err != nil {
		return false, err
	}
	c.reader = &reader

	return false, nil
}

func (c *SQLIterator) Decode(ctx context.Context) (flux.Table, error) {
	groupKey := execute.NewGroupKey(nil, nil)
	builder := execute.NewColListTableBuilder(groupKey, c.administration.Allocator())

	reader := *c.reader

	for i, dataType := range reader.ColumnTypes() {
		_, err := builder.AddCol(flux.ColMeta{Label: reader.ColumnNames()[i], Type: dataType})
		if err != nil {
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

	return builder.Table()
}

func (c *SQLIterator) Close() error {
	return c.db.Close()
}
