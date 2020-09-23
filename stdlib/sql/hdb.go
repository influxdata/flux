package sql

import (
	"database/sql"
	"fmt"
	"math/big"
	"strings"
	"time"

	_ "github.com/SAP/go-hdb/driver"
	hdb "github.com/SAP/go-hdb/driver"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/values"
)

// SAP HANA DB support.
// Notes:
// * the last version compatible with Go 1.12 is v0.14.1. In v0.14.3, they started using 1.13 type sql.NullTime
//   and added `go 1.14` directive to go.mod (https://github.com/SAP/go-hdb/releases/tag/v0.14.3)
// * HDB returns BOOLEAN as TINYINT
// * TIMESTAMP does not TZ info stored, but:
//   - it is "strongly discouraged" to store data in local time zone: https://blogs.sap.com/2018/03/28/trouble-with-time/
//   - more on timestamps in HDB: https://help.sap.com/viewer/f1b440ded6144a54ada97ff95dac7adf/2.4/en-US/a394f75dcbe64b42b7a887231af8f15f.html
//   Therefore TIMESTAMP is mapped to TTime and vice-versa here.
// * the hdb driver is rather strict, eg. does not convert date- or time-formatted string values to time.Time,
//   or float64 to Decimal on its own and just throws "unsupported conversion" error

type HdbRowReader struct {
	Cursor      *sql.Rows
	columns     []interface{}
	columnTypes []flux.ColType
	columnNames []string
	sqlTypes    []*sql.ColumnType
	NextFunc    func() bool
	CloseFunc   func() error
}

// Next prepares HdbRowReader to return rows
func (m *HdbRowReader) Next() bool {
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

func (m *HdbRowReader) GetNextRow() ([]values.Value, error) {
	row := make([]values.Value, len(m.columns))
	for i, column := range m.columns {
		switch value := column.(type) {
		case bool, int64, float64, string:
			row[i] = values.New(value)
		case time.Time:
			// DATE, TIME types get scanned to time.Time by the driver too,
			// but they have no counterpart in Flux therefore will be represented as string
			switch m.sqlTypes[i].DatabaseTypeName() {
			case "DATE":
				row[i] = values.NewString(value.Format(layoutDate))
			case "TIME":
				row[i] = values.NewString(value.Format(layoutTime))
			default:
				row[i] = values.NewTime(values.ConvertTime(value))
			}
		case []uint8:
			switch m.columnTypes[i] {
			case flux.TFloat:
				var out hdb.Decimal
				err := out.Scan(value)
				if err != nil {
					return nil, err
				}
				newFloat, _ := (*big.Rat)(&out).Float64()
				row[i] = values.NewFloat(newFloat)
			default:
				row[i] = values.NewString(string(value))
			}
		case nil:
			row[i] = values.NewNull(flux.SemanticType(m.columnTypes[i]))
		default:
			execute.PanicUnknownType(flux.TInvalid)
		}
	}
	return row, nil
}

func (m *HdbRowReader) InitColumnNames(names []string) {
	m.columnNames = names
}

func (m *HdbRowReader) InitColumnTypes(types []*sql.ColumnType) {
	fluxTypes := make([]flux.ColType, len(types))
	for i := 0; i < len(types); i++ {
		switch types[i].DatabaseTypeName() {
		case "TINYINT", "SMALLINT", "INTEGER", "BIGINT":
			fluxTypes[i] = flux.TInt
		case "REAL", "DOUBLE", "DECIMAL":
			fluxTypes[i] = flux.TFloat
		case "TIMESTAMP": // not exactly correct (see Notes)
			fluxTypes[i] = flux.TTime
		default:
			fluxTypes[i] = flux.TString
		}
	}
	m.columnTypes = fluxTypes
	m.sqlTypes = types
}

func (m *HdbRowReader) ColumnNames() []string {
	return m.columnNames
}

func (m *HdbRowReader) ColumnTypes() []flux.ColType {
	return m.columnTypes
}

func (m *HdbRowReader) SetColumnTypes(types []flux.ColType) {
	m.columnTypes = types
}

func (m *HdbRowReader) SetColumns(i []interface{}) {
	m.columns = i
}

func (m *HdbRowReader) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	if err := m.Cursor.Err(); err != nil {
		return err
	}
	return m.Cursor.Close()
}

func NewHdbRowReader(r *sql.Rows) (execute.RowReader, error) {
	reader := &HdbRowReader{
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

var fluxToHdb = map[flux.ColType]string{
	flux.TFloat:  "DOUBLE",
	flux.TInt:    "BIGINT",
	flux.TString: "NVARCHAR(5000)", // 5000 is the max
	flux.TBool:   "BOOLEAN",
	flux.TTime:   "TIMESTAMP", // not exactly correct (see Notes)
}

// HdbTranslateColumn translates flux colTypes into their corresponding SAP HANA column type
func HdbColumnTranslateFunc() translationFunc {
	return func(f flux.ColType, colName string) (string, error) {
		s, found := fluxToHdb[f]
		if !found {
			return "", errors.Newf(codes.Invalid, "SAP HANA does not support column type %s", f.String())
		}
		return colName + " " + s, nil
	}
}

// Template for conditional query by table existence check
var hdbDoIfTableNotExistsTemplate = `DO
BEGIN
	DECLARE SCHEMA_NAME NVARCHAR(%d) = '%s';
	DECLARE TABLE_NAME NVARCHAR(%d) = '%s';
    DECLARE X_EXISTS INT = 0;
    SELECT COUNT(*) INTO X_EXISTS FROM TABLES %s;
    IF :X_EXISTS = 0
    THEN
        %s;
    END IF;
END;
`

// Adds SAP HANA specific table existence check to CREATE TABLE statement.
func hdbAddIfNotExist(table string, query string) string {
	var where string
	var args []interface{}
	parts := strings.SplitN(table, ".", 2)
	if len(parts) == 2 { // fully-qualified table name
		where = "WHERE SCHEMA_NAME=UPPER(:SCHEMA_NAME) AND TABLE_NAME=UPPER(:TABLE_NAME)"
		args = append(args, len(parts[0]))
		args = append(args, parts[0])
		args = append(args, len(parts[1]))
		args = append(args, parts[1])
	} else { // table in user default schema
		where = "WHERE TABLE_NAME=UPPER(:TABLE_NAME)"
		args = append(args, len("default"))
		args = append(args, "default")
		args = append(args, len(table))
		args = append(args, table)
	}
	args = append(args, where)
	args = append(args, query)

	return fmt.Sprintf(hdbDoIfTableNotExistsTemplate, args...)
}
