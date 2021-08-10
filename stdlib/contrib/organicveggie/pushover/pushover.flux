package pushover

import "http"
import "json"

option pushoverURL = "https://api.pushover.net/1/messages.json"

// send sends a message to Pushover.
// `apiToken` - Your Pushover API token. Required.
// `userKey` - The user/group key (not e-mail address) of your Pushover user. Required.
// `content` - The message to display.
// `priority` - Defaults to 0. -2 to generate no notification/alert, -1 to send as a quiet notification, 
//              1 to display as high-priority and bypass the user's quiet hours, or 2 to also require
//              confirmation from the user
// `title` - Title for your message. If unspecificed, your Pushover app's name is used. Defaults to empty string.
// `device` - A user's device name, which cause the message to go directly to that device, rather than all of
//            the user's devices (multiple devices may be separated by a comma).
send = (apiToken, userKey, content, priority=0, title="", device="") => {
    data = {
        token: apiToken,
        user: userKey,
        message: content,
        priority: priority,
        title: title,
        device: device,
    }
    headers = {
        "Content-Type": "application/json"
    }
    encode = json.encode(v:data)

    return http.post(headers: headers, url: pushoverURL, data: encode)
}

// The returned factory function accepts a `mapFn` parameter.
// The `mapFn` must return an object with `content`, as defined in the `send` function arguments.
endpoint = (apiToken, userKey, priority=0, device="") => (mapFn) => (tables=<-) => tables
    |> map(
        fn: (r) => {
            obj = mapFn(r: r)

            return {r with
                _sent: string(
                    v: 2 == send(
                        apiToken: apiToken,
                        userKey: userKey,
                        priority: priority,
                        device: device,
                        content: obj.content,
                    ) / 100,
                ),
            }
        },
    )
