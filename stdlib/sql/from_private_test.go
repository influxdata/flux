package sql

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"testing"

	"github.com/influxdata/flux"
	fhttp "github.com/influxdata/flux/dependencies/http"
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
			Name: "ok vertica",
			Spec: &FromSQLProcedureSpec{
				DriverName:     "vertica",
				DataSourceName: "vertica://dbadmin:password@localhost:5433/VMart",
				Query:          "",
			},
			ErrMsg: "",
		}, {
			Name: "ok snowflake",
			Spec: &FromSQLProcedureSpec{
				DriverName:     "snowflake",
				DataSourceName: "username:password@accountname.us-east-1/dbname",
				Query:          "",
			},
			ErrMsg: "",
		}, {
			Name: "ok sqlserver (URL connection string)",
			Spec: &FromSQLProcedureSpec{
				DriverName:     "sqlserver",
				DataSourceName: "sqlserver://sa:mypass@localhost:1234?database=master",
				Query:          "",
			},
			ErrMsg: "",
		}, {
			Name: "ok sqlserver (ADO connection string)",
			Spec: &FromSQLProcedureSpec{
				DriverName:     "sqlserver",
				DataSourceName: "server=localhost;user id=sa;database=master;",
				Query:          "",
			},
			ErrMsg: "",
		}, {
			Name: "ok sqlserver (ADO connection string, Azure auth option)",
			Spec: &FromSQLProcedureSpec{
				DriverName:     "sqlserver",
				DataSourceName: "server=localhost;user id=sa;database=master;azure auth=ENV",
				Query:          "",
			},
			ErrMsg: "",
		}, {
			Name: "ok sqlserver (ADO connection string, Azure inline params)",
			Spec: &FromSQLProcedureSpec{
				DriverName:     "sqlserver",
				DataSourceName: "server=localhost;user id=sa;database=master;azure tenant id=77e7d537;azure client id=58879ce8;azure client secret=0123456789",
				Query:          "",
			},
			ErrMsg: "",
		}, {
			Name: "ok awsathena",
			Spec: &FromSQLProcedureSpec{
				DriverName:     "awsathena",
				DataSourceName: "s3://bucket/?accessID=ABCD123&region=us-west-1&secretAccessKey=PWD007&db=test",
				Query:          "",
			},
			ErrMsg: "",
		}, {
			Name: "ok bigquery",
			Spec: &FromSQLProcedureSpec{
				DriverName:     "bigquery",
				DataSourceName: "bigquery://project1/?dataset=dataset1",
				Query:          "",
			},
			ErrMsg: "",
		}, {
			Name: "ok hdb",
			Spec: &FromSQLProcedureSpec{
				DriverName:     "hdb",
				DataSourceName: "hdb://user:password@localhost:39013",
				Query:          "",
			},
			ErrMsg: "",
		}, {
			Name: "invalid driver",
			Spec: &FromSQLProcedureSpec{
				DriverName:     "voltdb",
				DataSourceName: "blablabla",
				Query:          "",
			},
			ErrMsg: "sql driver voltdb not supported",
		}, {
			Name: "invalid empty path",
			Spec: &FromSQLProcedureSpec{
				DriverName:     "blabla",
				DataSourceName: "",
				Query:          "",
			},
			ErrMsg: "invalid data source url: empty path supplied",
		}, {
			Name: "invalid mysql",
			Spec: &FromSQLProcedureSpec{
				DriverName:     "mysql",
				DataSourceName: "username:password@tcp(localhost:3306)/dbname?param=value",
				Query:          "",
			},
			V:      url.PrivateIPValidator{},
			ErrMsg: "no such host",
		}, {
			Name: "invalid postgres",
			Spec: &FromSQLProcedureSpec{
				DriverName:     "postgres",
				DataSourceName: "postgres://pqgotest:password@localhost/pqgotest?sslmode=verify-full",
				Query:          "",
			},
			V:      url.PrivateIPValidator{},
			ErrMsg: "no such host",
		}, {
			Name: "invalid bigquery",
			Spec: &FromSQLProcedureSpec{
				DriverName:     "bigquery",
				DataSourceName: "biqquery://project1/?dataset=dataset1",
				Query:          "",
			},
			ErrMsg: "invalid prefix",
		}, {
			Name: "invalid sqlmock",
			Spec: &FromSQLProcedureSpec{
				DriverName:     "sqlmock",
				DataSourceName: "sqlmock://test:password@localhost/pqgotest?sslmode=verify-full",
				Query:          "",
			},
			V:      url.PrivateIPValidator{},
			ErrMsg: "no such host",
		}, {
			Name: "no such host",
			Spec: &FromSQLProcedureSpec{
				DriverName: "sqlmock",
				// Using 'invalid.' for DNS name as its guaranteed not to exist
				// https://tools.ietf.org/html/rfc6761#section-6.4
				DataSourceName: "sqlmock://test:password@notfound.invalid./pqgotest?sslmode=verify-full",
				Query:          "",
			},
			V:      url.PrivateIPValidator{},
			ErrMsg: "data source did not pass url validation",
		}, {
			Name: "invalid mysql allowAllFiles parameter",
			Spec: &FromSQLProcedureSpec{
				DriverName:     "mysql",
				DataSourceName: "username:password@tcp(localhost:3306)/dbname?allowAllFiles=true",
				Query:          "",
			},
			ErrMsg: "invalid data source dsn: may not set allowAllFiles",
		},
	}
	// Scan over test cases and adjust any that the driver is disabled for
	for i := range testCases {
		if spec, ok := testCases[i].Spec.(*FromSQLProcedureSpec); ok {
			if err := disabledDriverError(spec.DriverName); err != nil {
				testCases[i].ErrMsg = err.Error()
			}
		}
	}
	testCases.Run(t, createFromSQLSource)
}

type mockDialer struct {
	err error
}

func (d *mockDialer) DialContext(_ context.Context, _, _ string) (net.Conn, error) {
	return nil, d.err
}

type mockDeps struct {
	flux.Deps
	dialer     flux.Dialer
	httpClient fhttp.Client
}

func (d mockDeps) Dialer() (flux.Dialer, error) {
	return d.dialer, nil
}

func (d mockDeps) HTTPClient() (fhttp.Client, error) {
	if d.httpClient != nil {
		return d.httpClient, nil
	}
	return d.Deps.HTTPClient()
}

func TestPostgresOpenFunctionDialer(t *testing.T) {
	expectErr := errors.New("test dial error")
	deps := mockDeps{
		Deps:   flux.NewDefaultDependencies(),
		dialer: &mockDialer{err: expectErr},
	}

	openFn := postgresOpenFunction("postgres://user:pass@localhost:5432/testdb")
	db, err := openFn(deps)
	if err != nil {
		t.Fatalf("unexpected error from open function: %v", err)
	}
	defer db.Close()

	// Ping triggers a real connection attempt, which will use our mock dialer.
	err = db.Ping()
	if err == nil {
		t.Fatal("expected error from Ping, got nil")
	}
	if !errors.Is(err, expectErr) {
		t.Fatalf("expected error %q, got: %v", expectErr, err)
	}
}

func TestMssqlOpenFunctionDialer(t *testing.T) {
	expectErr := errors.New("test dial error")
	deps := mockDeps{
		Deps:   flux.NewDefaultDependencies(),
		dialer: &mockDialer{err: expectErr},
	}

	openFn := mssqlOpenFunction("sqlserver://sa:password@localhost:1433?database=master")
	db, err := openFn(deps)
	if err != nil {
		t.Fatalf("unexpected error from open function: %v", err)
	}
	defer db.Close()

	// Ping triggers a real connection attempt, which will use our mock dialer.
	err = db.Ping()
	if err == nil {
		t.Fatal("expected error from Ping, got nil")
	}
	if !errors.Is(err, expectErr) {
		t.Fatalf("expected error %q, got: %v", expectErr, err)
	}
}

func TestMysqlOpenFunctionDialer(t *testing.T) {
	expectErr := errors.New("test dial error")
	deps := mockDeps{
		Deps:   flux.NewDefaultDependencies(),
		dialer: &mockDialer{err: expectErr},
	}

	openFn := mysqlOpenFunction("username:password@tcp(localhost:3306)/dbname")
	db, err := openFn(deps)
	if err != nil {
		t.Fatalf("unexpected error from open function: %v", err)
	}
	defer db.Close()

	// Ping triggers a real connection attempt, which will use our mock dialer.
	err = db.Ping()
	if err == nil {
		t.Fatal("expected error from Ping, got nil")
	}
	if !errors.Is(err, expectErr) {
		t.Fatalf("expected error %q, got: %v", expectErr, err)
	}
}

func TestHdbOpenFunctionDialer(t *testing.T) {
	expectErr := errors.New("test dial error")
	deps := mockDeps{
		Deps:   flux.NewDefaultDependencies(),
		dialer: &mockDialer{err: expectErr},
	}

	openFn := hdbOpenFunction("hdb://user:password@localhost:39013")
	db, err := openFn(deps)
	if err != nil {
		t.Fatalf("unexpected error from open function: %v", err)
	}
	defer db.Close()

	// Ping triggers a real connection attempt, which will use our mock dialer.
	err = db.Ping()
	if err == nil {
		t.Fatal("expected error from Ping, got nil")
	}
	if !errors.Is(err, expectErr) {
		t.Fatalf("expected error %q, got: %v", expectErr, err)
	}
}

func TestVerticaOpenFunctionUsesInjectedDialer(t *testing.T) {
	expectErr := errors.New("test dial error")
	deps := mockDeps{
		Deps:   flux.NewDefaultDependencies(),
		dialer: &mockDialer{err: expectErr},
	}

	openFn := verticaOpenFunction("vertica://dbadmin:password@localhost:5433/VMart")
	db, err := openFn(deps)
	if err != nil {
		t.Fatalf("unexpected error from open function: %v", err)
	}
	defer db.Close()

	// Ping triggers a real connection attempt, which will use our mock dialer.
	err = db.Ping()
	if err == nil {
		t.Fatal("expected error from Ping, got nil")
	}
	// The Vertica driver wraps the dial error in its own
	// connection failure message, so check that the original
	// error message is present.
	if !strings.Contains(err.Error(), expectErr.Error()) {
		t.Fatalf("expected error containing %q, got: %v", expectErr, err)
	}
}

func TestAthenaOpenFunctionDialer(t *testing.T) {
	var dialerCalled bool
	dialf := func(_ context.Context, _, _ string) (net.Conn, error) {
		dialerCalled = true
		return nil, errors.New("test dial error")
	}

	deps := mockDeps{
		Deps: flux.NewDefaultDependencies(),
		httpClient: &http.Client{
			Transport: fhttp.NewTransport(dialf),
		},
	}

	openFn := athenaOpenFunc("s3://bucket/?accessID=ABCD123&region=us-east-1&secretAccessKey=SECRET&db=test")
	db, err := openFn(deps)
	if err != nil {
		t.Fatalf("unexpected error from open function: %v", err)
	}
	defer db.Close()

	// Ping triggers a real connection attempt through the HTTP client,
	// which uses our custom dialer in its transport. The AWS SDK
	// wraps transport errors so the original error is not preserved,
	// but we can verify the dialer was called.
	_ = db.Ping()
	if !dialerCalled {
		t.Fatal("expected injected dialer to be called, but it was not")
	}
}

func TestFromSqliteUrlValidation(t *testing.T) {
	testCases := executetest.SourceUrlValidationTestCases{
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
