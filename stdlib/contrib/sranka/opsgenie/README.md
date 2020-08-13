# Opsgenie Package

Use this Flux Package to send alerts to Atlassian Opsgenie. The implementation is using opsgenie v2 API that is described
in https://docs.opsgenie.com/docs/alert-api#create-alert. 

See also
- https://docs.influxdata.com/kapacitor/v1.5/event_handlers/opsgenie/v2/
- https://github.com/influxdata/kapacitor/tree/master/services/opsgenie2 

## opsgenie.sendAlert

`sendAlert` sends a message that creates an alert in Opsgenie. Arguments:

| Name        | Type   | Description                                                       |
| ----        | ----   | -----------                                                       |
| url         | string | Opsgenie API URL. Defaults to "https://api.opsgenie.com/v2/alerts". |
| apiKey      | string | API Authorization key. |
| message     | string | Alert message text, at most 130 characters. |
| alias       | string | Opsgenie alias, at most 250 characters that are use to de-deduplicate alerts. Defaults to message. |
| description | string | Description field of an alert, at most 15000 characters. Optional. |
| priority    | string | "P1", "P2", "P3", "P4" or "P5". Defaults to "P3". |
| responders  | array  | Array of strings to identify responder teams or users, a 'user:' prefix is required for users, 'teams:' prefix for teams. Optional. |
| tags        | array  | Array of string tags. Optional. |
| entity      | string | Entity of the alert, used to specify a domain of the alert. Optional. |
| actions     | array  | Array of strings that specifies actions that will be available for the alert. Optional. |
| details     | string | Additional details of an alert, it must be a JSON-encoded map of key-value string pairs. Optional. |
| visibleTo   | array  | Arrays of teams and users that the alert will become visible to without sending any notification. Optional. |

Basic Example:

    import "contrib/sranka/opsgenie"

    lastReported =
      from(bucket: "example-bucket")
        |> range(start: -1m)
        |> filter(fn: (r) => r._measurement == "statuses")
        |> last()
        |> tableFind(fn: (key) => true)
        |> getRecord(idx: 0)

    opsgenie.createAlert(
      apiKey: "xxhhhssdd...",
      message: "Great Scott!- Disk usage is: ${lastReported.status}.",
      alias: "example-disk-usage",
      responders: ["user:scott@my.net","team:itcrowd"]
    )
## opsgenie.endpoint 

`endpoint` function creates a factory function that accepts a mapping function `mapFn` and creates a target function for pipeline `|>` that sends alert messages from table rows. The `mapFn` accepts a table row and returns an object with properties defined in the `opsgenie.sendAlert` function arguments expect url. apiKey and entity. Arguments:

| Name     | Type   | Description                                                         |
| ----     | ----   | -----------                                                         |
| url      | string | Opsgenie API URL. Defaults to "https://api.opsgenie.com/v2/alerts". | 
| apiKey   | string | API Authorization key. |
| entity   | string | Entity of the alert, used to specify domain of the alert. Optional. |

Basic Example:

    import "contrib/sranka/opsgenie"
    import "influxdata/influxdb/secrets"

    // this value can be stored in the secret-store()
    apiKey = secrets.get(key: "OPSGENIE_API_KEY")

    lastReported =
    from(bucket: "example-bucket")
      |> range(start: -1m)
      |> filter(fn: (r) => r._measurement == "statuses")
      |> last()
      |> tableFind(fn: (key) => true)
      |> opsgenie.endpoint(apiKey: apiKey)(mapFn: (r) => ({
              message: "Great Scott!- Disk usage is: ${r.status}.", 
              alias: "disk-usage-${r.status}",
              description: "",
              priority: "P3,
              responders: ["user:scott","team:itcrowd"],
              tags: [],
              entity: "my-lab",
              actions: [],
              details: "{}",
              visibleTo: []
            })
         )


## Contact

- Author: Pavel Zavora
- Email: pavel.zavora@bonitoo.io
- Github: [@sranka](https://github.com/sranka)
- Influx Slack: [@sranka](https://influxdata.com/slack)
