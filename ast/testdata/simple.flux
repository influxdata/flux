from(bucket:"testdb")
  |> range(start: 2018-05-23T13:09:22.885021542Z)
  |> filter(fn: (r) => r.user == "user1")
  |> derivative(unit:100ms)