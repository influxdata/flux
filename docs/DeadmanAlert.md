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
    period: 1h,
    every:  15m,
    ...
}
from(bucket: "telegraf/autogen")
    |> range(start: -(task.period + task.every))
    |> filter(fn: (r) => r._measurement == "cpu")
    |> deadman(off: task.every)
```

where the deadman function is defined as follows:
```
deadman(off, tables=<-) => {

    r = tables
        |> group_keys()

    s = tables
        |> filter(fn: (r) => r._time > now() - off)
        |> group_keys()

    return diff(r:r, s:s) |> alert(crit: (r) => true)
}
```

Note `group_keys` takes in a stream of tables and returns the group keys of those tables in the output:
```
group_keys = (tables=<-) =>
    |> keys()
    |> limit(n:1)
    |> drop(columns: ["_value"])
```

Lets walk through an example of a query that returns tables grouped by (`_measurement`, `host`).
Assume that query produces the following tables over the first period:

| _measurement | host | _value |
| ------------ | ---- | ------ |
| cpu          | A    | 56     |
| cpu          | B    | 17     |

But over the second period it doesn't receive a value for `host=A`:

| _measurement | host | _value |
| ------------ | ---- | ------ |
| cpu          | B    | 22     |

`r` is equal to:

| _measurement | host |
| ------------ | ---- |
| cpu          | A    |
| cpu          | B    |

`s` is equal to:

| _measurement | host |
| ------------ | ---- |
| cpu          | B    |

And `diff(r:r, s:s)` is equal to:

| _measurement | host |
| ------------ | ---- |
| cpu          | A    |

As a result, a non-empty stream is passed to the alert function and an alert is triggered.

Intuitively the deadman alert computes the set difference between the groups present in each of the two most recent intervals of data.
If the difference is non-empty, this means there is at least one group that stopped reporting in the most recent interval and an alert is fired.

The `diff` function takes two streams, `r` and `s`, and returns the rows of `r` that are not in `s`.
`diff` is equivalent to set difference and is defined as follows:
```
diff = (r, s) => join.leftAnti(left: r, right: s)
```

where `join.leftAnti` performs a left anti-join of its input tables.
That is, it returns all of the rows in `left` that do not join with any of the rows in `right`.
