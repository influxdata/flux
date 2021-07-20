// Package pushbullet provides functions for sending data to Pushbullet.
package pushbullet


import "http"
import "json"

option defaultURL = "https://api.pushbullet.com/v2/pushes"

// pushData sends a push notification to the Pushbullet API.
//
// ## Parameters
//
// - `url` is the URL of the PushBullet endpoint. Defaults to: "https://api.pushbullet.com/v2/pushes".
// - `token` is the api token string.  Defaults to: "".
// - `data` is the data to send to the endpoint. It will be encoded in JSON and sent to PushBullet's endpoint.
// For how to structure data, see https://docs.pushbullet.com/#create-push.
//
// ## Send the last reported status to Pushbullet
//
// ```
// import "pushbullet"
// import "influxdata/influxdb/secrets"
//
// token = secrets.get(key: "PUSHBULLET_TOKEN")
//
// lastReported =
//   from(bucket: "example-bucket")
//     |> range(start: -1m)
//     |> filter(fn: (r) => r._measurement == "statuses")
//     |> last()
//     |> tableFind(fn: (key) => true)
//     |> getRecord(idx: 0)
//
// pushbullet.pushData(
//   token: token,
//   data: {
//     "type": "link",
//     "title": "Last reported status",
//     "body": "${lastReported._time}: ${lastReported.status}."
//     "url": "${lastReported.statusURL}"
//   }
// )
// ```
//
pushData = (url=defaultURL, token="", data) => {
    headers = {
        "Access-Token": token,
        "Content-Type": "application/json",
    }
    enc = json.encode(v: data)

    return http.post(headers: headers, url: url, data: enc)
}

// pushNote sends a push notification of type note to the Pushbullet API.
//
// ## Parameters
//
// - `url` is the URL of the PushBullet endpoint. Defaults to: "https://api.pushbullet.com/v2/pushes".
// - `token` is the api token string.  Defaults to: "".
// - `title` is the title of the notification.
// - `text` is the text to display in the notification.
//
// ## Send the last reported status to Pushbullet
//
// ```
// import "pushbullet"
// import "influxdata/influxdb/secrets"
//
// token = secrets.get(key: "PUSHBULLET_TOKEN")
//
// lastReported =
//   from(bucket: "example-bucket")
//     |> range(start: -1m)
//     |> filter(fn: (r) => r._measurement == "statuses")
//     |> last()
//     |> tableFind(fn: (key) => true)
//     |> getRecord(idx: 0)
//
// pushbullet.pushNote(
//   token: token,
//   data: {
//     token: token,
//     title: "Last reported status",
//     text: "${lastReported._time}: ${lastReported.status}."
//   }
// )
// ```
pushNote = (url=defaultURL, token="", title, text) => {
    data = {
        type: "note",
        title: title,
        body: text,
    }

    return pushData(token: token, url: url, data: data)
}

// endpoint creates the endpoint for the Pushbullet API and sends a notification of type note.
//
// ## Parameters
//
// - `url` is the URL of the PushBullet endpoint. Defaults to: "https://api.pushbullet.com/v2/pushes".
// - `token` is the api token string.  Defaults to: "".
// - `Usage` pushbullet.endpoint is a factory function that outputs another function. The output function requires a mapFn parameter.
// - `mapFn` is a function that builds the record used to generate the API request. Requires an r parameter.
//
// ## Send the last reported status to Pushbullet
//
// ```
// import "pushbullet"
// import "influxdata/influxdb/secrets"
//
// token = secrets.get(key: "PUSHBULLET_TOKEN")
//
// lastReported =
//   from(bucket: "example-bucket")
//     |> range(start: -10m)
//     |> filter(fn: (r) => r._measurement == "statuses")
//     |> last()
//
// lastReported
//   |> e(mapFn: (r) => ({
//       r with
//       title: r.title,
//       text: "${string(v: r._time)}: ${r.status}."
//     })
//   )()
// ```
//
endpoint = (url=defaultURL, token="") => (mapFn) => (tables=<-) => tables
    |> map(
        fn: (r) => {
            obj = mapFn(r: r)

            return {r with
                _sent: string(
                    v: 2 == pushNote(
                        url: url,
                        token: token,
                        title: obj.title,
                        text: obj.text,
                    ) / 100,
                ),
            }
        },
    )
