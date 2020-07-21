# Discord Package


Use this Flux Package to send messages to a Discord channel using a webhook.

## discord.send

`discord.send` sends a single message to a Discord channel using a webhook. It has the following arguments:

| Name     | Type   | Description                                                       |
| ----     | ----   | -----------                                                       |
| webhookToken | string | secure token of the webhook. (Auto-gen from the WebhookURL)   |
| webhookID  | string | ID of the webhook. (Auto-gen from the WebhookURL)               |
| username | string | overrides the current username of the webhook.                    |
| content  | string | simple message, the message contains. (up to 2000 characters)     |
| avatar_url  | string | override the default avatar of the webhook. (_optional_)       |


Here's an example definition for the `discord.send()` function.

    import "contrib/chobbs/discord"
    import "influxdata/influxdb/secrets"

    //this value can be stored in the secret-store()
    token = secrets.get(key: "DISCORD_TOKEN")

    lastReported =
      from(bucket: "example-bucket")
        |> range(start: -1m)
        |> filter(fn: (r) => r._measurement == "statuses")
        |> last()
        |> tableFind(fn: (key) => true)
        |> getRecord(idx: 0)

    discord.send(
      webhookToken:token,
      webhookID:"1234567890",
      username:"chobbs",
      content: "Great Scott!- Disk usage is: \"${lastReported.status}\".",
      avatar_url:"https://staff-photos.net/pic.jpg"
      )

## discord.endpoint

`discord.endpoint` creates a factory function that creates a target function for pipeline `|>` to send messages 
to discord for each table row. The returned factory function accepts a `mapFn` parameter.
The `mapFn` accepts a row and returns an object with message `content`. Arguments:

| Name     | Type   | Description                                                       |
| ----     | ----   | -----------                                                       |
| webhookToken | string | secure token of the webhook. (Auto-gen from the WebhookURL)   |
| webhookID  | string | ID of the webhook. (Auto-gen from the WebhookURL)               |
| username | string | overrides the current username of the webhook.                    |
| avatar_url  | string | override the default avatar of the webhook. (_optional_)       |

Here's an example definition for the `discord.endpoint()` function

    import "contrib/chobbs/discord"
    import "influxdata/influxdb/secrets"

    //this value can be stored in the secret-store()
    token = secrets.get(key: "DISCORD_TOKEN")

    endpoint = discord.endpoint(
      webhookToken:token,
      webhookID:"1234567890",
      username:"chobbs",
      avatar_url:"https://staff-photos.net/pic.jpg"
      )
    
    from(bucket: "example-bucket")
      |> range(start: -1m)
      |> filter(fn: (r) => r._measurement == "statuses")
      |> last()
      |> tableFind(fn: (key) => true)
      |> endpoint(mapFn: (r) => ({
              content: "Great Scott!- Disk usage is: \"${r.status}\".",
            })
         )

## Contact

Provide a way for users to get in touch with you if they have questions or need help using your package. What information you give is up to you, but we encourage providing those below.

- Author: Craig Hobbs
- Email: craig@influxdata.com
- Github: [@chobbs](https://github.com/chobbs)
- Influx Slack: [@craig](https://influxdata.com/slack)
