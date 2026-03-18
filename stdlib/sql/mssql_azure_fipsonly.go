//go:build fipsonly

package sql

import (
	"database/sql"

	"github.com/influxdata/flux"
)

func mssqlOpenFunction(dataSourceName string) openFunc {
	return func(flux.Dependencies) (*sql.DB, error) { return nil, errMssqlDisabled }
}
