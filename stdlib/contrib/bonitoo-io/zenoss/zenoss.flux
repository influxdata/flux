package zenoss


import "http"
import "json"

// event sends event to Zenoss.
// `url` - string - events web service URL.
// `username` - string - username for HTTP BASIC authentication. Default is empty string (no auth).
// `password` - string - password for HTTP BASIC authentication. Default is empty string (no auth).
// `action` - string - routername. Default is "EventsRouter".
// `method` - string - router name. Default is "add_event".
// `type` - string - event type. Default is "rpc".
// `tid` - int - temporary transaction ID. Default is 1.
// `summary` - string - event summary. Default is empty string.
// `device` - string - related device. Default is empty string.
// `component` - string - related component.
// `severity` - string - severity of the event.
// `eventClass` - string - event class. Default is empty string.
// `eventClassKey` - string - event class key. Default is empty string.
// `collector` - string - collector. Default is empty string.
// `message` - string - event message. Default is empty string.
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
) => {
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
        data: [
            event,
        ],
        type: type,
        tid: tid,
    }
    headers = {
        "Authorization": http.basicAuth(u: username, p: password),
        "Content-Type": "application/json",
    }
    body = json.encode(v: payload)

    return http.post(headers: headers, url: url, data: body)
}

// endpoint return method for sending events to Zenoss.
// Parameters:
// `url` - string - events web service URL.
// `username` - string - username for HTTP BASIC authentication. Default is empty string (no auth).
// `password` - string - password for HTTP BASIC authentication. Default is empty string (no auth).
// `action` - string - routername. Default is "EventsRouter".
// `method` - string - router name. Default is "add_event".
// `type` - string - event type. Default is "rpc".
// `tid` - int - temporary transaction ID. Default is 1.
// The returned factory function accepts a `mapFn` parameter.
// The `mapFn` must return record with `summary`, `device`, `component`, `severity`, `eventClass`, `eventClassKey`, `collector` and `message` fields as defined in the `event` function arguments.
endpoint = (
        url,
        username,
        password,
        action="EventsRouter",
        method="add_event",
        type="rpc",
        tid=1,
) => (mapFn) => (tables=<-) => tables
    |> map(
        fn: (r) => {
            obj = mapFn(r: r)

            return {r with
                _sent: string(
                    v: 2 == event(
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
