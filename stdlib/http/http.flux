package http

//The Flux HTTP package provides functions for transferring data using the HTTP protocol.
//

// Post submits an HTTP post request to the specified URL with headers and data
// and returns the HTTP Status Code
//
// ## Parameters
//
// - `url` is the URL to post to
// - `headers` are the headers to include with the Post request
//
//      Header keys with special characters:
//      Wrap header keys that contain special characters in double quotes ("").
//
// - `data` is the data body to include with the Post request
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
//
// ```
//
// The http.basicAuth() function returns a Base64-encoded basic authentication header
// using a specified username and password combination.
//
// ## Parameters
//
// - `u` is the username to use in the basic authentication header.
// - `p` is the password to use in the basic authentication header.
//
## Set a basic authentication header in an HTTP POST request
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

builtin basicAuth : (u: string, p: string) => string

// PathEscape escapes the string so it can be safely placed inside a URL path segment
// replacing special characters (including /) with %XX sequences as needed.
builtin pathEscape : (inputString: string) => string

// Endpoint returns a function that acts as a notification endpoint,
// which will post the notification to the url.
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
