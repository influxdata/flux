// Package alerta provides functions that send alerts to [Alerta](https://alerta.io/).
//
// ## Metadata
// introduced: 0.115.0
package alerta


import "http"
import "json"
import "strings"

// alert sends an alert to [Alerta](https://alerta.io/).
//
// ## Parameters
//
// - url: (Required) Alerta URL.
// - apiKey: (Required) Alerta API key.
// - resource: (Required) Resource associated with the alert.
// - event: (Required) Event name.
// - environment: Alerta environment. Valid values: "Production", "Development" or empty string (default).
// - severity: (Required) Event severity. See [Alerta severities](https://docs.alerta.io/en/latest/api/alert.html#alert-severities).
// - service: List of affected services. Default is `[]`.
// - group: Alerta event group. Default is `""`.
// - value: Event value.  Default is `""`.
// - text: Alerta text description. Default is `""`.
// - tags: List of event tags. Default is `[]`.
// - attributes: (Required) Alert attributes.
// - origin: monitoring component.
// - type: Event type. Default is `""`.
// - timestamp: time alert was generated. Default is `now()`.
//
// ## Examples
// ### Send the last reported value and status to Alerta
// ```no_run
// import "contrib/bonitoo-io/alerta"
// import "influxdata/influxdb/secrets"
//
// apiKey = secrets.get(key: "ALERTA_API_KEY")
//
// lastReported =
//   from(bucket: "example-bucket")
//     |> range(start: -1m)
//     |> filter(fn: (r) =>
//       r._measurement == "example-measurement" and
//       r._field == "level"
//     )
//     |> last()
//     |> findRecord(fn: (key) => true, idx: 0)
//
// severity = if lastReported._value > 50 then "warning" else "ok"
//
// alerta.alert(
//   url: "https://alerta.io:8080/alert",
//   apiKey: apiKey,
//   resource: "example-resource",
//   event: "Example event",
//   environment: "Production",
//   severity: severity,
//   service: ["example-service"],
//   group: "example-group",
//   value: string(v: lastReported._value),
//   text: "Service is ${severity}. The last reported value was ${string(v: lastReported._value)}.",
//   tags: ["ex1", "ex2"],
//   attributes: {},
//   origin: "InfluxDB",
//   type: "exampleAlertType",
//   timestamp: now(),
// )
// ```
//
// ## Metadata
// tags: single notification
//
alert = (
    url,
    apiKey,
    resource,
    event,
    environment="",
    severity,
    service=[],
    group="",
    value="",
    text="",
    tags=[],
    attributes,
    origin="InfluxDB",
    type="",
    timestamp=now(),
) =>
{
    alert = {
        resource: resource,
        event: event,
        environment: environment,
        severity: severity,
        service: service,
        group: group,
        value: value,
        text: text,
        tags: tags,
        attributes: attributes,
        origin: origin,
        type: type,
        createTime: strings.substring(v: string(v: timestamp), start: 0, end: 23) + "Z",
            // Alerta supports ISO 8601 date format YYYY-MM-DDThh:mm:ss.sssZ only

    }
    headers = {"Authorization": "Key " + apiKey, "Content-Type": "application/json"}
    body = json.encode(v: alert)

    return http.post(headers: headers, url: url, data: body)
}

// endpoint sends alerts to Alerta using data from input rows.
//
// ## Parameters
//
// - url: (Required) Alerta URL.
// - apiKey: (Required) Alerta API key.
// - environment: Alert environment. Default is `""`.
//   Valid values: "Production", "Development" or empty string (default).
// - origin: Alert origin. Default is `"InfluxDB"`.
//
// ## Usage
// `alerta.endpoint` is a factory function that outputs another function.
//     The output function requires a `mapFn` parameter.
//
// ### mapFn
// A function that builds the object used to generate the POST request. Requires an `r` parameter.
//
// `mapFn` accepts a table row (`r`) and returns an object that must include the following fields:
//
// - `resource`
// - `event`
// - `severity`
// - `service`
// - `group`
// - `value`
// - `text`
// - `tags`
// - `attributes`
// - `type`
// - `timestamp`
//
// For more information, see `alerta.alert()` parameters.
//
// ## Examples
// ### Send critical alerts to Alerta
// ```no_run
// import "contrib/bonitoo-io/alerta"
// import "influxdata/influxdb/secrets"
//
// apiKey = secrets.get(key: "ALERTA_API_KEY")
// endpoint = alerta.endpoint(
//     url: "https://alerta.io:8080/alert",
//     apiKey: apiKey,
//     environment: "Production",
//     origin: "InfluxDB",
// )
//
// crit_events = from(bucket: "example-bucket")
//     |> range(start: -1m)
//     |> filter(fn: (r) => r._measurement == "statuses" and status == "crit")
//
// crit_events
//     |> endpoint(
//         mapFn: (r) => {
//             return {r with
//                 resource: "example-resource",
//                 event: "example-event",
//                 severity: "critical",
//                 service: r.service,
//                 group: "example-group",
//                 value: r.status,
//                 text: "Status is critical.",
//                 tags: ["ex1", "ex2"],
//                 attributes: {},
//                 type: "exampleAlertType",
//                 timestamp: now(),
//             }
//         },
//     )()
// ```
//
// ## Metadata
// tags: notification endpoints,transformations
//
endpoint = (url, apiKey, environment="", origin="") =>
    (mapFn) =>
        (tables=<-) =>
            tables
                |> map(
                    fn: (r) => {
                        obj = mapFn(r: r)

                        return {r with _sent:
                                string(
                                    v:
                                        2 == alert(
                                                url: url,
                                                apiKey: apiKey,
                                                resource: obj.resource,
                                                event: obj.event,
                                                environment: environment,
                                                severity: obj.severity,
                                                service: obj.service,
                                                group: obj.group,
                                                value: obj.value,
                                                text: obj.text,
                                                tags: obj.tags,
                                                attributes: obj.attributes,
                                                origin: origin,
                                                type: obj.type,
                                                timestamp: obj.timestamp,
                                            ) / 100,
                                ),
                        }
                    },
                )
