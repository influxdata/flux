supl = from(bucketID: "000000000000000a")
  |> range(start: 2018-10-02T17:55:11.520461Z)
  |> filter(fn: (r) => r._measurement == "records" and r.taskID == "02bac3c8f0f37000" )
  |> pivot(rowKey:["_time"], colKey: ["_field"], valueCol: "_value")
  |> group(by: ["runID"])
  |> yield(name:"r1")


main = from(bucketID: "000000000000000a")
  |> range(start: 2018-10-02T17:55:11.520461Z)
  |> filter(fn: (r) => r._measurement == "records" and r.taskID == "02bac3c8f0f37000" )
  |> pivot(rowKey:["_time"], colKey: ["_field"], valueCol: "_value")
  |> pivot(rowKey:["runID"], colKey: ["status"], valueCol: "_time")
  |> yield(name:"r2")

join(tables: {main: main, supl: supl}, on: ["_start", "_stop", "orgID", "taskID", "runID", "_measurement"])
 |> yield(name:"r3")

