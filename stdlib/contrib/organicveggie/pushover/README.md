# pushover package

Use this Flux package to send messages to Pushover.

## pushover.send

`pushover.send` sends a single message to Pushover. It has the following arguments:

| Name     | Type   | Description                                                                                                                                                                                               |
| -------- | ------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| apiToken | string | your application's API token                                                                                                                                                                              |
| userKey  | string | the user/group key (not e-mail address) of your user                                                                                                                                                      |
| content  | string | your message                                                                                                                                                                                              |
| priority | int    | send as -2 to generate no notification/alert, -1 to always send as a quiet notification, 1 to display as high-priority and bypass the user's quiet hours, or 2 to also require confirmation from the user |
| title    | string | your message's title, otherwise your app's name is used                                                                                                                                                   |
| device   | string | your user's device name to send the message directly to that device, rather than all of the user's devices (multiple devices may be separated by a comma)                                                 |

Here's a sample use of the `pushover.send()` function:

```flux
import "contrib/organicveggie/pushover"
import "influxdata/influxdb/secrets"

// These values can be stored in the secret-store()
token = secrets.get(key: "PUSHOVER_TOKEN")
user = secrets.get(key: "PUSHOVER_USER_KEY")

lastReported =
    from(bucket: "example-bucket")
    |> range(start: -1m)
    |> filter(fn: (r) => r._measurement == "statuses")
    |> last()
    |> tableFind(fn: (key) => true)
    |> getRecord(idx: 0)

pushover.send(
    apiToken: token,
    userKey: user,
    content: "Warning!- Disk usage is: \"${lastReported.status}\"."
)
```

## pushover.endpoint

`pushover.endpoint` creates a factory function that creates a target function for pipeline `|>` to send messages 
to Pushover for each table row. The returned factory function accepts a `mapFn` parameter.
The `mapFn` accepts a row and returns an object with message `content`. Arguments:

| Name     | Type   | Description                                                                                                                                                                                               |
| -------- | ------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| apiToken | string | your application's API token                                                                                                                                                                              |
| userKey  | string | the user/group key (not e-mail address) of your user                                                                                                                                                      |
| priority | int    | send as -2 to generate no notification/alert, -1 to always send as a quiet notification, 1 to display as high-priority and bypass the user's quiet hours, or 2 to also require confirmation from the user |
| title    | string | your message's title, otherwise your app's name is used                                                                                                                                                   |
| device   | string | your user's device name to send the message directly to that device, rather than all of the user's devices (multiple devices may be separated by a comma)                                                 |

Here's a sample use the `pushover.endpoint()` function:

```flux
import "contrib/organicveggie/pushover"
import "influxdata/influxdb/secrets"

// These values can be stored in the secret-store()
token = secrets.get(key: "PUSHOVER_TOKEN")
user = secrets.get(key: "PUSHOVER_USER_KEY")

endpoint = pushover.endpoint(
    apiToken: token,
    userKey: user,
    priority: 1,
    device: "pixel5"
)

from(bucket: "example-bucket")
    |> range(start: -1m)
    |> filter(fn: (r) => r._measurement == "statuses")
    |> last()
    |> tableFind(fn: (key) => true)
    |> endpoint(mapFn: (r) => ({
            content: "Warning!- Disk usage is: \"${r.status}\".",
        })
    )()
```

## Contact

- Author: Sean Laurent
- Github: [@organicveggie](https://github.com/organicveggie)
