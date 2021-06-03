package bigpanda


import "http"
import "json"
import "strings"

option defaultUrl = "https://api.bigpanda.io/data/v2/alerts"
option defaultTokenPrefix = "Bearer"

// `statusFromLevel` turns a level from the status object into a BigPanda status
// `level` - string - levels on status objects can be one of the following ok,info,warn,crit,unknown
// BigPanda accepts one of ok,critical,warning,acknowledged.
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

// `sendAlert` sends a single alert to BigPanda as described in https://docs.bigpanda.io/reference#alerts API. 
// `token` - string - BigPanda authorization Bearer token
// `url` - string - base URL of [BigPanda API](https://docs.bigpanda.io/reference#alerts).
// `appKey` - string - BigPanda App Key.
// `status` - string - Status of the BigPanda alert. One of ok, critical, warning, acknowledged.
// `rec` - record - additional data appended to alert
sendAlert = (
        url,
        token,
        appKey,
        status,
        rec,
) => {
    headers = {
        "Content-Type": "application/json; charset=utf-8",
        "Authorization": defaultTokenPrefix + " " + token,
    }
    data = {rec with app_key: appKey, status: status}

    return http.post(headers: headers, url: url, data: json.encode(v: data))
}

// `endpoint` creates a factory function that creates a target function for pipeline `|>` to send alert to BigPanda for each table row.
// `url` - string - base URL of [BigPanda API](https://docs.bigpanda.io/reference#alerts).
// `token` - string - BigPanda authorization Bearer token
// `appKey` - string - BigPanda App Key.
// The returned factory function accepts a `mapFn` parameter.
// The `mapFn` must return an object with all properties defined in the `sendAlert` function arguments (except url, apiKey and appKey).
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
