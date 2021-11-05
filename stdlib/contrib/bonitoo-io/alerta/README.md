# Alerta Package

The Flux `alerta` package provides functions that send alerts to [Alerta](https://alerta.io/).

## alerta.alert

The alerta.alert() function sends an alert to [Alerta](https://alerta.io/).

### Parameters

| Name | Type | Description |
| ---- | ---- | ----------- |
| url | string | Alerta URL. Usually `https://alert.victorops.com/integrations/generic/20131114/alert/$api_key/$routing_key`, use valid `api_key` and `routing_key`. |
| apiKey | string | Alerta API key. |
| resource | string | Resource associated with the alert. |
| event | string | Event name. |
| environment | string | Incident Alert environment. Default: `""`. Valid values are: `""`, `"Production"`, `"Development"`. |
| severity | string | Event severity. See [Alerta severities](https://docs.alerta.io/en/latest/api/alert.html#alert-severities). |
| service | array of strings | List of affected services. Default is `[]`. |
| group | string | Alerta event group. Default is `""`. |
| value | string | Event value. Default is `""`. |
| text | string | Alert text description. Default is `""`. |
| tags | array of strings | List of event tags. Default is `[]`. |
| attributes | record | Alert attributes (optional). |
| origin | string | Alert origin. Default is "InfluxDB". |
| type | string | Event type. Default is "". |
| timestamp | time | The time alert was generated. Default is `now()`. |

Example:

    import "contrib/bonitoo-io/alerta"
    import "influxdata/influxdb/secrets"
    
    apiKey = secrets.get(key: "ALERTA_API_KEY")
    
    lastReported =
      from(bucket: "example-bucket")
        |> range(start: -1m)
        |> filter(fn: (r) =>
          r._measurement == "example-measurement" and
          r._field == "level"
        )
        |> last()
        |> findRecord(fn: (key) => true, idx: 0)
    
    severity = if lastReported._value > 50 then "warning" else "ok"
    
    alerta.alert(
      url: "https://alerta.io:8080/alert",
      apiKey: apiKey,
      resource: "example-resource",
      event: "Example event",
      environment: "Production",
      severity: severity,
      service: ["example-service"],
      group: "example-group",
      value: string(v: lastReported._value),
      text: "Service is ${severity}. The last reported value was ${string(v: lastReported._value)}.",
      tags: ["ex1", "ex2"],
      type: "exampleAlertType",
      timestamp: now(),
    )

## alerta.endpoint

The `alerta.endpoint()` function sends alerts to [Alerta](https://alerta.io/) using data from input rows.

### Parameters

| Name | Type | Description |
| ---- | ---- | ----------- |
| url | string | Alerta URL. Usually `https://alert.victorops.com/integrations/generic/20131114/alert/$api_key/$routing_key`, use valid `api_key` and `routing_key`. |
| apiKey | string | Alerta API key. |
| environment | string | Incident Alert environment. Default: `""`. Valid values are: `""`, `"Production"`, `"Development"`. |
| origin | string | Alert origin. Default is "InfluxDB". |

The returned factory function accepts a `mapFn` parameter.
The `mapFn` accepts a row and returns a record with the following fields:

- resource
- event
- severity
- service
- group
- value
- text
- tags
- attributes
- type
- timestamp

_For more information, see `alerta.alert()` parameters._

Example:

    import "contrib/bonitoo-io/alerta"
    import "influxdata/influxdb/secrets"
    
    apiKey = secrets.get(key: "ALERTA_API_KEY")
    endpoint = alerta.endpoint(
      url: "https://alerta.io:8080/alert",
      apiKey: apiKey,
      environment: "Production",
      origin: "InfluxDB"
    )
    
    crit_events = from(bucket: "example-bucket")
      |> range(start: -1m)
      |> filter(fn: (r) => r._measurement == "statuses" and status == "crit")
    
    crit_events
      |> endpoint(mapFn: (r) => {
        return { r with
          resource: "example-resource",
          event: "example-event",
          severity: "critical",
          service: r.service,
          group: "example-group",
          value: r.status,
          text: "Status is critical.",
          tags: ["ex1", "ex2"],
          type: "exampleAlertType",
          timestamp: now(),
        }
      })()

## Contact

- Author: Ales Pour / Bonitoo
- Email: alespour@bonitoo.io
- Github: [@alespour](https://github.com/alespour), [@bonitoo-io](https://github.com/bonitoo-io)
- Influx Slack: [@Ales Pour](https://influxdata.com/slack)
