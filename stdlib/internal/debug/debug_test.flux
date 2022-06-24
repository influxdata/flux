package debug_test


import "array"
import "csv"
import "testing"
import "internal/debug"

inData =
    "
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
input = () =>
    csv.from(csv: inData)
        |> testing.load()
        |> range(start: 2018-12-18T20:00:00Z, stop: 2018-12-18T21:00:00Z)
        |> drop(columns: ["_start", "_stop"])

testcase slurp {
    got = input() |> debug.slurp()
    want = csv.from(csv: inData)

    testing.diff(got, want) |> yield()
}
testcase sink {
    got = input() |> debug.sink()
    want = csv.from(csv: inData) |> filter(fn: (r) => false)

    testing.diff(got, want) |> yield()
}

testcase get_option {
    got = debug.getOption(pkg: "internal/debug", name: "vectorize")
    want = false

    testing.diff(got: array.from(rows: [{v: got}]), want: array.from(rows: [{v: want}]))
        |> yield()
}

testcase get_option2 {
    option debug.vectorize = true

    got = debug.getOption(pkg: "internal/debug", name: "vectorize")
    want = true

    testing.diff(got: array.from(rows: [{v: got}]), want: array.from(rows: [{v: want}]))
        |> yield()
}

testcase get_option_in_map {
    option debug.vectorize = true

    got =
        array.from(rows: [{v: 123}])
            |> map(fn: (r) => ({v: debug.getOption(pkg: "internal/debug", name: "vectorize")}))
    want = true

    testing.diff(got: got, want: array.from(rows: [{v: want}]))
        |> yield()
}
