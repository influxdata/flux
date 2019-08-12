### Deadman Alert

The `deadman` function is used in the context of a task to alert when a group stops reporting data.
For example, say you have the following task definition:

```
option task = {
    ...
    every: 15m,
    ...
}
from(bucket: "telegraf/autogen")
    |> range(start: -1h)
    |> filter(fn: (r) => r._measurement == "cpu")
```

In order to be alerted when the above task stops reporting data for a group, you would define a deadman alert using the `deadman` function like so:
```
option task = {
    ...
    every: 15m,
    ...
}
from(bucket: "telegraf/autogen")
    |> range(start: -1h)
    |> filter(fn: (r) => r._measurement == "cpu")
    |> deadman(t: now() - task.every)
    |> alert(crit: (r) => r.dead)
```

The `deadman` function is defined as follows:
```
deadman = (t, tables=<-) => tables
    |> max(column: "_time")
    |> map(fn: (r) => ( {r with dead: r._time < t} ))
```

It takes a stream of tables and a time `t` and reports which groups or series were observed after time `t` and which were not.
The output stream can then be passed into an alert function to alert when a group or series has stopped reporting data.
For example, given a stream called `tables`, grouped by (`_measurement`, `host`):

| _time      | _measurement | host | _value |
| ---------- | ------------ | ---- | ------ |
| now() - 6s | cpu          | A    | 57     |
| now() - 5s | cpu          | A    | 56     |
| now() - 3s | cpu          | B    | 17     |
| now() - 2s | cpu          | B    | 18     |
| now() - 1s | cpu          | C    | 18     |
| now() - 0s | cpu          | C    | 25     |

`tables |> deadman(t: now()-4s)` produces the following stream:

| _time      | _measurement | host | _value | dead  |
| ---------- | ------------ | ---- | ------ | ----- |
| now() - 5s | cpu          | A    | 56     | true  |
| now() - 2s | cpu          | B    | 18     | false |
| now() - 0s | cpu          | C    | 25     | false |

And as a result `tables |> deadman(t: now()-4s) |> alert(crit: (r) => r.dead)` triggers an alert that the group defined by `_measurement=cpu,host=A` stopped reporting data.
