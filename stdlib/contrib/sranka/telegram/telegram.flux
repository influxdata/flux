// Package telegram provides functions for sending messages to [Telegram](https://telegram.org/)
// using the [Telegram Bot API](https://core.telegram.org/bots/api).
//
//
// ## Set up a Telegram bot
// The **Telegram Bot API** requires a **bot token** and a **channel ID**.
// To set up a Telegram bot and obtain the required bot token and channel ID:
//
// 1.  [Create a new Telegram account](https://telegram.org/) or use an existing account.
// 2.  [Create a Telegram bot](https://core.telegram.org/bots#creating-a-new-bot).
//     Telegram provides a **bot token** for the newly created bot.
// 3.  Use the **Telegram application** to create a new channel.
// 4.  [Add the new bot to the channel](https://stackoverflow.com/questions/33126743/how-do-i-add-my-bot-to-a-channel) as an **Administrator**.
//     Ensure the bot has permissions necessary to **post messages**.
// 5.  Send a message to bot in the channel.
// 6.  Send a request to `https://api.telegram.org/bot$token/getUpdates`.
//
//     ```sh
//     curl https://api.telegram.org/bot$token/getUpdates
//     ```
//
//     Find your **channel ID** in the `id` field of the response.
//
// ## Metadata
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
//
// ## Parameters
//
// - url: URL of the Telegram bot endpoint. Default is `https://api.telegram.org/bot`.
// - token: Telegram bot token.
// - channel: Telegram channel ID.
// - text: Message text.
// - parseMode: [Parse mode](https://core.telegram.org/bots/api#formatting-options)
//   of the message text.
//   Default is `MarkdownV2`.
// - disableWebPagePreview: Disable preview of web links in the sent message.
//   Default is `false`.
// - silent: Send message [silently](https://telegram.org/blog/channels-2-0#silent-messages).
//   Default is `true`.
//
// ## Examples
//
// ### Send the last reported status to Telegram
//
// ```no_run
// import "influxdata/influxdb/secrets"
// import "contrib/sranka/telegram"
//
// token = secrets.get(key: "TELEGRAM_TOKEN")
//
// lastReported = from(bucket: "example-bucket")
//     |> range(start: -1m)
//     |> filter(fn: (r) => r._measurement == "statuses")
//     |> last()
//     |> findRecord(fn: (key) => true, idx: 0)
//
// telegram.message(
//     token: token,
//     channel: "-12345",
//     text: "Disk usage is **${lastReported.status}**.",
// )
// ```
//
// ## Metadata
// tags: single notification
//
message = (
    url=defaultURL,
    token,
    channel,
    text,
    parseMode=defaultParseMode,
    disableWebPagePreview=defaultDisableWebPagePreview,
    silent=defaultSilent,
) =>
{
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
// - url: URL of the Telegram bot endpoint. Default is `https://api.telegram.org/bot`.
// - token: Telegram bot token.
// - parseMode: [Parse mode](https://core.telegram.org/bots/api#formatting-options)
//   of the message text.
//   Default is `MarkdownV2`.
// - disableWebPagePreview: Disable preview of web links in the sent message.
//   Default is false.
//
// The returned factory function accepts a `mapFn` parameter.
// The `mapFn` must return an record with the following properties:
//
// - `channel`
// - `text`
// - `silent`
//
// See `telegram.message` parameters for more information.
//
// ## Examples
// ### Send critical statuses to a Telegram channel
// ```no_run
// import "influxdata/influxdb/secrets"
// import "contrib/sranka/telegram"
//
// token = secrets.get(key: "TELEGRAM_TOKEN")
// endpoint = telegram.endpoint(token: token)
//
// crit_statuses = from(bucket: "example-bucket")
//     |> range(start: -1m)
//     |> filter(fn: (r) => r._measurement == "statuses" and status == "crit")
//
// crit_statuses
//     |> endpoint(
//         mapFn: (r) => ({
//             channel: "-12345",
//             text: "Disk usage is **${r.status}**.",
//             silent: true,
//         }),
//     )()
// ```
//
// ## Metadata
// tag: notification endpoints, transformations
endpoint = (url=defaultURL, token, parseMode=defaultParseMode, disableWebPagePreview=defaultDisableWebPagePreview) =>
    (mapFn) =>
        (tables=<-) =>
            tables
                |> map(
                    fn: (r) => {
                        obj = mapFn(r: r)

                        return {r with _sent:
                                string(
                                    v:
                                        2 == message(
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
