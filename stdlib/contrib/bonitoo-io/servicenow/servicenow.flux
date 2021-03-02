package servicenow

import "http"
import "json"

// event sends event to ServiceNow.
// `url` - string - web service URL. No default.
// `username` - string - username for HTTP BASIC authentication.
// `password` - string - password for HTTP BASIC authentication.
// `source` - string - source that generated the event. Defaults to "Flux".
// `node` - string - node name, IP address etc associated with the event. Defaults to empty string.
// `metricType` - string - metric type to which the event is related. Defaults to empty string.
// `resource` - string - node resource to which the event is related. Defaults to empty string.
// `metricName` - string - metric name. Defaults to empty string.
// `messageKey` - string - unique identification of the event, eg. InfluxDB alert ID. Default is empty string (ServiceNow will fill the value).
// `description` - string - The text describing the event.
// `severity` - severity - Severity of the event. One of "critical", "major", "minor", "warning", "info" or "clear".
// `additionalInfo` - record - Additional information.
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
    additionalInfo
) => {
    encodedInfo = string(v: json.encode(v: additionalInfo))
    event = {
        source: source,
        node: node,
        type: metricType,
        resource: resource,
        metric_name: metricName,
        message_key: messageKey,
        description: description,
        severity:
            if severity == "critical" then "1"
            else if severity == "major" then "2"
            else if severity == "minor" then "3"
            else if severity == "warning" then "4"
            else if severity == "info" then "5"
            else if severity == "clear" then "0"
            else "", // shouldn't happen
        additional_info: if encodedInfo == "{}" then "" else encodedInfo
    }
    payload = {
        records: [
            event
        ]
    }
    headers = {
        "Authorization": http.basicAuth(u: username, p: password),
        "Content-Type": "application/json",
    }
    body = json.encode(v:payload)

    return http.post(headers: headers, url: url, data: body)
}

// endpoint creates the endpoint for the ServiceNow external service.
// `url` - string - ServiceNow web service URL. No default.
// `username` - string - username for HTTP BASIC authentication. Default is empty string (no authentication).
// `password` - string - password for HTTP BASIC authentication. Default is empty string (no authentication).
// `source` - string - source that generated the event. Defaults to "Flux".
// The returned factory function accepts a `mapFn` parameter.
// The `mapFn` must return an object with `node`, `metricType`, `resource`, `metricName`, `messageKey`, `description`,
// `severity` and `additionalInfo` fields as defined in the `event` function arguments.
endpoint = (url, username, password, source="Flux") =>
    (mapFn) =>
        (tables=<-) => tables
            |> map(fn: (r) => {
                obj = mapFn(r: r)
                return {r with _sent: string(v: 2 == event(
                    url: url,
                    username:    username,
                    password:    password,
                    source:      source,
                    node:        obj.node,
                    metricType:  obj.metricType,
                    resource:    obj.resource,
                    metricName:  obj.metricName,
                    messageKey:  obj.messageKey,
                    description: obj.description,
                    severity:    obj.severity,
                    additionalInfo: obj.additionalInfo
                ) / 100)}
            })
