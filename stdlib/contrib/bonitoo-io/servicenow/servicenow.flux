// Package servicenow  provides functions for sending events to [ServiceNow](https://www.servicenow.com/).
//
// ## Metadata
// introduced: 0.136.0
// contributors: **GitHub**: [@alespour](https://github.com/alespour), [@bonitoo-io](https://github.com/bonitoo-io) | **InfluxDB Slack**: [@Ales Pour](https://influxdata.com/slack)
//
package servicenow


import "experimental/record"
import "http"
import "json"

// event sends an event to [ServiceNow](https://servicenow.com/).
//
// ServiceNow Event API fields are described in
// [ServiceNow Create Event documentation](https://docs.servicenow.com/bundle/paris-it-operations-management/page/product/event-management/task/t_EMCreateEventManually.html).
//
// ## Parameters
//
// - url: ServiceNow web service URL.
// - username: ServiceNow username to use for HTTP BASIC authentication.
// - password: ServiceNow password to use for HTTP BASIC authentication.
// - description: Event description.
// - severity: Severity of the event.
//
//   Supported values:
//   - `critical`
//   - `major`
//   - `minor`
//   - `warning`
//   - `info`
//   - `clear`
// - source: Source name. Default is `"Flux"`.
// - node: Node name or IP address related to the event.
//   Default is an empty string (`""`).
// - metricType: Metric type related to the event (for example, `CPU`).
//   Default is an empty string (`""`).
// - resource: Resource related to the event (for example, `CPU-1`).
//   Default is an empty string (`""`).
// - metricName: Metric name related to the event (for example, `usage_idle`).
//   Default is an empty string (`""`).
// - messageKey: Unique identifier of the event (for example, the InfluxDB alert ID).
//   Default is an empty string (`""`).
//   If an empty string, ServiceNow generates a value.
// - additionalInfo: Additional information to include with the event.
//
// ## Examples
// ### Send the last reported value and incident type to ServiceNow
// ```no_run
// import "contrib/bonitoo-io/servicenow"
// import "influxdata/influxdb/secrets"
//
// username = secrets.get(key: "SERVICENOW_USERNAME")
// password = secrets.get(key: "SERVICENOW_PASSWORD")
//
// lastReported = from(bucket: "example-bucket")
//     |> range(start: -1m)
//     |> filter(fn: (r) => r._measurement == "cpu" and r._field == "usage_idle")
//     |> last()
//     |> findRecord(fn: (key) => true, idx: 0)
//
// servicenow.event(
//     url: "https://tenant.service-now.com/api/global/em/jsonv2",
//     username: username,
//     password: password,
//     node: lastReported.host,
//     metricType: lastReported._measurement,
//     resource: lastReported.instance,
//     metricName: lastReported._field,
//     severity: if lastReported._value < 1.0 then
//         "critical"
//     else if lastReported._value < 5.0 then
//         "warning"
//     else
//         "info",
//     additionalInfo: {"devId": r.dev_id},
// )
// ```
//
// ## Metadata
// tags: single notification
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
    ) =>
    {
        event = {
            source: source,
            node: node,
            type: metricType,
            resource: resource,
            metric_name: metricName,
            message_key: messageKey,
            description: description,
            severity:
                if severity == "critical" then
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
            additional_info:
                if additionalInfo == record.any then
                    ""
                else
                    string(v: json.encode(v: additionalInfo)),
        }
        payload = {records: [event]}
        headers = {
            "Authorization": http.basicAuth(u: username, p: password),
            "Content-Type": "application/json",
        }
        body = json.encode(v: payload)

        return http.post(headers: headers, url: url, data: body)
    }

// endpoint sends events to [ServiceNow](https://servicenow.com/) using data from input rows.
//
// ### Usage
//
// `servicenow.endpoint` is a factory function that outputs another function.
// The output function requires a `mapFn` parameter.
//
// #### mapFn
// A function that builds the object used to generate the ServiceNow API request. Requires an `r` parameter.
//
// `mapFn` accepts a table row (`r`) and returns an object that must include the following properties:
//
// - `description`
// - `severity`
// - `source`
// - `node`
// - `metricType`
// - `resource`
// - `metricName`
// - `messageKey`
// - `additionalInfo`
//
// For more information, see `servicenow.event()` parameters.
//
// ## Parameters
//
// - url: ServiceNow web service URL.
// - username: ServiceNow username to use for HTTP BASIC authentication.
// - password: ServiceNow password to use for HTTP BASIC authentication.
// - source: Source name. Default is `"Flux"`.
//
// ## Examples
// ### Send critical events to ServiceNow
//
// ```no_run
// import "contrib/bonitoo-io/servicenow"
// import "influxdata/influxdb/secrets"
//
// username = secrets.get(key: "SERVICENOW_USERNAME")
// password = secrets.get(key: "SERVICENOW_PASSWORD")
//
// endpoint = servicenow.endpoint(
//     url: "https://example-tenant.service-now.com/api/global/em/jsonv2",
//     username: username,
//     password: password
// )
//
// crit_events = from(bucket: "example-bucket")
//     |> range(start: -1m)
//     |> filter(fn: (r) => r._measurement == "statuses" and status == "crit")
//
// crit_events
//     |> endpoint(mapFn: (r) => ({
//         node: r.host,
//         metricType: r._measurement,
//         resource: r.instance,
//         metricName: r._field,
//         severity: "critical",
//         additionalInfo: { "devId": r.dev_id }
//       })
//     )()
// ```
//
// ## Metadata
// tags: notification endpoints
endpoint = (url, username, password, source="Flux") =>
    (mapFn) =>
        (tables=<-) =>
            tables
                |> map(
                    fn: (r) => {
                        obj = mapFn(r: r)

                        return {r with _sent:
                                string(
                                    v:
                                        2 == event(
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
                                                additionalInfo:
                                                    record.get(
                                                        r: obj,
                                                        key: "additionalInfo",
                                                        default: record.any,
                                                    ),
                                            ) / 100,
                                ),
                        }
                    },
                )
