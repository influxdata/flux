from(bucket:"testdb")
    |> range(start: 2018-05-22T19:53:00Z)
    |> histogram(bins:[0.0,1.0,2.0])
