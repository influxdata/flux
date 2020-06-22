# Telegram Package

Use this Flux Package to send a message to a Telegram channel using https://core.telegram.org/bots/api#sendmessage API.

The telegram API requires you to know a bot token and a channel ID. The following steps were initially used to test this package:
   1. Create a new account in an application downloaded from https://telegram.org , a phone number is required.
   1. Create a new bot: https://core.telegram.org/bots#creating-a-new-bot , you will receive a bot *token* at the end of the registration process.
   1. Create a new channel from the telegram application,
   1. Add the new bot to the channel as an Administrator, the only required permission is to _post messages_. See https://stackoverflow.com/questions/33126743/how-do-i-add-my-bot-to-a-channel to know more.
   1. Send any message to @YOUR bot in the channel. Then `curl https://api.telegram.org/bot$token/getUpdates` and look for the _id_ of the channel, it is the *channel* argument in the telegram package functions.

## telegram.message

`message` function sends a single message to a Telegram channel using the API descibed in https://core.telegram.org/bots/api#sendmessage. Arguments:

| Name     | Type   | Description                                                       |
| ----     | ----   | -----------                                                       |
| url | string | URL of the telegram bot endpoint. Defaults to: "https://api.telegram.org/bot" |
| token  | string | Telegram bot token string, required. |
| channel | string | ID of the telegram channel, required. |
| text | string | The message text. |
| parseMode  | string | Parse mode of the message text per https://core.telegram.org/bots/api#formatting-options . Defaults to "MarkdownV2" |
| disableWebPagePreview  | bool | Disables preview of web links in the sent messages when "true". Defaults to "false" |
| silent  | bool | Messages are sent silently (https://telegram.org/blog/channels-2-0#silent-messages) when "true". Defaults to "true" |


Basic Example:

    import "contrib/sranka/telegram"
    import "influxdata/influxdb/secrets"

    //this value can be stored in the secret-store()
    token = secrets.get(key: "TELEGRAM_TOKEN")

    lastReported =
      from(bucket: "example-bucket")
        |> range(start: -1m)
        |> filter(fn: (r) => r._measurement == "statuses")
        |> last()
        |> tableFind(fn: (key) => true)
        |> getRecord(idx: 0)

    telegram.message(
      token:token,
      channel: "-12345"
      text: "Great Scott!- Disk usage is: *${lastReported.status}*.",
    )

## telegram.endpoint 

`endpoint` function creates a factory function that accepts a mapping function `mapFn` and creates a target function for pipeline `|>` that sends messages from table rows. The `mapFn` accepts a table row and returns an object with `channel`, `text`, and `silent` as defined in the `telegram.message` function arguments. Arguments:

| Name     | Type   | Description                                                       |
| ----     | ----   | -----------                                                       |
| url | string | URL of the telegram bot endpoint. Defaults to: "https://api.telegram.org/bot" |
| token  | string | Telegram bot token string, required. |
| parseMode  | string | Parse mode of the message text per https://core.telegram.org/bots/api#formatting-options . Defaults to "MarkdownV2" |
| disableWebPagePreview  | bool | Disables preview of web links in the sent messages when "true". Defaults to "false" |
"true" |

Basic Example:

    import "contrib/sranka/telegram"
    import "influxdata/influxdb/secrets"

    // this value can be stored in the secret-store()
    token = secrets.get(key: "TELEGRAM_TOKEN")

    lastReported =
    from(bucket: "example-bucket")
      |> range(start: -1m)
      |> filter(fn: (r) => r._measurement == "statuses")
      |> last()
      |> tableFind(fn: (key) => true)
      |> telegram.endpoint(token: token)(mapFn: (r) => ({
              channel: "-12345", 
              text: "Great Scott!- Disk usage is: **${r.status}**.", 
              silent: true
            })
         )

## Contact

- Author: Pavel Zavora
- Email: pavel.zavora@bonitoo.io
- Github: [@sranka](https://github.com/sranka)
- Influx Slack: [@sranka](https://influxdata.com/slack)
