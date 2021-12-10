// Package usage provides tools for collecting usage and usage limit data from
// **InfluxDB Cloud**.
// 
// introduced: 0.114.0
// 
package usage


import "csv"
import "experimental/influxdb"
import "experimental/json"
import "http"

// from returns usage data from an **InfluxDB Cloud** organization.
// 
// ### Output data schema
// - **http_request** measurement
//   - **req_bytes** field
//   - **resp_bytes** field
//   - **org_id** tag
//   - **endpoint** tag
//   - **status** tag
// - **query_count** measurement
//   - **req_bytes** field
//   - **endpoint** tag
//   - **orgID** tag
//   - **status** tag
// - **storage_usage_bucket_bytes** measurement
//   - **gauge** field
//   - **bucket_id** tag
//   - **org_id** tag
// 
// ## Parameters
// - start: Earliest time to include in results.
// - stop: Latest time to include in results.
// - host: [InfluxDB Cloud region URL](https://docs.influxdata.com/influxdb/cloud/reference/regions/).
//   Default is `""`.
// 
//   _(Required if executed outside of your InfluxDB Cloud organization or region)_.
// 
// - orgID: InfluxDB Cloud organization ID. Default is `""`.
// 
//   _(Required if executed outside of your InfluxDB Cloud organization or region)_.
// 
// - token: InfluxDB Cloud [API token](https://docs.influxdata.com/influxdb/cloud/security/tokens/).
//   Default is `""`.
// 
//   _(Required if executed outside of your InfluxDB Cloud organization or region)_.
// 
// - raw: Return raw, high resolution usage data instead of downsampled usage data.
//   Default is `false`.
// 
//     `usage.from()` can query the following time ranges:
//  
//     | Data resolution | Maximum time range |
//     | :-------------- | -----------------: |
//     | raw             |             1 hour |
//     | downsampled     |            30 days |
// 
// ## Examples
// 
// ### Query downsampled usage data for your InfluxDB Cloud organization
// ```no_run
// import "experimental/usage"
// import "influxdata/influxdb/secrets"
// 
// token = secrets.get(key: "INFLUX_TOKEN")
// 
// usage.from(start: -30d, stop: now())
// ```
// 
// ### Query raw usage data for your InfluxDB Cloud organization
// ```no_run
// import "experimental/usage"
// import "influxdata/influxdb/secrets"
// 
// token = secrets.get(key: "INFLUX_TOKEN")
// 
// usage.from(start: -1h, stop: now(), raw: true)
// ```
// 
// ### Query downsampled usage data for a different InfluxDB Cloud organization
// ```no_run
// import "experimental/usage"
// import "influxdata/influxdb/secrets"
// 
// token = secrets.get(key: "INFLUX_TOKEN")
// 
// usage.from(
//     start: -30d,
//     stop: now(),
//     host: "https://us-west-2-1.aws.cloud2.influxdata.com",
//     orgID: "x000X0x0xx0X00x0",
//     token: token,
// )
// ```
// 
// tags: inputs
// 
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
        query: ["start": string(v: start), "stop": string(v: stop), "raw": string(v: raw)],
    )

    return if response.statusCode > 299 then
        die(msg: "organization usage request returned status " + string(v: response.statusCode) + ": " + string(v: response.body))
    else
        csv.from(csv: string(v: response.body))
}

// limits returns a record containing usage limits for an **InfluxDB Cloud** organization.
// 
// ### Example output record
// ```
// {
//     orgID: "123",
//     rate: {
//         readKBs: 1000,
//         concurrentReadRequests: 0,
//         writeKBs: 17,
//         concurrentWriteRequests: 0,
//         cardinality: 10000,
//     },
//     bucket: {maxBuckets: 2, maxRetentionDuration: 2592000000000000},
//     task: {maxTasks: 5},
//     dashboard: {maxDashboards: 5},
//     check: {maxChecks: 2},
//     notificationRule: {maxNotifications: 2, blockedNotificationRules: "comma, delimited, list"},
//     notificationEndpoint: {blockedNotificationEndpoints: "comma, delimited, list"},
// }
// ```
// 
// ## Parameters
// - host: [InfluxDB Cloud region URL](https://docs.influxdata.com/influxdb/cloud/reference/regions/).
//   Default is `""`.
// 
//   _(Required if executed outside of your InfluxDB Cloud organization or region)_.
// 
// - orgID: InfluxDB Cloud organization ID. Default is `""`.
// 
//   _(Required if executed outside of your InfluxDB Cloud organization or region)_.
// 
// - token: InfluxDB Cloud [API token](https://docs.influxdata.com/influxdb/cloud/security/tokens/).
//   Default is `""`.
// 
//   _(Required if executed outside of your InfluxDB Cloud organization or region)_.
// 
// ## Examples
// ### Get rate limits for an InfluxDB Cloud organization
// ```no_run
// import "experimental/usage"
// import "influxdata/influxdb/secrets"
// 
// token = secrets.get(key: "INFLUX_TOKEN")
// 
// usage.limits(host: "https://us-west-2-1.aws.cloud2.influxdata.com", orgID: "x000X0x0xx0X00x0", token: token)
// ```
// 
limits = (host="", orgID="", token="") => {
    id = if orgID == "" then "{orgID}" else http.pathEscape(inputString: orgID)
    response = influxdb.api(method: "get", path: "/api/v2/orgs/" + id + "/limits", host: host, token: token)

    return if response.statusCode > 299 then
        die(msg: "organization limits request returned status " + string(v: response.statusCode) + ": " + string(v: response.body))
    else
        json.parse(data: response.body).limits
}
