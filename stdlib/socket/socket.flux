// Package socket provides tools for returning data from socket connections.
//
// ## Metadata
// introduced: 0.21.0
//
package socket


// from returns data from a socket connection and outputs a stream of tables
// given a specified decoder.
//
// The function produces a single table for everything that it receives from the
// start to the end of the connection.
//
// ## Parameters
// - url: URL to return data from.
//
//   **Supported URL schemes**:
//   - tcp
//   - unix
//
// - decoder: Decoder to use to parse returned data into a stream of tables.
//
//   **Supported decoders**:
//   - csv
//   - line
//
// ## Examples
//
// ### Query annotated CSV from a socket connection
// ```no_run
// import "socket"
//
// socket.from(url: "tcp://127.0.0.1:1234", decoder: "csv")
// ```
//
// ### Query line protocol from a socket connection
// ```no_run
// import "socket"
//
// socket.from(url: "tcp://127.0.0.1:1234", decoder: "line")
// ```
//
// ## Metadata
// tags: inputs
//
builtin from : (url: string, ?decoder: string) => stream[A]
