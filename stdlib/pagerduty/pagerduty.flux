package pagerduty

import "http"
import "json"
import "strings"

// `dedupKey` - adds a newline concatinated value of the sorted group key that is then sha256-hashed and hex-encoded to a column with the key `_pagerdutyDedupKey`.
builtin dedupKey

option defaultURL = "https://events.pagerduty.com/v2/enqueue"


// severity levels on status objects can be one of the following: ok,info,warn,crit,unknown
// but pagerduty only accepts critical, error, warning or info.
// formatSeverity turns a level from the status object into a pagerduty severity
formatSeverity = (severity) => {
    sev = strings.toLower(v:severity) 
    if sev == "warn" then "warning"
    else if sev == "crit" then "critical"
    else if sev == "info" then "info"
    else if sev == "ok" then "info"
    else "unknown"
}

// `actionFromSeverity` converts a severity to an action; "ok" becomes "resolve" everything else converts to "trigger".
actionFromSeverity = (severity)=> if strings.toLower(v:severity) == "ok" then "resolve" else "trigger"

// `sendEvent` sends an event to PagerDuty, the description of some of these parameters taken from the pagerduty documentation at https://v2.developer.pagerduty.com/docs/send-an-event-events-api-v2
// `pagerdutyURL` - sring - URL of the pagerduty endpoint.  Defaults to: `option defaultURL = "https://events.pagerduty.com/v2/enqueue"`
// `token` - string - auth for pagerduty.
// `routingKey` - string - routingKey.
// `client` - string - name of the client sending the alert.
// `clientURL` - string - url of the client sending the alert.
// `dedupkey` - string - a per alert ID. It acts as deduplication key, that allows you to ack or change the severity of previous messages. Supports a maximum of 255 characters.
// `class` - string - The class/type of the event, for example ping failure or cpu load.
// `group` - string - Logical grouping of components of a service, for example app-stack.
// `severity` - string - The perceived severity of the status the event is describing with respect to the affected system. This can be critical, error, warning or info.
// `component` - string - Component of the source machine that is responsible for the event, for example mysql or eth0.
// `source` - string - The unique location of the affected system, preferably a hostname or FQDN.
// `summary` - string - A brief text summary of the event, used to generate the summaries/titles of any associated alerts. The maximum permitted length of this property is 1024 characters.
// `timestamp` - string - The time at which the emitting tool detected or generated the event, in RFC 3339 nano format.
sendEvent = (pagerdutyURL=defaultURL,
    token="",
    routingKey,
    client,
    clientURL,
    dedupKey,
    class,
    group,
    severity,
    component,
    source,
    summary,
    timestamp) => {

    payload = {
            summary: summary,
            timestamp: timestamp,
            source: source,
            severity: formatSeverity(severity),
            component: component,
            group: group,
            class: class,
    }
    data = {
        payload: payload,
        routing_key: routingKey,
        dedup_key: dedupKey,
        event_action: actionFromSeverity(severity: severity),
        client: client,
        client_url: clientURL,
    }

    headers = {
        "Authorization": "Token token=" + token,
        "Accept": "application/vnd.pagerduty+json;version=2",
        "Content-Type": "application/json",
    }
    enc = json.encode(v: data)
    return http.post(headers: headers, url: pagerdutyURL, data: enc)
}

// `endpoint` creates the endpoint for the PagerDuty external service.
// `url` - string - URL of the slack endpoint. Defaults to: "https://events.pagerduty.com/v2/enqueue".
// `token` - string - token for the pagerduty endpoint.
// The returned factory function accepts a `mapFn` parameter.
// The `mapFn` parameter must be a function that returns an object with `routingKey`, `client`, `client_url`,`class`,`group`, `severity`, `component`, `source`, `summary`, and `timestamp` as defined in the sendEvent function.
// Note that while sendEvent accepts a dedup key, endpoint gets the dedupkey from the groupkey of the input table instead of it being handled by the `mapFn`.
endpoint = (url=defaultURL, token="") =>
    (mapFn) =>
        (tables=<-) => tables
            |> dedupKey()
            |> map(fn: (r) => {
                obj = mapFn(r: r)
                
                return {r with _sent: string(v: 2 == (sendEvent(pagerdutyURL: url,
                    token: token, 
                    routingKey: obj.routingKey, 
                    client: obj.client, 
                    clientURL: obj.clientURL, 
                    dedupKey: r._pagerdutyDedupKey,
                    class: obj.class,
                    group: obj.group,
                    severity: obj.severity, 
                    component: obj.component, 
                    source: obj.source, 
                    summary: obj.summary, 
                    timestamp: obj.timestamp,
                ) / 100))}
            })