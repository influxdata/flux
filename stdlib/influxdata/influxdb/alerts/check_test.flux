package alerts_test

import "influxdata/influxdb/alerts"
import "influxdata/influxdb/v1"
import "testing"

option now = () => 2018-05-22T19:54:20Z

inData = "
#datatype,string,long,dateTime:RFC3339,double,string,string,string,string
#group,false,false,false,false,true,true,true,true
#default,_result,,,,,,,
,result,table,_time,_value,_field,_measurement,cpu,host
,,0,2018-05-22T19:53:26Z,91.7364670583823,usage_idle,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:36Z,89.51118889861233,usage_idle,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:46Z,4.9,usage_idle,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:56Z,4.7,usage_idle,cpu,cpu-total,host.local
,,0,2018-05-22T19:54:06Z,7.0,usage_idle,cpu,cpu-total,host.local
,,0,2018-05-22T19:54:16Z,7.1,usage_idle,cpu,cpu-total,host.local
"

outData = "
#datatype,string,long,string,string,string,string,string,string,dateTime:RFC3339,dateTime:RFC3339,string,string,string,string,string,double
#group,false,false,true,true,true,true,false,true,false,false,true,true,true,true,true,false
#default,got,,,,,,,,,,,,,,,
,result,table,_check_id,_check_name,_level,_measurement,_message,_source_measurement,_source_timestamp,_time,_type,aaa,bbb,cpu,host,usage_idle
,,0,000000000000000a,cpu threshold check,crit,statuses,whoa!,cpu,2018-05-22T19:54:00Z,2018-05-22T19:54:20Z,threshold,vaaa,vbbb,cpu-total,host.local,4.800000000000001
,,1,000000000000000a,cpu threshold check,ok,statuses,whoa!,cpu,2018-05-22T19:53:40Z,2018-05-22T19:54:20Z,threshold,vaaa,vbbb,cpu-total,host.local,90.62382797849732
,,2,000000000000000a,cpu threshold check,warn,statuses,whoa!,cpu,2018-05-22T19:54:20Z,2018-05-22T19:54:20Z,threshold,vaaa,vbbb,cpu-total,host.local,7.05
"

data = {
    _check_id: "000000000000000a",
    _check_name: "cpu threshold check",
    _type: "threshold",
    tags: {aaa: "vaaa", bbb: "vbbb"}
}

crit = (r) => (r.usage_idle < 5.0)
warn = (r) => (r.usage_idle < 10.0)
info = (r) => (r.usage_idle < 25.0)

messageFn = (r) => "whoa!"

t_check = (table=<-) => table
    |> range(start: -1m)
    |> filter(fn: (r) => r._measurement == "cpu")
    |> filter(fn: (r) => r._field == "usage_idle")
    |> filter(fn: (r) => r.cpu == "cpu-total")
    |> v1.fieldsAsCols() // pivot data so there is a "usage_idle" column
    |> aggregateWindow(every: 20s, fn: mean, column: "usage_idle")
    |> alerts.check(
        data: data,
        messageFn: messageFn,
        info: info,
        warn: warn,
        crit: crit,
    )

test alerts_check = () =>
    ({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_check})
