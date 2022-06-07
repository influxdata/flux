package sample_test


import "influxdata/influxdb/sample"
import "testing"
import "csv"

option now = () => 2030-01-01T00:00:00Z

inData =
    "
datatype,string,long,dateTime:RFC3339,double,string,string,string
#group,false,false,false,false,true,true,true
#default,_result,,,,,,
,result,table,_time,_value,_field,_measurement,host
,,0,2018-05-22T19:53:26Z,50.12,used_percent,mem,host.local
,,0,2018-05-22T19:53:36Z,51.45,used_percent,mem,host.local
,,0,2018-05-22T19:53:46Z,48.3,used_percent,mem,host.local
,,0,2018-05-22T19:53:56Z,49.34,used_percent,mem,host.local
,,0,2018-05-22T19:54:06Z,49.06,used_percent,mem,host.local
,,0,2018-05-22T19:54:16Z,50.75,used_percent,mem,host.local
"

outData =
    "
#group,false,false,true,true,false,false,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host
,,0,2029-12-31T23:59:10Z,2030-01-01T00:00:44Z,2029-12-31T23:59:10Z,50.12,used_percent,mem,host.local
,,0,2029-12-31T23:59:10Z,2030-01-01T00:00:44Z,2029-12-31T23:59:20Z,51.45,used_percent,mem,host.local
,,0,2029-12-31T23:59:10Z,2030-01-01T00:00:44Z,2029-12-31T23:59:30Z,48.3,used_percent,mem,host.local
,,0,2029-12-31T23:59:10Z,2030-01-01T00:00:44Z,2029-12-31T23:59:40Z,49.34,used_percent,mem,host.local
,,0,2029-12-31T23:59:10Z,2030-01-01T00:00:44Z,2029-12-31T23:59:50Z,49.06,used_percent,mem,host.local
,,0,2029-12-31T23:59:10Z,2030-01-01T00:00:44Z,2030-01-01T00:00:00Z,50.75,used_percent,mem,host.local
"

testcase sample_alignToNow {
    got =
        csv.from(csv: inData)
            |> testing.load()
            |> range(start: 2018-05-22T19:53:26Z, stop: 2018-05-22T19:55:00Z)
            |> sample.alignToNow()
    want = csv.from(csv: outData)

    testing.diff(got, want)
}
