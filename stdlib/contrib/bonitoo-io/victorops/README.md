# VictorOps Package

Use this Flux package to send alerts to VictorOps (now Splunk On-Call).

Event fields are described in [REST Endpoint Integration Guide](https://help.victorops.com/knowledge-base/rest-endpoint-integration-guide/) documentation topic.

## victorops.event

`victorops.event` sends an alert to VictorOps. It has the following arguments:

| Name | Type | Description |
| ---- | ---- | ----------- |
| url | string | REST integration URL. Usually `https://alert.victorops.com/integrations/generic/20131114/alert/$api_key/$routing_key`, use valid `api_key` and `routing_key`. |
| monitoringTool | string | Monitoring agent name. Default: `""` |
| messageType | string | Alert behavior. Valid values: `"CRITICAL"`, `"WARNING"`, `"INFO"`. |
| entityID | string | Incident  ID. Default: `""`. |
| entityDisplayName | string | Incident summary. Default: `""`. |
| stateMessage | string | Verbose message. Default: `""`. |
| timestamp | string | Incident timestamp. Default: `now()` |

Example:

    import "contrib/bonitoo-io/victorops"
    import "influxdata/influxdb/secrets"
    import "strings"

    apiKey = secrets.get(key: "VICTOROPS_APIKEY")
    routingKey = secrets.get(key: "VICTOROPS_ROUTING")

    lastReported =
      from(bucket: "example-bucket")
        |> range(start: -1m)
        |> filter(fn: (r) => r._measurement == "cpu" and r._field == "usage_idle")
        |> last()
        |> tableFind(fn: (key) => true)
        |> getRecord(idx: 0)

    victorops.event(
        url: "https://alert.victorops.com/integrations/generic/20131114/alert/${apiKey}/${routingKey}",
        messageType:
            if lastReported._value < 1.0 then "CRITICAL"
            else if lastReported._value < 5.0 then "WARNING"
            else "INFO",
        entityID: "custom-alert-1",
        entityDisplayName: "Custom Alert 1",
        stateMessage: "last cpu idle alert"
    )

## victorops.endpoint

`victorops.endpoint` creates a factory function that creates a target function for pipeline `|>` to send events 
to VictorOps for each row.

| Name | Type | Description |
| ---- | ---- | ----------- |
| url | string | REST integration URL. Usually `https://alert.victorops.com/integrations/generic/20131114/alert/$api_key/$routing_key`, use valid `api_key` and `routing_key`. |

The returned factory function accepts a `mapFn` parameter.
The `mapFn` accepts a row and returns a record with the following events fields:

| Name | Type | Description |
| ---- | ---- | ----------- |
| url | string | REST integration URL. Usually `https://alert.victorops.com/integrations/generic/20131114/alert/$api_key/$routing_key`, use valid `api_key` and `routing_key`. |
| monitoringTool | string | Monitoring agent name. |
| messageType | string | Alert behavior. Valid values: `"CRITICAL"`, `"WARNING"`, `"INFO"`. |
| entityID | string | Incident  ID. |
| entityDisplayName | string | Incident summary. |
| stateMessage | string | Verbose message. |
| timestamp | string | Incident timestamp. |

Example:

    import "contrib/bonitoo-io/victorops"
    import "influxdata/influxdb/secrets"
    import "strings"

    apiKey = secrets.get(key: "VICTOROPS_APIKEY")
    routingKey = secrets.get(key: "VICTOROPS_ROUTING")

    endpoint = victorops.endpoint(
        url: "https://alert.victorops.com/integrations/generic/20131114/alert/${apiKey}/${routingKey}",
    )
    
    from(bucket: "example-bucket")
      |> range(start: -1m)
      |> filter(fn: (r) => r._measurement == "cpu" and r._field == "usage_idle")
      |> last()
      |> endpoint(mapFn: (r) => ({
          messageType:
              if r._value < 1.0 then "CRITICAL"
              else if r._value < 5.0 then "WARNING"
              else "INFO",
          entityID: "custom-alert-1",
          entityDisplayName: "Custom Alert 1",
          stateMessage: "last cpu idle alert"
        })
      )()

## Contact

- Author: Ales Pour / Bonitoo
- Email: alespour@bonitoo.io
- Github: [@alespour](https://github.com/alespour), [@bonitoo-io](https://github.com/bonitoo-io)
- Influx Slack: [@Ales Pour](https://influxdata.com/slack)
