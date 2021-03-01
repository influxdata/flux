# ServiceNow Package

Use this Flux package to send alerts to ServiceNow as events.

Event fields are described in [Create Event](https://docs.servicenow.com/bundle/paris-it-operations-management/page/product/event-management/task/t_EMCreateEventManually.html) ServiceNow documentation topic.

## servicenow.event

`servicenow.event` sends an alert to ServiceNow as event. It has the following arguments:

| Name | Type | Description |
| ---- | ---- | ----------- |
| url | string | ServiceNow web service URL. |
| username  | string | HTTP BASIC authentication username. |
| password | string | HTTP BASIC authentication username. |
| source | string | Source name. Default: "Flux" |
| node | string | Node name or IP address related to the event. Default is empty string. |
| metricType | string | Metric type related to the event (eg. CPU). Default is empty string. |
| resource | string | Resource related to the event (eg. CPU-1). Default is empty string. |
| metricName | string | Metric name related to the event (eg. usage_idle). Default is empty string. |
| messageKey | string | Unique identifier of the alert / event (eg. InfluxDB alert ID). Default is empty string (ServiceNow fills in the value). |
| description | string | Alert or event description. |
| severity | string | Severity of the alert or event. Possible values: "critical", "major", "minor", "warning", "info", "clear". |
| additionalInfo | record | More information about the event.

Example:

    import "contrib/bonitoo-io/servicenow"
    import "influxdata/influxdb/secrets"
    import "strings"

    username = secrets.get(key: "SERVICENOW_USERNAME")
    password = secrets.get(key: "SERVICENOW_PASSWORD")

    lastReported =
      from(bucket: "example-bucket")
        |> range(start: -1m)
        |> filter(fn: (r) => r._measurement == "cpu" and r._field == "usage_idle")
        |> last()
        |> tableFind(fn: (key) => true)
        |> getRecord(idx: 0)

    servicenow.event(
        url: "https://tenant.service-now.com/api/global/em/jsonv2",
        username: username,
        username: password,
        node: lastReported.host,
        metricType: strings.toUpper(v: lastReported._measurement),
        resource: lastReported.instance,
        metricName: lastReported._field,
        severity:
            if lastReported._value < 1.0 then "critical"
            else if lastReported._value < 5.0 then "warning"
            else "info",
        additionalInfo: {}
    )

## servicenow.endpoint

`servicenow.endpoint` creates a factory function that creates a target function for pipeline `|>` to send events 
to ServiceNow for each row.
The returned factory function accepts a `mapFn` parameter.
The `mapFn` accepts a row and returns an object with events fields. Arguments:

| Name | Type | Description |
| ---- | ---- | ----------- |
| url | string | ServiceNow web service URL. |
| username  | string | HTTP BASIC authentication username. |
| password | string | HTTP BASIC authentication username. |
| source | string | Source name. Default: "Flux" |
| node | string | Node name or IP address related to the event. Default is empty string. |
| metricType | string | Metric type related to the event (eg. CPU). Default is empty string. |
| resource | string | Resource related to the event (eg. CPU-1). Default is empty string. |
| metricName | string | Metric name related to the event (eg. usage_idle). Default is empty string. |
| messageKey | string | Unique identifier of the alert / event (eg. InfluxDB alert ID). Default is empty string (ServiceNow fills in the value). |
| description | string | Alert or event description. |
| severity | string | Severity of the alert or event. Possible values: "critical", "major", "minor", "warning", "info", "clear". |
| additionalInfo | record | More information about the event.

Example:

    import "contrib/bonitoo-io/servicenow"
    import "influxdata/influxdb/secrets"
    import "strings"

    username = secrets.get(key: "SERVICENOW_USERNAME")
    password = secrets.get(key: "SERVICENOW_PASSWORD")

    endpoint = servicenow.endpoint(
        url: "https://tenant.service-now.com/api/global/em/jsonv2",
        username: username,
        username: password
    )
    
    from(bucket: "example-bucket")
      |> range(start: -1m)
      |> filter(fn: (r) => r._measurement == "cpu" and r._field == "usage_idle")
      |> last()
      |> endpoint(mapFn: (r) => ({
          node: r.host,
          metricType: strings.toUpper(v: r._measurement),
          resource: r.instance,
          metricName: r._field,
          severity:
              if r._value < 1.0 then "critical"
              else if r._value < 5.0 then "warning"
              else "info",
          additionalInfo: { "devId": r.dev_id }
        })
      )()

## Contact

Provide a way for users to get in touch with you if they have questions or need help using your package. What information you give is up to you, but we encourage providing those below.

- Author: Ales Pour / Bonitoo
- Email: alespour@bonitoo.io
- Github: [@alespour](https://github.com/alespour), [@bonitoo-io](https://github.com/bonitoo-io)
- Influx Slack: [@Ales Pour](https://influxdata.com/slack)
