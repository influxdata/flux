// Package discord provides functions for sending messages to [Discord](https://discord.com/).
//
// ## Metadata
// introduced: 0.69.0
// contributors: **GitHub**: [@chobbs](https://github.com/chobbs) | **InfluxDB Slack**: [@chobbs](https://influxdata.com/slack)
//
package discord


import "http"
import "json"

// discordURL is the Discord webhook URL.
// Default is `https://discordapp.com/api/webhooks/`.
option discordURL = "https://discordapp.com/api/webhooks/"

// send sends a single message to a Discord channel using a Discord webhook.
//
// ## Parameters
//
// - webhookToken: Discord [webhook token](https://discord.com/developers/docs/resources/webhook).
// - webhookID: Discord [webhook ID](https://discord.com/developers/docs/resources/webhook).
// - username: Override the Discord webhook’s default username.
// - content: Message to send to Discord (2000 character limit).
// - avatar_url: Override the Discord webhook’s default avatar.
//
// ## Examples
// ### Send the last reported status to Discord
// ```no_run
// import "contrib/chobbs/discord"
// import "influxdata/influxdb/secrets"
//
// token = secrets.get(key: "DISCORD_TOKEN")
//
// lastReported = from(bucket: "example-bucket")
//     |> range(start: -1m)
//     |> filter(fn: (r) => r._measurement == "statuses")
//     |> last()
//     |> findRecord(fn: (key) => true, idx: 0)
//
// discord.send(
//     webhookToken: token,
//     webhookID: "1234567890",
//     username: "chobbs",
//     content: "The current status is \"${lastReported.status}\".",
//     avatar_url: "https://staff-photos.net/pic.jpg",
// )
// ```
//
// ## Metadata
// tags: single notification
//
send = (
        webhookToken,
        webhookID,
        username,
        content,
        avatar_url="",
    ) =>
    {
        data = {username: username, content: content, avatar_url: avatar_url}
        headers = {"Content-Type": "application/json"}
        encode = json.encode(v: data)

        return
            http.post(
                headers: headers,
                url: discordURL + webhookID + "/" + webhookToken,
                data: encode,
            )
    }

// endpoint sends a single message to a Discord channel using a
// [Discord webhook](https://support.discord.com/hc/en-us/articles/228383668-Intro-to-Webhooks&?page=3)
// and data from table rows.
//
// ### Usage
// `discord.endpoint` is a factory function that outputs another function.
// The output function requires a `mapFn` parameter.
//
// #### mapFn
// A function that builds the record used to generate the Discord webhook request.
// Requires an `r` parameter.
//
// `mapFn` accepts a table row (`r`) and returns a record that must include the following field:
//
// - `content`
//
// For more information, see the `discord.send()` `content` parameter.
//
// ## Parameters
// - webhookToken: Discord [webhook token](https://discord.com/developers/docs/resources/webhook).
// - webhookID: Discord [webhook ID](https://discord.com/developers/docs/resources/webhook).
// - username: Override the Discord webhook’s default username.
// - avatar_url: Override the Discord webhook’s default avatar.
//
// ## Examples
// ### Send critical statuses to a Discord channel
// ```no_run
// import "influxdata/influxdb/secrets"
// import "contrib/chobbs/discord"
//
// discordToken = secrets.get(key: "DISCORD_TOKEN")
// endpoint = telegram.endpoint(
//     webhookToken: discordToken,
//     webhookID: "123456789",
//     username: "critBot",
// )
//
// crit_statuses = from(bucket: "example-bucket")
//     |> range(start: -1m)
//     |> filter(fn: (r) => r._measurement == "statuses" and status == "crit")
//
// crit_statuses
//     |> endpoint(
//         mapFn: (r) => ({
//             content: "The status is critical!",
//         }),
//     )()
// ```
//
// ## Metadata
// tags: notifcation endpoints, transformations
//
endpoint = (webhookToken, webhookID, username, avatar_url="") =>
    (mapFn) =>
        (tables=<-) =>
            tables
                |> map(
                    fn: (r) => {
                        obj = mapFn(r: r)

                        return {r with _sent:
                                string(
                                    v:
                                        2 == send(
                                                webhookToken: webhookToken,
                                                webhookID: webhookID,
                                                username: username,
                                                avatar_url: avatar_url,
                                                content: obj.content,
                                            ) / 100,
                                ),
                        }
                    },
                )
