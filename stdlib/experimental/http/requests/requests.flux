// Package requests provides functions for transferring data using the HTTP protocol.
//
// introduced: 0.152.0
// tags: http
package requests


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
// import "experimental/http/requests"
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
// ```no_run
// import "experimental/http/requests"
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
// requests.get(url:"http://example.com", config: config)
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
    ) => {statusCode: int, body: bytes, headers: [string:string]}

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
//
// ## Examples
//
// ### Make a GET request
//
// ```no_run
// import "experimental/http/requests"
//
// resp = requests.do(url:"http://example.com", method: "GET")
// ```
//
// ### Make a GET request that needs authorization
//
// ```no_run
// import "experimental/http/requests"
// import "influxdata/influxdb/secrets"
//
// token = secrets.get(key:"TOKEN")
//
// resp = requests.do(
//     method: "GET",
//     url: "http://example.com",
//     headers: ["Authorization": "token ${token}"],
// )
// ```
//
// ### Make a GET request with query parameters
//
// ```no_run
// import "experimental/http/requests"
//
// resp = requests.do(
//     method: "GET",
//     url: "http://example.com",
//     params: ["start": ["100"]],
// )
// ```
//
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
// ### Make a POST request
//
// ```no_run
// import "json"
// import "experimental/http/requests"
//
// resp = requests.post(url:"http://example.com", body: json.encode(v: {data: {x:1, y: 2, z:3}))
// ```
//
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
// ### Make a GET request
//
// ```no_run
// import "experimental/http/requests"
//
// resp = requests.get(url:"http://example.com")
// ```
//
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
