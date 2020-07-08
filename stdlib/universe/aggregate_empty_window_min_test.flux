package universe_test
 
import "testing"

inData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#group,false,false,false,false,false,false,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,cpu,host
,,0,2018-05-22T19:53:24.421470485Z,2018-05-22T19:54:24.421470485Z,2018-05-22T19:53:26Z,0,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:24.421470485Z,2018-05-22T19:54:24.421470485Z,2018-05-22T19:53:36Z,0,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:24.421470485Z,2018-05-22T19:54:24.421470485Z,2018-05-22T19:53:46Z,0,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:24.421470485Z,2018-05-22T19:54:24.421470485Z,2018-05-22T19:53:56Z,0,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:24.421470485Z,2018-05-22T19:54:24.421470485Z,2018-05-22T19:54:06Z,0,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:24.421470485Z,2018-05-22T19:54:24.421470485Z,2018-05-22T19:54:16Z,0,usage_guest,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:24.421470485Z,2018-05-22T19:54:24.421470485Z,2018-05-22T19:53:26Z,0,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:24.421470485Z,2018-05-22T19:54:24.421470485Z,2018-05-22T19:53:36Z,0,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:24.421470485Z,2018-05-22T19:54:24.421470485Z,2018-05-22T19:53:46Z,0,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:24.421470485Z,2018-05-22T19:54:24.421470485Z,2018-05-22T19:53:56Z,0,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:24.421470485Z,2018-05-22T19:54:24.421470485Z,2018-05-22T19:54:06Z,0,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:24.421470485Z,2018-05-22T19:54:24.421470485Z,2018-05-22T19:54:16Z,0,usage_guest_nice,cpu,cpu-total,host.local
,,2,2018-05-22T19:53:24.421470485Z,2018-05-22T19:54:24.421470485Z,2018-05-22T19:53:26Z,91.7364670583823,usage_idle,cpu,cpu-total,host.local
,,2,2018-05-22T19:53:24.421470485Z,2018-05-22T19:54:24.421470485Z,2018-05-22T19:53:36Z,89.51118889861233,usage_idle,cpu,cpu-total,host.local
,,2,2018-05-22T19:53:24.421470485Z,2018-05-22T19:54:24.421470485Z,2018-05-22T19:53:46Z,91.0977744436109,usage_idle,cpu,cpu-total,host.local
,,2,2018-05-22T19:53:24.421470485Z,2018-05-22T19:54:24.421470485Z,2018-05-22T19:53:56Z,91.02836436336374,usage_idle,cpu,cpu-total,host.local
,,2,2018-05-22T19:53:24.421470485Z,2018-05-22T19:54:24.421470485Z,2018-05-22T19:54:06Z,68.304576144036,usage_idle,cpu,cpu-total,host.local
,,2,2018-05-22T19:53:24.421470485Z,2018-05-22T19:54:24.421470485Z,2018-05-22T19:54:16Z,87.88598574821853,usage_idle,cpu,cpu-total,host.local
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,string,string,double
#group,false,false,true,true,false,true,true,true,true,false
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_field,_measurement,cpu,host,_value
,,0,2018-05-22T19:53:26Z,2018-05-22T19:55:00Z,2018-05-22T19:53:30Z,usage_guest,cpu,cpu-total,host.local,0
,,0,2018-05-22T19:53:26Z,2018-05-22T19:55:00Z,2018-05-22T19:54:00Z,usage_guest,cpu,cpu-total,host.local,0
,,0,2018-05-22T19:53:26Z,2018-05-22T19:55:00Z,2018-05-22T19:54:30Z,usage_guest,cpu,cpu-total,host.local,0
,,1,2018-05-22T19:53:26Z,2018-05-22T19:55:00Z,2018-05-22T19:53:30Z,usage_guest_nice,cpu,cpu-total,host.local,0
,,1,2018-05-22T19:53:26Z,2018-05-22T19:55:00Z,2018-05-22T19:54:00Z,usage_guest_nice,cpu,cpu-total,host.local,0
,,1,2018-05-22T19:53:26Z,2018-05-22T19:55:00Z,2018-05-22T19:54:30Z,usage_guest_nice,cpu,cpu-total,host.local,0
,,2,2018-05-22T19:53:26Z,2018-05-22T19:55:00Z,2018-05-22T19:53:30Z,usage_idle,cpu,cpu-total,host.local,91.7364670583823
,,2,2018-05-22T19:53:26Z,2018-05-22T19:55:00Z,2018-05-22T19:54:00Z,usage_idle,cpu,cpu-total,host.local,89.51118889861233
,,2,2018-05-22T19:53:26Z,2018-05-22T19:55:00Z,2018-05-22T19:54:30Z,usage_idle,cpu,cpu-total,host.local,68.304576144036
"

test aggregate_window_empty_min = () => ({
    input: testing.loadStorage(csv: inData),
    want: testing.loadMem(csv: outData),
    fn: (table=<-) =>
        table
            |> range(start: 2018-05-22T19:53:26Z, stop: 2018-05-22T19:55:00Z)
            |> aggregateWindow(every: 30s, fn: min),
})

