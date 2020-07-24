package sql

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/values"
)

type MySQLRowReader struct {
	Cursor      *sql.Rows
	columns     []interface{}
	columnTypes []flux.ColType
	columnNames []string
	NextFunc    func() bool
	CloseFunc   func() error
}

// Next prepares MySQLRowReader to return rows
func (m *MySQLRowReader) Next() bool {
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

func (m *MySQLRowReader) GetNextRow() ([]values.Value, error) {
	row := make([]values.Value, len(m.columns))
	for i, col := range m.columns {
		switch col := col.(type) {
		case bool, int64, uint64, float64, string:
			row[i] = values.New(col)
		case []uint8:
			// Hack for MySQL, might need to work with charset?
			// Can't do boolean with MySQL - stores BOOLEANs as TINYINTs (0 or 1)
			// No way to distinguish if intended int or bool
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
			// This works, but you can also just just add the DSN parameter parseTime=true (see line 136)
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

func (m *MySQLRowReader) InitColumnNames(names []string) {
	m.columnNames = names
}

func (m *MySQLRowReader) InitColumnTypes(types []*sql.ColumnType) {
	stringTypes := make([]flux.ColType, len(types))
	for i := 0; i < len(types); i++ {
		switch types[i].DatabaseTypeName() {
		case "INT", "BIGINT", "SMALLINT", "TINYINT":
			stringTypes[i] = flux.TInt
		case "FLOAT", "DOUBLE":
			stringTypes[i] = flux.TFloat
		case "DATETIME":
			stringTypes[i] = flux.TTime
		default:
			stringTypes[i] = flux.TString
		}
	}
	m.columnTypes = stringTypes
}

func (m *MySQLRowReader) ColumnNames() []string {
	return m.columnNames
}

func (m *MySQLRowReader) ColumnTypes() []flux.ColType {
	return m.columnTypes
}

func (m *MySQLRowReader) SetColumnTypes(types []flux.ColType) {
	m.columnTypes = types
}

func (m *MySQLRowReader) SetColumns(i []interface{}) {
	m.columns = i
}

func (m *MySQLRowReader) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	if err := m.Cursor.Err(); err != nil {
		return err
	}
	return m.Cursor.Close()
}

func UInt8ToFloat(a []uint8) (float64, error) {
	str := string(a)
	s, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return s, err
	}
	return s, nil
}

func UInt8ToInt64(a []uint8) (int64, error) {
	str := string(a)
	s, err := strconv.ParseInt(str, 0, 64)
	if err != nil {
		return s, err
	}
	return s, nil
}

func NewMySQLRowReader(r *sql.Rows) (execute.RowReader, error) {
	reader := &MySQLRowReader{
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

var fluxToMySQL = map[flux.ColType]string{
	flux.TFloat:  "FLOAT",
	flux.TInt:    "BIGINT",
	flux.TUInt:   "BIGINT UNSIGNED",
	flux.TString: "TEXT(16383)",
	flux.TTime:   "DATETIME",
	flux.TBool:   "BOOL",
	// BOOL is a synonym supplied by MySQL for "convenience", and MYSQL turns this into a TINYINT type under the hood
	// which means that looking at the schema afterwards shows the columntype as TINYINT, and not bool!
}

// MysqlTranslateColumn translates flux colTypes into their corresponding MySQL column type
func MysqlColumnTranslateFunc() translationFunc {
	return func(f flux.ColType, colName string) (string, error) {
		s, found := fluxToMySQL[f]
		if !found {
			return "", errors.Newf(codes.Internal, "MySQL does not support column type %s", f.String())
		}
		return colName + " " + s, nil
	}
}
