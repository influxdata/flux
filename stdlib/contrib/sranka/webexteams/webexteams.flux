package webexteams


import "http"
import "json"

// `message` sends a single message to Webex Teams as described in [Webex Message API](https://developer.webex.com/docs/api/v1/messages/create-a-message). 
// `url` - string - base URL of Webex API endpoint without a trailing slash, "https://webexapis.com" by default.
// `token` - string - [Webex API access token](https://developer.webex.com/docs/api/getting-started).
// `roomId` - string - The room ID of the message, required.
// `text` - string - the message, in plain text.
// `markdown` - string - the message, in markdown format as explained in https://developer.webex.com/docs/api/basics
message = (
        url="https://webexapis.com",
        token,
        roomId,
        text,
        markdown,
) => {
    data = {
        text: text,
        markdown: markdown,
        roomId: roomId,
    }
    headers = {
        "Content-Type": "application/json; charset=utf-8",
        "Authorization": "Bearer " + token,
    }

    content = json.encode(v: data)

    return http.post(headers: headers, url: url + "/v1/messages", data: content)
}

// `endpoint` creates a factory function that creates a target function for pipeline `|>` to send message to Webex teams for each table row.
// `url` - string - base URL of Webex API endpoint without a trailing slash, "https://webexapis.com" by default.
// `token` - string - [Webex API access token](https://developer.webex.com/docs/api/getting-started).
// The returned factory function accepts a `mapFn` parameter.
// The `mapFn` must return an object with `roomId`, `text` and `markdown` properties  as defined in the `message` function arguments.
endpoint = (url="https://webexapis.com", token) => (mapFn) => (tables=<-) => tables
    |> map(
        fn: (r) => {
            obj = mapFn(r: r)

            return {r with
                _sent: string(
                    v: 2 == message(
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
