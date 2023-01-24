// Package http provides functions for transferring data using HTTP protocol.
//
// **Deprecated**: This package is deprecated in favor of [`requests`](https://docs.influxdata.com/flux/v0.x/stdlib/http/requests/).
//
// ## Metadata
// introduced: 0.39.0
// deprecated: 0.173.0
// tags: http
//
package http


// get submits an HTTP GET request to the specified URL and returns the HTTP
// status code, response body, and response headers.
//
// **Deprecated**: Experimental `http.get()` is deprecated in favor of [`requests.get()`](https://docs.influxdata.com/flux/v0.x/stdlib/http/requests/get/).
//
// ## Response format
// `http.get()` returns a record with the following properties:
//
// - **statusCode**: HTTP status code returned by the GET request (int).
// - **body**: HTTP response body (bytes).
// - **headers**: HTTP response headers (record).
//
// ## Parameters
// - url: URL to send the GET request to.
// - headers: Headers to include with the GET request.
// - timeout: Timeout for the GET request. Default is `30s`.
//
// ## Examples
// ### Get the status of an InfluxDB OSS instance
// ```no_run
// import "experimental/http"
//
// http.get(url: "http://localhost:8086/health", headers: {Authorization: "Token mY5up3RS3crE7t0k3N", Accept: "application/json"})
// ```
//
// ## Metadata
// tags: http,inputs
//
builtin get : (
        url: string,
        ?headers: A,
        ?timeout: duration,
    ) => {statusCode: int, body: bytes, headers: B}
    where
    A: Record,
    B: Record
