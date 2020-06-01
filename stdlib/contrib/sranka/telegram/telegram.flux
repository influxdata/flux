package telegram

import "http"
import "json"

option defaultURL = "https://api.telegram.org/bot"
option defaultParseMode = "MarkdownV2"
option defaultDisableWebPagePreview = false
option defaultSilent = true

// `message` sends a single message to a Telegram channel using the API descibed in https://core.telegram.org/bots/api#sendmessage
// `url` - string - URL of the telegram bot endpoint. Defaults to: "https://api.telegram.org/bot"
// `token` - string - Required telegram bot token string, such as 123456789:AAxSFgij0ln9C7zUKnr4ScDi5QXTGF71S
// `channel` - string - Required id of the telegram channel.
// `text` - string - The text to display.
// `parseMode` - string - Parse mode of the message text per https://core.telegram.org/bots/api#formatting-options . Defaults to "MarkdownV2"
// `disableWebPagePreview` - bool - Disables preview of web links in the sent messages when "true". Defaults to "false"
// `silent` - bool - Messages are sent silently (https://telegram.org/blog/channels-2-0#silent-messages) when "true". Defaults to "true".
message = (url=defaultURL, token, channel, text, parseMode=defaultParseMode, disableWebPagePreview=defaultDisableWebPagePreview, silent=defaultSilent) => {
    data = {
        chat_id: channel,
        text: text,
        parse_mode: parseMode,
        disable_web_page_preview: disableWebPagePreview,
        disable_notification: silent,
    }

    headers = {
        "Content-Type": "application/json; charset=utf-8",
    }
    enc = json.encode(v:data)
    return http.post(headers: headers, url: url + token + "/sendMessage", data: enc)
}

// `endpoint` creates a factory function that creates a target function for pipeline `|>` to send messages to telegram for each table row.
// `url` - string - URL of the telegram bot endpoint. Defaults to: "https://api.telegram.org/bot"
// `token` - string - Required telegram bot token string, such as 123456789:AAxSFgij0ln9C7zUKnr4ScDi5QXTGF71S
// `parseMode` - string - Parse mode of the message text per https://core.telegram.org/bots/api#formatting-options . Defaults to "MarkdownV2"
// `disableWebPagePreview` - bool - Disables preview of web links in the sent messages when "true". Defaults to "false"
// The returned factory function accepts a `mapFn` parameter.
// The `mapFn` must return an object with `channel`, `text`, and `silent`, as defined in the `message` function arguments.
endpoint = (url=defaultURL, token, parseMode=defaultParseMode, disableWebPagePreview=defaultDisableWebPagePreview) =>
    (mapFn) =>
        (tables=<-) => tables
            |> map(fn: (r) => {
                obj = mapFn(r: r)
                return {r with _sent: string(v: 2 == message(
                    url: url,
                    token: token,
                    channel: obj.channel,
                    text: obj.text,
                    parseMode: parseMode,
                    disableWebPagePreview: disableWebPagePreview, 
                    silent: obj.silent,
                ) / 100)}
            })
