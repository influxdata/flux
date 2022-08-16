// Package http provides functions for transferring data using the HTTP protocol.
//
// ## Metadata
// introduced: 0.39.0
package http


import "experimental"

// post sends an HTTP POST request to the specified URL with headers and data
// and returns the HTTP status code.
//
// ## Parameters
//
// - url: URL to send the POST request to.
// - headers: Headers to include with the POST request.
//
//    **Header keys with special characters:**
//    Wrap header keys that contain special characters in double quotes (`""`).
//
// - data: Data body to include with the POST request.
//
// ## Examples
//
// ### Send the last reported status to a URL
// ```no_run
// import "json"
// import "http"
//
// lastReported = from(bucket: "example-bucket")
//     |> range(start: -1m)
//     |> filter(fn: (r) => r._measurement == "statuses")
//     |> last()
//     |> findColumn(fn: (key) => true, column: "_level")
//
// http.post(
//     url: "http://myawsomeurl.com/api/notify",
//     headers: {
//         Authorization: "Bearer mySuPerSecRetTokEn",
//         "Content-type": "application/json",
//     },
//     data: json.encode(v: lastReported[0]),
// )
// ```
//
// ## Metadata
// introduced: 0.40.0
// tags: single notification
//
builtin post : (url: string, ?headers: A, ?data: bytes) => int where A: Record

// basicAuth returns a Base64-encoded basic authentication header
// using a specified username and password combination.
//
// ## Parameters
//
// - u: Username to use in the basic authentication header.
// - p: Password to use in the basic authentication header.
//
// ## Examples
//
// ### Set a basic authentication header in an HTTP POST request
// ```no_run
// import "http"
//
// username = "myawesomeuser"
// password = "mySupErSecRetPasSW0rD"
//
// http.post(
//     url: "http://myawesomesite.com/api/",
//     headers: {Authorization: http.basicAuth(u: username, p: password)},
//     data: bytes(v: "Something I want to send."),
// )
// ```
//
// ## Metadata
// introduced: 0.44.0
//
builtin basicAuth : (u: string, p: string) => string

// pathEscape escapes special characters in a string (including `/`)
// and replaces non-ASCII characters with hexadecimal representations (`%XX`).
//
// ## Parameters
// - inputString: String to escape.
//
// ## Examples
//
// ### URL-encode a string
// ```no_run
// import "http"
//
// http.pathEscape(inputString: "Hello world!")
//
// // Returns "Hello%20world%21"
// ```
//
// ### URL-encode strings in a stream of tables
// ```
// import "http"
// import "sampledata"
//
// < sampledata.string()
//     |> map(
//         fn: (r) => ({r with
//             _value: http.pathEscape(inputString: r._value),
//         }),
// >     )
// ```
//
// ## Metadata
// introduced: 0.71.0
//
builtin pathEscape : (inputString: string) => string

// endpoint iterates over input data and sends a single POST request per input row to
// a specficied URL.
//
// This function is designed to be used with `monitor.notify()`.
//
// `http.endpoint()` outputs a function that requires a `mapFn` parameter.
// `mapFn` is the function that builds the record used to generate the POST request.
// It accepts a table row (`r`) and returns a record that must include the
// following properties:
//
// - `headers`
// - `data`
//
// _For information about properties, see `http.post`._
//
// ## Parameters
// - url: URL to send the POST reqeust to.
//
// ## Examples
//
// ### Send an HTTP POST request for each row
// ```no_run
// import "http"
// import "sampledata"
//
// endpoint = http.endpoint(url: "http://example.com/")(
//     mapfn: (r) => ({
//         headers: {header1: "example1", header2: "example2"},
//         data: bytes(v: "The value is ${r._value}"),
//     }),
// )
//
// sampledata.int()
//     |> endpoint()
// ```
//
// ## Metadata
// tags: notification endpoints
//
endpoint = (url) =>
    (mapFn) =>
        (tables=<-) =>
            tables
                |> map(
                    fn: (r) => {
                        obj = mapFn(r: r)

                        return {r with _sent:
                                string(
                                    v: 200 == post(url: url, headers: obj.headers, data: obj.data),
                                ),
                        }
                    },
                )
                |> experimental.group(mode: "extend", columns: ["_sent"])
