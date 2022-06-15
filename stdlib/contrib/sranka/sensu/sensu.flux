// Package sensu provides functions for sending events to [Sensu Go](https://docs.sensu.io/sensu-go/latest/).
//
// ## Sensu API Key authentication
//
// The Flux Sensu package only supports [Sensu API key authentication](https://docs.sensu.io/sensu-go/latest/api/#authenticate-with-an-api-key).
// All `sensu` functions require an `apiKey` parameter to successfully authenticate with your Sensu service.
// For information about managing Sensu API keys, see the [Sensu APIKeys API documentation](https://docs.sensu.io/sensu-go/latest/api/apikeys/).
//
// ## Metadata
// introduced: 0.90.0
package sensu


import "http"
import "json"

// toSensuName translates a string value to a Sensu name
// by replacing non-alphanumeric characters (`[a-zA-Z0-9_.-]`) with underscores (`_`).
//
// ## Parameters
// - v: String to operate on.
//
// ## Examples
//
// ### Convert a string into a Sensu name
// ```no_run
// import "contrib/sranka/sensu"
//
// sensu.toSensuName(v: "Example name conversion")
//
// // Returns "Example_name_conversion"
// ```
//
builtin toSensuName : (v: string) => string

// event sends a single event to the [Sensu Events API](https://docs.sensu.io/sensu-go/latest/api/events/#create-a-new-event).
//
// ## Parameters
//
// - url: Base URL of [Sensu API](https://docs.sensu.io/sensu-go/latest/migrate/#architecture)
//   without a trailing slash.
//
//   Example: `http://localhost:8080`
//
// - apiKey: Sensu [API Key](https://docs.sensu.io/sensu-go/latest/operations/control-access/).
// - checkName: Check name.
//
//   Use alphanumeric characters, underscores (`_`), periods (`.`), and hyphens (`-`).
//   All other characters are replaced with an underscore.
//
// - text: Event text.
//
//   Mapped to `output` in the Sensu Events API request.
//
// - handlers: Sensu handlers to execute. Default is `[]`.
// - status: Event status code that indicates [state](https://docs.influxdata.com/flux/v0.x/stdlib/contrib/sranka/sensu/event/#state).
//   Default is `0`.
//
//   | Status code     | State                   |
//   | :-------------- | :---------------------- |
//   | 0               | OK                      |
//   | 1               | WARNING                 |
//   | 2               | CRITICAL                |
//   | Any other value | UNKNOWN or custom state |
//
// - state: Event state.
//   Default is `"passing"` for `0` [status](https://docs.influxdata.com/flux/v0.x/stdlib/contrib/sranka/sensu/event/#status) and `"failing"` for other statuses.
//
//   **Accepted values**:
//   - `"failing"`
//   - `"passing"`
//   - `"flapping"`
//
// - namespace: [Sensu namespace](https://docs.sensu.io/sensu-go/latest/reference/rbac/).
//   Default is `"default"`.
// - entityName: Event source.
//   Default is `influxdb`.
//
//   Use alphanumeric characters, underscores (`_`), periods (`.`), and hyphens (`-`).
//   All other characters are replaced with an underscore.
//
// ## Examples
// ### Send the last reported status to Sensu
// ```no_run
// import "influxdata/influxdb/secrets"
// import "contrib/sranka/sensu"
//
// apiKey = secrets.get(key: "SENSU_API_KEY")
//
// lastReported = from(bucket: "example-bucket")
//     |> range(start: -1m)
//     |> filter(fn: (r) => r._measurement == "statuses")
//     |> last()
//     |> findRecord(fn: (key) => true, idx: 0)
//
// sensu.event(
//     url: "http://localhost:8080",
//     apiKey: apiKey,
//     checkName: "diskUsage",
//     text: "Disk usage is **${lastReported.status}**.",
// )
// ```
//
// ## Metadata
// tags: single notification
//
event = (
        url,
        apiKey,
        checkName,
        text,
        handlers=[],
        status=0,
        state="",
        namespace="default",
        entityName="influxdb",
    ) =>
    {
        data = {
            entity: {entity_class: "proxy", metadata: {name: toSensuName(v: entityName)}},
            check: {
                output: text,
                state: if state != "" then state else if status == 0 then "passing" else "failing",
                status: status,
                handlers: handlers,
                // required
                interval: 60,
                metadata: {name: toSensuName(v: checkName)},
            },
        }
        headers = {"Content-Type": "application/json; charset=utf-8", "Authorization": "Key " + apiKey}
        enc = json.encode(v: data)

        return http.post(headers: headers, url: url + "/api/core/v2/namespaces/" + namespace + "/events", data: enc)
    }

// endpoint sends an event
// to the [Sensu Events API](https://docs.sensu.io/sensu-go/latest/api/events/#create-a-new-event)
// using data from table rows.
//
// ## Parameters
// - url: Base URL of [Sensu API](https://docs.sensu.io/sensu-go/latest/migrate/#architecture)
//   *without a trailing slash*.
//   Example: `http://localhost:8080`.
// - apiKey: Sensu [API Key](https://docs.sensu.io/sensu-go/latest/operations/control-access/).
// - handlers: [Sensu handlers](https://docs.sensu.io/sensu-go/latest/reference/handlers/) to execute.
//   Default is `[]`.
// - namespace: [Sensu namespace](https://docs.sensu.io/sensu-go/latest/reference/rbac/).
//   Default is `default`.
// - entityName: Event source.
//   Default is `influxdb`.
//
//   Use alphanumeric characters, underscores (`_`), periods (`.`), and hyphens (`-`).
//   All other characters are replaced with an underscore.
//
// ## Usage
// `sensu.endpoint()` is a factory function that outputs another function.
// The output function requires a `mapFn` parameter.
//
// `mapFn`
// A function that builds the object used to generate the POST request. Requires an `r` parameter.
//
// `mapFn` accepts a table row (`r`) and returns an object that must include the following fields:
//
// - `checkName`
// - `text`
// - `status`
//
// For more information, see `sensu.event()` parameters.
//
// ## Examples
// ### Send critical status events to Sensu
// ```no_run
// import "influxdata/influxdb/secrets"
// import "contrib/sranka/sensu"
//
// token = secrets.get(key: "TELEGRAM_TOKEN")
// endpoint = sensu.endpoint(
//     url: "http://localhost:8080",
//     apiKey: apiKey,
// )
//
// crit_statuses = from(bucket: "example-bucket")
//     |> range(start: -1m)
//     |> filter(fn: (r) => r._measurement == "statuses" and status == "crit")
//
// crit_statuses
//     |> endpoint(
//         mapFn: (r) => ({
//             checkName: "critStatus",
//             text: "Status is critical",
//             status: 2,
//         }),
//     )()
// ```
//
// ## Metadata
// tags: notification endpoints,transformations
endpoint = (
    url,
    apiKey,
    handlers=[],
    namespace="default",
    entityName="influxdb",
) =>
    (mapFn) =>
        (tables=<-) =>
            tables
                |> map(
                    fn: (r) => {
                        obj = mapFn(r: r)

                        return {r with _sent:
                                string(
                                    v:
                                        2 == event(
                                                url: url,
                                                apiKey: apiKey,
                                                checkName: obj.checkName,
                                                text: obj.text,
                                                handlers: handlers,
                                                status: obj.status,
                                                namespace: namespace,
                                                entityName: entityName,
                                            ) / 100,
                                ),
                        }
                    },
                )
