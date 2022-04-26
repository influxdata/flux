# teams.endpoint
This Repository in based on https://github.com/influxdata/flux/tree/v0.113.1/stdlib/contrib/sranka teams code and adds extra features like mentions and buttons.

Basic Example:
```
url= "https://..."
endpoint = teams.endpoint(url: url)
mentions = teams.addMention(name : "team user name",id:"team user ID")
button = teams.addButton(type: "Action.OpenUrl", title: "Go To Google.com", url:"google.com" )
crit_statuses =from(bucket: "bucket")
  |> range(start: -15s)
  |> filter(fn: (r) => r["_measurement"] == "win_cpu")
  |> endpoint(mapFn: (r) => ({
      title: "Memory Usage",
      text: "<at>team user name</at>: ${r.host}: Process uses ${r._value} GB",
      summary: "Alert",
      mention: mentions,
      button : button + button
    }),
  )()
```
