package alerta


import "http"
import "json"
import "strings"

// alert sends an alert to Alerta.
// `url` - string - Alerta URL.
// `apiKey` - string - Alerta API key.
// `resource` - string - resource under alarm.
// `event` - string - event name.
// `environment` - string - environment. Valid values: "Production", "Development" or empty string (default).
// `severity` - string - event severity. See https://docs.alerta.io/en/latest/api/alert.html#alert-severities.
// `service` - arrays of string - list of affected services.
// `group` - string - event group.
// `value` - string - event value.
// `text` - string - text description.
// `type` - string - event type.
// `origin` - string - monitoring component.
// `timestamp` - time - time alert was generated.
// `timeout` - int - seconds before alert is considered stale.
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
) => {
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
    headers = {
        "Authorization": "Key " + apiKey,
        "Content-Type": "application/json",
    }
    body = json.encode(v: alert)

    return http.post(headers: headers, url: url, data: body)
}

// endpoint creates the endpoint for the Alerta.
// `url` - string - VictorOps REST endpoint URL. No default.
// `apiKey` - string - Alerta API key.
// `environment` - string - environment. Valid values: "Production", "Development" or empty string (default).
// `origin` - string - monitoring component.
// The returned factory function accepts a `mapFn` parameter.
// The `mapFn` must return an object with `resource`, `event`, `severity`, `service`, `group`, `value`, `text`,
// `tags`, `attributes`, `origin`, `type` and `timestamp` fields as defined in the `alert` function arguments.
endpoint = (url, apiKey, environment="", origin="") => (mapFn) => (tables=<-) => tables
    |> map(
        fn: (r) => {
            obj = mapFn(r: r)

            return {r with
                _sent: string(
                    v: 2 == alert(
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
