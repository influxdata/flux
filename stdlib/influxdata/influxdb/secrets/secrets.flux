// Flux InfluxDB Secrets package provides functions and tools for working
// with sensitive secrets managed by InfluxDB.
package secrets


// get is a function that retrieves a secret from the InfluxDB
//  secret store.
//
// ## Parameters
// - `key` is the secret key to retrieve.
//
// ## Example
//
// ```
// import "influxdata/influxdb/secrets"
//
// secrets.get(key: "KEY_NAME")
// ```
//
// ## Populate sensitive credentials with secrets
//
// ```
// import "sql"
// import "influxdata/influxdb/secrets"
//
// username = secrets.get(key: "POSTGRES_USERNAME")
// password = secrets.get(key: "POSTGRES_PASSWORD")
//
// sql.from(
//   driverName: "postgres",
//   dataSourceName: "postgresql://${username}:${password}@localhost",
//   query:"SELECT * FROM example-table"
// )
// ```
builtin get : (key: string) => string
