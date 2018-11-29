from(bucket:"testdb")
  |> range(start: 2018-05-22T19:53:26Z)
  |> filter(fn: (r) => r._measurement == "diskio" and r._field == "io_time")
  |> group(none: true)
