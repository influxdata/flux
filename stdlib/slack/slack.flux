// Package slack provides functions for sending messages to [Slack](https://slack.com/).
//
// ## Metadata
// introduced: 0.41.0
//
package slack


import "http"
import "json"

// validateColorString ensures a string contains a valid hex color code.
//
// ## Parameters
// - color: Hex color code.
//
// ## Examples
//
// ### Validate a hex color code string
// ```no_run
// import "slack"
//
// slack.validateColorString(color: "#fff")
// ```
//
builtin validateColorString : (color: string) => string

// defaultURL defines the default Slack API URL used by functions in the `slack` package.
option defaultURL = "https://slack.com/api/chat.postMessage"

// message sends a single message to a Slack channel and returns the HTTP
// response code of the request.
//
// The function works with either with the `chat.postMessage` API or with a Slack webhook.
//
// ## Parameters
//
// - url: Slack API URL.
//   Default is `https://slack.com/api/chat.postMessage`.
//
//   If using the Slack webhook API, this URL is provided ine Slack webhook setup process.
//
// - token: Slack API token. Default is `""`.
//
//   If using the Slack Webhook API, a token is not required.
//
// - channel: Slack channel or user to send the message to.
// - text: Message text.
// - color: Slack message color.
//
//     Valid values:
//     - good
//     - warning
//     - danger
//     - Any hex RGB color code
//
// ## Examples
//
// ### Send a message to Slack using a Slack webhook
// ```no_run
// import "slack"
//
// slack.message(
//     url: "https://hooks.slack.com/services/EXAMPLE-WEBHOOK-URL",
//     channel: "#example-channel",
//     text: "Example slack message",
//     color: "warning",
// )
// ```
//
// ### Send a message to Slack using chat.postMessage API
// ```no_run
// import "slack"
//
// slack.message(
//     url: "https://slack.com/api/chat.postMessage",
//     token: "mySuPerSecRetTokEn",
//     channel: "#example-channel",
//     text: "Example slack message",
//     color: "warning",
// )
// ```
//
// ## Metadata
// tags: single notification
//
message = (
    url=defaultURL,
    token="",
    channel,
    text,
    color,
) =>
{
    attachments = [{color: validateColorString(color), text: string(v: text), mrkdwn_in: ["text"]}]
    data = {channel: channel, attachments: attachments}
    headers = {"Authorization": "Bearer " + token, "Content-Type": "application/json"}
    enc = json.encode(v: data)

    return http.post(headers: headers, url: url, data: enc)
}

// endpoint returns a function that can be used to send a message to Slack per input row.
//
// Each output row includes a `_sent` column that indicates if the message for
// that row was sent successfully.
//
// ## Parameters
//
// - url: Slack API URL. Default is  `https://slack.com/api/chat.postMessage`.
//
//   If using the Slack webhook API, this URL is provided ine Slack webhook setup process.
//
// - token: Slack API token. Default is `""`.
//
//   If using the Slack Webhook API, a token is not required.
//
// ## Usage
// `slack.endpoint()` is a factory function that outputs another function.
// The output function requires a `mapFn` parameter.
//
// ### mapFn
// A function that builds the record used to generate the POST request.
//
// `mapFn` accepts a table row (`r`) and returns a record that must include the
// following properties:
//
// - channel
// - color
// - text
//
// ## Examples
//
// ### Send status alerts to a Slack endpoint
// ```no_run
// import "sampledata"
// import "slack"
//
// data = sampledata.int()
//     |> map(fn: (r) => ({r with status: if r._value > 15 then "alert" else "ok"}))
//     |> filter(fn: (r) => r.status == "alert")
//
// data
//     |> slack.endpoint(token: "mY5uP3rSeCr37T0kEN")(mapFn: (r) => ({channel: "Alerts", text: r._message, color: "danger"}))()
// ```
//
// ## Metadata
// tags: notification endpoints, transformations
//
endpoint = (url=defaultURL, token="") =>
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
                                                color: obj.color,
                                            ) / 100,
                                ),
                        }
                    },
                )
