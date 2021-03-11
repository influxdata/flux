# Zenoss Package

Use this Flux package to send events to Zenoss.

Event fields are described in [Event Fields](https://help.zenoss.com/zsd/RM/administering-resource-manager/event-management/event-fields) Zenoss documentation topic.

## zenoss.event

`zenoss.event` sends an event to Zenos. It has the following arguments:

| Name | Type | Description | Default value |
| ---- | ---- | ----------- | --- |
| url | string | Zenoss events endpoint URL. | |
| username  | string | HTTP BASIC authentication username. | "" (no auth) | 
| password | string | HTTP BASIC authentication username. | "" (no auth) |
| action | string | Router name. | "EventsRouter" |
| method | string | Router method | "add_event" |
| type | string | Event type | "rpc" |
| tid | int | Temporary request transaction ID | 1 |
| summary | string | Event summary | "" |
| device | string | Related device | "" |
| component | string | Related component | "" |
| severity | string | Event severity | |
| eventClass | string | Event class | "" |
| eventClassKey | string | Event class key | "" |
| collector | string | Collector |  |
| message | string | Related message |  |

Supported severity values are: `"Critical"`, `"Warning"`, `"Info"`, `"Clear"`.

Example:

    import "contrib/bonitoo-io/zenoss"
    import "influxdata/influxdb/secrets"
    import "strings"

    username = secrets.get(key: "ZENOSS_USERNAME")
    password = secrets.get(key: "ZENOSS_PASSWORD")

    lastReported =
      from(bucket: "example-bucket")
        |> range(start: -1m)
        |> filter(fn: (r) => r._measurement == "cpu" and r._field == "usage_idle")
        |> last()
        |> tableFind(fn: (key) => true)
        |> getRecord(idx: 0)

    zenoss.event(
        url: "https://tenant.zenoss.io:8080/zport/dmd/evconsole_router",
        username: username,
        username: password,
        device: lastReported.host,
        component: "CPU",
        eventClass: "/App,
        severity:
            if lastReported._value < 1.0 then "Critical"
            else if lastReported._value < 5.0 then "Warning"
            else "Info",
    )

## zenoss.endpoint

`zenoss.endpoint` creates a factory function that creates a target function for pipeline `|>` to send event 
to Zenoss for each row.

| Name | Type | Description | Default value |
| ---- | ---- | ----------- | --- |
| url | string | Zenoss events endpoint URL. | |
| username  | string | HTTP BASIC authentication username. | | 
| password | string | HTTP BASIC authentication username. | |
| action | string | Router name. | "EventsRouter" |
| method | string | Router method | "add_event" |
| type | string | Event type | "rpc" |
| tid | int | Temporary request transaction ID | 1 |

The returned factory function accepts a `mapFn` parameter.
The `mapFn` accepts a row and returns a record with the following events fields:

| Name | Type | Description |
| ---- | ---- | ----------- |
| summary | string | Event summary |
| device | string | Related device |
| component | string | Related component |
| severity | string | Event severity |
| eventClass | string | Event class |
| eventClassKey | string | Event class key |
| collector | string | Collector |
| message | string | Related message |

Example:

    import "contrib/bonitoo-io/zenoss"
    import "influxdata/influxdb/secrets"
    import "strings"

    username = secrets.get(key: "ZENOSS_USERNAME")
    password = secrets.get(key: "ZENOSS_PASSWORD")

    endpoint = zenoss.endpoint(
        url: "https://tenant.zenoss.io:8080/zport/dmd/evconsole_router",
        username: username,
        username: password
    )
    
    from(bucket: "example-bucket")
      |> range(start: -1m)
      |> filter(fn: (r) => r._measurement == "cpu" and r._field == "usage_idle")
      |> last()
      |> endpoint(mapFn: (r) => ({
          device: lastReported.host,
          component: "CPU",
          eventClass: "/App,
          severity:
              if r._value < 1.0 then "Critical"
              else if r._value < 5.0 then "Warning"
              else "Info",
        })
      )()

## Contact

- Author: Ales Pour / Bonitoo
- Email: alespour@bonitoo.io
- Github: [@alespour](https://github.com/alespour), [@bonitoo-io](https://github.com/bonitoo-io)
- Influx Slack: [@Ales Pour](https://influxdata.com/slack)
