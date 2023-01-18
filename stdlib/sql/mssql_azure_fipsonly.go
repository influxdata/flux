//go:build fipsonly

package sql

import "database/sql"

func mssqlOpenFunction(driverName, dataSourceName string) openFunc {
	return func() (*sql.DB, error) { return nil, errMssqlDisabled }
}
