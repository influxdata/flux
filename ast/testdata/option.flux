option task = {
    name: "foo",
    every: 1h,
    delay: 10m,
    cron: "0 2 * * *",
    retry: 5,
}

from(bucket: "test")
    |> range(start:2018-05-22T19:53:26Z)
    |> window(every: task.every)
    |> group(by: ["_field", "host"])
    |> sum()
    |> to(bucket: "test", tagColumns:["host", "_field"])
