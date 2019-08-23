package monitor_test

import "influxdata/influxdb/monitor"
import "testing"
import "experimental"

option now = () => 2018-05-22T20:00:00Z

inData = "
#datatype,string,long,dateTime:RFC3339,long,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,0,2018-05-22T19:30:00Z,11,A,cpu
,,0,2018-05-22T18:30:00Z,11,A,cpu
,,0,2018-05-22T17:30:00Z,11,A,cpu
,,0,2018-05-22T16:30:00Z,11,A,cpu
,,0,2018-05-22T15:30:00Z,11,A,cpu
,,1,2018-05-22T15:30:00Z,11,B,cpu
,,1,2018-05-22T16:30:00Z,11,B,cpu
,,1,2018-05-22T17:30:00Z,11,B,cpu
,,1,2018-05-22T18:30:00Z,11,B,cpu
,,1,2018-05-22T19:30:00Z,11,B,cpu
,,2,2018-05-22T18:30:00Z,11,C,cpu
,,2,2018-05-22T14:30:00Z,11,C,cpu
,,2,2018-05-22T17:30:00Z,11,C,cpu
,,2,2018-05-22T15:30:00Z,11,C,cpu
,,2,2018-05-22T16:30:00Z,11,C,cpu
,,3,2018-05-22T18:30:00Z,11,D,cpu
,,3,2018-05-22T15:30:00Z,11,D,cpu
,,3,2018-05-22T19:30:00Z,11,D,cpu
,,3,2018-05-22T16:30:00Z,11,D,cpu
,,3,2018-05-22T17:30:00Z,11,D,cpu
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,boolean
#group,false,false,true,true,false,false,true,true,false
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,dead
,,0,2018-05-22T15:00:00Z,2018-05-22T20:00:00Z,2018-05-22T19:30:00Z,11,A,cpu,false
,,1,2018-05-22T15:00:00Z,2018-05-22T20:00:00Z,2018-05-22T19:30:00Z,11,B,cpu,false
,,2,2018-05-22T15:00:00Z,2018-05-22T20:00:00Z,2018-05-22T18:30:00Z,11,C,cpu,true
,,3,2018-05-22T15:00:00Z,2018-05-22T20:00:00Z,2018-05-22T19:30:00Z,11,D,cpu,false
"

t_deadman_add = (table=<-) => table
    |> range(start: -5h)
    |> monitor.deadman(t: experimental.addDuration(d: -1h, to: now()))

test deadman_add = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_deadman_add})
