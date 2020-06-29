package sql

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/values"
)

// Snowflake DB support.
// Notes:
// * type mapping
//     - see https://pkg.go.dev/github.com/snowflakedb/gosnowflake
//     - current mappings are valid for v1.3.4

type SnowflakeRowReader struct {
	Cursor      *sql.Rows
	columns     []interface{}
	columnTypes []flux.ColType
	columnNames []string
	sqlTypes    []*sql.ColumnType
	NextFunc    func() bool
	CloseFunc   func() error
}

const (
	layoutDate         = "2006-01-02"
	layoutTime         = "15:04:05"
	layoutTimeStampNtz = "2006-01-02T15:04:05.0000000000"
)

// Next prepares SnowflakeRowReader to return rows
func (m *SnowflakeRowReader) Next() bool {
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

func (m *SnowflakeRowReader) GetNextRow() ([]values.Value, error) {
	row := make([]values.Value, len(m.columns))
	for i, column := range m.columns {
		switch value := column.(type) {
		case bool, int64, float64: // never happens with scan into []*interface{}
			row[i] = values.New(value)
		case string:
			switch m.columnTypes[i] {
			case flux.TFloat:
				f, err := strconv.ParseFloat(value, 64)
				if err != nil {
					return nil, err
				}
				row[i] = values.NewFloat(f)
			case flux.TInt:
				d, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					return nil, err
				}
				row[i] = values.NewInt(d)
			case flux.TBool:
				b, err := strconv.ParseBool(value)
				if err != nil {
					return nil, err
				}
				row[i] = values.NewBool(b)
			default:
				row[i] = values.New(value)
			}
		case time.Time:
			// DATE, TIME and TIMESTAMP_NTZ types get scanned to time.Time by the driver,
			// but they have no counterpart in Flux therefore will be represented as string
			switch m.sqlTypes[i].DatabaseTypeName() {
			case "DATE":
				row[i] = values.NewString(value.Format(layoutDate))
			case "TIME":
				row[i] = values.NewString(value.Format(layoutTime))
			case "TIMESTAMP_NTZ":
				row[i] = values.NewString(value.Format(layoutTimeStampNtz))
			default:
				row[i] = values.NewTime(values.ConvertTime(value))
			}
		case nil:
			row[i] = values.NewNull(flux.SemanticType(m.columnTypes[i]))
		default:
			execute.PanicUnknownType(flux.TInvalid)
		}
	}
	return row, nil
}

func (m *SnowflakeRowReader) InitColumnNames(names []string) {
	m.columnNames = names
}

func (m *SnowflakeRowReader) InitColumnTypes(types []*sql.ColumnType) {
	fluxTypes := make([]flux.ColType, len(types))
	for i := 0; i < len(types); i++ {
		switch types[i].DatabaseTypeName() {
		case "FIXED", "NUMBER": // FIXED is reported by Snowflake driver
			_, scale, ok := types[i].DecimalSize()
			if ok && scale > 0 {
				fluxTypes[i] = flux.TFloat
			} else {
				fluxTypes[i] = flux.TInt
			}
		case "REAL", "FLOAT": // REAL is reported by Snowflake driver
			fluxTypes[i] = flux.TFloat
		case "BOOLEAN":
			fluxTypes[i] = flux.TBool
		case "TIMESTAMP_TZ", "TIMESTAMP_LTZ": // "TIMESTAMP_NTZ", "DATE" and "TIME" will be represented as string
			fluxTypes[i] = flux.TTime
		default:
			fluxTypes[i] = flux.TString
		}
	}
	m.columnTypes = fluxTypes
	m.sqlTypes = types
}

func (m *SnowflakeRowReader) ColumnNames() []string {
	return m.columnNames
}

func (m *SnowflakeRowReader) ColumnTypes() []flux.ColType {
	return m.columnTypes
}

func (m *SnowflakeRowReader) SetColumnTypes(types []flux.ColType) {
	m.columnTypes = types
}

func (m *SnowflakeRowReader) SetColumns(i []interface{}) {
	m.columns = i
}

func (m *SnowflakeRowReader) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	if err := m.Cursor.Err(); err != nil {
		return err
	}
	return m.Cursor.Close()
}

func NewSnowflakeRowReader(r *sql.Rows) (execute.RowReader, error) {
	reader := &SnowflakeRowReader{
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

var fluxToSnowflake = map[flux.ColType]string{
	flux.TFloat:  "FLOAT",
	flux.TInt:    "NUMBER",
	flux.TUInt:   "NUMBER",
	flux.TString: "TEXT",
	flux.TBool:   "BOOLEAN",
	flux.TTime:   "TIMESTAMP_LTZ",
}

// SnowflakeTranslateColumn translates flux colTypes into their corresponding Snowflake column type
func SnowflakeColumnTranslateFunc() translationFunc {
	return func(f flux.ColType, colName string) (string, error) {
		s, found := fluxToSnowflake[f]
		if !found {
			return "", errors.Newf(codes.Internal, "Snowflake does not support column type %s", f.String())
		}
		return colName + " " + s, nil
	}
}
