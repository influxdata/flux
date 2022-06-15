// Package pagerduty provides functions for sending data to PagerDuty.
//
// ## Metadata
// introduced: 0.43.0
//
package pagerduty


import "experimental/record"
import "http"
import "json"
import "strings"

// dedupKey uses the group key of an input table to generate and store a
// deduplication key in the `_pagerdutyDedupKey`column.
// The function sorts, newline-concatenates, SHA256-hashes, and hex-encodes the
// group key to create a unique deduplication key for each input table.
//
// ## Parameters
// - exclude: Group key columns to exclude when generating the deduplication key.
//   Default is ["_start", "_stop", "_level"].
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Add a PagerDuty deduplication key to output data
// ```
// import "pagerduty"
// import "sampledata"
//
// < sampledata.int()
// >     |> pagerduty.dedupKey()
// ```
//
builtin dedupKey : (<-tables: stream[A], ?exclude: [string]) => stream[{A with _pagerdutyDedupKey: string}]

// defaultURL is the default PagerDuty URL used by functions in the `pagerduty` package.
option defaultURL = "https://events.pagerduty.com/v2/enqueue"

// severityFromLevel converts an InfluxDB status level to a PagerDuty severity.
//
// | Status level | PagerDuty severity |
// | :----------- | :----------------- |
// | crit         | critical           |
// | warn         | warning            |
// | info         | info               |
// | ok           | info               |
//
// ## Parameters
// - level: InfluxDB status level to convert to a PagerDuty severity.
//
// ## Examples
//
// ### Convert a status level to a PagerDuty serverity
// ```no_run
// import "pagerduty"
//
// pagerduty.severityFromLevel(level: "crit") // Returns critical
// ```
//
severityFromLevel = (level) => {
    lvl = strings.toLower(v: level)
    sev =
        if lvl == "warn" then
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

// actionFromSeverity converts a severity to a PagerDuty action.
//
// - `ok` converts to `resolve`.
// - All other severities convert to `trigger`.
//
// ## Parameters
// - severity: Severity to convert to a PagerDuty action.
//
// ## Examples
//
// ### Convert a severity to a PagerDuty action
// ```no_run
// import "pagerduty"
//
// pagerduty.actionFromSeverity(severity: "crit") // Returns trigger
// ```
//
actionFromSeverity = (severity) =>
    if strings.toLower(v: severity) == "ok" then
        "resolve"
    else
        "trigger"

// actionFromLevel converts a monitoring level to a PagerDuty action.
//
// - `ok` converts to `resolve`.
// - All other levels convert to `trigger`.
//
// ## Parameters
// - level: Monitoring level to convert to a PagerDuty action.
//
// ## Examples
//
// ### Convert a monitoring level to a PagerDuty action
// ```no_run
// import "pagerduty"
//
// pagerduty.actionFromLevel(level: "crit") // Returns trigger
// ```
//
actionFromLevel = (level) => if strings.toLower(v: level) == "ok" then "resolve" else "trigger"

// sendEvent sends an event to PagerDuty and returns the HTTP response code of the request.
//
// ## Parameters
// - pagerdutyURL: PagerDuty endpoint URL.
//
//      Default is https://events.pagerduty.com/v2/enqueue.
//
// - routingKey: Routing key generated from your PagerDuty integration.
// - client: Name of the client sending the alert.
// - clientURL: URL of the client sending the alert.
// - dedupKey: Per-alert ID that acts as deduplication key and allows you to
//   acknowledge or change the severity of previous messages.
//   Supports a maximum of 255 characters.
// - class: Class or type of the event.
//
//      Classes are user-defined.
//      For example, `ping failure` or `cpu load`.
//
// - group: Logical grouping used by PagerDuty.
//
//      Groups are user-defined.
//      For example, `app-stack`.
//
// - severity: Severity of the event.
//
//      Valid values:
//
//      - `critical`
//      - `error`
//      - `warning`
//      - `info`
//
// - eventAction: Event type to send to PagerDuty.
//
//      Valid values:
//
//      - `trigger`
//      - `resolve`
//      - `acknowledge`
//
// - source: Unique location of the affected system.
//   For example, the hostname or fully qualified domain name (FQDN).
// - component: Component responsible for the event.
// - summary: Brief text summary of the event used as the summaries or titles of associated alerts.
//   The maximum permitted length is 1024 characters.
// - timestamp: Time the detected event occurred in RFC3339nano format.
// - customDetails: Record with additional details about the event.
//
// ## Examples
//
// ### Send an event to PagerDuty
// ```no_run
// import "pagerduty"
// import "pagerduty"
//
// pagerduty.sendEvent(
//     routingKey: "example-routing-key",
//     client: "example-client",
//     clientURL: "http://example-url.com",
//     class: "example-class",
//     eventAction: "trigger",
//     group: "example-group",
//     severity: "crit",
//     component: "example-component",
//     source: "example-source",
//     component: "example-component",
//     summary: "example-summary",
//     timestamp: now(),
//     customDetails: {"example-key": "example value"},
// )
// ```
//
// ## Metadata
// tags: single notification
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
        component="",
        summary,
        timestamp,
        customDetails=record.any,
    ) =>
    {
        payload = {
            summary: summary,
            timestamp: timestamp,
            source: source,
            component: component,
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
        headers = {"Accept": "application/vnd.pagerduty+json;version=2", "Content-Type": "application/json"}
        enc =
            if customDetails == record.any then
                json.encode(v: data)
            else
                json.encode(v: {data with payload: {payload with custom_details: customDetails}})

        return http.post(headers: headers, url: pagerdutyURL, data: enc)
    }

// endpoint returns a function that sends a message to PagerDuty that includes output data.
//
// ## Parameters
// - url: PagerDuty v2 Events API URL.
//
//      Default is `https://events.pagerduty.com/v2/enqueue`.
//
// ## Usage
// `pagerduty.endpoint()` is a factory function that outputs another function.
//  The output function requires a `mapFn` parameter.
//
// ### mapFn
// Function that builds the record used to generate the POST request.
// Requires an `r` parameter.
//
// `mapFn` accepts a table row (`r`) and returns a record that must include the
// following properties:
//
// - routingKey
// - client
// - client_url
// - class
// - eventAction
// - group
// - severity
// - source
// - component
// - summary
// - timestamp
// - customDetails
//
// ## Examples
//
// ### Send critical statuses to a PagerDuty endpoint
// ```no_run
// import "pagerduty"
// import "influxdata/influxdb/secrets"
//
// routingKey = secrets.get(key: "PAGERDUTY_ROUTING_KEY")
// toPagerDuty = pagerduty.endpoint()
//
// crit_statuses = from(bucket: "example-bucket")
//     |> range(start: -1m)
//     |> filter(fn: (r) => r._measurement == "statuses" and r.status == "crit")
//
// crit_statuses
//     |> toPagerDuty(
//         mapFn: (r) => ({r with
//             routingKey: routingKey,
//             client: r.client,
//             clientURL: r.clientURL,
//             class: r.class,
//             eventAction: r.eventAction,
//             group: r.group,
//             severity: r.severity,
//             source: r.source,
//             component: r.component,
//             summary: r.summary,
//             timestamp: r._time,
//             customDetails: {"ping time": r.ping, load: r.load},
//         }),
//     )()
// ```
//
// ## Metadata
// tags: notification endpoints, transformations
//
endpoint = (url=defaultURL) =>
    (mapFn) =>
        (tables=<-) =>
            tables
                |> dedupKey()
                |> map(
                    fn: (r) => {
                        obj = mapFn(r: r)

                        status =
                            sendEvent(
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
                                component: record.get(r: obj, key: "component", default: ""),
                                summary: obj.summary,
                                timestamp: obj.timestamp,
                                customDetails: record.get(r: obj, key: "customDetails", default: record.any),
                            )

                        return {r with _sent: string(v: 2 == status / 100), _status: string(v: status)}
                    },
                )
