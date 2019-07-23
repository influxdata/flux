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

In order to be alerted when the above task stops reporting data for a group, you would define a deadman alert like so:
```
option task = {
    ...
    every: 15m,
    ...
}
f = (start, stop) =>
    from(bucket: "telegraf/autogen")
        |> range(start: start, stop: stop)
        |> filter(fn: (r) => r._measurement == "cpu")

deadman(f: f, period: 1h, offset: task.every)
```

where the deadman function is defined as follows:
```
deadman = (f, period, offset) => {

    r = f(start: now() - offset - period, stop: now() - offset)
        |> keys()
        |> drop(columns: ["_value"])

    s = f(start: now() - period, stop: now())
        |> keys()
        |> drop(columns: ["_value"])

    return diff(r: r, s: s) |> alert(crit: (r) => true)
}
```

Intuitively the deadman alert computes the set difference between the groups present in each of the two most recent intervals of data.
If the difference is non-empty, this means there is at least one group that stopped reporting in the most recent interval and an alert is fired.

The `diff` function takes two streams, `r` and `s`, and returns the rows of `r` that are not in `s`.
`diff` is equivalent to set difference and is defined as follows:
```
diff = (r, s) => join.leftAnti(left: r, right: s)
```

where `join.leftAnti` performs a left anti-join of its input tables.
That is, it returns all of the rows in `left` that do not join with any of the rows in `right`.
