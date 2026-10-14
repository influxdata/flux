//go:build !flux_nosqlite

package sql

// isSqliteDriverEnabled indicates if the sqlite3 driver is enabled.
func isSqliteDriverEnabled() bool {
	return true
}

// sqliteDisabledDriverError returns the error the sqlite3 driver gives
// when it is disabled, or nil if it is enabled.
func sqliteDisabledDriverError() error {
	return nil
}

// sqliteOpenFunction returns a function that opens a sqlite3 database.
func sqliteOpenFunction(dataSourceName string) openFunc {
	return defaultOpenFunction("sqlite3", dataSourceName)
}
