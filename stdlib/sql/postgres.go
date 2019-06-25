package sql

import (
	"database/sql"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/values"
	"time"
)

type PostgresRowReader struct {
	Cursor      *sql.Rows
	columns     []interface{}
	columnTypes []flux.ColType
	columnNames []string
}

func (m *PostgresRowReader) Next() bool {
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

func (m *PostgresRowReader) GetNextRow() ([]values.Value, error) {
	row := make([]values.Value, len(m.columns))
	for i, col := range m.columns {
		switch col := col.(type) {
		case bool, int64, uint64, float64, string:
			row[i] = values.New(col)
		case time.Time:
			row[i] = values.NewTime(values.ConvertTime(col))
		case nil:
			row[i] = values.NewNull(flux.SemanticType(m.columnTypes[i]))
		default:
			execute.PanicUnknownType(flux.TInvalid)
		}
	}
	return row, nil
}

func (m *PostgresRowReader) InitColumnNames(n []string) {
	m.columnNames = n
}

func (m *PostgresRowReader) InitColumnTypes(types []*sql.ColumnType) {
	stringTypes := make([]flux.ColType, len(types))
	for i := 0; i < len(types); i++ {
		switch types[i].DatabaseTypeName() {
		case "INT", "BIGINT", "SMALLINT", "TINYINT", "INT2", "INT4", "INT8":
			stringTypes[i] = flux.TInt
		case "FLOAT", "DOUBLE":
			stringTypes[i] = flux.TFloat
		case "DATE", "TIME", "TIMESTAMP":
			stringTypes[i] = flux.TTime
		case "BOOL":
			stringTypes[i] = flux.TBool
		default:
			stringTypes[i] = flux.TString
		}
	}
	m.columnTypes = stringTypes
}

func (m *PostgresRowReader) ColumnNames() []string {
	return m.columnNames
}

func (m *PostgresRowReader) ColumnTypes() []flux.ColType {
	return m.columnTypes
}

func (m *PostgresRowReader) SetColumns(i []interface{}) {
	m.columns = i
}

func NewPostgresRowReader(r *sql.Rows) (execute.RowReader, error) {
	reader := &PostgresRowReader{
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
