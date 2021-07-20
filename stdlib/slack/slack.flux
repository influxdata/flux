// Package slack provides functions for sending data to Slack.
package slack


import "http"
import "json"

builtin validateColorString : (color: string) => string

option defaultURL = "https://slack.com/api/chat.postMessage"

// message sends a single message to a Slack channel.
// The function works with either with the chat.postMessage API or with a Slack webhook.
//
// ## Parameters
//
// - `url` is the URL of the slack endpoint.
//
//      Defaults to: "https://slack.com/api/chat.postMessage", if one uses the webhook api this must be acquired as part of the slack API setup.
//      This URL will be secret. Don't worry about secrets for the initial implementation.
//
// - `token` is the api token string.
//
//      Defaults to: "", and can be ignored if one uses the webhook api URL.
//
// - `channel` is the name of channel in which to post the message. No default.
// - `text` is the text to display.
// - `color` is the color to give message: one of good, warning, and danger, or any hex rgb color value ex. #439FE0.
//
// ## Send the last reported status to Slack using a Slack webhook
//
// ```
// import "slack"
//
// lastReported =
//   from(bucket: "example-bucket")
//     |> range(start: -1m)
//     |> filter(fn: (r) => r._measurement == "statuses")
//     |> last()
//     |> findRecord(fn: (key) => true, idx: 0)
//
// slack.message(
//   url: "https://hooks.slack.com/services/EXAMPLE-WEBHOOK-URL",
//   channel: "#system-status",
//   text: "The last reported status was \"${lastReported.status}\"."
//   color: "warning"
// )
// ```
//
// ## Send the last reported status to Slack using chat.postMessage API
//
// ```
// import "slack"
//
// lastReported =
//   from(bucket: "example-bucket")
//     |> range(start: -1m)
//     |> filter(fn: (r) => r._measurement == "statuses")
//     |> tableFind(fn: (key) => true)
//     |> getRecord(idx: 0)
//
// slack.message(
//   url: "https://slack.com/api/chat.postMessage",
//   token: "mySuPerSecRetTokEn",
//   channel: "#system-status",
//   text: "The last reported status was \"${lastReported.status}\"."
//   color: "warning"
// )
// ```
//
message = (
        url=defaultURL,
        token="",
        channel,
        text,
        color,
) => {
    attachments = [
        {color: validateColorString(color), text: string(v: text), mrkdwn_in: ["text"]},
    ]
    data = {
        channel: channel,
        attachments: attachments,
    }
    headers = {
        "Authorization": "Bearer " + token,
        "Content-Type": "application/json",
    }
    enc = json.encode(v: data)

    return http.post(headers: headers, url: url, data: enc)
}

// endpoint sends a message to Slack that includes output data.
//
// ## Parameters
//
// - `url` is the API URL of the slack endpoint. Defaults to https://slack.com/api/chat.postMessage.
//
//      If using a Slack webhook, you’ll receive a Slack webhook URL when you create an incoming webhook.
//
// - `token` is the Slack API token used to interact with Slack. Defaults to "".
// - `Usage`: slack.endpoint is a factory function that outputs another function. The output function requires a mapFn parameter.
// - `mapFn` is a function that builds the record used to generate the POST request. Requires an r parameter.
//
// ## Send critical statuses to a Slack endpoint
//
// ```
// import "slack"
//
// toSlack = slack.endpoint(url: https://hooks.slack.com/services/EXAMPLE-WEBHOOK-URL)
//
// crit_statuses = from(bucket: "example-bucket")
//   |> range(start: -1m)
//   |> filter(fn: (r) => r._measurement == "statuses" and r.status == "crit")
//
// crit_statuses
//   |> toSlack(mapFn: (r) => ({
//       channel: "Alerts",
//       text: r._message,
//       color: "danger",
//    })
//   )()
// ```
endpoint = (url=defaultURL, token="") => (mapFn) => (tables=<-) => tables
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
                        color: obj.color,
                    ) / 100,
                ),
            }
        },
    )
