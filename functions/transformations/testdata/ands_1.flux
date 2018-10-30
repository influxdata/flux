from(bucket: "test")
    |> range(start: 0, stop: 19)
    |> filter(fn: (r) => r._measurement == "ctr"
        AND r._field == "n"
        AND r._value > 11
        AND r._value < 18)