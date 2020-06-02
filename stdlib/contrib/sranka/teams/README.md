# Teams Package

Use this Flux Package to send a message to a Microsoft Teams channel using an incoming webhook. See https://docs.microsoft.com/en-us/microsoftteams/platform/webhooks-and-connectors/how-to/connectors-using#setting-up-a-custom-incoming-webhook .

## teams.message

`message` sends a single message to Microsoft Teams via incoming web hook. Arguments:

| Name     | Type   | Description                                                       |
| ----     | ----   | -----------                                                       |
| url      | string | Incoming web hook URL. |
| title    | string | Message card title. |
| text     | string | Message card text. |
| summary  | string | Message card summary, it can be an empty string to generate summary from text. |

All text fields can be formatted using basic [Markdown ](https://docs.microsoft.com/en-us/outlook/actionable-messages/message-card-reference#text-formatting).

Basic Example:

    import "contrib/sranka/teams"

    lastReported =
      from(bucket: "example-bucket")
        |> range(start: -1m)
        |> filter(fn: (r) => r._measurement == "statuses")
        |> last()
        |> tableFind(fn: (key) => true)
        |> getRecord(idx: 0)

    teams.message(
      url: "https://outlook.office.com/webhook/12345678-1234-...",
      title: "Disk Usage"
      text: "Great Scott!- Disk usage is: *${lastReported.status}*.",
      summary: "Disk Usage is ${lastReported.status}"
    )

## teams.endpoint 

`endpoint` function creates a factory function that accepts a mapping function `mapFn` and creates a target function for pipeline `|>` operator that sends messages from table rows. The `mapFn` accepts a table row and returns an object with `title`, `text`, and `summary` as defined in the `teams.message` function arguments. Arguments:

| Name     | Type   | Description                                                       |
| ----     | ----   | -----------                                                       |
| url      | string | Incoming web hook URL. |

Basic Example:

    import "contrib/sranka/teams"

    url = "https://outlook.office.com/webhook/12345678-1234-..."

    lastReported =
    from(bucket: "example-bucket")
      |> range(start: -1m)
      |> filter(fn: (r) => r._measurement == "statuses")
      |> last()
      |> tableFind(fn: (key) => true)
      |> teams.endpoint(url: url)(mapFn: (r) => ({
              title: "Disk Usage"
              text: "Great Scott!- Disk usage is: **${r.status}**.",
              summary: "Disk Usage is ${r.status}"
            })
         )

## Contact

- Author: Pavel Zavora
- Email: pavel.zavora@bonitoo.io
- Github: [@sranka](https://github.com/sranka)
- Influx Slack: [@sranka](https://influxdata.com/slack)
