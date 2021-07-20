// Package http provides functions for transferring data using the HTTP protocol.
package http


import "experimental"

// post submits an HTTP POST request to the specified URL with headers and data
// and returns the HTTP Status Code
//
// ## Parameters
//
// - `url` is the URL to POST to
// - `headers` are the headers to include with the POST request
//
//      Header keys with special characters:
//          Wrap header keys that contain special characters in double quotes ("").
//
// - `data` is the data body to include with the POST request
//
// ## Send the last reported status to a URL
//
// ```
// import "json"
// import "http"
//
// lastReported =
//  from(bucket: "example-bucket")
//    |> range(start: -1m)
//    |> filter(fn: (r) => r._measurement == "statuses")
//    |> last()
//    |> findColumn(fn: (key) => true, column: "_level")
//
//  http.post(
//  url: "http://myawsomeurl.com/api/notify",
//  headers: {
//    Authorization: "Bearer mySuPerSecRetTokEn",
//    "Content-type": "application/json"
//  },
//  data: json.encode(v: lastReported[0])
// )
// ```
//
builtin post : (url: string, ?headers: A, ?data: bytes) => int where A: Record

// basicAuth returns a Base64-encoded basic authentication header
// using a specified username and password combination.
//
// ## Parameters
//
// - `u` is the username to use in the basic authentication header.
// - `p` is the password to use in the basic authentication header.
//
// ## Set a basic authentication header in an HTTP POST request
//
// ```
// import "json"
//
// from(bucket: "example-bucket")
//   |> range(start: -1h)
//   |> map(fn: (r) => ({
//       r with _value: json.encode(v: r._value)
//   }))
// ```
//
builtin basicAuth : (u: string, p: string) => string

// pathEscape() escapes special characters in a string (including /)
// and replaces non-ASCII characters with hexadecimal representations (%XX).
//
// ## Parameters
//
// - `inputString` is the string to escape.
//
// ## Set a basic authentication header in an HTTP POST request
//
// ```
// import "http"
//
// data
//   |> map(fn: (r) => ({ r with
//     path: http.pathEscape(inputString: r.path)
//   }))
// ```
//
builtin pathEscape : (inputString: string) => string

// endpoint sends output data
// to an HTTP URL using the POST request method.
//
// ## Parameters
//
// - `url` is the URL to POST to.
// - `mapFn` is a function that builds the record used to generate the POST request.
//     - mapFn accepts a table row (r) and returns a record that must include the following fields:
//          - `headers`
//          - `data`
//
// See influxdata/influxdb/monitor.notify
endpoint = (url) => (mapFn) => (tables=<-) => tables
    |> map(
        fn: (r) => {
            obj = mapFn(r: r)

            return {r with
                _sent: string(v: 200 == post(url: url, headers: obj.headers, data: obj.data)),
            }
        },
    )
    |> experimental.group(mode: "extend", columns: ["_sent"])
