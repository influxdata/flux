from(bucket:"testdb")
  |> range(start: 2018-05-22T19:53:26Z)
  |> stateCount(fn:(r) => r._value > 80)
