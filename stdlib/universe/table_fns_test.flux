package universe_test


import "array"
import "csv"
import "testing"

option now = () => 2020-02-22T18:00:00Z

csvdata =
    "
#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,location,state
,,0,2018-04-06T10:49:41.565Z,2020-04-06T11:49:41.564Z,2020-02-22T15:01:00Z,50,bottom_degrees,h2o_temperature,santa_monica,CA
,,0,2018-04-06T10:49:41.565Z,2020-04-06T11:49:41.564Z,2020-02-22T15:31:00Z,49,bottom_degrees,h2o_temperature,santa_monica,CA
,,0,2018-04-06T10:49:41.565Z,2020-04-06T11:49:41.564Z,2020-02-22T16:01:00Z,49,bottom_degrees,h2o_temperature,santa_monica,CA
,,0,2018-04-06T10:49:41.565Z,2020-04-06T11:49:41.564Z,2020-02-22T16:31:00Z,49,bottom_degrees,h2o_temperature,santa_monica,CA
,,0,2018-04-06T10:49:41.565Z,2020-04-06T11:49:41.564Z,2020-02-22T17:01:00Z,48,bottom_degrees,h2o_temperature,santa_monica,CA
,,0,2018-04-06T10:49:41.565Z,2020-04-06T11:49:41.564Z,2020-02-22T17:31:00Z,48,bottom_degrees,h2o_temperature,santa_monica,CA
,,0,2018-04-06T10:49:41.565Z,2020-04-06T11:49:41.564Z,2020-02-22T17:46:00Z,48,bottom_degrees,h2o_temperature,santa_monica,CA
"
data =
    csv.from(csv: csvdata)
        |> range(start: -3h)
col =
    data
        |> tableFind(fn: (key) => true)
        |> getColumn(column: "_value")

testcase table_fns {
    got =
        data
            |> filter(fn: (r) => contains(value: r._value, set: col))
    want = data

    testing.diff(got, want)
}

testcase findColumnWithGroup {
    // The result of grouping and then ungrouping
    // will be for a single table to be spread over
    // four buffers.
    // https://github.com/influxdata/flux/issues/4884
    input =
        array.from(
            rows: [
                {m: "m", k: "north", v: 10, _time: 2020-02-20T00:00:00Z},
                {m: "m", k: "south", v: 20, _time: 2020-02-20T00:00:00Z},
                {m: "m", k: "east", v: 30, _time: 2020-02-20T00:00:00Z},
                {m: "m", k: "west", v: 40, _time: 2020-02-20T00:00:00Z},
            ],
        )
            |> group(columns: ["k"])
            |> range(start: -100y)
            |> group()

    arr = input |> findColumn(fn: (key) => true, column: "v")

    // Verifying the output is tricky here:
    // - The first call to group() will not produce a reliable order of groups
    // - We cannot sort after the second call to group() since doing so would
    //   consolidate the separate buffers into one
    // - Arrays cannot yet be sorted in Flux
    // Therefore, we need to get creative with how we test for the expected output.
    gotObj = {
        len: arr |> length(),
        has10: contains(set: arr, value: 10),
        has20: contains(set: arr, value: 20),
        has30: contains(set: arr, value: 30),
        has40: contains(set: arr, value: 40),
    }
    wantObj = {
        len: 4,
        has10: true,
        has20: true,
        has30: true,
        has40: true,
    }
    got = array.from(rows: [{value: display(v: gotObj)}])
    want = array.from(rows: [{value: display(v: wantObj)}])

    testing.diff(got, want)
}

testcase findColumnKeepEmpty {
    input =
        array.from(rows: [{m: "m", k: "north", v: 10, _time: 2020-02-20T00:00:00Z}])
            |> filter(fn: (r) => r.m == "n", onEmpty: "keep")
    arr = input |> findColumn(fn: (key) => true, column: "v")
    got = array.from(rows: [{value: display(v: arr)}])
    want = array.from(rows: [{value: "[]"}])

    testing.diff(got, want)
}

testcase findColumnDropEmpty {
    input =
        array.from(rows: [{m: "m", k: "north", v: 10, _time: 2020-02-20T00:00:00Z}])
            |> filter(fn: (r) => r.m == "n", onEmpty: "drop")
    arr = input |> findColumn(fn: (key) => true, column: "v")
    got = array.from(rows: [{value: display(v: arr)}])
    want = array.from(rows: [{value: "[]"}])

    testing.diff(got, want)
}
