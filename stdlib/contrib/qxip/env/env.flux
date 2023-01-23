// Package env provides a function for reading environment variables starting with the `FLUX_` prefix.
//
// ## Metadata
// introduced: NEXT
// tags: secrets,security,env
//
package env


// get retrieves an environment variable from the process ENV.
//
// ## Parameters
// - key: ENV key to retrieve.
//
// ## Examples
//
// ### Retrieve an environment variable
// ```no_run
// import "contrib/qxip/env"
//
// env.get(key: "FLUX_KEY_NAME")
// ```
//
// ### Populate sensitive credentials with ENV variables
// ```no_run
// import "sql"
// import "contrib/qxip/env"
//
// username = env.get(key: "FLUX_USERNAME")
// password = env.get(key: "FLUX_PASSWORD")
//
// sql.from(
//     driverName: "postgres",
//     dataSourceName: "postgresql://${username}:${password}@localhost",
//     query: "SELECT * FROM example-table",
// )
// ```
//
builtin get : (key: string) => string
