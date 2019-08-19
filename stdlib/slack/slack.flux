package slack

import "http"
import "json"

builtin validateColorString

option defaultURL = "https://slack.com/api/chat.postMessage"

// `message` sends a single message to a Slack channel.
// url - string - URL of the slack endpoint. No default value, this must be acquired as part of the slack API setup. This URL will be secret. Don't worry about secrets for the initial implementation.
// token - the api token string
// username - string - Username posting the message
// channel - string - Name of channel in which to post the message. No default
// workspace - string - Name of the slack workspace to use if there are multiple. Defaults to empty string
// text - string - The text to display
// iconEmoji - string - Name of : enclose emoji to use as image of user when posting message.
// color - string - Color to give message: one of good, warning, and danger, or any hex rgb color value ex. #439FE0.
message = (url, token, username, channel, workspace, text, iconEmoji, color) => {
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
        icon_emoji: iconEmoji,
    }
    headers = {
        "Authorization": token,
        "Content Type": "application/json",
    }
    enc = json.encode(v:data)
    return http.post(headers: headers, url: url, data: enc)
}

// `endpoint` creates the endpoint for the Slack external service.
// The returned factory function accepts a `mapFn` parameter.
// The `mapFn` must return an object with `channel` and `text` fields.
endpoint = (url=defaultURL, token) =>
    (mapFn) =>
        (tables=<-) => tables
            |> map(fn: (r) => {
                obj = mapFn(r: r)
                return {r with status: message(url: url, token: token, username: obj.username, channel: obj.channel, workspace: obj.workspace, text: obj.text, iconEmoji: obj.iconEmoji, color: obj.color)}
            })
