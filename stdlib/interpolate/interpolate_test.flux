package interpolate_test

import "testing"
import "interpolate"

inData = "
#datatype,string,long,dateTime:RFC3339,string,string,double
#group,false,false,false,true,true,false
#default,_result,,,,,
,result,table,_time,_measurement,_field,_value
,,0,2014-01-01T01:00:00Z,_m,FF,0
,,0,2014-01-01T01:02:00Z,_m,FF,2
,,0,2014-01-01T01:04:00Z,_m,FF,4
,,0,2014-01-01T01:06:00Z,_m,FF,6
,,0,2014-01-01T01:08:00Z,_m,FF,8
,,0,2014-01-01T01:10:00Z,_m,FF,10
,,1,2014-01-01T01:01:00Z,_m,QQ,11
,,1,2014-01-01T01:03:00Z,_m,QQ,9
,,1,2014-01-01T01:05:00Z,_m,QQ,7
,,1,2014-01-01T01:07:00Z,_m,QQ,5
,,1,2014-01-01T01:09:00Z,_m,QQ,3
,,1,2014-01-01T01:11:00Z,_m,QQ,1
"

outData = "
#datatype,string,long,dateTime:RFC3339,string,string,double
#group,false,false,false,true,true,false
#default,_result,,,,,
,result,table,_time,_measurement,_field,_value
,,0,2014-01-01T01:00:00Z,_m,FF,0
,,0,2014-01-01T01:01:00Z,_m,FF,1
,,0,2014-01-01T01:02:00Z,_m,FF,2
,,0,2014-01-01T01:03:00Z,_m,FF,3
,,0,2014-01-01T01:04:00Z,_m,FF,4
,,0,2014-01-01T01:05:00Z,_m,FF,5
,,0,2014-01-01T01:06:00Z,_m,FF,6
,,0,2014-01-01T01:07:00Z,_m,FF,7
,,0,2014-01-01T01:08:00Z,_m,FF,8
,,0,2014-01-01T01:09:00Z,_m,FF,9
,,0,2014-01-01T01:10:00Z,_m,FF,10
,,1,2014-01-01T01:01:00Z,_m,QQ,11
,,1,2014-01-01T01:02:00Z,_m,QQ,10
,,1,2014-01-01T01:03:00Z,_m,QQ,9
,,1,2014-01-01T01:04:00Z,_m,QQ,8
,,1,2014-01-01T01:05:00Z,_m,QQ,7
,,1,2014-01-01T01:06:00Z,_m,QQ,6
,,1,2014-01-01T01:07:00Z,_m,QQ,5
,,1,2014-01-01T01:08:00Z,_m,QQ,4
,,1,2014-01-01T01:09:00Z,_m,QQ,3
,,1,2014-01-01T01:10:00Z,_m,QQ,2
,,1,2014-01-01T01:11:00Z,_m,QQ,1
"

interpolateFn = (table=<-) => table
    |> range(start: 2014-01-01T01:00:00Z, stop: 2014-01-01T02:00:00Z)
    |> interpolate.linear(every: 1m)
    |> drop(columns: ["_start", "_stop"])

test interpolate_test = () =>
    ({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: interpolateFn})
