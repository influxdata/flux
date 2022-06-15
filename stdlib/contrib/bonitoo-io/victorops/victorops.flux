// Package victorops provides functions that send events to [VictorOps](https://victorops.com/).
//
// **Note**: VictorOps is now Splunk On-Call
//
//
// ## Set up VictorOps
// To send events to VictorOps with Flux:
//
// 1. [Enable the VictorOps REST Endpoint Integration](https://help.victorops.com/knowledge-base/rest-endpoint-integration-guide/).
// 2. [Create a REST integration routing key](https://help.victorops.com/knowledge-base/routing-keys/).
// 3. [Create a VictorOps API key](https://help.victorops.com/knowledge-base/api/).
//
// ## Metadata
// introduced: 0.108.0
package victorops


import "http"
import "json"

// alert sends an alert to VictorOps.
//
// ## Parameters
//
// - url: VictorOps REST endpoint integration URL.
//
//    Example: `https://alert.victorops.com/integrations/generic/00000000/alert/<api_key>/<routing_key>`
//    Replace `<api_key>` and `<routing_key>` with valid VictorOps API and routing keys.
//
// - monitoringTool: Monitoring agent name. Default is `""`.
// - messageType: VictorOps message type (alert behavior).
//
//   **Valid values**:
//   - `CRITICAL`
//   - `WARNING`
//   - `INFO`
// - entityID: Incident ID. Default is `""`.
// - entityDisplayName: Incident display name or summary. Default is `""`.
// - stateMessage: Verbose incident message. Default is `""`.
// - timestamp: Incident start time. Default is `now()`.
//
// ## Examples
// ### Send the last reported value and incident type to VictorOps
//
// ```no_run
// import "contrib/bonitoo-io/victorops"
// import "influxdata/influxdb/secrets"
//
// apiKey = secrets.get(key: "VICTOROPS_API_KEY")
// routingKey = secrets.get(key: "VICTOROPS_ROUTING_KEY")
//
// lastReported = from(bucket: "example-bucket")
//     |> range(start: -1m)
//     |> filter(fn: (r) => r._measurement == "cpu" and r._field == "usage_idle")
//     |> last()
//     |> findRecord(fn: (key) => true, idx: 0)
//
// victorops.alert(
//     url: "https://alert.victorops.com/integrations/generic/00000000/alert/${apiKey}/${routingKey}",
//     messageType: if lastReported._value < 1.0 then
//         "CRITICAL"
//     else if lastReported._value < 5.0 then
//         "WARNING"
//     else
//         "INFO",
//     entityID: "example-alert-1",
//     entityDisplayName: "Example Alert 1",
//     stateMessage: "Last reported cpu_idle was ${string(v: r._value)}.",
// )
// ```
//
// ## Metadata
// tags: single notification
//
alert = (
    url,
    messageType,
    entityID="",
    entityDisplayName="",
    stateMessage="",
    timestamp=now(),
    monitoringTool="InfluxDB",
) =>
{
    alert = {
        message_type: messageType,
        entity_id: entityID,
        entity_display_name: entityDisplayName,
        state_message: stateMessage,
        // required in seconds
        state_start_time: uint(v: timestamp) / uint(v: 1000000000),
        monitoring_tool: monitoringTool,
    }
    headers = {"Content-Type": "application/json"}
    body = json.encode(v: alert)

    return http.post(headers: headers, url: url, data: body)
}

// endpoint sends events to VictorOps using data from input rows.
//
// ## Parameters
// - url: VictorOps REST endpoint integration URL.
//
//   Example: `https://alert.victorops.com/integrations/generic/00000000/alert/<api_key>/<routing_key>`
//   Replace `<api_key>` and `<routing_key>` with valid VictorOps API and routing keys.
//
// - monitoringTool: Tool to use for monitoring.
//   Default is `InfluxDB`.
//
// ## Usage
// `victorops.endpoint` is a factory function that outputs another function.
// The output function requires a `mapFn` parameter.
//
// ### mapFn
// A function that builds the object used to generate the POST request. Requires an `r` parameter.
//
// `mapFn` accepts a table row (`r`) and returns an object that must include the following fields:
//
// - monitoringTool
// - messageType
// - entityID
// - entityDisplayName
// - stateMessage
// - timestamp
//
// For more information, see `victorops.event()` parameters.
//
// ## Examples
// ### Send critical events to VictorOps
//
// ```no_run
// import "contrib/bonitoo-io/victorops"
// import "influxdata/influxdb/secrets"
//
// apiKey = secrets.get(key: "VICTOROPS_API_KEY")
// routingKey = secrets.get(key: "VICTOROPS_ROUTING_KEY")
// url = "https://alert.victorops.com/integrations/generic/00000000/alert/${apiKey}/${routingKey}"
// endpoint = victorops.endpoint(url: url)
//
// crit_events = from(bucket: "example-bucket")
//     |> range(start: -1m)
//     |> filter(fn: (r) => r._measurement == "statuses" and status == "crit")
//
// crit_events
//     |> endpoint(
//         mapFn: (r) => ({
//             monitoringTool: "InfluxDB",
//             messageType: "CRITICAL",
//             entityID: "${r.host}-${r._field}-critical",
//             entityDisplayName: "Critical alert for ${r.host}",
//             stateMessage: "${r.host} is in a critical state. ${r._field} is ${string(v: r._value)}.",
//             timestamp: now(),
//         }),
//     )()
// ```
//
// ## Metadata
// tags: notification endpoints,transformations
//
endpoint = (url, monitoringTool="InfluxDB") =>
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
                                                messageType: obj.messageType,
                                                entityID: obj.entityID,
                                                entityDisplayName: obj.entityDisplayName,
                                                stateMessage: obj.stateMessage,
                                                timestamp: obj.timestamp,
                                                monitoringTool: monitoringTool,
                                            ) / 100,
                                ),
                        }
                    },
                )
