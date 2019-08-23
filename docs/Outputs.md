# Flux Outputs

This documentation contains the design of how Flux will produce outputs to external services.
This doc has the aim of defining the tiny bits of Go functions that must be implemented to allow a pure Flux implementation for external service communication.
 
## Endpoints, Encoders, and Transports

There are three key components for an external service:
 - an _endpoint_: it is the component that takes a table and pushes it to the external service.
 - an _encoder_: it is the component that translates tables into something understandable by the external service.
 - the _transport_: it is the communication protocol used to communicate with the external service.

In a nutshell, endpoints use encoders to encode data and a transport to push data to external services.

Encoders and transports are written in Go as `values.Function`s.
For example there will be `json.encode`, as well as `csv.encode` for encoders, and `http.post` and `tcp.send` for transports.

The minimal set that must be implemented is `http.post` and `json.encode`.

Endpoints are implemented in pure Flux.

## Flux

Encoders are provided in Flux as `encode` functions.
For example the JSON encoder is part of the `json` package.
Encoders take an object `data` and return its encoding as a string:

```flux
encode: (data:object) -> bytes
```

Communication protocols are also implemented (in Go), so that one can, for example, send an HTTP request with JSON data in pure Flux:

```flux
import "http"
import "json"

data = {a: 1, b: 2}
json_enc = json.encode(data)
http.post(header: {"Content Type": "application-json"}, url: "http://some/url/", data: json_enc)
```

External services in Flux provide an `endpoint` function for building an endpoint.
The `endpoint` function returns a factory function for building a specialized transformation for sending tables to the external service.
The factory function is used to create multiple ways of sending data to the service, as such it accepts a `mapFn` to map each record in a table to a proper object.
The fields in that object are used by the endpoint to build the message for the external service.
Every external service must define how that object must be shaped for the communication to happen.

For example:

```flux
import ext "external_service"

someFilter = (r) => ...
someOtherFilter = (r) => ...

// Obtain the factory function with this configuration for the endpoint.
endpoint = ext.endpoint(url: "https://url/to/external/service")

someMapFn = (r) => {
   // Generate payload from record.
   return {title: r.someTag, body: string(v: r._value)}
})
someOtherMapFn = (r) => {
   // Generate payload from record.
   return {title: r.someOtherTag, body: string(v: r._value)}
})

// Invoke the factory to get custom transformations.
to1 = endpoint(mapFn: someMapFn)
to2 = endpoint(mapFn: someOtherMapFn)

from(...)
    |> range(...)
    |> filter(fn: someFilter)
    |> to1()

from(...)
    |> range(...)
    |> filter(fn: someOtherFilter)
    |> to2()
```

The output of `to` functions contains the result of the communication with the external service and varies on a per-service base.


### An Example: Slack

We provide below an example implementation for a Slack endpoint.
The package provides an `endpoint` function that return the Slack endpoint.
It also provides helpers to send a single message in a channel and build the request header.

```flux
package slack

import "http"
import "json"

// This can be globally overridden in another package with:
// slack.defaultURL = "https://slack.com/api/chat.anotherURL"
option defaultURL = "https://slack.com/api/chat.postMessage"

// `message` sends a single message to a Slack channel.
message = (url=defaultURL, token, channel, text) => {
    data = {
        channel: channel,
        text: text
    }
    header = {
        "Authorization": token,
        "Content Type": application/json"
    }
    enc = json.encode(data)
    return http.post(header: header, url: url, data: enc)
}

// `endpoint` creates the endpoint for the Slack external service.
// The returned factory function accepts a `mapFn` parameter.
// The `mapFn` must return an object with `channel` and `text` fields.
endpoint = (url=defaultURL, token) =>
    // The factory function.
    (mapFn) =>
        (tables=<-) => tables
            |> map(fn: (r) => {
                obj = mapFn(r)
                resp = message(url: url, token: token, channel: obj.channel, text: obj.text)
                return {r with status: resp.status}
            })
```

One could use the package to send a single message to Slack:

```flux
import "slack"
import "secret" // imagine we have a package for secrets

tok = secret.get("SLACK_TOKEN")
text = "hello @query_owners!"
slack.message(token: tok, channel: "#flux", text: text)
```

Or to send data processed by a query:

```flux
import "slack"
import "secret" // imagine we have a package for secrets

token = secret.get("SLACK_TOKEN")
ep = slack.endpoint(token: token)

from(bucket: "telegraf")
    |> range(start: -5m)
    |> filter(fn: (r) => r._measurement == "cpu" and r._field == "usage_system")
    |> ep(mapFn: (r) => ({
        channel: "#flux", // this is static, it could have been dynamically based on `r`
        text: r._measurement + "@" + r.host + ": " + string(v: r._value)
    }))() // need to call the returned transformation
```

### Notifications

The package `influxdata/influxdb/monitor` provides a `notify` function that accepts the endpoint `name` and an `endpoint` transformation for pushing data to the external service.
The `notify` function checks that its input has at least `_time` and `_level` as columns.
It adds a `_endpoint:string` column with the endpoint `name` that will be part of the group key.
It filters the `unknown` level (or non-positive level values). It writes the results of the operations to a bucket.

Even if it cannot be implemented in pure Flux, we provide a pseudo-implementation:

```flux
notify = (tables=<-, name, endpoint) => tables
    |> filter(fn: (r) => exists r._time and exists r._level and r._level >= 0)
    |> endpoint()
    |> set(key: "_endpoint", value: ep.name)
    |> group(columns: ...) // cannot extend group key
    |> to(bucket: "notifications")
```

Example script using multiple notification services:

```flux
import "influxdata/influxdb/monitor"

import "slack"
import "file"
import pgr "pagerduty"

slack_ep = slack.endpoint(token)
file_ep = file.endpoint(path: "/var/log/notifications")
pgr_ep = pgr.endpoint(routingKey: "https://route/to/pages")

checks = from(bucket: "telegraf")
    |> range(start: -5m)
    |> filter(fn: (r) => r._measurement == "cpu" and r._field == "usage_system")
    |> monitor.check(name: "cpu_usage",
    	crit: (r) => r._value > 90,
        warn: (r) => r._value > 80,
    )

checks |> monitor.notify(name: "slack", endpoint: slack_ep(mapFn: (r) => ({channel: ..., text: ...})))
checks |> monitor.notify(name: "fileLog", endpoint: file_ep(mapFn: (r) => ...))
checks |> monitor.notify(name: "pagerDuty", endpoint: pgr_ep(mapFn: (r) => ...))
```
