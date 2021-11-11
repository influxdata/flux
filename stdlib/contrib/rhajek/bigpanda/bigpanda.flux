// Package bigpanda provides functions for sending alerts to [BigPanda](https://www.bigpanda.io/).
package bigpanda


import "http"
import "json"
import "strings"

// defaultUrl is the default [BigPanda alerts API URL](https://docs.bigpanda.io/reference#alerts-how-it-works)
// for functions in the `bigpanda` package.
// Default is `https://api.bigpanda.io/data/v2/alerts`.
option defaultUrl = "https://api.bigpanda.io/data/v2/alerts"

// defaultTokenPrefix is the default HTTP authentication scheme to use when authenticating with BigPanda.
// Default is `Bearer`.
option defaultTokenPrefix = "Bearer"

// statusFromLevel converts an alert level to a BigPanda status.
// 
// BigPanda accepts one of ok, warning, or critical,.
//
// ## Parameters
// 
// - level: Alert level.
//
//   ##### Supported alert levels
//
//   | Alert level | BigPanda status |
//   | :---------- | :--------------|
//   | crit        | critical        |
//   | warn        | warning         |
//   | info        | ok              |
//   | ok          | ok              |
//
//   _All other alert levels return a `critical` BigPanda status._
//
// ## Examples
// ### Convert an alert level to a BigPanda status
// ```
// import "contrib/rhajek/bigpanda"
//
// bigpanda.statusFromLevel(level: "crit")
//
// // Returns "critical"
// ```
//
// ### Convert alert levels in a stream of tables to BigPanda statuses
// Use `map()` to iterate over rows in a stream of tables and convert alert levels to Big Panda statuses.
//
// ```
// # import "array"
// import "contrib/rhajek/bigpanda"
//
// # data = array.from(
// #     rows: [
// #         {_time: 2021-01-01T00:00:00Z, _level: "ok"},
// #         {_time: 2021-01-01T00:01:00Z, _level: "info"},
// #         {_time: 2021-01-01T00:02:00Z, _level: "warn"},
// #         {_time: 2021-01-01T00:03:00Z, _level: "crit"},
// #     ],
// # )
// # 
// < data
//     |> map(
//         fn: (r) => ({r with
//             big_panda_status: bigpanda.statusFromLevel(level: r._level),
//         }),
// >     )
// ```
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

// sendAlert sends an alert to [BigPanda](https://www.bigpanda.io/). 
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
