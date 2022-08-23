// Package requests provides functions for transferring data using the HTTP protocol.
//
// ## Metadata
// introduced: 0.173.0
// tags: http
package requests


import "array"

_emptyBody = bytes(v: "")

// defaultConfig is the global default for all http requests using the requests package.
// Changing this config will affect all other packages using the requests package.
// To change the config for a single request, pass a new config directly into the corresponding function.
//
//
// ## Examples
//
// ### Change global configuration
//
// Modify the defaultConfig option to change all consumers of the request package.
//
// ```no_run
// import "http/requests"
//
// option requests.defaultConfig = {
//  // Set a default timeout of 10s for all requests
//  timeout: 10s,
//  insecureSkipVerify: true,
// }
// ```
//
// ### Change configuration for a single request
//
// Change the configuration for a single request. Change only the configuration values
// you need by extending the default configuration with your changes.
//
// ```
// import "http/requests"
//
// // NOTE: Flux syntax does not yet let you specify anything but an identifier
// // as the record to extend. As a workaround, this example rebinds the default configuration to a new name.
// // See https://github.com/influxdata/flux/issues/3655
// defaultConfig = requests.defaultConfig
// config = {defaultConfig with
//      // Change the timeout to 60s for this request
//      // NOTE: We don't have to specify any other properites of the config because we're
//      // extending the default.
//      timeout: 60s,
// }
// response = requests.get(url:"http://example.com", config: config)
// requests.peek(response: response)
// ```
option defaultConfig = {
    // Timeout on the request. If the timeout is zero no timeout is applied
    timeout: 0s,
    // insecureSkipVerify If true, TLS verification will not be performed. This is insecure.
    insecureSkipVerify: false,
}

// Internal method used to perform the actual request
builtin _do : (
        method: string,
        url: string,
        ?params: [string:[string]],
        ?headers: [string:string],
        ?body: bytes,
        config: {A with timeout: duration, insecureSkipVerify: bool},
    ) => {statusCode: int, body: bytes, headers: [string:string], duration: duration}

// do makes an http request.
//
// ## Parameters
// - method: method of the http request.
//      Supported methods: DELETE, GET, HEAD, PATCH, POST, PUT.
// - url: URL to request. This should not include any query parameters.
// - params: Set of key value pairs to add to the URL as query parameters.
//     Query parameters will be URL encoded.
//     All values for a key will be appended to the query.
// - headers: Set of key values pairs to include on the request.
// - body: Data to send with the request.
// - config: Set of options to control how the request should be performed.
//
// The returned response contains the following properties:
//
// - statusCode: HTTP status code returned from the request.
// - body: Contents of the request. A maximum size of 100MB will be read from the response body.
// - headers: Headers present on the response.
// - duration: Duration of request.
//
// ## Examples
//
// ### Make a GET request
//
// ```
// import "http/requests"
//
// response = requests.do(url:"http://example.com", method: "GET")
// requests.peek(response: response)
// ```
//
// ### Make a GET request that needs authorization
//
// ```no_run
// import "http/requests"
// import "influxdata/influxdb/secrets"
//
// token = secrets.get(key:"TOKEN")
//
// response = requests.do(
//     method: "GET",
//     url: "http://example.com",
//     headers: ["Authorization": "token ${token}"],
// )
//
// requests.peek(response: response)
// ```
//
// ### Make a GET request with query parameters
//
// ```no_run
// import "http/requests"
//
// response = requests.do(
//     method: "GET",
//     url: "http://example.com",
//     params: ["start": ["100"]],
// )
//
// requests.peek(response: response)
// ```
//
// ## Metadata
// tags: http,inputs
do = (
    method,
    url,
    params=[:],
    headers=[:],
    body=_emptyBody,
    config=defaultConfig,
) =>
    _do(
        method: method,
        url: url,
        params: params,
        headers: headers,
        body: body,
        config: config,
    )

// post makes a http POST request. This identical to calling `request.do(method: "POST", ...)`.
//
// ## Parameters
// - url: URL to request. This should not include any query parameters.
// - params: Set of key value pairs to add to the URL as query parameters.
//     Query parameters will be URL encoded.
//     All values for a key will be appended to the query.
// - headers: Set of key values pairs to include on the request.
// - body: Data to send with the request.
// - config: Set of options to control how the request should be performed.
//
// ## Examples
//
// ### Make a POST request with a JSON body and decode JSON response
//
// ```
// import "http/requests"
// import ejson "experimental/json"
// import "json"
// import "array"
//
// response =
//     requests.post(
//         url: "https://goolnk.com/api/v1/shorten",
//         body: json.encode(v: {url: "http://www.influxdata.com"}),
//         headers: ["Content-Type": "application/json"],
//     )
//
// data = ejson.parse(data: response.body)
//
// > array.from(rows: [data])
// ```
//
// ### Make a POST request with query parameters
//
// ```
// import "http/requests"
//
// response = requests.post(
//     url: "http://example.com",
//     params: ["start": ["100"], "interval": ["1h", "1d"]],
// )
// // Full URL: http://example.com?start=100&interval=1h&interval=1d
//
// > requests.peek(response: response)
// ```
//
// ## Metadata
// tags: http,inputs
post = (
    url,
    params=[:],
    headers=[:],
    body=_emptyBody,
    config=defaultConfig,
) =>
    do(
        method: "POST",
        url: url,
        params: params,
        headers: headers,
        body: body,
        config: config,
    )

// get makes a http GET request. This identical to calling `request.do(method: "GET", ...)`.
//
// ## Parameters
// - url: URL to request. This should not include any query parameters.
// - params: Set of key value pairs to add to the URL as query parameters.
//     Query parameters will be URL encoded.
//     All values for a key will be appended to the query.
// - headers: Set of key values pairs to include on the request.
// - body: Data to send with the request.
// - config: Set of options to control how the request should be performed.
//
// ## Examples
//
// ### Make a GET request
//
// ```no_run
// import "http/requests"
//
// response = requests.get(url:"http://example.com")
//
// requests.peek(response: response)
// ```
//
// ### Make a GET request and decode the JSON response
//
// ```
// import "http/requests"
// import "experimental/json"
// import "array"
//
// response = requests.get(
//     url: "https://api.agify.io",
//     params: ["name": ["nathaniel"]],
// )
//
// // api.agify.io returns JSON with the form
// //
// // {
// //    name: string,
// //    age: number,
// //    count: number,
// // }
// //
// // Define a data variable that parses the JSON response body into a Flux record.
// data = json.parse(data: response.body)
//
// // Use array.from() to construct a table with one row containing our response data.
// // We do not care about the count so only include name and age.
// array.from(rows:[{
//      name: data.name,
//      age: data.age,
// > }])
// ```
//
// ## Metadata
// tags: http,inputs
get = (
    url,
    params=[:],
    headers=[:],
    body=_emptyBody,
    config=defaultConfig,
) =>
    do(
        method: "GET",
        url: url,
        params: params,
        headers: headers,
        body: body,
        config: config,
    )

// peek converts an HTTP response into a table for easy inspection.
//
// The output table includes the following columns:
//  - **body** with the response body as a string
//  - **statusCode** with the returned status code as an integer
//  - **headers** with a string representation of the headers
//  - **duration** the duration of the request as a number of nanoseconds
//
// To customize how the response data is structured in a table, use `array.from()`
// with a function like `json.parse()`. Parse the response body into a set of values
// and then use `array.from()` to construct a table from those values.
//
// ## Parameters
//
// - response: Response data from an HTTP request.
//
// ## Examples
//
// ### Inspect the response of an HTTP request
//
// ```
// import "http/requests"
//
// requests.peek(response: requests.get(
//     url: "https://api.agify.io",
//     params: ["name": ["natalie"]],
// ))
// #     // We don't want the duration of the request to change
// #     // each time we run the example so set it to a static value
// #>    |> map(fn: (r) => ({r with duration: int(v:100ms)}))
// ```
peek = (response) =>
    array.from(
        rows: [
            {
                statusCode: response.statusCode,
                body: string(v: response.body),
                headers: display(v: response.headers),
                duration: int(v: response.duration),
            },
        ],
    )
