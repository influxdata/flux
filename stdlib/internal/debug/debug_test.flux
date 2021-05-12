package debug_test


import "array"
import "testing"
import "internal/debug"

inData = "
#datatype,string,long,string,string,dateTime:RFC3339,double
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_measurement,_field,_time,_value
,,0,iZquGj,ei77f8T,2018-12-18T20:52:33Z,-61.68790887989735
,,0,iZquGj,ei77f8T,2018-12-18T20:52:43Z,-6.3173755351186465
,,0,iZquGj,ei77f8T,2018-12-18T20:52:53Z,-26.049728557657513
,,0,iZquGj,ei77f8T,2018-12-18T20:53:03Z,114.285955884979
,,0,iZquGj,ei77f8T,2018-12-18T20:53:13Z,16.140262630578995
,,0,iZquGj,ei77f8T,2018-12-18T20:53:23Z,29.50336437998469

#datatype,string,long,string,string,dateTime:RFC3339,long
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_measurement,_field,_time,_value
,,1,iZquGj,ucyoZ,2018-12-18T20:52:33Z,-66
,,1,iZquGj,ucyoZ,2018-12-18T20:52:43Z,59
,,1,iZquGj,ucyoZ,2018-12-18T20:52:53Z,64
,,1,iZquGj,ucyoZ,2018-12-18T20:53:03Z,84
,,1,iZquGj,ucyoZ,2018-12-18T20:53:13Z,68
,,1,iZquGj,ucyoZ,2018-12-18T20:53:23Z,49
"
input = () => testing.loadStorage(csv: inData)
    |> range(start: 2018-12-18T20:00:00Z, stop: 2018-12-18T21:00:00Z)
    |> drop(columns: ["_start", "_stop"])

testcase slurp {
    got = input() |> debug.slurp()
    want = testing.loadMem(csv: inData)

    testing.diff(got, want) |> yield()
}
testcase sink {
    got = input() |> debug.sink()
    want = testing.loadMem(csv: inData) |> filter(fn: (r) => false)

    testing.diff(got, want) |> yield()
}
