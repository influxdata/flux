package sql

import (
	neturl "net/url"
	"strings"

	"github.com/bonitoo-io/go-sql-bigquery"
	"github.com/go-sql-driver/mysql"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/dependencies/url"
	"github.com/influxdata/flux/internal/errors"
	"github.com/snowflakedb/gosnowflake"
)

// helper function to validate the data source url (postgres, sqlmock) / dsn (mysql, snowflake) using the URLValidator.
func validateDataSource(validator url.Validator, driverName string, dataSourceName string) error {

	/*
		NOTE: some parsers don't return an error for an "empty path" (a path consisting of nothing at all, or only whitespace) - not an error as such, but here we rely on the driver implementation "doing the right thing"
		better not to, and flag this as an error because calling any SQL DB with an empty DSN is likely wrong.
	*/
	if strings.TrimSpace(dataSourceName) == "" {
		return errors.Newf(codes.Invalid, "invalid data source url: %v", "empty path supplied")
	}

	var u *neturl.URL
	var err error

	switch driverName {
	case "mysql":
		// an example is: username:password@tcp(localhost:3306)/dbname?param=value
		cfg, err := mysql.ParseDSN(dataSourceName)
		if err != nil {
			return errors.Newf(codes.Invalid, "invalid data source dsn: %v", err)
		}
		u = &neturl.URL{
			Scheme: cfg.Net,
			User:   neturl.UserPassword(cfg.User, cfg.Passwd),
			Host:   cfg.Addr,
		}
	case "postgres", "sqlmock":
		// an example for postgres data source is: postgres://pqgotest:password@localhost/pqgotest?sslmode=verify-full
		// this follows the URI semantics
		u, err = neturl.Parse(dataSourceName)
		if err != nil {
			return errors.Newf(codes.Invalid, "invalid data source url: %v", err)
		}
	case "sqlite3":
		/*
			example SQLite is: file:test.db?cache=shared&mode=memory
			SQLite supports a superset of DSNs, including several special cases that net/url will flag as errors:
			:memory:
			file::memory:

			so we need to check for these, otherwise will flag as an error
		*/
		if dataSourceName == ":memory:" || dataSourceName == "file::memory:" {
			return nil
		}
		// we have a dsn that MIGHT be valid, so need to parse it - if it fails here, it is likely to be invalid
		u, err = neturl.Parse(dataSourceName)
		if err != nil {
			return errors.Newf(codes.Invalid, "invalid data source url: %v", err)
		}
	case "snowflake":
		// an example is: username:password@accountname/dbname/testschema?warehouse=mywh
		cfg, err := gosnowflake.ParseDSN(dataSourceName)
		if err != nil {
			return errors.Newf(codes.Invalid, "invalid data source dsn: %v", err)
		}
		u = &neturl.URL{
			Scheme: cfg.Protocol,
			User:   neturl.UserPassword(cfg.User, cfg.Password),
			Host:   cfg.Host,
		}
	case "mssql", "sqlserver":
		// URL example: sqlserver://sa:mypass@localhost:1234?database=master
		// ADO example: server=localhost;user id=sa;database=master
		cfg, err := mssqlParseDSN(dataSourceName)
		if err != nil {
			return errors.Newf(codes.Invalid, "invalid data source dsn: %v", err)
		}
		u = &neturl.URL{
			Scheme: cfg.Scheme,
			User:   neturl.UserPassword(cfg.User, cfg.Password),
			Host:   cfg.Host,
		}
	case "awsathena":
		// an example is: s3://bucketname/?region=us-west-1&db=dbname&accessID=AKI...&secretAccessKey=NnQ7...
		u, err = neturl.Parse(dataSourceName)
		if err != nil {
			return errors.Newf(codes.Invalid, "invalid data source url: %v", err)
		}
	case "bigquery":
		// an example is: bigquery://projectid/location?dataset=datasetid
		cfg, err := bigquery.ConfigFromConnString(dataSourceName)
		if err != nil {
			return errors.Newf(codes.Invalid, "invalid data source dsn: %v", err)
		}
		u = &neturl.URL{
			Scheme: "bigquery",
			Host:   cfg.ProjectID,
			Path:   cfg.Location,
		}
	case "hdb": // SAP HANA
		// an example is: hdb://user:password@host:port
		u, err = neturl.Parse(dataSourceName)
		if err != nil {
			return errors.Newf(codes.Invalid, "invalid data source url: %v", err)
		}
	default:
		return errors.Newf(codes.Invalid, "sql driver %s not supported", driverName)
	}

	if err = validator.Validate(u); err != nil {
		return errors.Newf(codes.Invalid, "data source did not pass url validation: %v", err)
	}
	return nil

}
