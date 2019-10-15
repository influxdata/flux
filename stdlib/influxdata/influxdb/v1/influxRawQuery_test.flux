package v1_test

import "testing"
import "influxdata/influxdb/v1"

option now = () => 2030-01-01T00:00:00Z

inData = "
#datatype,string,long,dateTime:RFC3339,double,string,string,string
#group,false,false,false,false,true,true,true
#default,_result,,,,,,
,result,table,_time,_value,_field,_measurement,host
,,0,2018-05-22T19:53:26Z,1.83,load1,system,host.local
,,0,2018-05-22T19:53:36Z,1.7,load1,system,host.local
,,0,2018-05-22T19:53:46Z,1.74,load1,system,host.local
,,0,2018-05-22T19:53:56Z,1.63,load1,system,host.local
,,0,2018-05-22T19:54:06Z,1.91,load1,system,host.local
,,0,2018-05-22T19:54:16Z,1.84,load1,system,host.local
,,1,2018-05-22T19:53:26Z,1.98,load15,system,host.local
,,1,2018-05-22T19:53:36Z,1.97,load15,system,host.local
,,1,2018-05-22T19:53:46Z,1.97,load15,system,host.local
,,1,2018-05-22T19:53:56Z,1.96,load15,system,host.local
,,1,2018-05-22T19:54:06Z,1.98,load15,system,host.local
,,1,2018-05-22T19:54:16Z,1.97,load15,system,host.local
,,2,2018-05-22T19:53:26Z,1.95,load5,system,host.local
,,2,2018-05-22T19:53:36Z,1.92,load5,system,host.local
,,2,2018-05-22T19:53:46Z,1.92,load5,system,host.local
,,2,2018-05-22T19:53:56Z,1.89,load5,system,host.local
,,2,2018-05-22T19:54:06Z,1.94,load5,system,host.local
,,2,2018-05-22T19:54:16Z,1.93,load5,system,host.local
"
outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,double,double,double
#group,false,false,true,true,false,true,true,false,false,false
#default,0,,,,,,,,,
,result,table,_start,_stop,_time,_measurement,host,load1,load15,load5
,,0,2018-05-22T19:53:26Z,2018-05-22T19:54:17Z,2018-05-22T19:53:26Z,system,host.local,1.83,1.98,1.95
,,0,2018-05-22T19:53:26Z,2018-05-22T19:54:17Z,2018-05-22T19:53:36Z,system,host.local,1.7,1.97,1.92
,,0,2018-05-22T19:53:26Z,2018-05-22T19:54:17Z,2018-05-22T19:53:46Z,system,host.local,1.74,1.97,1.92
,,0,2018-05-22T19:53:26Z,2018-05-22T19:54:17Z,2018-05-22T19:53:56Z,system,host.local,1.63,1.96,1.89
,,0,2018-05-22T19:53:26Z,2018-05-22T19:54:17Z,2018-05-22T19:54:06Z,system,host.local,1.91,1.98,1.94
,,0,2018-05-22T19:53:26Z,2018-05-22T19:54:17Z,2018-05-22T19:54:16Z,system,host.local,1.84,1.97,1.93
"

// select load1, load15
rawQuery = (stream=<-, start, stop, measurement, fields=[], groupBy=["_time", "_value"], groupMode="except", every=inf, period=0s) =>
  stream
    |> range(start:start, stop: stop)
    |> filter(fn: (r) => r._measurement == measurement and contains(value: r._field, set: fields))
    |> group(columns: groupBy, mode:groupMode)
    |> v1.fieldsAsCols()
    |> window(every: every, period: period)

test influx_raw_query = () => ({
    input: testing.loadStorage(csv: inData),
    want: testing.loadMem(csv: outData),
    fn: (table=<-) => table |> rawQuery(measurement: "system",fields: ["load1", "load15", "load5"], start: 2018-05-22T19:53:26Z, stop: 2018-05-22T19:54:17Z)
})


