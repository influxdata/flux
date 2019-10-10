package sql

import (
	neturl "net/url"

	"github.com/go-sql-driver/mysql"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/dependencies/url"
	"github.com/influxdata/flux/internal/errors"
)

func validateDataSource(validator url.Validator, driverName string, dataSourceName string) error {
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
	default:
		return errors.Newf(codes.Invalid, "sql driver %s not supported", driverName)
	}

	if err = validator.Validate(u); err != nil {
		return errors.Newf(codes.Invalid, "data source did not url pass validation: %v", err)
	} else {
		return nil
	}
}
