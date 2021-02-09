package sql

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/godror/godror"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/values"
)

// Oracle database support.
// Notes:
//   * uses Go SQL driver https://github.com/godror/godror
//   * supports Easy Connect string format: https://www.orafaq.com/wiki/EZCONNECT
//     Syntax: username/password@[//]host[:port][/service_name]
//     Examples:
//       "scott/tiger@sales-server:1521/ORCL"
//       "scott/tiger@//75.184.3.22:1522/xepdb1"
//   * ODPI-C wrapper is required at compile time https://github.com/oracle/odpi
//   * Oracle Instant Client is required at runtime
//     - https://www.oracle.com/database/technologies/instant-client.html
//     - client version should match the server, or it should be ensured that timezone files match.
//       Otherwise, errors such as "ORA-01805: possible error in date/time operation"
//       may occur when working with columns of data types with time zone information.
// Issues:
//   * BOOLEAN type does not exist in Oracle. Output mapping uses most(?) recommended approach (CHAR(1)).
//   * INTERVAL types are scanned to time.Duration, but there is no corresponding flux.ColType (yet?),
//     so these types are mapped converted to string. However, duration string representation in Flux
//     is nothing like in Oracle, so an attempt to write such value directly to interval type column fails
//     with "ORA-01867: the interval is invalid".
//   * TIMESTAMP types workaround https://github.com/godror/godror#timestamp is used (in sql.ExecuteQueries()),
//     as they are not properly supported by the driver. It also seems to solve time zone modification for
//     TIMESTAMP WITH TIME ZONE type.

type OracleRowReader struct {
	Cursor      *sql.Rows
	columns     []interface{}
	columnTypes []flux.ColType
	columnNames []string
	sqlTypes    []*sql.ColumnType
	NextFunc    func() bool
	CloseFunc   func() error
}

// Next prepares OracleRowReader to return rows
func (m *OracleRowReader) Next() bool {
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

func (m *OracleRowReader) GetNextRow() ([]values.Value, error) {
	row := make([]values.Value, len(m.columns))
	for i, column := range m.columns {
		switch value := column.(type) {
		case bool, int64, float64:
			row[i] = values.New(value)
		case int:
			row[i] = values.NewInt(int64(value))
		case int8:
			row[i] = values.NewInt(int64(value))
		case int16:
			row[i] = values.NewInt(int64(value))
		case int32:
			row[i] = values.NewInt(int64(value))
		case float32:
			row[i] = values.NewFloat(float64(value))
		case []uint8: // NUMBER in go-ora
			switch m.columnTypes[i] {
			case flux.TInt:
				newInt, err := UInt8ToInt64(value)
				if err != nil {
					return nil, err
				}
				row[i] = values.NewInt(newInt)
			case flux.TFloat:
				newFloat, err := UInt8ToFloat(value)
				if err != nil {
					return nil, err
				}
				row[i] = values.NewFloat(newFloat)
			default:
				row[i] = values.NewString(string(value))
			}
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
			default:
				row[i] = values.New(value)
			}
		case time.Duration: // INTERVAL ... types
			row[i] = values.NewString(value.String())
		case time.Time:
			// DATE, TIMESTAMP and TIMESTAMP WITH LOCAL TIME ZONE types are scanned to time.Time by the driver,
			// but they have no counterpart in Flux therefore will be represented as string
			switch m.sqlTypes[i].DatabaseTypeName() {
			case "DATE":
				row[i] = values.NewString(value.Format(layoutDate))
			case "TIMESTAMP":
				row[i] = values.NewString(value.Format(layoutTimeStampNtz))
			case "TIMESTAMP WITH LOCAL TIME ZONE":
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

func (m *OracleRowReader) InitColumnNames(names []string) {
	m.columnNames = names
}

func (m *OracleRowReader) InitColumnTypes(types []*sql.ColumnType) {
	fluxTypes := make([]flux.ColType, len(types))
	for i := 0; i < len(types); i++ {
		switch types[i].DatabaseTypeName() {
		case "NUMBER":
			_, scale, ok := types[i].DecimalSize()
			if ok && scale == 0 {
				fluxTypes[i] = flux.TInt
			} else {
				fluxTypes[i] = flux.TFloat
			}
		case "FLOAT", "DOUBLE":
			fluxTypes[i] = flux.TFloat
		case "BOOLEAN":
			fluxTypes[i] = flux.TBool
		case "TIMESTAMP WITH TIME ZONE": // "DATE", "TIMESTAMP" and "TIMESTAMP WITH LOCAL TIME ZONE" will be represented as string
			fluxTypes[i] = flux.TTime
		default:
			fluxTypes[i] = flux.TString
		}
	}
	m.columnTypes = fluxTypes
	m.sqlTypes = types
}

func (m *OracleRowReader) ColumnNames() []string {
	return m.columnNames
}

func (m *OracleRowReader) ColumnTypes() []flux.ColType {
	return m.columnTypes
}

func (m *OracleRowReader) SetColumnTypes(types []flux.ColType) {
	m.columnTypes = types
}

func (m *OracleRowReader) SetColumns(i []interface{}) {
	m.columns = i
}

func (m *OracleRowReader) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	if err := m.Cursor.Err(); err != nil {
		return err
	}
	return m.Cursor.Close()
}

func NewOracleRowReader(r *sql.Rows) (execute.RowReader, error) {
	reader := &OracleRowReader{
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

var fluxToOracle = map[flux.ColType]string{
	flux.TFloat:  "BINARY_DOUBLE", // double-precision (64-bit) IEEE 754 floating-point
	flux.TInt:    "NUMBER(38)",
	flux.TUInt:   "NUMBER(38)",
	flux.TString: "VARCHAR2(4000)",
	flux.TBool:   "CHAR(1)", // most(?) recommended solution ('Y', 'N')
	flux.TTime:   "TIMESTAMP WITH TIME ZONE",
}

// OracleTranslateColumn translates flux colTypes into their corresponding Oracle column type
func OracleColumnTranslateFunc() translationFunc {
	return func(f flux.ColType, colName string) (string, error) {
		s, found := fluxToOracle[f]
		if !found {
			return "", errors.Newf(codes.Internal, "Oracle does not support column type %s", f.String())
		}
		return colName + " " + s, nil
	}
}

// oracleOpenFunction opens connection to Oracle DB and sets date and timestamp formats.
// The driver name for user is "oracle", but real driver is "godror".
// For Go older then 1.14.6, it is recommended to disable pooling (https://godror.github.io/godror/doc/connection.html#-oracle-session-pooling).
// Flux can be built with 1.12, but InfluxDb requires 1.15, so we do nothing here about it (the pooling can be affected with connection string option as well).
func oracleOpenFunction(driverName, dataSourceName string) openFunc {
	return func() (*sql.DB, error) {
		P, err := godror.ParseDSN(dataSourceName)
		if err != nil {
			return nil, err
		}
		P.SetSessionParamOnInit("NLS_DATE_FORMAT", "YYYY-MM-DD")
		P.SetSessionParamOnInit("NLS_TIMESTAMP_FORMAT", "YYYY-MM-DD\"T\"HH24:MI:SS.FF")
		P.SetSessionParamOnInit("NLS_TIMESTAMP_TZ_FORMAT", "YYYY-MM-DD\"T\"HH24:MI:SS.FFTZH:TZM")
		connector := godror.NewConnector(P)
		db := sql.OpenDB(connector)
		return db, nil
	}
}

// template for conditional query by table existence check
var oracleDoIfTableNotExistsTemplate = `DECLARE
    v_count BINARY_INTEGER;
BEGIN
    SELECT COUNT(*) INTO v_count FROM user_tables WHERE table_name = UPPER('%s');
    IF (v_count = 0)
    THEN
        EXECUTE IMMEDIATE '%s';
    END IF;
END;
`

// oracleAddIfNotExist adds Oracle specific table existence check to CREATE TABLE statement.
func oracleAddIfNotExist(table string, query string) string {
	return fmt.Sprintf(oracleDoIfTableNotExistsTemplate, []interface{}{table, query}...)
}
