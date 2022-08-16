// Package webexteams provides functions that send messages
// to [Webex Teams](https://www.webex.com/team-collaboration.html).
//
// ## Metadata
// introduced: 0.125.0
// contributors: **GitHub**: [@sranka](https://github.com/sranka) | **InfluxDB Slack**: [@sranka](https://influxdata.com/slack)
//
package webexteams


import "http"
import "json"

// message sends a single message to Webex
// using the [Webex messages API](https://developer.webex.com/docs/api/v1/messages/create-a-message).
//
// ## Parameters
//
// - url: Base URL of Webex API endpoint (without a trailing slash).
//   Default is `https://webexapis.com`.
// - token: [Webex API access token](https://developer.webex.com/docs/api/getting-started).
// - roomId: Room ID to send the message to.
// - text: Plain text message.
// - markdown: [Markdown formatted message](https://developer.webex.com/docs/api/basics#formatting-messages).
//
// ## Examples
// ### Send the last reported status to Webex Teams
// ```no_run
// import "contrib/sranka/webexteams"
// import "influxdata/influxdb/secrets"
//
// apiToken = secrets.get(key: "WEBEX_API_TOKEN")
//
// lastReported = from(bucket: "example-bucket")
//     |> range(start: -1m)
//     |> filter(fn: (r) => r._measurement == "statuses")
//     |> last()
//     |> findRecord(fn: (key) => true, idx: 0)
//
// webexteams.message(
//     token: apiToken,
//     roomId: "Y2lzY29zcGFyazovL3VzL1JPT00vYmJjZWIxYWQtNDNmMS0zYjU4LTkxNDctZjE0YmIwYzRkMTU0",
//     text: "Disk usage is ${lastReported.status}.",
//     markdown: "Disk usage is **${lastReported.status}**.",
// )
// ```
//
// ## Metadata
// tags: single notification
//
message = (
        url="https://webexapis.com",
        token,
        roomId,
        text,
        markdown,
    ) =>
    {
        data = {text: text, markdown: markdown, roomId: roomId}
        headers = {
            "Content-Type": "application/json; charset=utf-8",
            "Authorization": "Bearer " + token,
        }

        content = json.encode(v: data)

        return http.post(headers: headers, url: url + "/v1/messages", data: content)
    }

// endpoint returns a function that sends a message that includes data from input rows to a Webex room.
//
// ### Usage
// `webexteams.endpoint` is a factory function that outputs another function.
// The output function requires a `mapFn` parameter.
//
// #### mapFn
// A function that builds the object used to generate the POST request. Requires an `r` parameter.
//
// `mapFn` accepts a table row (`r`) and returns an object that must include the following fields:
//
// - `roomId`
// - `text`
// - `markdown`
//
// For more information, see `webexteams.message` parameters.
//
// ## Parameters
//
// - url: Base URL of Webex API endpoint (without a trailing slash).
//   Default is `https://webexapis.com`.
// - token: [Webex API access token](https://developer.webex.com/docs/api/getting-started).
//
// ## Examples
// ### Send the last reported status to Webex Teams
// ```no_run
// import "contrib/sranka/webexteams"
// import "influxdata/influxdb/secrets"
//
// token = secrets.get(key: "WEBEX_API_KEY")
//
// from(bucket: "example-bucket")
//     |> range(start: -1m)
//     |> filter(fn: (r) => r._measurement == "statuses")
//     |> last()
//     |> tableFind(fn: (key) => true)
//     |> webexteams.endpoint(token: token)(
//         mapFn: (r) => ({
//             roomId: "Y2lzY29zcGFyazovL3VzL1JPT00vYmJjZWIxYWQtNDNmMS0zYjU4LTkxNDctZjE0YmIwYzRkMTU0",
//             text: "",
//             markdown: "Disk usage is **${r.status}**.",
//         }),
//     )()
// ```
//
// ## Metadata
// tags: notification endpoints, transformations
endpoint = (url="https://webexapis.com", token) =>
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
                                                roomId: obj.roomId,
                                                text: obj.text,
                                                markdown: obj.markdown,
                                            ) / 100,
                                ),
                        }
                    },
                )
