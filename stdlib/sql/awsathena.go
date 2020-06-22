package sql

import (
	"database/sql"
	"time"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/values"
	_ "github.com/uber/athenadriver/go"
)

// AWS Athena support.
// Notes:
// * transactions are not supported by Athena
// * only `sql.from` is supported, as the purpose of the database is to query existing data
//     - tables are are typically created like CREATE TABLE ... AS SELECT FROM. It is still possible
//       to create external tables CREATE EXTERNAL TABLE ..., but it may be rather tricky.
//     - typical insert is INSERT INTO .. SELECT FROM. But INSERT INTO .. (...) VALUES (...) is possible.
// * connection string
//   - see https://uber.github.io/athenadriver/ for details
//   - minimal connection string contains S3 bucket, access ID, secret key and AWS region
//   - the bucket for query results must exist
//   - parameters:
//     * region - AWS region(eg. us-west-1)
//     * db - database name
//     * accessID - AWS IAM access ID
//     * secretAccessKey - AWS IAM secret key
//     * WGRemoteCreation - controls creating of workgroup and tags (true by default)
//     * missingAsDefault - controls whether missing data in S3 files are returned with default values
//     * missingAsEmptyString - controls whether missing data in S3 files are returned as empty string (default)
//   - examples:
//     "s3://myorgqueryresults/?accessID=AKIAJLO3F...&region=us-west-1&secretAccessKey=NnQ7MUMp9PYZsmD47c%2BSsXGOFsd%2F..."
//     "s3://myorgqueryresults/?accessID=AKIAJLO3F...&db=dbname&missingAsDefault=false&missingAsEmptyString=false&region=us-west-1&secretAccessKey=NnQ7MUMp9PYZsmD47c%2BSsXGOFsd%2F...&WGRemoteCreation=false"

type AwsAthenaRowReader struct {
	Cursor      *sql.Rows
	columns     []interface{}
	columnTypes []flux.ColType
	columnNames []string
	sqlTypes    []*sql.ColumnType
	NextFunc    func() bool
	CloseFunc   func() error
}

// Next prepares AwsAthenaRowReader to return rows
func (m *AwsAthenaRowReader) Next() bool {
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

func (m *AwsAthenaRowReader) GetNextRow() ([]values.Value, error) {
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
		case string:
			row[i] = values.New(value)
		case time.Time:
			switch m.sqlTypes[i].DatabaseTypeName() {
			case "date":
				row[i] = values.NewString(value.Format(layoutDate))
			case "time":
				row[i] = values.NewString(value.Format(layoutTime))
			case "timestamp":
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

func (m *AwsAthenaRowReader) InitColumnNames(names []string) {
	m.columnNames = names
}

func (m *AwsAthenaRowReader) InitColumnTypes(types []*sql.ColumnType) {
	fluxTypes := make([]flux.ColType, len(types))
	for i := 0; i < len(types); i++ {
		switch types[i].DatabaseTypeName() {
		case "tinyint", "smallint", "int", "integer", "bigint":
			fluxTypes[i] = flux.TInt
		case "float", "double", "real":
			fluxTypes[i] = flux.TFloat
		case "boolean":
			fluxTypes[i] = flux.TBool
		case "timestamp with time zone": // "timestamp", "date" and "time" will be represented as string
			fluxTypes[i] = flux.TTime
		default:
			fluxTypes[i] = flux.TString
		}
	}
	m.columnTypes = fluxTypes
	m.sqlTypes = types
}

func (m *AwsAthenaRowReader) ColumnNames() []string {
	return m.columnNames
}

func (m *AwsAthenaRowReader) ColumnTypes() []flux.ColType {
	return m.columnTypes
}

func (m *AwsAthenaRowReader) SetColumnTypes(types []flux.ColType) {
	m.columnTypes = types
}

func (m *AwsAthenaRowReader) SetColumns(i []interface{}) {
	m.columns = i
}

func (m *AwsAthenaRowReader) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	if err := m.Cursor.Err(); err != nil {
		return err
	}
	return m.Cursor.Close()
}

func NewAwsAthenaRowReader(r *sql.Rows) (execute.RowReader, error) {
	reader := &AwsAthenaRowReader{
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
