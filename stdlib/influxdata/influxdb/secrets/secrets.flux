// Package secrets functions for working with sensitive secrets managed by InfluxDB.
//
// ## Metadata
// introduced: 0.41.0
// tags: secrets,security
//
package secrets


// get retrieves a secret from the InfluxDB secret store.
//
// ## Parameters
// - key: Secret key to retrieve.
//
// ## Examples
//
// ### Retrive a key from the InfluxDB secret store
// ```no_run
// import "influxdata/influxdb/secrets"
//
// secrets.get(key: "KEY_NAME")
// ```
//
// ### Populate sensitive credentials with secrets//
// ```no_run
// import "sql"
// import "influxdata/influxdb/secrets"
//
// username = secrets.get(key: "POSTGRES_USERNAME")
// password = secrets.get(key: "POSTGRES_PASSWORD")
//
// sql.from(
//     driverName: "postgres",
//     dataSourceName: "postgresql://${username}:${password}@localhost",
//     query: "SELECT * FROM example-table",
// )
// ```
//
builtin get : (key: string) => string
