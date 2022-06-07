package universe_test


import "testing"
import "csv"

inData =
    "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,dateTime:RFC3339,double
#group,false,false,true,true,true,true,false,false
#default,_result,,,,,,,
,result,table,_start,_stop,_field,_measurement,_time,_value
,,0,2015-08-22T22:12:00.000000000Z,2015-08-28T03:01:00.000000000Z,water_level,water,2015-08-22T22:12:00.000000000Z,4.948
"
outData = "
#datatype,string,long
#group,false,false
#default,_result,
,result,table
,,0
"

testcase holt_winters_panic {
    got =
        csv.from(csv: inData)
            |> testing.load()
            |> range(start: 2015-08-22T22:12:00Z, stop: 2015-08-28T03:00:00Z)
            |> window(every: 379m, offset: 348m)
            |> first()
            // InfluxQL associates the value of the beginning of the window
            // to the result of the function applied to it.
            // So, we overwrite "_time" with "_start" in order to make timestamps
            // of the starting dataset to match between InfluxQL and Flux.
            |> duplicate(column: "_start", as: "_time")
            |> window(every: inf)
            |> holtWinters(n: 10, seasonality: 4, interval: 379m)
            |> keep(columns: ["_time", "_value"])
    want = csv.from(csv: outData)

    testing.diff(got, want)
}
