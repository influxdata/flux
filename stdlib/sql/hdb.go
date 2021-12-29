package sql

import (
	"database/sql"
	"fmt"
	"math/big"
	"strings"
	"time"

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
// * Per naming conventions rules (https://documentation.sas.com/?cdcId=pgmsascdc&cdcVersion=9.4_3.5&docsetId=acreldb&docsetTarget=p1k98908uh9ovsn1jwzl3jg05exr.htm&locale=en),
//   `sql.to` target table and column names are assumed in / converted to uppercase.

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
			default: // flux.TString
				switch m.sqlTypes[i].DatabaseTypeName() {
				case "BINARY", "VARBINARY":
					return nil, errors.Newf(codes.Invalid, "Flux does not support column type %s", m.sqlTypes[i].DatabaseTypeName())
				}
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
		// quote the column name for safety also convert to uppercase per HDB naming conventions
		return hdbEscapeName(colName, true) + " " + s, nil
	}
}

// Template for conditional query by table existence check
var hdbDoIfTableNotExistsTemplate = `DO
BEGIN
    DECLARE SCHEMA_NAME NVARCHAR(%d) = %s;
    DECLARE TABLE_NAME NVARCHAR(%d) = %s;
    DECLARE X_EXISTS INT = 0;
    SELECT COUNT(*) INTO X_EXISTS FROM TABLES %s;
    IF :X_EXISTS = 0
    THEN
        %s;
    END IF;
END;
`

// hdbAddIfNotExist adds SAP HANA specific table existence check to CREATE TABLE statement.
func hdbAddIfNotExist(table string, query string) string {
	var where, schema, tbl string
	var args []interface{}
	// Schema and table name assumed uppercase in HDB by default (see Notes)
	// XXX: since we are currently forcing identifiers to be UPPER CASE.
	//  Shadowing the table param ensures we use the UPPER CASE form regardless
	//  of which branch we land in for the `if` below.
	table = strings.ToUpper(table)
	parts := strings.SplitN(table, ".", 2)

	// XXX: maybe we should panic if len(parts) is greater than 2?
	if len(parts) == 2 {
		// When there are 2 parts, we assume a fully-qualified table name (ex: `schema.tbl`)
		schema = parts[0]
		tbl = parts[1]
		where = "WHERE SCHEMA_NAME=ESCAPE_DOUBLE_QUOTES(:SCHEMA_NAME) AND TABLE_NAME=ESCAPE_DOUBLE_QUOTES(:TABLE_NAME)"
	} else {
		// Otherwise we assume there's only one part (table, with an implicit default schema).
		where = "WHERE TABLE_NAME=ESCAPE_DOUBLE_QUOTES(:TABLE_NAME)"
		schema = "default"
		tbl = table
	}
	args = append(args, len(schema))
	args = append(args, singleQuote(schema))
	args = append(args, len(tbl))
	args = append(args, singleQuote(tbl))

	args = append(args, where)
	args = append(args, query)

	return fmt.Sprintf(hdbDoIfTableNotExistsTemplate, args...)
}

// hdbEscapeName escapes name in double quotes and convert it to uppercase per HDB naming conventions
func hdbEscapeName(name string, toUpper bool) string {
	// XXX(onelson): Seems like it would be better to *just* quote/escape without
	// the case transformation. If the mandate is to "always quote identifiers"
	// as an SQL injection mitigation step, the case transformation feels like
	// an unexpected twist on what otherwise might be easier to explain.
	// Eg: "We quote all identifiers as a security precaution. Quoted identifiers are case-sensitive."
	// Currently, it is (arbitrarily) impossible for Flux to reference objects
	// in HDB that don't have an UPPER CASE identifier (which is perfectly valid).

	// truncate `name` to the first interior nul byte (if one is present).
	end := strings.IndexRune(name, 0)
	if end > -1 {
		name = name[:end]
	}

	parts := strings.Split(name, ".")
	for i := range parts {
		if toUpper {
			parts[i] = strings.ToUpper(parts[i])
		}
		parts[i] = doubleQuote(strings.Trim(parts[i], "\""))
	}
	return strings.Join(parts, ".")
}
