package monitor_test

import "influxdata/influxdb/monitor"
import "influxdata/influxdb/v1"
import "testing"
import "experimental"

option now = () => 2018-05-22T19:54:40Z

option monitor.log = (tables=<-) => tables |> drop(columns:["_start", "_stop"])

// Note this input data is identical to the output data of the check test case, post pivot.
inData = "
#datatype,string,long,string,string,string,string,string,dateTime:RFC3339,string,string,string,string,string,string,double
#group,false,false,true,true,true,true,true,false,true,true,true,true,true,true,false
#default,got,,,,,,,,,,,,,,
,result,table,_check_id,_check_name,_level,_measurement,_source_measurement,_time,_type,aaa,bbb,cpu,host,_field,_value
,,0,000000000000000a,cpu threshold check,crit,statuses,cpu,2018-05-22T19:54:20Z,threshold,vaaa,vbbb,cpu-total,host.local,usage_idle,4.800000000000001
,,1,000000000000000a,cpu threshold check,ok,statuses,cpu,2018-05-22T19:54:21Z,threshold,vaaa,vbbb,cpu-total,host.local,usage_idle,90.62382797849732
,,2,000000000000000a,cpu threshold check,warn,statuses,cpu,2018-05-22T19:54:22Z,threshold,vaaa,vbbb,cpu-total,host.local,usage_idle,7.05

#datatype,string,long,string,string,string,string,string,dateTime:RFC3339,string,string,string,string,string,string,string
#group,false,false,true,true,true,true,true,false,true,true,true,true,true,true,false
#default,got,,,,,,,,,,,,,,
,result,table,_check_id,_check_name,_level,_measurement,_source_measurement,_time,_type,aaa,bbb,cpu,host,_field,_value
,,0,000000000000000a,cpu threshold check,crit,statuses,cpu,2018-05-22T19:54:20Z,threshold,vaaa,vbbb,cpu-total,host.local,_message,whoa!
,,1,000000000000000a,cpu threshold check,ok,statuses,cpu,2018-05-22T19:54:21Z,threshold,vaaa,vbbb,cpu-total,host.local,_message,whoa!
,,2,000000000000000a,cpu threshold check,warn,statuses,cpu,2018-05-22T19:54:22Z,threshold,vaaa,vbbb,cpu-total,host.local,_message,whoa!

#datatype,string,long,string,string,string,string,string,dateTime:RFC3339,string,string,string,string,string,string,long
#group,false,false,true,true,true,true,true,false,true,true,true,true,true,true,false
#default,got,,,,,,,,,,,,,,
,result,table,_check_id,_check_name,_level,_measurement,_source_measurement,_time,_type,aaa,bbb,cpu,host,_field,_value
,,0,000000000000000a,cpu threshold check,crit,statuses,cpu,2018-05-22T19:54:20Z,threshold,vaaa,vbbb,cpu-total,host.local,_source_timestamp,1527018840000000000
,,1,000000000000000a,cpu threshold check,ok,statuses,cpu,2018-05-22T19:54:21Z,threshold,vaaa,vbbb,cpu-total,host.local,_source_timestamp,1527018820000000000
,,2,000000000000000a,cpu threshold check,warn,statuses,cpu,2018-05-22T19:54:22Z,threshold,vaaa,vbbb,cpu-total,host.local,_source_timestamp,1527018860000000000
"


outData = "
#datatype,string,long,string,string,string,string,string,string,long,dateTime:RFC3339,string,string,string,string,string,double
#group,false,false,true,true,true,true,false,true,false,false,true,true,true,true,true,false
#default,got,,,,,,,,,,,,,,,
,result,table,_check_id,_check_name,_level,_measurement,_message,_source_measurement,_source_timestamp,_time,_type,aaa,bbb,cpu,host,usage_idle
,,1,000000000000000a,cpu threshold check,ok,statuses,whoa!,cpu,1527018820000000000,2018-05-22T19:54:21Z,threshold,vaaa,vbbb,cpu-total,host.local,90.62382797849732
,,2,000000000000000a,cpu threshold check,warn,statuses,whoa!,cpu,1527018860000000000,2018-05-22T19:54:22Z,threshold,vaaa,vbbb,cpu-total,host.local,7.05
"

t_state_changes_any_to_any = (table=<-) => table
    |> range(start: -1m)
    |> v1.fieldsAsCols()
    |> monitor.stateChanges(
        fromLevel: "any",
        toLevel: "any",
    )
    |> drop(columns: ["_start","_stop"])

test monitor_state_changes_any_to_any = () =>
    ({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_state_changes_any_to_any})
