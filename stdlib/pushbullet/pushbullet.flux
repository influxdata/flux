package pushbullet

import "http"
import "json"

option defaultURL = "https://api.pushbullet.com/v2/pushes"

// `pushData` sends a push notification using PushBullet's APIs.
// `url` - string - URL of the PushBullet endpoint. Defaults to: "https://api.pushbullet.com/v2/pushes".
// `token` - string - the api token string.  Defaults to: "".
// `data` - object - The data to send to the endpoint. It will be encoded in JSON and sent to PushBullet's endpoint.
// For how to structure data, see https://docs.pushbullet.com/#create-push.
pushData = (url=defaultURL, token="", data) => {
    headers = {
        "Access-Token": token,
        "Content-Type": "application/json",
    }
    enc = json.encode(v: data)
    return http.post(headers: headers, url: url, data: enc)
}

// `pushNote` sends a push notification of type `note` using PushBullet's APIs.
// `url` - string - URL of the PushBullet endpoint. Defaults to: "https://api.pushbullet.com/v2/pushes".
// `token` - string - the api token string.  Defaults to: "".
// `title` - string - The title of the notification.
// `text` - string - The text to display in the notification.
pushNote = (url=defaultURL, token="", title, text) => {
    data = {
        type: "note",
        title: title,
        body: text,
    }
    return pushData(token: token, url: url, data: data)
}

// `genericEndpoint` does not work for now for a bug in type inference in the compiler.
// // `genericEndpoint` creates the endpoint for the PushBullet external service.
// // `url` - string - URL of the PushBullet endpoint. Defaults to: "https://api.pushbullet.com/v2/pushes".
// // `token` - string - token for the PushBullet endpoint.
// // The returned factory function accepts a `mapFn` parameter.
// // The `mapFn` must return an object that will be used as payload as defined in `pushData` function arguments.
// genericEndpoint = (url=defaultURL, token="") =>
//     (mapFn) =>
//         (tables=<-) => tables
//             |> map(fn: (r) => {
//                 obj = mapFn(r: r)
//                 return {r with _sent: string(v: 2 == pushData(
//                   url: url,
//                   token: token,
//                   data: obj,
//                 ) / 100)}
//             })


// `endpoint` creates the endpoint for the PushBullet external service.
// It will push notifications of type `note`.
// If you want to push something else, see `genericEndpoint`.
// `url` - string - URL of the PushBullet endpoint. Defaults to: "https://api.pushbullet.com/v2/pushes".
// `token` - string - token for the PushBullet endpoint.
// The returned factory function accepts a `mapFn` parameter.
// The `mapFn` must return an object with `title` and `text` fields as defined in the `pushNote` function arguments.
endpoint = (url=defaultURL, token="") =>
    (mapFn) =>
        (tables=<-) => tables
            |> map(fn: (r) => {
                obj = mapFn(r: r)
                return {r with _sent: string(v: 2 == pushNote(
                    url: url,
                    token: token,
                    title: obj.title,
                    text: obj.text,
                ) / 100)}
            })
