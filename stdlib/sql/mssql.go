package sql

import (
	"database/sql"
	"fmt"
	neturl "net/url"
	"strings"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/values"
)

// Microsoft SQL Server support.

type MssqlRowReader struct {
	Cursor      *sql.Rows
	columns     []interface{}
	columnTypes []flux.ColType
	columnNames []string
	sqlTypes    []*sql.ColumnType
	NextFunc    func() bool
	CloseFunc   func() error
}

// Next prepares MssqlRowReader to return rows
func (m *MssqlRowReader) Next() bool {
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

func (m *MssqlRowReader) GetNextRow() ([]values.Value, error) {
	row := make([]values.Value, len(m.columns))
	for i, column := range m.columns {
		switch value := column.(type) {
		case bool, int64, float64, string:
			row[i] = values.New(value)
		case []uint8:
			// DECIMAL, MONEY and SMALLMONEY is scanned to []uint8 by the driver
			switch m.columnTypes[i] {
			case flux.TFloat:
				newFloat, err := UInt8ToFloat(value)
				if err != nil {
					return nil, err
				}
				row[i] = values.NewFloat(newFloat)
			default:
				row[i] = values.NewString(string(value))
			}
		case time.Time:
			// DATETIME, DATETIME2, DATE, TIME and others types get scanned to time.Time by the driver,
			// but they have no counterpart in Flux therefore will be represented as string
			switch m.sqlTypes[i].DatabaseTypeName() {
			case "DATE":
				row[i] = values.NewString(value.Format(layoutDate))
			case "TIME":
				row[i] = values.NewString(value.Format(layoutTime))
			case "DATETIME", "DATETIME2", "SMALLDATETIME":
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

func (m *MssqlRowReader) InitColumnNames(names []string) {
	m.columnNames = names
}

func (m *MssqlRowReader) InitColumnTypes(types []*sql.ColumnType) {
	fluxTypes := make([]flux.ColType, len(types))
	for i := 0; i < len(types); i++ {
		switch types[i].DatabaseTypeName() {
		case "INT", "TINYINT", "SMALLINT", "BIGINT":
			fluxTypes[i] = flux.TInt
		case "DECIMAL", "REAL", "FLOAT", "MONEY", "SMALLMONEY":
			fluxTypes[i] = flux.TFloat
		case "BIT":
			fluxTypes[i] = flux.TBool
		case "DATETIMEOFFSET": // other date/time types will be represented as string because they do not have tz
			fluxTypes[i] = flux.TTime
		default:
			fluxTypes[i] = flux.TString
		}
	}
	m.columnTypes = fluxTypes
	m.sqlTypes = types
}

func (m *MssqlRowReader) ColumnNames() []string {
	return m.columnNames
}

func (m *MssqlRowReader) ColumnTypes() []flux.ColType {
	return m.columnTypes
}

func (m *MssqlRowReader) SetColumnTypes(types []flux.ColType) {
	m.columnTypes = types
}

func (m *MssqlRowReader) SetColumns(i []interface{}) {
	m.columns = i
}

func (m *MssqlRowReader) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	if err := m.Cursor.Err(); err != nil {
		return err
	}
	return m.Cursor.Close()
}

func NewMssqlRowReader(r *sql.Rows) (execute.RowReader, error) {
	reader := &MssqlRowReader{
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

var fluxToSQLServer = map[flux.ColType]string{
	flux.TFloat:  "FLOAT",
	flux.TInt:    "BIGINT",
	flux.TUInt:   "BIGINT",
	flux.TString: "VARCHAR(MAX)",
	flux.TBool:   "BIT",
	flux.TTime:   "DATETIMEOFFSET",
}

// MssqlTranslateColumn translates flux colTypes into their corresponding SQL Server column type
func MssqlColumnTranslateFunc() translationFunc {
	return func(f flux.ColType, colName string) (string, error) {
		s, found := fluxToSQLServer[f]
		if !found {
			return "", errors.Newf(codes.Internal, "SQLServer does not support column type %s", f.String())
		}
		return colName + " " + s, nil
	}
}

// Checks if the driver is SQL Server.
func isMssqlDriver(driverName string) bool {
	return driverName == "mssql" || driverName == "sqlserver"
}

// Config is a (minimalistic) configuration parsed from a DSN string.
// Some other drivers (such as MySQL or Snowflake) have such type.
type mssqlConfig struct {
	Scheme   string
	Host     string
	User     string
	Password string
	Database string
	// Azure auth
	AzureAuth string
	*AzureConfig
}

// Parses DSN connection string to a Config.
// Unlike other drivers, go-mssqldb does not provide ParseDSN() or similar method.
// Connection string options: https://github.com/denisenkom/go-mssqldb#connection-parameters-and-dsn
func mssqlParseDSN(dsn string) (cfg *mssqlConfig, err error) {
	cfg = &mssqlConfig{
		Scheme: "sqlserver",
	}
	if strings.HasPrefix(dsn, "sqlserver://") {
		u, err := neturl.Parse(dsn)
		if err != nil {
			return nil, errors.Newf(codes.Invalid, "invalid data source dsn: %v", err)
		}
		cfg.Host = u.Host
		cfg.User = u.User.Username()
		cfg.Password, _ = u.User.Password()
		v, err := neturl.ParseQuery(u.RawQuery)
		if err != nil {
			return nil, err
		}
		cfg.Database = v.Get("database")
		// set Azure AD auth configuration (if any)
		mssqlSetAzureConfig(v, cfg)
	} else { // ADO or ODBC style connection string
		if len(dsn) == 0 {
			return nil, errors.Newf(codes.Invalid, "invalid data source dsn: %v", err)
		}
		dsn = strings.TrimPrefix(dsn, "odbc:") // ODBC is very much like just prefixed ADO conn string, so use simplistic approach
		params := make(neturl.Values, 6)
		pairs := strings.Split(dsn, ";")
		params.Set("port", "1433") // default that may be omitted in the dsn
		for _, pair := range pairs {
			if len(pair) == 0 {
				continue
			}
			lst := strings.SplitN(pair, "=", 2)
			if len(lst) < 2 {
				continue
			}
			name := strings.TrimSpace(strings.ToLower(lst[0]))
			if len(name) == 0 {
				continue
			}
			value := strings.TrimSpace(lst[1])
			params.Set(name, value)
		}
		cfg.Host = fmt.Sprintf("%s:%s", params.Get("server"), params.Get("port"))
		cfg.User = params.Get("user id")
		cfg.Password = params.Get("password")
		cfg.Database = params.Get("database")
		// set Azure AD auth configuration (if any)
		mssqlSetAzureConfig(params, cfg)
	}
	return cfg, nil
}

//
// Some operations should only be allowed when explicitly enabled by the user.
//

// Mssql driver specific options.
const (
	// SQL Server cannot insert into table with IDENTITY set unless explicitly enabled.
	// https://docs.microsoft.com/en-us/sql/t-sql/statements/set-identity-insert-transact-sql
	mssqlIdentityInsertEnabled = "identity insert=on"
)

// Checks if a parameter is set.
// Very simplistic for now.
func mssqlCheckParameter(dsn string, option string) bool {
	raw, err := neturl.QueryUnescape(dsn)
	if err != nil {
		raw = dsn
	}
	return strings.Contains(strings.ToLower(raw), option)
}
