package sensu

import "http"
import "json"

// toSensuName translates a string value to a Sensu name.
// Characters not being [a-zA-Z0-9_.\-] are replaced by underscore.
builtin toSensuName : (v: string) => string

// `event` sends a single event to Sensu as described in https://docs.sensu.io/sensu-go/latest/api/events/#create-a-new-event API. 
// `url` - string - base URL of [Sensu API](https://docs.sensu.io/sensu-go/latest/migrate/#architecture) without a trailing slash, for example "http://localhost:8080" .
// `apiKey` - string - Sensu [API Key](https://docs.sensu.io/sensu-go/latest/operations/control-access/).
// `checkName` - string - Check name, it can contain [a-zA-Z0-9_.\-] characters, other characters are replaced by underscore.
// `text` - string - The event text (named output in a Sensu Event).
// `handlers` - array<string> - Sensu handlers to execute, optional.
// `status` - int - The event status, 0 (default) indicates "OK", 1 indicates "WARNING", 2 indicates "CRITICAL", any other value indicates an “UNKNOWN” or custom status.
// `state` - string - The event state can be "failing", "passing" or "flapping". Defaults to "passing" for 0 status, "failing" otherwise. 
// `namespace` - string - The Sensu namespace. Defaults to "default".
// `entityName` - string - Source of the event, it can contain [a-zA-Z0-9_.\-] characters, other characters are replaced by underscore. Defaults to "influxdb".
event = (url, apiKey, checkName, text, handlers = [], status=0, state="", namespace="default", entityName="influxdb") => {
    data = {
        entity: {
            entity_class: "proxy",
            metadata: {
                name: toSensuName(v:entityName),
            }
        },
        check: {
            output: text,
            state: if state != "" then state else if status == 0 then "passing" else "failing",
            status: status,
            handlers: handlers,
            interval: 60, // required
            metadata: {
                name: toSensuName(v:checkName)
            }
        }
    }

    headers = {
        "Content-Type": "application/json; charset=utf-8",
        "Authorization": "Key " + apiKey,
    }
    enc = json.encode(v:data)
    return http.post(headers: headers, url: url + "/api/core/v2/namespaces/" + namespace + "/events", data: enc)
}

// `endpoint` creates a factory function that creates a target function for pipeline `|>` to send event to Sensu for each table row.
// `url` - string - base URL of [Sensu API](https://docs.sensu.io/sensu-go/latest/migrate/#architecture) without a trailing slash, for example "http://localhost:8080" .
// `apiKey` - string - Sensu [API Key](https://docs.sensu.io/sensu-go/latest/operations/control-access/).
// `handlers` - array<string> - Sensu handlers to execute.
// `namespace` - string - The Sensu namespace. Defaults to "default".
// `entityName` - string - Source of the event, it can contain [a-zA-Z0-9_.\-] characters, other characters are replaced by underscore. Defaults to "influxdb".
// The returned factory function accepts a `mapFn` parameter.
// The `mapFn` must return an object with `checkName`, `text`, and `status`, as defined in the `event` function arguments.
endpoint = (url, apiKey, handlers = [], namespace="default", entityName="influxdb") =>
    (mapFn) =>
        (tables=<-) => tables
            |> map(fn: (r) => {
                obj = mapFn(r: r)
                return {r with _sent: string(v: 2 == event(
                    url: url,
                    apiKey: apiKey,
                    checkName: obj.checkName,
                    text: obj.text,
                    handlers: handlers,
                    status: obj.status,
                    namespace: namespace,
                    entityName: entityName,
                ) / 100)}
            })
