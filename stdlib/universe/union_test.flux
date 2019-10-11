package universe_test

import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,double,string,string,string,string
#group,false,false,false,false,true,true,true,true
#default,_result,,,,,,,
,result,table,_time,_value,_field,_measurement,cpu,host
,,0,2018-05-22T19:53:26Z,0,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:36Z,0,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:46Z,0,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:56Z,0,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:54:06Z,0,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:54:16Z,0,usage_guest,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:26Z,0,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:36Z,0,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:46Z,0,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:56Z,0,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:54:06Z,0,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:54:16Z,0,usage_guest_nice,cpu,cpu-total,host.local
,,2,2018-05-22T19:53:26Z,91.7364670583823,usage_idle,cpu,cpu-total,host.local
,,2,2018-05-22T19:53:36Z,89.51118889861233,usage_idle,cpu,cpu-total,host.local
,,2,2018-05-22T19:53:46Z,91.0977744436109,usage_idle,cpu,cpu-total,host.local
,,2,2018-05-22T19:53:56Z,91.02836436336374,usage_idle,cpu,cpu-total,host.local
,,2,2018-05-22T19:54:06Z,68.304576144036,usage_idle,cpu,cpu-total,host.local
,,2,2018-05-22T19:54:16Z,87.88598574821853,usage_idle,cpu,cpu-total,host.local
"

outData = "
#datatype,string,long,dateTime:RFC3339,double,string,string,string,string
#group,false,false,false,false,true,true,true,true
#default,_result,,,,,,,
,result,table,_time,_value,_field,_measurement,cpu,host
,,0,2018-05-22T19:53:26Z,0,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:36Z,0,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:46Z,0,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:56Z,0,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:54:06Z,0,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:54:16Z,0,usage_guest,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:26Z,0,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:36Z,0,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:46Z,0,usage_guest_nice,cpu,cpu-total,host.local
,,2,2018-05-22T19:53:56Z,91.02836436336374,usage_idle,cpu,cpu-total,host.local
,,2,2018-05-22T19:54:06Z,68.304576144036,usage_idle,cpu,cpu-total,host.local
,,2,2018-05-22T19:54:16Z,87.88598574821853,usage_idle,cpu,cpu-total,host.local
"

t_union = (table=<-) => {
    t1 = table
        |> range(start: 2018-05-22T19:53:00Z, stop: 2018-05-22T19:53:50Z)
        |> filter(fn: (r) => r._field == "usage_guest" or r._field == "usage_guest_nice")
        |> drop(columns: ["_start", "_stop"])

    t2 = table
        |> range(start: 2018-05-22T19:53:50Z, stop: 2018-05-22T19:54:20Z)
        |> filter(fn: (r) => r._field == "usage_guest" or r._field == "usage_idle")
        |> drop(columns: ["_start", "_stop"])

    return union(tables: [t1, t2]) |> sort(columns: ["_time"])
}

test _union = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_union})
