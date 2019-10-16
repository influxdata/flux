package sql

import (
	"testing"

	"github.com/influxdata/flux/dependencies/url"
	"github.com/influxdata/flux/querytest"
)

func TestFromSqlUrlValidation(t *testing.T) {
	testCases := querytest.SourceUrlValidationTestCases{
		{
			Name: "ok mysql",
			Spec: &FromSQLProcedureSpec{
				DriverName:     "mysql",
				DataSourceName: "username:password@tcp(localhost:12345)/dbname?param=value",
				Query:          "",
			},
			ErrMsg: "",
		}, {
			Name: "ok postgres",
			Spec: &FromSQLProcedureSpec{
				DriverName:     "postgres",
				DataSourceName: "postgres://pqgotest:password@localhost:12345/pqgotest?sslmode=verify-full",
				Query:          "",
			},
			ErrMsg: "",
		}, {
			Name: "invalid driver",
			Spec: &FromSQLProcedureSpec{
				DriverName:     "voltdb",
				DataSourceName: "",
				Query:          "",
			},
			ErrMsg: "sql driver voltdb not supported",
		}, {
			Name: "invalid mysql",
			Spec: &FromSQLProcedureSpec{
				DriverName:     "mysql",
				DataSourceName: "username:password@tcp(localhost:3306)/dbname?param=value",
				Query:          "",
			},
			V:      url.PrivateIPValidator{},
			ErrMsg: "it connects to a private IP",
		}, {
			Name: "invalid postgres",
			Spec: &FromSQLProcedureSpec{
				DriverName:     "postgres",
				DataSourceName: "postgres://pqgotest:password@localhost/pqgotest?sslmode=verify-full",
				Query:          "",
			},
			V:      url.PrivateIPValidator{},
			ErrMsg: "it connects to a private IP",
		}, {
			Name: "invalid sqlmock",
			Spec: &FromSQLProcedureSpec{
				DriverName:     "sqlmock",
				DataSourceName: "sqlmock://test:password@localhost/pqgotest?sslmode=verify-full",
				Query:          "",
			},
			V:      url.PrivateIPValidator{},
			ErrMsg: "it connects to a private IP",
		}, {
			Name: "no such host",
			Spec: &FromSQLProcedureSpec{
				DriverName:     "sqlmock",
				DataSourceName: "sqlmock://test:password@notfound/pqgotest?sslmode=verify-full",
				Query:          "",
			},
			V:      url.PrivateIPValidator{},
			ErrMsg: "no such host",
		},
	}
	testCases.Run(t, createFromSQLSource)
}

func TestFromSqliteUrlValidation(t *testing.T) {
	testCases := querytest.SourceUrlValidationTestCases{
		{
			Name: "ok sqlite path1",
			Spec: &FromSQLProcedureSpec{
				DriverName:     "sqlite3",
				DataSourceName: "file::memory:",
				Query:          "",
			},
			ErrMsg: "",
		}, {
			Name: "ok sqlite path2",
			Spec: &FromSQLProcedureSpec{
				DriverName:     "sqlite3",
				DataSourceName: ":memory:",
				Query:          "",
			},
			ErrMsg: "",
		}, {
			Name: "ok sqlite path3",
			Spec: &FromSQLProcedureSpec{
				DriverName:     "sqlite3",
				DataSourceName: "bananas.db",
				Query:          "",
			},
			ErrMsg: "",
		}, {
			Name: "ok sqlite path4",
			Spec: &FromSQLProcedureSpec{
				DriverName:     "sqlite3",
				DataSourceName: "bananas?cool_pragma",
				Query:          "",
			},
			ErrMsg: "",
		}, {
			Name: "ok sqlite path5",
			Spec: &FromSQLProcedureSpec{
				DriverName:     "sqlite3",
				DataSourceName: "file:test.db?cache=shared&mode=memory",
				Query:          "",
			},
			ErrMsg: "",
		}, {
			Name: "ok sqlite path6",
			Spec: &FromSQLProcedureSpec{
				DriverName:     "sqlite3",
				DataSourceName: "bananas?cool_pragma&even_better=true",
				Query:          "",
			},
			ErrMsg: "",
		}, {
			Name: "ok sqlite path7",
			Spec: &FromSQLProcedureSpec{
				DriverName:     "sqlite3",
				DataSourceName: "file:test.db",
				Query:          "",
			},
			ErrMsg: "",
		}, {
			Name: "bad sqlite driver",
			Spec: &FromSQLProcedureSpec{
				DriverName:     "sqlite4",
				DataSourceName: "bananas?cool_pragma",
				Query:          "",
			},
			ErrMsg: "sql driver sqlite4 not supported",
		}, {
			Name: "bad sqlite path1",
			Spec: &FromSQLProcedureSpec{
				DriverName:     "sqlite3",
				DataSourceName: ":cool_pragma",
				Query:          "",
			},
			ErrMsg: "invalid data source url: parse :cool_pragma: missing protocol scheme",
		}, {
			Name: "bad sqlite path2",
			Spec: &FromSQLProcedureSpec{
				DriverName:     "sqlite3",
				DataSourceName: "",
				Query:          "",
			},
			ErrMsg: "invalid data source url: empty path supplied",
		}, {
			Name: "bad sqlite path3",
			Spec: &FromSQLProcedureSpec{
				DriverName:     "sqlite3",
				DataSourceName: "    ",
				Query:          "",
			},
			ErrMsg: "invalid data source url: empty path supplied",
		},
	}
	testCases.Run(t, createFromSQLSource)
}
