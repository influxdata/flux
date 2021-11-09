// Package telegram provides functions for sending messages to [Telegram](https://telegram.org/)
// using the [Telegram Bot API](https://core.telegram.org/bots/api).
//
// FIXME:Include this content?
// https://docs.influxdata.com/flux/v0.x/stdlib/contrib/sranka/telegram/#set-up-a-telegram-bot
// 
// introduced: 0.70.0
package telegram


import "http"
import "json"

// defaultURL is the default Telegram bot URL. Default is `https://api.telegram.org/bot`.
option defaultURL = "https://api.telegram.org/bot"

// defaultParseMode is the default [Telegram parse mode](https://core.telegram.org/bots/api#formatting-options). Default is `MarkdownV2`.
option defaultParseMode = "MarkdownV2"

// defaultDisableWebPagePreview - Use Telegram web page preview by default. Default is `false`.
option defaultDisableWebPagePreview = false

// defaultSilent - Send silent Telegram notifications by default. Default is `true`.
option defaultSilent = true

// message sends a single message to a Telegram channel
// using the [`sendMessage`](https://core.telegram.org/bots/api#sendmessage) method of the Telegram Bot API.
//
// ```
// import "contrib/sranka/telegram"
// 
// telegram.message(
//   url: "https://api.telegram.org/bot",
//   token: "S3crEtTel3gRamT0k3n",
//   channel: "-12345",
//   text: "Example message text",
//   parseMode: "MarkdownV2",
//   disableWebPagePreview: false,
//   silent: true
// )
// ```
//
// ## Parameters
// 
// - url: - string - URL of the Telegram bot endpoint. Default is `https://api.telegram.org/bot`.
// - token: - string - (Required) Telegram bot token.
// - channel: - string - (Required) Telegram channel ID.
// - text: - string - Message text.
// - parseMode: - string - [Parse mode](https://core.telegram.org/bots/api#formatting-options)
//   of the message text.
//   Default is `MarkdownV2`.
// - disableWebPagePreview: - bool - Disable preview of web links in the sent message.
//   Default is `false`.
// - silent: - bool - Send message [silently](https://telegram.org/blog/channels-2-0#silent-messages).
//   Default is `true`.
// 
// ## Examples
// 
// ### Send the last reported status to Telegram
// 
// ```
// import "influxdata/influxdb/secrets"
// import "contrib/sranka/telegram"
// 
// token = secrets.get(key: "TELEGRAM_TOKEN")
// 
// lastReported =
//   from(bucket: "example-bucket")
//     |> range(start: -1m)
//     |> filter(fn: (r) => r._measurement == "statuses")
//     |> last()
//     |> findRecord(fn: (key) => true, idx: 0)
// 
//     telegram.message(
//       token: token,
//       channel: "-12345"
//       text: "Disk usage is **${lastReported.status}**.",
//     )
// ```
message = (
    url=defaultURL,
    token,
    channel,
    text,
    parseMode=defaultParseMode,
    disableWebPagePreview=defaultDisableWebPagePreview,
    silent=defaultSilent,
) => {
    data = {
        chat_id: channel,
        text: text,
        parse_mode: parseMode,
        disable_web_page_preview: disableWebPagePreview,
        disable_notification: silent,
    }
    headers = {"Content-Type": "application/json; charset=utf-8"}
    enc = json.encode(v: data)

    return http.post(headers: headers, url: url + token + "/sendMessage", data: enc)
}

// endpoint sends a message to a Telegram channel using data from table rows.
// 
// ```
// import "contrib/sranka/telegram"
// 
// telegram.endpoint(
//   url: "https://api.telegram.org/bot",
//   token: "S3crEtTel3gRamT0k3n",
//   parseMode: "MarkdownV2",
//   disableWebPagePreview: false,
// )
// ```
// 
// ## Usage
// 
// `telegram.endpoint` is a factory function that outputs another function.
// The output function requires a `mapFn` parameter.
// 
// ### `mapFn`
// A function that builds the object used to generate the POST request. Requires an `r` parameter.
// 
// `mapFn` accepts a table row (`r`) and returns an object that must include the following fields:
// 
// - `channel`
// - `text`
// - `silent`
// 
// For more information, see `telegram.message()` parameters.
//
// ## Parameters
// 
// - url: - string - URL of the Telegram bot endpoint. Default is `https://api.telegram.org/bot`.
// - token: - string - (Required) Telegram bot token.
// - parseMode: - string - [Parse mode](https://core.telegram.org/bots/api#formatting-options)
//   of the message text.
//   Default is `MarkdownV2`.
// - disableWebPagePreview: - bool - Disable preview of web links in the sent message.
//   Default is false.
// 
//   The returned factory function accepts a `mapFn` parameter.
//   The `mapFn` must return an object with `channel`, `text`, and `silent`,
//   as defined in the `message` function arguments.
//
// ## Examples
// ### Send critical statuses to a Telegram channel
// ```
// import "influxdata/influxdb/secrets"
// import "contrib/sranka/telegram"
// 
// token = secrets.get(key: "TELEGRAM_TOKEN")
// endpoint = telegram.endpoint(token: token)
// 
// crit_statuses = from(bucket: "example-bucket")
//   |> range(start: -1m)
//   |> filter(fn: (r) => r._measurement == "statuses" and status == "crit")
// 
// crit_statuses
//   |> endpoint(mapFn: (r) => ({
//       channel: "-12345",
//       text: "Disk usage is **${r.status}**.",
//       silent: true
//     })
//   )()
// ```
// 
// tag: notification-endpoints
endpoint = (url=defaultURL, token, parseMode=defaultParseMode, disableWebPagePreview=defaultDisableWebPagePreview) => (mapFn) => (tables=<-) => tables
    |> map(
        fn: (r) => {
            obj = mapFn(r: r)
    
            return {r with
                _sent: string(
                    v: 2 == message(
                        url: url,
                        token: token,
                        channel: obj.channel,
                        text: obj.text,
                        parseMode: parseMode,
                        disableWebPagePreview: disableWebPagePreview,
                        silent: obj.silent,
                    ) / 100,
                ),
            }
        },
    )
