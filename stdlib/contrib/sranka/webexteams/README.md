# Webexteams Package

Use this Flux Package to send a message to [Webex Teams](https://www.webex.com/team-collaboration.html).

## webexteams.message

`message` function sends a single message to Webex as described in https://developer.webex.com/docs/api/v1/messages/create-a-message API. 
See [webexteams.flux](./webexteams.flux) for details

Basic Example:

    import "contrib/sranka/webexteams"
    import "influxdata/influxdb/secrets"

    //this value can be stored in the secret-store()
    apiToken = secrets.get(key: "WEBEX_API_TOKEN")

    lastReported =
      from(bucket: "example-bucket")
        |> range(start: -1m)
        |> filter(fn: (r) => r._measurement == "statuses")
        |> last()
        |> tableFind(fn: (key) => true)
        |> getRecord(idx: 0)

    webexteams.message(
      // url: "https://webexapis.com",
      token: apiToken,
      roomId: "Y2lzY2.....",
      text: "Great Scott!- Disk usage is: ${lastReported.status}."
    )

## webexteams.endpoint 

`endpoint` function creates a factory function that accepts a mapping function `mapFn` and creates a target function for pipeline `|>` that sends messages from table rows. The `mapFn` accepts a table row and returns an object with `roomId`, `text` and `markdown` properties as defined in the `webexteams.message` function arguments. Arguments:

| Name     | Type   | Description                                                       |
| ----     | ----   | -----------                                                       |
| url      | string | base URL of Webex API endpoint without a trailing slash, by default "https://webexapis.com" |
| apiKey   | string | [Webex API access token](https://developer.webex.com/docs/api/getting-started), required. |

Basic Example:

    import "contrib/sranka/webexteams"
    import "influxdata/influxdb/secrets"

    // this value can be stored in the secret-store()
    token = secrets.get(key: "WEBEX_API_KEY")

    lastReported =
    from(bucket: "example-bucket")
      |> range(start: -1m)
      |> filter(fn: (r) => r._measurement == "statuses")
      |> last()
      |> tableFind(fn: (key) => true)
      |> webexteams.endpoint(token: token)(mapFn: (r) => ({
              roomId: "Y2lzY2.....",
              text: "",
              markdown: "Great Scott! Disk usage is: **${r.status}**.", 
            })
         )()


## Contact

- Author: Pavel Zavora
- Email: pavel.zavora@bonitoo.io
- Github: [@sranka](https://github.com/sranka)
- Influx Slack: [@sranka](https://influxdata.com/slack)
