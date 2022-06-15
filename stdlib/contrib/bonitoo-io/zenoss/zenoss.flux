// Package zenoss provides functions that send events to [Zenoss](https://www.zenoss.com/).
//
// ## Metadata
// introduced: 0.108.0
package zenoss


import "http"
import "json"

// event sends an event to [Zenoss](https://www.zenoss.com/).
//
// ## Parameters
// - url: Zenoss [router endpoint URL](https://help.zenoss.com/zsd/RM/configuring-resource-manager/enabling-access-to-browser-interfaces/creating-and-changing-public-endpoints).
// - username: Zenoss username to use for HTTP BASIC authentication.
//   Default is `""` (no authentication).
// - password: Zenoss password to use for HTTP BASIC authentication.
//   Default is `""` (no authentication).
// - action: Zenoss router name.
//   Default is "EventsRouter".
// - method: [EventsRouter method](https://help.zenoss.com/dev/collection-zone-and-resource-manager-apis/codebase/routers/router-reference/eventsrouter).
//   Default is "add_event".
// - type: Event type.
//   Default is "rpc".
// - tid: Temporary request transaction ID.
//   Default is `1`.
// - summary: Event summary.
//   Default is `""`.
// - device: Related device.
//   Default is `""`.
// - component: Related component.
//   Default is `""`.
// - severity: [Event severity level](https://help.zenoss.com/zsd/RM/administering-resource-manager/event-management/event-severity-levels).
//
//   **Supported values**:
//   - Critical
//   - Warning
//   - Info
//   - Clear
//
// - eventClass: [Event class](https://help.zenoss.com/zsd/RM/administering-resource-manager/event-management/understanding-event-classes).
//   Default is `""`.
// - eventClassKey: Event [class key](https://help.zenoss.com/zsd/RM/administering-resource-manager/event-management/event-fields).
//   Default is `""`.
// - collector: Zenoss [collector](https://help.zenoss.com/zsd/RM/administering-resource-manager/event-management/event-fields).
//   Default is `""`.
// - message: Related message.
//   Default is `""`.
//
// ## Examples
// ### Send the last reported value and severity to Zenoss
//
// ```no_run
// import "contrib/bonitoo-io/zenoss"
// import "influxdata/influxdb/secrets"
//
// username = secrets.get(key: "ZENOSS_USERNAME")
// password = secrets.get(key: "ZENOSS_PASSWORD")
//
// lastReported = from(bucket: "example-bucket")
//     |> range(start: -1m)
//     |> filter(fn: (r) => r._measurement == "cpu" and r._field == "usage_idle")
//     |> last()
//     |> findRecord(fn: (key) => true, idx: 0)
//
// zenoss.event(
//     url: "https://tenant.zenoss.io:8080/zport/dmd/evconsole_router",
//     username: username,
//     username: password,
//     device: lastReported.host,
//     component: "CPU",
//     eventClass: "/App",
//     severity: if lastReported._value < 1.0 then
//         "Critical"
//     else if lastReported._value < 5.0 then
//         "Warning"
//     else if lastReported._value < 20.0 then
//         "Info"
//     else
//         "Clear",
// )
// ```
//
// ## Metadata
// tags: single notification
event = (
        url,
        username,
        password,
        action="EventsRouter",
        method="add_event",
        type="rpc",
        tid=1,
        summary="",
        device="",
        component="",
        severity,
        eventClass="",
        eventClassKey="",
        collector="",
        message="",
    ) =>
    {
        event = {
            summary: summary,
            device: device,
            component: component,
            severity: severity,
            evclass: eventClass,
            evclasskey: eventClassKey,
            collector: collector,
            message: message,
        }
        payload = {
            action: action,
            method: method,
            data: [event],
            type: type,
            tid: tid,
        }
        headers = {"Authorization": http.basicAuth(u: username, p: password), "Content-Type": "application/json"}
        body = json.encode(v: payload)

        return http.post(headers: headers, url: url, data: body)
    }

// endpoint sends events to Zenoss using data from input rows.
//
// ## Parameters
//
// - url: Zenoss [router endpoint URL](https://help.zenoss.com/zsd/RM/configuring-resource-manager/enabling-access-to-browser-interfaces/creating-and-changing-public-endpoints).
// - username: Zenoss username to use for HTTP BASIC authentication.
//   Default is `""` (no authentication).
// - password: Zenoss password to use for HTTP BASIC authentication.
//   Default is `""` (no authentication).
// - action: Zenoss router name.
//   Default is `"EventsRouter"`.
// - method: EventsRouter method.
//   Default is `"add_event"`.
// - type: Event type. Default is `"rpc"`.
// - tid: Temporary request transaction ID.
//   Default is `1`.
//
// ## Usage
// `zenoss.endpoint` is a factory function that outputs another function.
// The output function requires a `mapFn` parameter.
//
// ### mapFn
// A function that builds the object used to generate the POST request. Requires an `r` parameter.
//
// `mapFn` accepts a table row (`r`) and returns an object that must include the following fields:
//
// - summary
// - device
// - component
// - severity
// - eventClass
// - eventClassKey
// - collector
// - message
//
// For more information, see zenoss.event() parameters.
//
// ## Examples
// ### Send critical events to Zenoss
//
// ```no_run
// import "contrib/bonitoo-io/zenoss"
// import "influxdata/influxdb/secrets"
//
// url = "https://tenant.zenoss.io:8080/zport/dmd/evconsole_router"
// username = secrets.get(key: "ZENOSS_USERNAME")
// password = secrets.get(key: "ZENOSS_PASSWORD")
// endpoint = zenoss.endpoint(
//     url: url,
//     username: username,
//     password: password,
// )
//
// crit_events = from(bucket: "example-bucket")
//     |> range(start: -1m)
//     |> filter(fn: (r) => r._measurement == "statuses" and status == "crit")
//
// crit_events
//     |> endpoint(
//         mapFn: (r) => ({
//             summary: "Critical event for ${r.host}",
//             device: r.deviceID,
//             component: r.host,
//             severity: "Critical",
//             eventClass: "/App",
//             eventClassKey: "",
//             collector: "",
//             message: "${r.host} is in a critical state.",
//         }),
//     )()
// ```
//
// ## Metadata
// tags: notification endpoints,transformations
//
endpoint = (
    url,
    username,
    password,
    action="EventsRouter",
    method="add_event",
    type="rpc",
    tid=1,
) =>
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
                                                action: action,
                                                method: method,
                                                type: type,
                                                tid: tid,
                                                summary: obj.summary,
                                                device: obj.device,
                                                component: obj.component,
                                                severity: obj.severity,
                                                eventClass: obj.eventClass,
                                                eventClassKey: obj.eventClassKey,
                                                collector: obj.collector,
                                                message: obj.message,
                                            ) / 100,
                                ),
                        }
                    },
                )
