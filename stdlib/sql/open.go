package sql

import (
	"database/sql"
)

// There are cases when database connection cannot be open with `sql.Open(driverName, dsn)`.
// When Azure SQL server is to be accessed, it is necessary to call sql.Open(connector) function instead,
// because the connector instance has access to authorization token required for the Azure SQL database access.
// See [https://github.com/denisenkom/go-mssqldb#azure-active-directory-authentication---preview].
// Example:
//
// conn, err := mssql.NewAccessTokenConnector(
//     "Server=test.database.windows.net;Database=testdb",
//     tokenProvider)
// if err != nil {
//     handle errors in DSN
// }
// db := sql.OpenDB(conn)
//
// When user requests Azure AD authentication in SQL server connection string,
// `getOpenFunc` returns a function with body like the above example.

type openFunc func() (*sql.DB, error)

// Returns function that calls `sql.Open(driverName, dataSourceName)`
func defaultOpenFunction(driverName, dataSourceName string) openFunc {
	return func() (*sql.DB, error) {
		return sql.Open(driverName, dataSourceName)
	}
}

// Returns function that opens DB connection. For databases other than SQL Server,
// and for SQL server with SQL or Windows authentication, it simply returns `defaultOpenFunction`.
// For Azure SQL Server with Azure AD authentication, it returns a function that authenticates
// against Azure AD and uses connector with access token to open DB connection.
func getOpenFunc(driverName, dataSourceName string) openFunc {
	switch driverName {
	case "mssql", "sqlserver":
		return mssqlOpenFunction(driverName, dataSourceName)
	default:
		return defaultOpenFunction(driverName, dataSourceName)
	}
}
