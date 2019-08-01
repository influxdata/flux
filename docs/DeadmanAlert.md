### Deadman Alert

The `deadman` function is used in the context of a task to alert when a group stops reporting data.
For example, say you have the following task definition:

```
option task = {
    ...
    period: 1h,
    every:  15m,
    ...
}
from(bucket: "telegraf/autogen")
    |> range(start: -period)
    |> filter(fn: (r) => r._measurement == "cpu")
```

In order to be alerted when the above task stops reporting data for a group, you would define a deadman alert using the `deadman` function like so:
```
option task = {
    ...
    period: 1h,
    every:  15m,
    ...
}
from(bucket: "telegraf/autogen")
    |> range(start: -perod)
    |> filter(fn: (r) => r._measurement == "cpu")
    |> deadman(d: task.every)
    |> alert(crit: (r) => true)
```

where the `deadman` function is defined as follows:
```
deadman = (d, tables=<-) => tables
    |> sort(columns: ["_time"])
    |> last()
    |> filter(fn: (r) => r._time < now() - d)
```

Note the `deadman` function takes a stream of tables and a duration and returns all groups **not** observed within the interval defined by `[now() - d, now()]`.
For example, given a stream called `tables`, grouped by (`_measurement`, `host`):

| _time      | _measurement | host | _value |
| ---------- | ------------ | ---- | ------ |
| now() - 5s | cpu          | A    | 56     |
| now() - 3s | cpu          | B    | 17     |
| now() - 1s | cpu          | C    | 18     |

`tables |> deadman(d: 4s)` produces a non-empty result:

| _time      | _measurement | host | _value |
| ---------- | ------------ | ---- | ------ |
| now() - 5s | cpu          | A    | 56     |

And therefore `tables |> deadman(d: 4s) |> alert(crit: (r) => true)` triggers an alert that the group defined by `_measurement=cpu,host=A` stopped reporting data.
