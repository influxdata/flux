//go:build flux_nosqlite

package sql

import (
	"database/sql"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
)

// errSqliteDisabled indicates the sqlite3 driver has been disabled with the flux_nosqlite build tag.
var errSqliteDisabled = errors.Wrap(ErrorDriverDisabled, codes.Unimplemented, "sqlite3")

// isSqliteDriverEnabled indicates if the sqlite3 driver is enabled.
func isSqliteDriverEnabled() bool {
	return false
}

// sqliteDisabledDriverError returns the error the sqlite3 driver gives
// when it is disabled, or nil if it is enabled.
func sqliteDisabledDriverError() error {
	return errSqliteDisabled
}

// sqliteOpenFunction returns a function that always fails with
// errSqliteDisabled, because sqlite is disabled with the flux_nosqlite build tag.
func sqliteOpenFunction(dataSourceName string) openFunc {
	return func(flux.Dependencies) (*sql.DB, error) {
		return nil, errSqliteDisabled
	}
}
