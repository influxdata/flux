from(bucket: "test")
    |> range(start: 0, stop: 20h)
    |> group(by: ["_tag"])
    |> sum()