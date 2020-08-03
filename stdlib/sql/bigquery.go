package sql

import (
	"database/sql"
	"math/big"
	"time"

	"cloud.google.com/go/civil"
	_ "github.com/bonitoo-io/go-sql-bigquery"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/values"
)

// Google BigQuery support.
// Notes:
// * data types
//   https://cloud.google.com/bigquery/docs/reference/standard-sql/data-types
// * connection string
//   https://github.com/bonitoo-io/go-sql-bigquery#connection-string

type BigQueryRowReader struct {
	Cursor      *sql.Rows
	columns     []interface{}
	columnTypes []flux.ColType
	columnNames []string
	sqlTypes    []*sql.ColumnType
	NextFunc    func() bool
	CloseFunc   func() error
}

// Next prepares BigQueryRowReader to return rows
func (m *BigQueryRowReader) Next() bool {
	if m.NextFunc != nil {
		return m.NextFunc()
	}
	next := m.Cursor.Next()
	if next {
		columnNames, err := m.Cursor.Columns()
		if err != nil {
			return false
		}
		m.columns = make([]interface{}, len(columnNames))
		columnPointers := make([]interface{}, len(columnNames))
		for i := 0; i < len(columnNames); i++ {
			columnPointers[i] = &m.columns[i]
		}
		if err := m.Cursor.Scan(columnPointers...); err != nil {
			return false
		}
	}
	return next
}

func (m *BigQueryRowReader) GetNextRow() ([]values.Value, error) {
	row := make([]values.Value, len(m.columns))
	for i, column := range m.columns {
		switch value := column.(type) {
		case bool, int64, float64, string:
			row[i] = values.New(value)
		case civil.Date:
			row[i] = values.NewString(value.String())
		case civil.DateTime:
			row[i] = values.NewString(value.String())
		case civil.Time:
			row[i] = values.NewString(value.String())
		case time.Time:
			row[i] = values.NewTime(values.ConvertTime(value))
		case *big.Int:
			row[i] = values.NewInt(value.Int64())
		case *big.Float:
			f, _ := value.Float64()
			row[i] = values.NewFloat(f)
		case *big.Rat:
			f, _ := value.Float64()
			row[i] = values.NewFloat(f)
		case nil:
			row[i] = values.NewNull(flux.SemanticType(m.columnTypes[i]))
		default:
			execute.PanicUnknownType(flux.TInvalid)
		}
	}
	return row, nil
}

func (m *BigQueryRowReader) InitColumnNames(names []string) {
	m.columnNames = names
}

func (m *BigQueryRowReader) InitColumnTypes(types []*sql.ColumnType) {
	fluxTypes := make([]flux.ColType, len(types))
	for i := 0; i < len(types); i++ {
		switch types[i].DatabaseTypeName() {
		case "INTEGER":
			fluxTypes[i] = flux.TInt
		case "FLOAT", "NUMERIC":
			fluxTypes[i] = flux.TFloat
		case "BOOLEAN":
			fluxTypes[i] = flux.TBool
		case "TIMESTAMP": // "DATE", "TIME" and "DATETIME" will be represented as string because TZ is unknown
			fluxTypes[i] = flux.TTime
		default:
			fluxTypes[i] = flux.TString
		}
	}
	m.columnTypes = fluxTypes
	m.sqlTypes = types
}

func (m *BigQueryRowReader) ColumnNames() []string {
	return m.columnNames
}

func (m *BigQueryRowReader) ColumnTypes() []flux.ColType {
	return m.columnTypes
}

func (m *BigQueryRowReader) SetColumnTypes(types []flux.ColType) {
	m.columnTypes = types
}

func (m *BigQueryRowReader) SetColumns(i []interface{}) {
	m.columns = i
}

func (m *BigQueryRowReader) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	if err := m.Cursor.Err(); err != nil {
		return err
	}
	return m.Cursor.Close()
}

func NewBigQueryRowReader(r *sql.Rows) (execute.RowReader, error) {
	reader := &BigQueryRowReader{
		Cursor: r,
	}
	cols, err := r.Columns()
	if err != nil {
		return nil, err
	}
	reader.InitColumnNames(cols)

	types, err := r.ColumnTypes()
	if err != nil {
		return nil, err
	}
	reader.InitColumnTypes(types)

	return reader, nil
}

var fluxToBigQuery = map[flux.ColType]string{
	flux.TFloat:  "FLOAT64",
	flux.TInt:    "INT64",
	flux.TString: "STRING",
	flux.TBool:   "BOOL",
	flux.TTime:   "TIMESTAMP",
}

// BigQueryTranslateColumn translates flux colTypes into their corresponding BigQuery column type
func BigQueryColumnTranslateFunc() translationFunc {
	return func(f flux.ColType, colName string) (string, error) {
		s, found := fluxToBigQuery[f]
		if !found {
			return "", errors.Newf(codes.Internal, "BigQuery does not support column type %s", f.String())
		}
		return colName + " " + s, nil
	}
}
