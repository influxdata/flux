package usage


import "csv"
import "experimental/influxdb"
import "experimental/json"
import "http"

// from returns an organization's usage data. The time range to query is
// bounded by start and stop arguments. Optional orgID, host and token arguments
// allow cross-org and/or cross-cluster queries. Setting the raw parameter will
// return raw usage data rather than the downsampled data returned by default.
// Note that unlike the range function, the stop argument is required here,
// pending implementation of https://github.com/influxdata/flux/issues/3629.
from = (
        start,
        stop,
        host="",
        orgID="",
        token="",
        raw=false,
) => {
    id = if orgID == "" then "{orgID}" else http.pathEscape(inputString: orgID)
    response = influxdb.api(
        method: "get",
        path: "/api/v2/orgs/" + id + "/usage",
        host: host,
        token: token,
        query: [
            "start": string(v: start),
            "stop": string(v: stop),
            "raw": string(v: raw),
        ],
    )

    return if response.statusCode > 299 then
        die(msg: "organization usage request returned status " + string(v: response.statusCode) + ": " + string(v: response.body))
    else
        csv.from(csv: string(v: response.body))
}

// limits returns an organization's usage limits. Optional orgID, host
// and token arguments allow cross-org and/or cross-cluster calls.
limits = (host="", orgID="", token="") => {
    id = if orgID == "" then "{orgID}" else http.pathEscape(inputString: orgID)
    response = influxdb.api(
        method: "get",
        path: "/api/v2/orgs/" + id + "/limits",
        host: host,
        token: token,
    )

    return if response.statusCode > 299 then
        die(msg: "organization limits request returned status " + string(v: response.statusCode) + ": " + string(v: response.body))
    else
        json.parse(data: response.body).limits
}
