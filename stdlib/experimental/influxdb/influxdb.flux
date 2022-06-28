// Package influxdb provides tools for working with the InfluxDB API.
//
// ## Metadata
// introduced: 0.114.0
//
package influxdb


// api submits an HTTP request to the specified InfluxDB API path and returns a
// record containing the HTTP status code, response headers, and the response body.
//
// **Note**: `influxdb.api()` uses the authorization of the specified `token` or, if executed
// from the InfluxDB UI, the authorization of the InfluxDB user that invokes the script.
// Authorization permissions and limits apply to each request.
//
// ## Response format
// `influxdb.api()` returns a record with the following properties:
//
// - **statusCode**: HTTP status code returned by the GET request (int).
// - **headers**: HTTP response headers (dict).
// - **body**: HTTP response body (bytes).
//
// ## Parameters
// - method: HTTP request method.
// - path: InfluxDB API path.
// - host: InfluxDB host URL _(Required when executed outside of InfluxDB)_.
//   Default is `""`.
// - token: [InfluxDB API token](https://docs.influxdata.com/influxdb/cloud/security/tokens/)
//   _(Required when executed outside of InfluxDB)_.
//   Default is `""`.
// - headers: HTTP request headers.
// - query: URL query parameters.
// - timeout: HTTP request timeout. Default is `30s`.
// - body: HTTP request body as bytes.
//
// ## Examples
// ### Retrieve the health of an InfluxDB OSS instance
// ```no_run
// import "experimental/influxdb"
// import "influxdata/influxdb/secrets"
//
// token = secrets.get(key: "INFLUX_TOKEN")
//
// response = influxdb.api(method: "get", path: "/health", host: "http://localhost:8086", token: token)
//
// string(v: response.body)
// ```
//
// ### Create a bucket through the InfluxDB Cloud API
// ```no_run
// import "experimental/influxdb"
// import "json"
// import "influxdata/influxdb/secrets"
//
// token = secrets.get(key: "INFLUX_TOKEN")
//
// influxdb.api(
//     method: "post",
//     path: "/api/v2/buckets",
//     host: "https://us-west-2-1.aws.cloud2.influxdata.com",
//     token: token,
//     body: json.encode(v: {name: "example-bucket", description: "This is an example bucket.", orgID: "x000X0x0xx0X00x0", retentionRules: []}),
// )
// ```
//
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
