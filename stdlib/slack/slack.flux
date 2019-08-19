package slack

import "http"
import "json"

builtin validateColorString

option defaultURL = "https://slack.com/api/chat.postMessage"

// `message` sends a single message to a Slack channel. It will work either with the chat.postMessage API or with a slack webhook.
// `url` - string - URL of the slack endpoint. Defaults to: "https://slack.com/api/chat.postMessage", if one uses the webhook api this must be acquired as part of the slack API setup. This URL will be secret. Don't worry about secrets for the initial implementation.
// `token` - string - the api token string.  Defaults to: "", and can be ignored if one uses the webhook api URL.
// `username` - string - Username posting the message.
// `channel` - string - Name of channel in which to post the message. No default.
// `workspace` - string - Name of the slack workspace to use if there are multiple. Defaults to empty string.
// `text` - string - The text to display.
// `iconEmoji` - string - Name of : enclose emoji to use as image of user when posting message, will not show as the avatar icon with a slack webhook.
// `color` - string - Color to give message: one of good, warning, and danger, or any hex rgb color value ex. #439FE0.
message = (url=defaultURL, token="", username, channel, workspace, text, iconEmoji, color) => {
    attachments = [{
        color: validateColorString(color),
        text: string(v: text),
        mrkdwn_in: ["text"],
    }]

    data = {
        username: username,
        channel: channel,
        workspace: workspace,
        attachments: attachments,
        as_user: false,
        icon_emoji: iconEmoji,
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
// The `mapFn` must return an object with `username`, `channel`, `workspace`, `text`, `iconEmoji`, and `color` fields as defined in the `message` function arguments.
endpoint = (url=defaultURL, token="") =>
    (mapFn) =>
        (tables=<-) => tables
            |> map(fn: (r) => {
                obj = mapFn(r: r)
                return {r with status: message(url: url, token: token, username: obj.username, channel: obj.channel, workspace: obj.workspace, text: obj.text, iconEmoji: obj.iconEmoji, color: obj.color)}
            })
