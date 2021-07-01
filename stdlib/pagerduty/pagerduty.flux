// Package pagerduty provides functions for sending data to PagerDuty.
package pagerduty


import "http"
import "json"
import "strings"

// dedupKey uses the group key of an input table to generate and store a deduplication key in the _pagerdutyDedupKey column.
// The function sorts, newline-concatenates, SHA256-hashes, and hex-encodes the group key to create a unique deduplication key for each input table.
//
// ## Parameters
// - `exclude` is the group key columns to exclude when generating the deduplication key. Default is ["_start", "_stop", "_level"].
//
// ## Add a PagerDuty deduplication key to output data
// ```
// import "pagerduty"
//
// from(bucket: "default")
//   |> range(start: -5m)
//   |> filter(fn: (r) => r._measurement == "mem")
//   |> pagerduty.dedupKey()
// ```
//
builtin dedupKey : (<-tables: [A], ?exclude: [string]) => [{A with _pagerdutyDedupKey: string}]

option defaultURL = "https://events.pagerduty.com/v2/enqueue"

// severityFromLevel converts an InfluxDB status level to a PagerDuty severity.
//
//
//  Status level	PagerDuty severity
//  crit	        critical
//  warn	        warning
//  info	        info
//  ok	            info
//
// ## Parameters
// - `level` is the InfluxDB status level to convert to a PagerDuty severity.
//
severityFromLevel = (level) => {
    lvl = strings.toLower(v: level)
    sev = if lvl == "warn" then
        "warning"
    else if lvl == "crit" then
        "critical"
    else if lvl == "info" then
        "info"
    else if lvl == "ok" then
        "info"
    else
        "error"

    return sev
}

// actionFromSeverity converts a severity to a PagerDuty action. ok converts to resolve. All other severities convert to trigger.
//
// ## Parameters
// - `severity` is the severity to convert to a PagerDuty action.
//
actionFromSeverity = (severity) => if strings.toLower(v: severity) == "ok" then
    "resolve"
else
    "trigger"

// `actionFromLevel` converts a monitoring level to an action; "ok" becomes "resolve" everything else converts to "trigger".
actionFromLevel = (level) => if strings.toLower(v: level) == "ok" then "resolve" else "trigger"

// sendEvent sends an event to PagerDuty.
//
// ## Parameters
// - `pagerdutyURL` is the URL of the PagerDuty endpoint.
//
//      Defaults to https://events.page rduty.com/v2/enqueue.
//
// - `routingKey` is the routing key generated from your PagerDuty integration.
// - `client` is the name of the client sending the alert.
// - `clientURL` is the URL of the client sending the alert.
// - `dedupkey` is a per-alert ID that acts as deduplication key and allows you to acknowledge or change the severity of previous messages. Supports a maximum of 255 characters.
// - `class` is the class or type of the event.
//
//      Classes are user-defined.
//      For example, ping failure or cpu load.
//
// - `group` is a logical grouping used by PagerDuty.
//
//      Groups are user-defined.
//      For example, app-stack.
//
// - `severity` is the severity of the event.
//
//      Valid values include:
//
//        critical
//        error
//        warning
//        info
//
// - `eventAction` is the event type to send to PagerDuty.
//
//      Valid values include:
//
//        trigger
//        resolve
//        acknowledge
//
// - `source` is the unique location of the affected system. For example, the hostname or fully qualified domain name (FQDN).
// - `summary` is a brief text summary of the event used as the summaries or titles of associated alerts. The maximum permitted length is 1024 characters.
// - `timestamp` is the time the detected event occurred in RFC3339nano format.
//
sendEvent = (
        pagerdutyURL=defaultURL,
        routingKey,
        client,
        clientURL,
        dedupKey,
        class,
        group,
        severity,
        eventAction,
        source,
        summary,
        timestamp,
) => {
    payload = {
        summary: summary,
        timestamp: timestamp,
        source: source,
        severity: severity,
        group: group,
        class: class,
    }
    data = {
        payload: payload,
        routing_key: routingKey,
        dedup_key: dedupKey,
        event_action: eventAction,
        client: client,
        client_url: clientURL,
    }
    headers = {
        "Accept": "application/vnd.pagerduty+json;version=2",
        "Content-Type": "application/json",
    }
    enc = json.encode(v: data)

    return http.post(headers: headers, url: pagerdutyURL, data: enc)
}

// endpoint returns a function that can be used to send a message to PagerDuty that includes output data.
//
// ## Parameters
// - `url` is the The PagerDuty v2 Events API URL.
//
//      Defaults to https://events.pagerduty.com/v2/enqueue.
//
// - `Usage` the pagerduty.endpoint is a factory function that outputs another function.
//
//      The output function requires a mapFn parameter.
//      See the PagerDuty v2 Events API documentation for more information about these parameters.
//
// - `mapFn` is a function that builds the record used to generate the POST request. Requires an r parameter.
//
//      mapFn accepts a table row (r) and returns a record that must include the following fields:
//         routingKey
//         client
//         client_url
//         class
//         eventAction
//         group
//         severity
//         component
//         source
//         summary
//         timestamp
//
// ## Send critical statuses to a PagerDuty endpoint
// ```
// import "pagerduty"
// import "influxdata/influxdb/secrets"
//
// routingKey = secrets.get(key: "PAGERDUTY_ROUTING_KEY")
// toPagerDuty = pagerduty.endpoint()
//
// crit_statuses = from(bucket: "example-bucket")
//   |> range(start: -1m)
//   |> filter(fn: (r) => r._measurement == "statuses" and r.status == "crit")
//
// crit_statuses
//   |> toPagerDuty(mapFn: (r) => ({ r with
//       routingKey: routingKey,
//       client: r.client,
//       clientURL: r.clientURL,
//       class: r.class,
//       eventAction: r.eventAction,
//       group: r.group,
//       severity: r.severity,
//       component: r.component,
//       source: r.source,
//       summary: r.summary,
//       timestamp: r._time,
//     })
//   )()
// ```
//
endpoint = (url=defaultURL) => (mapFn) => (tables=<-) => tables
    |> dedupKey()
    |> map(
        fn: (r) => {
            obj = mapFn(r: r)

            return {r with
                _sent: string(
                    v: 2 == sendEvent(
                        pagerdutyURL: url,
                        routingKey: obj.routingKey,
                        client: obj.client,
                        clientURL: obj.clientURL,
                        dedupKey: r._pagerdutyDedupKey,
                        class: obj.class,
                        group: obj.group,
                        severity: obj.severity,
                        eventAction: obj.eventAction,
                        source: obj.source,
                        summary: obj.summary,
                        timestamp: obj.timestamp,
                    ) / 100,
                ),
            }
        },
    )
