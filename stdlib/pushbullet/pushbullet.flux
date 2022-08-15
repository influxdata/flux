// Package pushbullet provides functions for sending data to Pushbullet.
//
// ## Metadata
// introduced: 0.66.0
//
package pushbullet


import "http"
import "json"

// defaultURL is the default Pushbullet API URL used by functions in the `pushbullet` package.
option defaultURL = "https://api.pushbullet.com/v2/pushes"

// pushData sends a push notification to the Pushbullet API.
//
// ## Parameters
//
// - url: URL of the PushBullet endpoint. Default is `"https://api.pushbullet.com/v2/pushes"`.
// - token: API token string.  Default is `""`.
// - data: Data to send to the endpoint. Data is JSON-encoded and sent to the Pushbullet's endpoint.
//
//   For how to structure data, see the [Pushbullet API documentation](https://docs.pushbullet.com/#create-push).
//
// ## Examples
//
// ### Send a push notification to Pushbullet
// ```no_run
// import "pushbullet"
//
// pushbullet.pushData(token: "mY5up3Rs3Cre7T0k3n", data: {"type": "link", "title": "Example title", "body": "Example nofication body", "url": "http://example-url.com"})
// ```
//
// ## Metadata
// tags: single notification
//
pushData = (url=defaultURL, token="", data) => {
    headers = {"Access-Token": token, "Content-Type": "application/json"}
    enc = json.encode(v: data)

    return http.post(headers: headers, url: url, data: enc)
}

// pushNote sends a push notification of type "note" to the Pushbullet API.
//
// ## Parameters
//
// - url: URL of the PushBullet endpoint. Default is `"https://api.pushbullet.com/v2/pushes"`.
// - token: API token string.  Defaults to: `""`.
// - title: Title of the notification.
// - text: Text to display in the notification.
//
// ## Examples
//
// ### Send a push notification note to Pushbullet
// ```no_run
// import "pushbullet"
//
// pushbullet.pushNote(token: "mY5up3Rs3Cre7T0k3n", data: {"type": "link", "title": "Example title", "text": "Example note text"})
// ```
//
// ## Metadata
// tags: single notification
//
pushNote = (url=defaultURL, token="", title, text) => {
    data = {type: "note", title: title, body: text}

    return pushData(token: token, url: url, data: data)
}

// endpoint creates the endpoint for the Pushbullet API and sends a notification of type note.
//
// ### Usage
// `pushbullet.endpoint()` is a factory function that outputs another function.
// The output function requires a mapFn parameter.
//
// #### mapFn
// A function that builds the record used to generate the API request.
// Requires an `r` parameter.
//
// `mapF`n accepts a table row (`r`) and returns a record that must include the
// following properties (as defined in `pushbullet.pushNote()`):
//
// - title
// - text
//
// ## Parameters
//
// - url: PushBullet API endpoint URL. Default is `"https://api.pushbullet.com/v2/pushes"`.
// - token: Pushbullet API token string.  Default is `""`.
//
// ## Examples
//
// ### Send push notifications to Pushbullet
// ```no_run
// import "pushbullet"
// import "influxdata/influxdb/secrets"
//
// token = secrets.get(key: "PUSHBULLET_TOKEN")
//
// crit_statuses = from(bucket: "example-bucket")
//     |> range(start: -1m)
//     |> filter(fn: (r) => r._measurement == "statuses" and r.status == "crit")
//
// crit_statuses
//     |> pushbullet.endpoint(token: token)(mapFn: (r) => ({title: "${r.component} is critical", text: "${r.component} is critical. {$r._field} is {r._value}."}))()
// ```
//
// ## Metadata
// tags: notification endpoints, transformations
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
                                        2 == pushNote(
                                                url: url,
                                                token: token,
                                                title: obj.title,
                                                text: obj.text,
                                            ) / 100,
                                ),
                        }
                    },
                )
