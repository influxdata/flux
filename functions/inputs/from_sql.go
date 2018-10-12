package inputs

import (
	"database/sql"
	"fmt"

	"reflect"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	plan "github.com/influxdata/flux/planner"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
	_ "github.com/lib/pq"
)

const FromSQLKind = "fromSQL"

type FromSQLOpSpec struct {
	DriverName     string `json:"driverName,omitempty"`
	DataSourceName string `json:"dataSourceName,omitempty"`
	Query          string `json:"query,omitempty"`
}

var fromSQLSignature = semantic.FunctionSignature{
	Params: map[string]semantic.Type{
		"driverName":     semantic.String,
		"dataSourceName": semantic.String,
		"query":          semantic.String,
	},
	ReturnType: flux.TableObjectType,
}

func init() {
	flux.RegisterFunction(FromSQLKind, createFromSQLOpSpec, fromSQLSignature)
	flux.RegisterOpSpec(FromSQLKind, newFromSQLOp)
	plan.RegisterProcedureSpec(FromSQLKind, newFromSQLProcedure, FromSQLKind)
	execute.RegisterSource(FromSQLKind, createFromSQLSource)
}

func createFromSQLOpSpec(args flux.Arguments, administration *flux.Administration) (flux.OperationSpec, error) {
	spec := new(FromSQLOpSpec)

	if driverName, err := args.GetRequiredString("driverName"); err != nil {
		return nil, err
	} else  {
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

	if spec.DriverName != "postgres" && spec.DriverName != "mysql" {
		return nil, fmt.Errorf("sql driver %s not supported", spec.DriverName)
	}

	SQLIterator := SQLIterator{id: dsid, spec: spec, administration: a}

	return CreateSourceFromDecoder(&SQLIterator, dsid, a)
}

type SQLIterator struct {
	id             execute.DatasetID
	data           flux.Result
	ts             []execute.Transformation
	administration execute.Administration
	spec           *FromSQLProcedureSpec
	db             *sql.DB
	rows           *sql.Rows
}

func (c *SQLIterator) Connect() error {
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

func (c *SQLIterator) Fetch() (bool, error) {
	rows, err := c.db.Query(c.spec.Query)
	if err != nil {
		return false, err
	}
	c.rows = rows

	return false, nil
}

func (c *SQLIterator) Decode() (flux.Table, error) {
	groupKey := execute.NewGroupKey(nil, nil)
	builder := execute.NewColListTableBuilder(groupKey, c.administration.Allocator())

	firstRow := true
	for c.rows.Next() {
		columnNames, err := c.rows.Columns()
		if err != nil {
			return nil, err
		}
		columns := make([]interface{}, len(columnNames))
		columnPointers := make([]interface{}, len(columnNames))
		for i := 0; i < len(columnNames); i++ {
			columnPointers[i] = &columns[i]
		}
		if err := c.rows.Scan(columnPointers...); err != nil {
			return nil, err
		}

		if firstRow {
			for i, col := range columns {
				var dataType flux.DataType
				switch col.(type) {
				case bool:
					dataType = flux.TBool
				case int64:
					dataType = flux.TInt
				case uint64:
					dataType = flux.TUInt
				case float64:
					dataType = flux.TFloat
				case string:
					dataType = flux.TString
				case []uint8:
					// Hack for MySQL, might need to work with charset?
					dataType = flux.TString
				case time.Time:
					dataType = flux.TTime
				default:
					fmt.Println(i, reflect.TypeOf(col))
					execute.PanicUnknownType(flux.TInvalid)
				}

				builder.AddCol(flux.ColMeta{Label: columnNames[i], Type: dataType})
			}
			firstRow = false
		}

		for i, col := range columns {
			switch col.(type) {
			case bool:
				builder.AppendBool(i, col.(bool))
			case int64:
				builder.AppendInt(i, col.(int64))
			case uint64:
				builder.AppendUInt(i, col.(uint64))
			case float64:
				builder.AppendFloat(i, col.(float64))
			case string:
				builder.AppendString(i, col.(string))
			case []uint8:
				// Hack for MySQL, might need to work with charset?
				builder.AppendString(i, string(col.([]uint8)))
			case time.Time:
				builder.AppendTime(i, values.ConvertTime(col.(time.Time)))
			default:
				execute.PanicUnknownType(flux.TInvalid)
			}
		}
	}

	return builder.Table()
}
