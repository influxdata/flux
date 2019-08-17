package testdata_test

import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,double,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,0,2018-05-22T00:00:00Z,110.46,used_percent,disk
,,0,2018-05-22T00:00:10Z,109.80,used_percent,disk
,,0,2018-05-22T00:00:20Z,110.17,used_percent,disk
,,0,2018-05-22T00:00:30Z,109.82,used_percent,disk
,,0,2018-05-22T00:00:40Z,110.15,used_percent,disk
,,0,2018-05-22T00:00:50Z,109.31,used_percent,disk
,,0,2018-05-22T00:01:00Z,109.05,used_percent,disk
,,0,2018-05-22T00:01:10Z,107.94,used_percent,disk
,,0,2018-05-22T00:01:20Z,107.76,used_percent,disk
,,0,2018-05-22T00:01:30Z,109.24,used_percent,disk
,,0,2018-05-22T00:01:40Z,109.40,used_percent,disk
,,0,2018-05-22T00:01:50Z,108.50,used_percent,disk
,,0,2018-05-22T00:02:00Z,107.96,used_percent,disk
,,0,2018-05-22T00:02:10Z,108.55,used_percent,disk
,,0,2018-05-22T00:02:20Z,108.85,used_percent,disk
,,0,2018-05-22T00:02:30Z,110.44,used_percent,disk
,,0,2018-05-22T00:02:40Z,109.89,used_percent,disk
,,0,2018-05-22T00:02:50Z,110.70,used_percent,disk
,,0,2018-05-22T00:03:00Z,110.79,used_percent,disk
,,0,2018-05-22T00:03:10Z,110.22,used_percent,disk
,,0,2018-05-22T00:03:20Z,110.00,used_percent,disk
,,0,2018-05-22T00:03:30Z,109.27,used_percent,disk
,,0,2018-05-22T00:03:40Z,106.69,used_percent,disk
,,0,2018-05-22T00:03:50Z,107.07,used_percent,disk
,,0,2018-05-22T00:04:00Z,107.92,used_percent,disk
,,0,2018-05-22T00:04:10Z,107.95,used_percent,disk
,,0,2018-05-22T00:04:20Z,107.70,used_percent,disk
,,0,2018-05-22T00:04:30Z,107.97,used_percent,disk
,,0,2018-05-22T00:04:40Z,106.09,used_percent,disk
"

outData = "
#datatype,string,long,dateTime:RFC3339,double,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,0,2018-05-22T00:01:40Z,109.24,used_percent,disk
,,0,2018-05-22T00:01:50Z,109.22,used_percent,disk
,,0,2018-05-22T00:02:00Z,109.12,used_percent,disk
,,0,2018-05-22T00:02:10Z,109.10,used_percent,disk
,,0,2018-05-22T00:02:20Z,109.09,used_percent,disk
,,0,2018-05-22T00:02:30Z,109.12,used_percent,disk
,,0,2018-05-22T00:02:40Z,109.14,used_percent,disk
,,0,2018-05-22T00:02:50Z,109.28,used_percent,disk
,,0,2018-05-22T00:03:00Z,109.44,used_percent,disk
,,0,2018-05-22T00:03:10Z,109.46,used_percent,disk
,,0,2018-05-22T00:03:20Z,109.47,used_percent,disk
,,0,2018-05-22T00:03:30Z,109.46,used_percent,disk
,,0,2018-05-22T00:03:40Z,109.39,used_percent,disk
,,0,2018-05-22T00:03:50Z,109.32,used_percent,disk
,,0,2018-05-22T00:04:00Z,109.29,used_percent,disk
,,0,2018-05-22T00:04:10Z,109.18,used_percent,disk
,,0,2018-05-22T00:04:20Z,109.08,used_percent,disk
,,0,2018-05-22T00:04:30Z,108.95,used_percent,disk
,,0,2018-05-22T00:04:40Z,108.42,used_percent,disk
"

kama = (table=<-) =>
    (table
        |> range(start:2018-05-22T00:00:00Z)
        |> drop(columns: ["_start", "_stop"])
        |> kaufmansAMA(n: 10)
        |> map(fn: (r) => ({r with _value: (math.round(x: r._value*100.0)/100.0)}))
    )

test _kama = () =>
    ({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: kama})