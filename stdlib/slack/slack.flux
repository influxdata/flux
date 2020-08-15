package slack

import "http"
import "json"

builtin validateColorString : (color: string) => string

option defaultURL = "https://slack.com/api/chat.postMessage"

// `message` sends a single message to a Slack channel. It will work either with the chat.postMessage API or with a slack webhook.
// `url` - string - URL of the slack endpoint. Defaults to: "https://slack.com/api/chat.postMessage", if one uses the webhook api this must be acquired as part of the slack API setup. This URL will be secret. Don't worry about secrets for the initial implementation.
// `token` - string - the api token string.  Defaults to: "", and can be ignored if one uses the webhook api URL.
// `channel` - string - Name of channel in which to post the message. No default.
// `text` - string - The text to display.
// `color` - string - Color to give message: one of good, warning, and danger, or any hex rgb color value ex. #439FE0.
message = (url=defaultURL, token="", channel, text, color) => {
    attachments = [{
        color: validateColorString(color),
        text: string(v: text),
        mrkdwn_in: ["text"],
    }]

    data = {
        channel: channel,
        attachments: attachments,
        as_user: false,
    }

    headers = {
        "Authorization": "Bearer " + token,
        "Content-Type": "application/json",
    }
    enc = json.encode(v:data)
    return http.post(headers: headers, url: url, data: enc)
}

// `endpoint` creates the endpoint for the Slack external service.
// `url` - string - URL of the slack endpoint. Defaults to: "https://slack.com/api/chat.postMessage", if one uses the webhook api this must be acquired as part of the slack API setup, and this URL will be secret.
// `token` - string - token for the slack endpoint.  This can be ignored if one uses the webhook url acquired as part of the slack API setup, but must be supplied if the chat.postMessage API is used.
// The returned factory function accepts a `mapFn` parameter.
// The `mapFn` must return an object with `channel`, `text`, and `color` fields as defined in the `message` function arguments.
endpoint = (url=defaultURL, token="") =>
    (mapFn) =>
        (tables=<-) => tables
            |> map(fn: (r) => {
                obj = mapFn(r: r)
                return {r with _sent: string(v: 2 == message(
                    url: url,
                    token: token,
                    channel: obj.channel,
                    text: obj.text,
                    color: obj.color,
                ) / 100)}
            })
