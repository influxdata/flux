// Package servicenow provides functions for sending data to ServiceNow.
package servicenow


import "experimental/record"
import "http"
import "json"

// event sends event to ServiceNow.
//
// ## Parameters
//
// - `url` - web service URL. No default.
// - `username` - username for HTTP BASIC authentication.
// - `password` - password for HTTP BASIC authentication.
// - `source` - source that generated the event. Default is "Flux".
// - `node` - node name, IP address etc associated with the event. Default is empty string.
// - `metricType` - metric type to which the event is related. Default is empty string.
// - `resource` - node resource to which the event is related. Default is empty string.
// - `metricName` - metric name. Default is empty string.
// - `messageKey` - unique identification of the event, eg. InfluxDB alert ID. Default is empty string (ServiceNow will fill the value).
// - `description` - text describing the event.
// - `severity` - severity of the event. One of "critical", "major", "minor", "warning", "info" or "clear".
// - `additionalInfo` - additional information (optional).
//
// ## Example
//
// ```
//  import "contrib/bonitoo-io/servicenow"
//  import "influxdata/influxdb/secrets"
//  import "strings"
//
//  username = secrets.get(key: "SERVICENOW_USERNAME")
//  password = secrets.get(key: "SERVICENOW_PASSWORD")
//
//  lastReported =
//    from(bucket: "example-bucket")
//      |> range(start: -1m)
//      |> filter(fn: (r) => r._measurement == "cpu" and r._field == "usage_idle")
//      |> last()
//      |> tableFind(fn: (key) => true)
//      |> getRecord(idx: 0)
//
//  servicenow.event(
//      url: "https://tenant.service-now.com/api/global/em/jsonv2",
//      username: username,
//      password: password,
//      node: lastReported.host,
//      metricType: strings.toUpper(v: lastReported._measurement),
//      resource: lastReported.instance,
//      metricName: lastReported._field,
//      severity:
//          if lastReported._value < 1.0 then "critical"
//          else if lastReported._value < 5.0 then "warning"
//          else "info",
//  )
// ```
//
event = (
    url,
    username,
    password,
    source="Flux",
    node="",
    metricType="",
    resource="",
    metricName="",
    messageKey="",
    description,
    severity,
    additionalInfo=record.any,
) => {
    event = {
        source: source,
        node: node,
        type: metricType,
        resource: resource,
        metric_name: metricName,
        message_key: messageKey,
        description: description,
        severity: if severity == "critical" then
            "1"
        else if severity == "major" then
            "2"
        else if severity == "minor" then
            "3"
        else if severity == "warning" then
            "4"
        else if severity == "info" then
            "5"
        else if severity == "clear" then
            "0"
        else
            "",
        // shouldn't happen
        additional_info: if additionalInfo == record.any then "" else string(v: json.encode(v: additionalInfo)),
    }
    payload = {records: [event]}
    headers = {"Authorization": http.basicAuth(u: username, p: password), "Content-Type": "application/json"}
    body = json.encode(v: payload)

    return http.post(headers: headers, url: url, data: body)
}

// endpoint creates the endpoint for the ServiceNow external service.
//
// ## Parameters
//
// - `url` ServiceNow web service URL.
// - `username` username for HTTP BASIC authentication. Default is empty string (no authentication).
// - `password` password for HTTP BASIC authentication. Default is empty string (no authentication).
// - `source` source that generated the event. Default is "Flux".
//
// The returned factory function accepts a `mapFn` parameter.
// The `mapFn` must return an object with `node`, `metricType`, `resource`, `metricName`, `messageKey`, `description`,
// `severity` and `additionalInfo` fields as defined in the `event` function arguments.
//
// ## Example
//
// ```
//  import "contrib/bonitoo-io/servicenow"
//  import "influxdata/influxdb/secrets"
//  import "strings"
//
//  username = secrets.get(key: "SERVICENOW_USERNAME")
//  password = secrets.get(key: "SERVICENOW_PASSWORD")
//
//  endpoint = servicenow.endpoint(
//      url: "https://tenant.service-now.com/api/global/em/jsonv2",
//      username: username,
//      password: password
//  )
//
//  from(bucket: "example-bucket")
//    |> range(start: -1m)
//    |> filter(fn: (r) => r._measurement == "cpu" and r._field == "usage_idle")
//    |> last()
//    |> endpoint(mapFn: (r) => ({
//        node: r.host,
//        metricType: strings.toUpper(v: r._measurement),
//        resource: r.instance,
//        metricName: r._field,
//        severity:
//            if r._value < 1.0 then "critical"
//            else if r._value < 5.0 then "warning"
//            else "info",
//        additionalInfo: { "devId": r.dev_id }
//      })
//    )()
// ```
//
endpoint = (url, username, password, source="Flux") => (mapFn) => (tables=<-) => tables
    |> map(
        fn: (r) => {
            obj = mapFn(r: r)
    
            return {r with
                _sent: string(
                    v: 2 == event(
                        url: url,
                        username: username,
                        password: password,
                        source: source,
                        node: obj.node,
                        metricType: obj.metricType,
                        resource: obj.resource,
                        metricName: obj.metricName,
                        messageKey: obj.messageKey,
                        description: obj.description,
                        severity: obj.severity,
                        additionalInfo: record.get(r: obj, key: "additionalInfo", default: record.any),
                    ) / 100,
                ),
            }
        },
    )
