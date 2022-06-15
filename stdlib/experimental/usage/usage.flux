// Package usage provides tools for collecting usage and usage limit data from
// **InfluxDB Cloud**.
//
// ## Metadata
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
// ### Query number of bytes in requests to the /api/v2/write endpoint
// ```no_run
// import "experimental/usage"
//
// usage.from(start: -30d, stop: now())
//     |> filter(fn: (r) => r._measurement == "http_request")
//     |> filter(fn: (r) => r._field == "req_bytes")
//     |> filter(fn: (r) => r.endpoint == "/api/v2/write")
//     |> group(columns: ["_time"])
//     |> sum()
//     |> group()
// ```
//
// ### Query number of bytes returned from the /api/v2/query endpoint
// ```no_run
// import "experimental/usage"
//
// usage.from(start: -30d, stop: now())
//     |> filter(fn: (r) => r._measurement == "http_request")
//     |> filter(fn: (r) => r._field == "resp_bytes")
//     |> filter(fn: (r) => r.endpoint == "/api/v2/query")
//     |> group(columns: ["_time"])
//     |> sum()
//     |> group()
// ```
//
// ### Query the query count for InfluxDB Cloud query endpoints
// The following query returns query counts for the following query endpoints:
//
// - **/api/v2/query**: Flux queries
// - **/query**: InfluxQL queries
//
// ```no_run
// import "experimental/usage"
//
// usage.from(start: -30d, stop: now())
//     |> filter(fn: (r) => r._measurement == "query_count")
//     |> sort(columns: ["_time"])
// ```
//
// ### Compare usage metrics to organization usage limits
// The following query compares the amount of data written to and queried from your
// InfluxDB Cloud organization to your organization's rate limits.
// It appends a `limitReached` column to each row that indicates if your rate
// limit was exceeded.
//
// ```no_run
// import "experimental/usage"
//
// limits = usage.limits()
//
// checkLimit = (tables=<-, limit) => tables
//     |> map(fn: (r) => ({r with _value: r._value / 1000, limit: int(v: limit) * 60 * 5}))
//     |> map(fn: (r) => ({r with limitReached: r._value > r.limit}))
//
// read = usage.from(start: -30d, stop: now())
//     |> filter(fn: (r) => r._measurement == "http_request")
//     |> filter(fn: (r) => r._field == "resp_bytes")
//     |> filter(fn: (r) => r.endpoint == "/api/v2/query")
//     |> group(columns: ["_time"])
//     |> sum()
//     |> group()
//     |> checkLimit(limit: limits.rate.readKBs)
//
// write = usage.from(start: -30d, stop: now())
//     |> filter(fn: (r) => r._measurement == "http_request")
//     |> filter(fn: (r) => r._field == "req_bytes")
//     |> filter(fn: (r) => r.endpoint == "/api/v2/write")
//     |> group(columns: ["_time"])
//     |> sum()
//     |> group()
//     |> checkLimit(limit: limits.rate.writeKBs)
//
// union(tables: [read, write])
// ```
//
// ## Metadata
// tags: inputs
//
from = (
    start,
    stop,
    host="",
    orgID="",
    token="",
    raw=false,
) =>
{
    id = if orgID == "" then "{orgID}" else http.pathEscape(inputString: orgID)
    response =
        influxdb.api(
            method: "get",
            path: "/api/v2/orgs/" + id + "/usage",
            host: host,
            token: token,
            query: ["start": string(v: start), "stop": string(v: stop), "raw": string(v: raw)],
        )

    return
        if response.statusCode > 299 then
            die(
                msg:
                    "organization usage request returned status " + string(v: response.statusCode) + ": " + string(
                            v: response.body,
                        ),
            )
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
//
// ### Get rate limits for your InfluxDB Cloud organization
// ```no_run
// import "experimental/usage"
//
// usage.limits()
// ```
//
// ### Get rate limits for a different InfluxDB Cloud organization
// ```no_run
// import "experimental/usage"
// import "influxdata/influxdb/secrets"
//
// token = secrets.get(key: "INFLUX_TOKEN")
//
// usage.limits(host: "https://us-west-2-1.aws.cloud2.influxdata.com", orgID: "x000X0x0xx0X00x0", token: token)
// ```
//
// ### Output organization limits in a table
// ```no_run
// import "array"
// import "experimental/usage"
//
// limits = usage.limits()
//
// array.from(
//     rows: [
//         {orgID: limits.orgID, limitGroup: "rate", limitName: "Read (kb/s)", limit: limits.rate.readKBs},
//         {orgID: limits.orgID, limitGroup: "rate", limitName: "Concurrent Read Requests", limit: limits.rate.concurrentReadRequests},
//         {orgID: limits.orgID, limitGroup: "rate", limitName: "Write (kb/s)", limit: limits.rate.writeKBs},
//         {orgID: limits.orgID, limitGroup: "rate", limitName: "Concurrent Write Requests", limit: limits.rate.concurrentWriteRequests},
//         {orgID: limits.orgID, limitGroup: "rate", limitName: "Cardinality", limit: limits.rate.cardinality},
//         {orgID: limits.orgID, limitGroup: "bucket", limitName: "Max Buckets", limit: limits.bucket.maxBuckets},
//         {orgID: limits.orgID, limitGroup: "bucket", limitName: "Max Retention Period (ns)", limit: limits.bucket.maxRetentionDuration},
//         {orgID: limits.orgID, limitGroup: "task", limitName: "Max Tasks", limit: limits.task.maxTasks},
//         {orgID: limits.orgID, limitGroup: "dashboard", limitName: "Max Dashboards", limit: limits.dashboard.maxDashboards},
//         {orgID: limits.orgID, limitGroup: "check", limitName: "Max Checks", limit: limits.check.maxChecks},
//         {orgID: limits.orgID, limitGroup: "notificationRule", limitName: "Max Notification Rules", limit: limits.notificationRule.maxNotifications},
//     ],
// )
// ```
//
// ### Output current cardinality with your cardinality limit
// ```no_run
// import "experimental/usage"
// import "influxdata/influxdb"
//
// limits = usage.limits()
// bucketCardinality = (bucket) => (influxdb.cardinality(bucket: bucket, start: time(v: 0))
//     |> findColumn(fn: (key) => true, column: "_value"))[0]
//
// buckets()
//     |> filter(fn: (r) => not r.name =~ /^_/)
//     |> map(fn: (r) => ({bucket: r.name, Cardinality: bucketCardinality(bucket: r.name)}))
//     |> sum(column: "Cardinality")
//     |> map(fn: (r) => ({r with "Cardinality Limit": limits.rate.cardinality}))
// ```
//
limits = (host="", orgID="", token="") => {
    id = if orgID == "" then "{orgID}" else http.pathEscape(inputString: orgID)
    response = influxdb.api(method: "get", path: "/api/v2/orgs/" + id + "/limits", host: host, token: token)

    return
        if response.statusCode > 299 then
            die(
                msg:
                    "organization limits request returned status " + string(v: response.statusCode) + ": " + string(
                            v: response.body,
                        ),
            )
        else
            json.parse(data: response.body).limits
}
