package sql

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
)

// represents unsupported type
var unsupportedType flux.ColType = 666

// additional and seperate tests that can be run without needing functions to be Exported in sql, just to be testable
func TestCorrectBatchSize(t *testing.T) {
	// given the combination of row width and supplied batchSize argument from user, verify that it is modified as required
	userBatchSize := 1000
	rowWidth := 10
	correctedSize := correctBatchSize(userBatchSize, rowWidth)
	if !cmp.Equal(99, correctedSize) {
		t.Log(cmp.Diff(90, correctedSize))
		t.Fail()
	}

	// verify that the batchSoze is not lower than the width of a single row - if it ever is, we have a big problem
	userBatchSize = 1
	correctedSize = correctBatchSize(userBatchSize, rowWidth)
	if !cmp.Equal(10, correctedSize) {
		t.Log(cmp.Diff(10, correctedSize))
		t.Fail()
	}

	userBatchSize = -1
	correctedSize = correctBatchSize(userBatchSize, rowWidth)
	if !cmp.Equal(10, correctedSize) {
		t.Log(cmp.Diff(10, correctedSize))
		t.Fail()
	}
}

func TestTranslationDriverReturn(t *testing.T) {

	// verify invalid return error
	_, err := getTranslationFunc("bananas")
	assertErrorMsg(t, "invalid driverName: bananas", err)

	// verify that valid returns expected happiness for SQLITE
	_, err = getTranslationFunc("sqlite3")
	assertErrorIsNil(t, err)

	// verify that valid returns expected happiness for Postgres
	_, err = getTranslationFunc("postgres")
	assertErrorIsNil(t, err)

	// verify that valid returns expected happiness for MySQL
	_, err = getTranslationFunc("mysql")
	assertErrorIsNil(t, err)

	// verify that valid returns expected happiness for Snowflake
	_, err = getTranslationFunc("snowflake")
	assertErrorIsNil(t, err)

	// verify that valid returns expected happiness for Mssql
	_, err = getTranslationFunc("sqlserver")
	assertErrorIsNil(t, err)

	// verify that valid returns expected happiness for BigQuery
	_, err = getTranslationFunc("bigquery")
	assertErrorIsNil(t, err)

	// verify that valid returns expected error for AWS Athena (yes, error)
	_, err = getTranslationFunc("awsathena")
	assertErrorMsg(t, "writing is not supported for awsathena", err)

	// verify that valid returns expected happiness for SAP HANA
	_, err = getTranslationFunc("hdb")
	assertErrorIsNil(t, err)

}

func TestSqliteTranslation(t *testing.T) {
	sqliteTypeTranslations := map[string]flux.ColType{
		"FLOAT":    flux.TFloat,
		"INT":      flux.TInt,
		"TEXT":     flux.TString,
		"DATETIME": flux.TTime,
	}
	columnLabel := "apples"
	quote, err := getQuoteIdentFunc("sqlite3")
	assertErrorIsNil(t, err)

	sqlT, err := getTranslationFunc("sqlite3")
	assertErrorIsNil(t, err)

	for dbTypeString, fluxType := range sqliteTypeTranslations {
		v, err := sqlT()(fluxType, columnLabel)
		assertErrorIsNil(t, err)
		assertTranslation(t, quote, columnLabel, dbTypeString, v)
	}

	// as SQLITE has NO BOOLEAN column type, we need to return an error rather than doing implicit conversions
	_, err = sqlT()(flux.TBool, columnLabel)
	assertErrorMsg(t, "SQLite does not support column type bool", err)
}

func TestPostgresTranslation(t *testing.T) {
	postgresTypeTranslations := map[string]flux.ColType{
		"FLOAT":     flux.TFloat,
		"TEXT":      flux.TString,
		"BIGINT":    flux.TInt,
		"TIMESTAMP": flux.TTime,
		"BOOL":      flux.TBool,
	}

	columnLabel := "apples"
	// verify that valid returns expected happiness for postgres
	quote, err := getQuoteIdentFunc("postgres")
	assertErrorIsNil(t, err)

	sqlT, err := getTranslationFunc("postgres")
	assertErrorIsNil(t, err)

	for dbTypeString, fluxType := range postgresTypeTranslations {
		v, err := sqlT()(fluxType, columnLabel)
		assertErrorIsNil(t, err)
		assertTranslation(t, quote, columnLabel, dbTypeString, v)
	}

	// test no match
	_, err = sqlT()(unsupportedType, columnLabel)
	assertErrorMsg(t, "PostgreSQL does not support column type unknown", err)
}

func TestMysqlTranslation(t *testing.T) {
	mysqlTypeTranslations := map[string]flux.ColType{
		"FLOAT":       flux.TFloat,
		"BIGINT":      flux.TInt,
		"TEXT(16383)": flux.TString,
		"DATETIME":    flux.TTime,
		"BOOL":        flux.TBool,
	}

	columnLabel := "apples"
	// verify that valid returns expected happiness for mysql
	quote, err := getQuoteIdentFunc("mysql")
	assertErrorIsNil(t, err)

	sqlT, err := getTranslationFunc("mysql")
	assertErrorIsNil(t, err)

	for dbTypeString, fluxType := range mysqlTypeTranslations {
		v, err := sqlT()(fluxType, columnLabel)
		assertErrorIsNil(t, err)
		assertTranslation(t, quote, columnLabel, dbTypeString, v)

	}

	// test no match
	_, err = sqlT()(unsupportedType, columnLabel)
	assertErrorMsg(t, "MySQL does not support column type unknown", err)
}

func TestSnowflakeTranslation(t *testing.T) {
	if !isDriverEnabled("snowflake") {
		t.Skip("snowflake is disabled, skipping test")
	}
	snowflakeTypeTranslations := map[string]flux.ColType{
		"FLOAT":         flux.TFloat,
		"NUMBER":        flux.TInt,
		"TEXT":          flux.TString,
		"TIMESTAMP_LTZ": flux.TTime,
		"BOOLEAN":       flux.TBool,
	}

	columnLabel := "apples"
	// verify that valid returns expected happiness for snowflake
	quote, err := getQuoteIdentFunc("snowflake")
	assertErrorIsNil(t, err)

	sqlT, err := getTranslationFunc("snowflake")
	assertErrorIsNil(t, err)

	for dbTypeString, fluxType := range snowflakeTypeTranslations {
		v, err := sqlT()(fluxType, columnLabel)
		assertErrorIsNil(t, err)
		assertTranslation(t, quote, columnLabel, dbTypeString, v)
	}

	// test no match
	_, err = sqlT()(unsupportedType, columnLabel)
	assertErrorMsg(t, "Snowflake does not support column type unknown", err)
}

func TestMssqlTranslation(t *testing.T) {
	if !isDriverEnabled("sqlserver") {
		t.Skip("mssql is disabled")
	}
	mssqlTypeTranslations := map[string]flux.ColType{
		"FLOAT":          flux.TFloat,
		"BIGINT":         flux.TInt,
		"VARCHAR(MAX)":   flux.TString,
		"DATETIMEOFFSET": flux.TTime,
		"BIT":            flux.TBool,
	}

	columnLabel := "apples"
	// verify that valid returns expected happiness for mssql
	quote, err := getQuoteIdentFunc("sqlserver")
	assertErrorIsNil(t, err)

	sqlT, err := getTranslationFunc("sqlserver")
	assertErrorIsNil(t, err)

	for dbTypeString, fluxType := range mssqlTypeTranslations {
		v, err := sqlT()(fluxType, columnLabel)
		assertErrorIsNil(t, err)
		assertTranslation(t, quote, columnLabel, dbTypeString, v)
	}

	// test no match
	_, err = sqlT()(unsupportedType, columnLabel)
	assertErrorMsg(t, "SQLServer does not support column type unknown", err)
}

func TestBigQueryTranslation(t *testing.T) {
	bigqueryTypeTranslations := map[string]flux.ColType{
		"FLOAT64":   flux.TFloat,
		"INT64":     flux.TInt,
		"STRING":    flux.TString,
		"TIMESTAMP": flux.TTime,
		"BOOL":      flux.TBool,
	}

	columnLabel := "apples"
	// verify that valid returns expected happiness for bigquery
	quote, err := getQuoteIdentFunc("bigquery")
	assertErrorIsNil(t, err)

	sqlT, err := getTranslationFunc("bigquery")
	assertErrorIsNil(t, err)

	for dbTypeString, fluxType := range bigqueryTypeTranslations {
		v, err := sqlT()(fluxType, columnLabel)
		assertErrorIsNil(t, err)
		assertTranslation(t, quote, columnLabel, dbTypeString, v)
	}

	// test no match
	_, err = sqlT()(unsupportedType, columnLabel)
	assertErrorMsg(t, "BigQuery does not support column type unknown", err)
}

func TestHdbTranslation(t *testing.T) {
	hdbTypeTranslations := map[string]flux.ColType{
		"DOUBLE":         flux.TFloat,
		"BIGINT":         flux.TInt,
		"NVARCHAR(5000)": flux.TString,
		"TIMESTAMP":      flux.TTime,
		"BOOLEAN":        flux.TBool,
	}

	columnLabel := "apples"
	// verify that valid returns expected happiness for hdb
	quote, err := getQuoteIdentFunc("hdb")
	assertErrorIsNil(t, err)
	sqlT, err := getTranslationFunc("hdb")
	assertErrorIsNil(t, err)
	for dbTypeString, fluxType := range hdbTypeTranslations {
		v, err := sqlT()(fluxType, columnLabel)
		assertErrorIsNil(t, err)
		assertTranslation(t, quote, columnLabel, dbTypeString, v)
	}

	// test no match
	var _unsupportedType flux.ColType = 666
	_, err = sqlT()(_unsupportedType, columnLabel)
	assertErrorMsg(t, "SAP HANA does not support column type unknown", err)
}

// Ensure an error has the expected message.
func assertErrorMsg(t *testing.T, want string, got error) {
	if got == nil {
		t.Error("expected error, got nil")
	}
	if diff := cmp.Diff(want, got.Error()); diff != "" {
		t.Errorf("expected error does not match: (-want/+got):\n%s", diff)
	}
}

// Ensure the output from the traslate func uses the appropriate quote func
func assertTranslation(t *testing.T, quoteFunc quoteIdentFunc, label string, dbType string, got string) {
	if diff := cmp.Diff(quoteFunc(label)+" "+dbType, got); diff != "" {
		t.Errorf("quoted ident does not match: (-want/+got):\n%s", diff)
	}
}

// Check if an error is nil, failing the test when it isn't
func assertErrorIsNil(t *testing.T, got error) {
	if got != nil {
		t.Errorf("expected error to be nil, got %v", got)
	}
}
