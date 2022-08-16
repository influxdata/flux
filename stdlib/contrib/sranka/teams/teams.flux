// Package teams (Microsoft Teams) provides functions
// for sending messages to a [Microsoft Teams](https://www.microsoft.com/microsoft-365/microsoft-teams/group-chat-software)
// channel using an [incoming webhook](https://docs.microsoft.com/microsoftteams/platform/webhooks-and-connectors/how-to/add-incoming-webhook).
//
// ## Metadata
// introduced: 0.70.0
// contributors: **GitHub**: [@sranka](https://github.com/sranka) | **InfluxDB Slack**: [@sranka](https://influxdata.com/slack)
//
package teams


import "http"
import "json"
import "strings"

// summaryCutoff is the limit for message summaries.
// Default is `70`.
option summaryCutoff = 70

// message sends a single message to a Microsoft Teams channel using an
// [incoming webhook](https://docs.microsoft.com/microsoftteams/platform/webhooks-and-connectors/how-to/add-incoming-webhook).
//
// ## Parameters
//
// - url: Incoming webhook URL.
// - title: Message card title.
// - text: Message card text.
// - summary: Message card summary.
//   Default is `""`.
//
//   If no summary is provided, Flux generates the summary from the message text.
//
// ## Examples
// ### Send the last reported status to a Microsoft Teams channel
// ```no_run
// import "contrib/sranka/teams"
//
// lastReported = from(bucket: "example-bucket")
//     |> range(start: -1m)
//     |> filter(fn: (r) => r._measurement == "statuses")
//     |> last()
//     |> findRecord(fn: (key) => true, idx: 0)
//
// teams.message(
//     url: "https://outlook.office.com/webhook/example-webhook",
//     title: "Disk Usage",
//     text: "Disk usage is: *${lastReported.status}*.",
//     summary: "Disk usage is ${lastReported.status}",
// )
// ```
//
// ## Metadata
// tags: single notification
//
message =
    (url, title, text, summary="") =>
        {
            headers = {"Content-Type": "application/json; charset=utf-8"}

            // see https://docs.microsoft.com/en-us/outlook/actionable-messages/message-card-reference#card-fields
            // using string body, object cannot be used because '@' is an illegal character in the object property key
            summary2 =
                if summary == "" then
                    text
                else
                    summary
            shortSummary =
                if strings.strlen(v: summary2) > summaryCutoff then
                    "${strings.substring(v: summary2, start: 0, end: summaryCutoff)}..."
                else
                    summary2
            body = "{
\"@type\": \"MessageCard\",
\"@context\": \"http://schema.org/extensions\",
\"title\": ${string(v: json.encode(v: title))},
\"text\": ${string(v: json.encode(v: text))},
\"summary\": ${string(v: json.encode(v: shortSummary))}
}"

            return http.post(headers: headers, url: url, data: bytes(v: body))
        }

// endpoint sends a message to a Microsoft Teams channel using data from table rows.
//
// ### Usage
// `teams.endpoint` is a factory function that outputs another function.
// The output function requires a `mapFn` parameter.
//
// #### mapFn
// A function that builds the object used to generate the POST request. Requires an `r` parameter.
//
// `mapFn` accepts a table row (`r`) and returns an object that must include the following fields:
//
// - `title`
// - `text`
// - `summary`
//
// For more information, see `teams.message` parameters.
//
// ## Parameters
// - url: Incoming webhook URL.
//
// ## Examples
// ### Send critical statuses to a Microsoft Teams channel
// ```no_run
// import "contrib/sranka/teams"
//
// url = "https://outlook.office.com/webhook/example-webhook"
// endpoint = teams.endpoint(url: url)
//
// crit_statuses = from(bucket: "example-bucket")
//     |> range(start: -1m)
//     |> filter(fn: (r) => r._measurement == "statuses" and status == "crit")
//
// crit_statuses
//     |> endpoint(
//         mapFn: (r) => ({
//             title: "Disk Usage",
//             text: "Disk usage is: **${r.status}**.",
//             summary: "Disk usage is ${r.status}",
//         }),
//     )()
// ```
//
// ## Metadata
// tags: notification endpoints, transformations
//
endpoint = (url) =>
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
                                                title: obj.title,
                                                text: obj.text,
                                                summary:
                                                    if exists obj.summary then obj.summary else "",
                                            ) / 100,
                                ),
                        }
                    },
                )
