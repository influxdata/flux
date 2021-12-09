// Package influxdb provides tools for working with the InfluxDB API. 
// 
// introduced: 0.114.0
// 
package influxdb


// api submits an HTTP request to the specified InfluxDB API path and returns a
// record containing the HTTP status code, response headers, and the response body.
// 
// ## Response format
// `influxdb.api()` returns a record with the following properties:
// 
// - **statusCode**: HTTP status code returned by the GET request (int).
// - **headers**: HTTP response headers (dict).
// - **body**: HTTP response body (bytes). 
// 
// ## Parameters
builtin api : (
    method: string,
    path: string,
    ?host: string,
    ?token: string,
    ?body: bytes,
    ?headers: [string:string],
    ?query: [string:string],
    ?timeout: duration,
) => {statusCode: int, body: bytes, headers: [string:string]}
