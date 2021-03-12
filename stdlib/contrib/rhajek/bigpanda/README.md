# BigPanda Package

Use this Flux Package to send alerts to BigPanda. The implementation is using bigpanda v2 API that is described
in https://docs.bigpanda.com/docs/alert-api#create-alert. 

See also
- https://docs.influxdata.com/kapacitor/v1.5/event_handlers/bigpanda/v2/
- https://github.com/influxdata/kapacitor/tree/master/services/bigpanda 

## bigpanda.sendAlert

`sendAlert` sends a message that creates an alert in BigPanda. Arguments:

| Name        | Type   | Description                                                       |
| ----        | ----   | -----------                                                       |
| token       | string | BigPanda API Authorization token. Required. |
| appKey      | string | BigPanda App Key. Required.  |
| status      | string | Alert status, one of  `ok, critical, warning, acknowledged` Required.|
| rec         | record | Alert data. Required. |
| url         | string | BigPanda API URL. Defaults to "https://api.bigpanda.io/data/v2/alerts". Optional |

Basic Example:

    import "contrib/rhajek/bigpanda"

    data = {
        host: "my-host",
        check: "my-check",
        description: "Great Scott!- Disk usage is: ${lastReported.status}.",
    }

    bigpanda.sendAlert(url: url, appKey: appKey, token: token, status: "critical", rec: data)

## bigpanda.endpoint 

`endpoint` function creates a factory function that accepts a mapping function `mapFn` and creates a target function for pipeline `|>` that sends alert messages from table rows. The `mapFn` accepts a table row and returns an object with properties defined in the `bigpanda.sendAlert` function arguments expect appKey, token and url. Arguments:

| Name     | Type   | Description                                                         |
| ----     | ----   | -----------                                                         |
| token   | string | API Authorization key. |
| appKey   | string | BigPanda AppKey used to specify domain of the alert. |
| url      | string | BigPanda API URL. Optional. Default is "https://api.bigpanda.io/data/v2/alerts". | 

Basic Example:

    import "contrib/rhajek/bigpanda"
    import "influxdata/influxdb/secrets"

    // this value can be stored in the secret-store()
    token = secrets.get(key: "BIG_PANDA_TOKEN")

    lastReported =
    from(bucket: "example-bucket")
        |> range(start: -1m)
        |> filter(fn: (r) => r._measurement == "statuses")
        |> last()
        |> tableFind(fn: (key) => true)
        |> bigpanda.endpoint(appKey: "my-appkey", token: token)(mapFn: (r) => {
            return { r with 
                status:  bigpanda.statusFromLevel(level: r.level),
                description: "Great Scott!- Disk usage is: ${r.level}." 
            }
        })()

BigPanda alert timestamps are optional and use standard Unix time format.

Example with timestamp:

        from(...)
        |> bigpanda.endpoint(appKey: "my-appkey", token: "...") (mapFn: (r) => {
            return { r with
                status: bigpanda.statusFromLevel(level: r.level),
                description: "Great Scott!- Disk usage is: ${r.level}.",
                timestamp: int(v:now())/1000000000
        }
        })()

## bigpanda.statusFromLevel

`bigpanda.statusFromLevel()` function converts InfluxDB status level to a BigPanda status.

    import "bigpanda"
    
    bigpanda.statusFromLevel(
        level: "crit"
    )
    // returns "critical"

| Status level  | BigPanda status
| ----          | ----                                              |
| crit          | critical
| warn          | warning
| info          | ok
| ok            | ok

## Contact

- Author: Robert Hajek
- Email: robert.hajek@bonitoo.io
- Github: [@rhajek](https://github.com/rhajek)

