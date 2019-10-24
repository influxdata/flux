package influxql

epoch = 1970-01-01T00:00:00Z
minTime = 1677-09-21T00:12:43.145224194Z
maxTime = 2262-04-11T23:47:16.854775806Z

setTime = (tables=<-,time) =>
	tables |> map(fn: (r) => ({r with time: time}))
