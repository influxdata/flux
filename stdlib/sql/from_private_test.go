package sql

import (
	"testing"

	"github.com/influxdata/flux/dependencies/url"
	"github.com/influxdata/flux/execute/executetest"
)

func TestFromSqlUrlValidation(t *testing.T) {
	testCases := executetest.SourceUrlValidationTestCases{
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
