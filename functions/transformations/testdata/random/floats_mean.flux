from(bucket: "test")
    |> range(start:0, stop:20h)
    |> mean()