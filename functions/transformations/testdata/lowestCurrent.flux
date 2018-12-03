from(bucket:"test")
    |> range(start: 2018-11-07T00:00:00Z)
    |> lowestCurrent(n: 3, groupColumns: ["_measurement", "host"])
