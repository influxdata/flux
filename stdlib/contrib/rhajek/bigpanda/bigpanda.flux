// Package bigpanda provides functions to interact with the BigPana.
package bigpanda


import "http"
import "json"
import "strings"

// defaultUrl The default [BigPanda alerts API URL](https://docs.bigpanda.io/reference#alerts-how-it-works)
// for functions in the BigPanda package.
// Default is `https://api.bigpanda.io/data/v2/alerts`.
option defaultUrl = "https://api.bigpanda.io/data/v2/alerts"

// defaultTokenPrefix default HTTP authentication schema to use when authenticating with BigPanda.
// Default is `Bearer`.
option defaultTokenPrefix = "Bearer"

// statusFromLevel creates BigPanda status from a given level.
// 
// BigPanda accepts one of ok,critical,warning,acknowledged.
//
// ## Parameters
// 
// - level: levels on status objects can be one of the following:
//   `ok`, `info`, `warn`, `crit`, `unknown`.
statusFromLevel = (level) => {
    lvl = strings.toLower(v: level)
    sev = if lvl == "warn" then
        "warning"
    else if lvl == "crit" then
        "critical"
    else if lvl == "info" then
        "ok"
    else if lvl == "ok" then
        "ok"
    else
        "critical"

    return sev
}

// sendAlert sends a single alert to BigPanda as described in the
// bigpanda [API reference](https://docs.bigpanda.io/reference#alerts API). 
//
// ## Parameters
// 
// - token: BigPanda authorization Bearer token
// - url: base URL of [BigPanda API](https://docs.bigpanda.io/reference#alerts).
// - appKey: BigPanda App Key.
// - status: Status of the BigPanda alert. One of ok, critical, warning, acknowledged.
// - rec: Additional data appended to alert.
sendAlert = (
    url,
    token,
    appKey,
    status,
    rec,
) => {
    headers = {"Content-Type": "application/json; charset=utf-8", "Authorization": defaultTokenPrefix + " " + token}
    data = {rec with app_key: appKey, status: status}

    return http.post(headers: headers, url: url, data: json.encode(v: data))
}

// endpoint sends alerts to [BigPanda](https://www.bigpanda.io/) using data from input rows.
// 
// ## Parameters
// 
// - url: BigPanda [alerts API URL](https://docs.bigpanda.io/reference#alerts-how-it-works).
//   Default is the value of the `bigpanda.defaultURL` option.
// - token: [BigPanda API Authorization token (API key)](https://docs.bigpanda.io/docs/api-key-management).
// - appKey: BigPanda [App Key](https://docs.bigpanda.io/reference#integrating-monitoring-systems).
// 
//   The returned factory function accepts a `mapFn` parameter.
//   The `mapFn` must return an object with all properties defined
//   in the `sendAlert` function arguments (except url, apiKey and appKey).
endpoint = (url=defaultUrl, token, appKey) => (mapFn) => (tables=<-) => tables
    |> map(
        fn: (r) => {
            obj = mapFn(r: r)
    
            return {r with
                _sent: string(
                    v: 2 == sendAlert(
                        url: url,
                        appKey: appKey,
                        token: token,
                        status: obj.status,
                        rec: obj,
                    ) / 100,
                ),
            }
        },
    )
