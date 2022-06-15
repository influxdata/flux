// Package opsgenie provides functions that send alerts to
// [Atlassian Opsgenie](https://www.atlassian.com/software/opsgenie)
// using the [Opsgenie v2 API](https://docs.opsgenie.com/docs/alert-api#create-alert).
//
// ## Metadata
// introduced: 0.84.0
package opsgenie


import "http"
import "json"
import "strings"

// respondersToJSON converts an array of Opsgenie responder strings
// to a string-encoded JSON array that can be embedded in an alert message.
//
// ## Parameters
// - v: (Required) Array of Opsgenie responder strings.
//   Responder strings must begin with
//   `user: `, `team: `, `escalation: `, or `schedule: `.
builtin respondersToJSON : (v: [string]) => string

// sendAlert sends an alert message to Opsgenie.
//
// ## Parameters
//
// - url: Opsgenie API URL. Defaults to `https://api.opsgenie.com/v2/alerts`.
// - apiKey: (Required) Opsgenie API authorization key.
// - message: (Required) Alert message text.
//   130 characters or less.
// - alias: Opsgenie alias usee to de-deduplicate alerts.
//   250 characters or less.
//   Defaults to [message](https://docs.influxdata.com/flux/v0.x/stdlib/contrib/sranka/opsgenie/sendalert/#message).
// - description: Alert description. 15000 characters or less.
// - priority: Opsgenie alert priority.
//
//   Valid values include:
//   - `P1`
//   - `P2`
//   - `P3` (default)
//   - `P4`
//   - `P5`
// - responders: List of responder teams or users.
//   Use the `user: ` prefix for users and `teams: ` prefix for teams.
// - tags: Alert tags.
// - entity: Alert entity used to specify the alert domain.
// - actions: List of actions available for the alert.
// - details: Additional alert details. Must be a JSON-encoded map of key-value string pairs.
// - visibleTo: List of teams and users the alert will be visible to without sending notifications.
//   Use the `user: ` prefix for users and `teams: ` prefix for teams.
//
// ## Examples
// ### Send the last reported status to a Opsgenie
// ```no_run
// import "influxdata/influxdb/secrets"
// import "contrib/sranka/opsgenie"
//
// apiKey = secrets.get(key: "OPSGENIE_APIKEY")
//
// lastReported =
//   from(bucket: "example-bucket")
//     |> range(start: -1m)
//     |> filter(fn: (r) => r._measurement == "statuses")
//     |> last()
//     |> findRecord(fn: (key) => true, idx: 0)
//     opsgenie.sendAlert(
//       apiKey: apiKey,
//       message: "Disk usage is: ${lastReported.status}.",
//       alias: "example-disk-usage",
//       responders: ["user:john@example.com", "team:itcrowd"]
//     )
// ```
//
// ## Metadata
// tags: single notification
sendAlert = (
    url="https://api.opsgenie.com/v2/alerts",
    apiKey,
    message,
    alias="",
    description="",
    priority="P3",
    responders=[],
    tags=[],
    entity="",
    actions=[],
    visibleTo=[],
    details="{}",
) =>
{
    headers = {"Content-Type": "application/json; charset=utf-8", "Authorization": "GenieKey " + apiKey}
    cutEncode = (v, max, defV="") => {
        v2 = if strings.strlen(v: v) != 0 then v else defV

        return
            if strings.strlen(v: v2) > max then
                string(v: json.encode(v: "${strings.substring(v: v2, start: 0, end: max)}"))
            else
                string(v: json.encode(v: v2))
    }
    body = "{
\"message\": ${cutEncode(v: message, max: 130)},
\"alias\": ${cutEncode(v: alias, max: 512, defV: message)},
\"description\": ${cutEncode(v: description, max: 15000)},
\"responders\": ${respondersToJSON(v: responders)},
\"visibleTo\": ${respondersToJSON(v: visibleTo)},
\"actions\": ${string(v: json.encode(v: actions))},
\"tags\": ${string(v: json.encode(v: tags))},
\"details\": ${details},
\"entity\": ${cutEncode(v: entity, max: 512)},
\"priority\": ${cutEncode(v: priority, max: 2)}
}"

    return http.post(headers: headers, url: url, data: bytes(v: body))
}

// endpoint sends an alert message to Opsgenie using data from table rows.
//
// ## Parameters
// - url: Opsgenie API URL. Defaults to `https://api.opsgenie.com/v2/alerts`.
// - apiKey: (Required) Opsgenie API authorization key.
// - entity: Alert entity used to specify the alert domain.
//
// ## Usage
// opsgenie.endpoint is a factory function that outputs another function.
// The output function requires a `mapFn` parameter.
//
// ### mapFn
// A function that builds the record used to generate the POST request. Requires an `r` parameter.
//
// `mapFn` accepts a table row (`r`) and returns a record that must include the following fields:
//
// - message
// - alias
// - description
// - priority
// - responders
// - tags
// - actions
// - details
// - visibleTo
//
// For more information, see `opsgenie.sendAlert`.
//
// ## Examples
// ### Send critical statuses to Opsgenie
// ```no_run
// import "influxdata/influxdb/secrets"
// import "contrib/sranka/opsgenie"
//
// apiKey = secrets.get(key: "OPSGENIE_APIKEY")
// endpoint = opsgenie.endpoint(apiKey: apiKey)
//
// crit_statuses = from(bucket: "example-bucket")
//   |> range(start: -1m)
//   |> filter(fn: (r) => r._measurement == "statuses" and status == "crit")
//
// crit_statuses
//   |> endpoint(mapFn: (r) => ({
//     message: "Great Scott!- Disk usage is: ${r.status}.",
//       alias: "disk-usage-${r.status}",
//       description: "",
//       priority: "P3",
//       responders: ["user:john@example.com", "team:itcrowd"],
//       tags: [],
//       entity: "my-lab",
//       actions: [],
//       details: "{}",
//       visibleTo: []
//     })
//   )()
// ```
//
// ## Metadata
// tags: notification endpoints, transformations
endpoint = (url="https://api.opsgenie.com/v2/alerts", apiKey, entity="") =>
    (mapFn) =>
        (tables=<-) =>
            tables
                |> map(
                    fn: (r) => {
                        obj = mapFn(r: r)

                        return {r with _sent:
                                string(
                                    v:
                                        2 == sendAlert(
                                                url: url,
                                                apiKey: apiKey,
                                                entity: entity,
                                                message: obj.message,
                                                alias: obj.alias,
                                                description: obj.description,
                                                priority: obj.priority,
                                                responders: obj.responders,
                                                tags: obj.tags,
                                                actions: obj.actions,
                                                visibleTo: obj.visibleTo,
                                                details: obj.details,
                                            ) / 100,
                                ),
                        }
                    },
                )
