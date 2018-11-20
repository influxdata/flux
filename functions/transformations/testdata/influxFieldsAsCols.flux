from(bucket:"testdb")
  |> range(start: 2018-05-22T19:53:26Z, stop: 2018-05-22T19:54:17Z)
  |> influxFieldsAsCols()
  |> yield(name:"0")