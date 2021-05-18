package sql

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/values"
)

type SqliteRowReader struct {
	Cursor      *sql.Rows
	columns     []interface{}
	columnTypes []flux.ColType
	columnNames []string
	NextFunc    func() bool
}

func (m *SqliteRowReader) Next() bool {
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

func (m *SqliteRowReader) GetNextRow() ([]values.Value, error) {
	row := make([]values.Value, len(m.columns))
	for i, col := range m.columns {
		switch col := col.(type) {
		case bool, int64, uint64, float64, string:
			row[i] = values.New(col)
		case []uint8:
			// this allows easier testing using existing methods
			switch m.columnTypes[i] {
			case flux.TInt:
				newInt, err := UInt8ToInt64(col)
				if err != nil {
					return nil, err
				}
				row[i] = values.NewInt(newInt)
			case flux.TFloat:
				newFloat, err := UInt8ToFloat(col)
				if err != nil {
					return nil, err
				}
				row[i] = values.NewFloat(newFloat)
			case flux.TTime:
				t, err := time.Parse(layout, string(col))
				if err != nil {
					fmt.Print(err)
				}
				row[i] = values.NewTime(values.ConvertTime(t))
			default:
				row[i] = values.NewString(string(col))
			}
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

func (m *SqliteRowReader) InitColumnNames(n []string) {
	m.columnNames = n
}

func (m *SqliteRowReader) InitColumnTypes(types []*sql.ColumnType) {
	stringTypes := make([]flux.ColType, len(types))
	for i := 0; i < len(types); i++ {
		switch types[i].DatabaseTypeName() {
		case "INT", "INTEGER", "BIGINT", "SMALLINT", "TINYINT":
			stringTypes[i] = flux.TInt
		case "FLOAT", "DOUBLE":
			stringTypes[i] = flux.TFloat
		case "DATETIME", "TIMESTAMP", "DATE":
			stringTypes[i] = flux.TTime
		case "TEXT":
			stringTypes[i] = flux.TString
		case "BOOL", "BOOLEAN":
			stringTypes[i] = flux.TInt
		default:
			stringTypes[i] = flux.TString
		}
	}
	m.columnTypes = stringTypes
}

func (m *SqliteRowReader) ColumnNames() []string {
	return m.columnNames
}

func (m *SqliteRowReader) ColumnTypes() []flux.ColType {
	return m.columnTypes
}

func (m *SqliteRowReader) SetColumnTypes(types []flux.ColType) {
	m.columnTypes = types
}

func (m *SqliteRowReader) SetColumns(i []interface{}) {
	m.columns = i
}

func (m *SqliteRowReader) Close() error {
	if err := m.Cursor.Err(); err != nil {
		return err
	}
	return m.Cursor.Close()
}

func NewSqliteRowReader(r *sql.Rows) (execute.RowReader, error) {
	reader := &SqliteRowReader{
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

var fluxToSQLite = map[flux.ColType]string{
	flux.TFloat:  "FLOAT",
	flux.TInt:    "INT",
	flux.TUInt:   "INT",
	flux.TString: "TEXT",
	flux.TTime:   "DATETIME",
}

// SqliteTranslateColumn translates flux colTypes into their corresponding SQLite column type
func SqliteColumnTranslateFunc() translationFunc {
	return func(f flux.ColType, colName string) (string, error) {
		s, found := fluxToSQLite[f]
		if !found {
			return "", errors.Newf(codes.Invalid, "SQLite does not support column type %s", f.String())
		}
		return colName + " " + s, nil
	}
}
