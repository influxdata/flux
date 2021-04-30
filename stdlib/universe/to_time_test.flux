package universe_test


import "testing"

option now = () => 2030-01-01T00:00:00Z

inData = "
#datatype,string,long,dateTime:RFC3339,long,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,0,2018-05-22T19:53:00Z,0,k0,m
,,0,2018-05-22T19:53:01Z,1527018806033000000,k0,m
,,0,2018-05-22T19:53:02Z,1527018806033066000,k0,m
,,0,2018-05-22T19:53:03Z,1527012000000000000,k0,m
#datatype,string,long,dateTime:RFC3339,string,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,1,2018-05-22T19:53:10Z,2018-05-22T19:53:26Z,k1,m
,,1,2018-05-22T19:53:11Z,2018-05-22T19:53:26.033Z,k1,m
,,1,2018-05-22T19:53:12Z,2018-05-22T19:53:26.033066Z,k1,m
,,1,2018-05-22T19:53:13Z,2018-05-22T20:00:00+01:00,k1,m
,,1,2018-05-22T19:53:14Z,2018-05-22T20:00:00.000+01:00,k1,m
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,2,2018-05-22T19:53:20Z,1970-01-01T00:00:00Z,k2,m
,,2,2018-05-22T19:53:21Z,2600-07-21T23:34:33.709551615Z,k2,m
,,2,2018-05-22T19:53:22Z,2018-05-22T19:53:26.033Z,k2,m
,,2,2018-05-22T19:53:23Z,2018-05-22T20:00:00+01:00,k2,m
,,2,2018-05-22T19:53:24Z,2018-05-22T20:00:00.000-01:00,k2,m
#datatype,string,long,dateTime:RFC3339,unsignedLong,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,3,2018-05-22T19:53:30Z,0,k3,m
,,3,2018-05-22T19:53:31Z,18446744073709551615,k3,m
,,3,2018-05-22T19:53:32Z,1527018806033066000,k3,m
,,3,2018-05-22T19:53:33Z,1527012000000000000,k3,m
"
outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,dateTime:RFC3339
#group,false,false,true,true,false,true,true,false
#default,want,,,,,,,
,result,table,_start,_stop,_time,_field,_measurement,_value
,,0,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:00Z,k0,m,1970-01-01T00:00:00Z
,,0,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:01Z,k0,m,2018-05-22T19:53:26.033Z
,,0,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:02Z,k0,m,2018-05-22T19:53:26.033066Z
,,0,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:03Z,k0,m,2018-05-22T18:00:00Z
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:10Z,k1,m,2018-05-22T19:53:26Z
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:11Z,k1,m,2018-05-22T19:53:26.033Z
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:12Z,k1,m,2018-05-22T19:53:26.033066Z
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:13Z,k1,m,2018-05-22T19:00:00Z
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:14Z,k1,m,2018-05-22T19:00:00Z
,,2,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:20Z,k2,m,1970-01-01T00:00:00Z
,,2,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:21Z,k2,m,2600-07-21T23:34:33.709551615Z
,,2,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:22Z,k2,m,2018-05-22T19:53:26.033Z
,,2,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:23Z,k2,m,2018-05-22T19:00:00Z
,,2,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:24Z,k2,m,2018-05-22T21:00:00Z
,,3,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:30Z,k3,m,1970-01-01T00:00:00Z
,,3,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:31Z,k3,m,2554-07-21T23:34:33.709551615Z
,,3,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:32Z,k3,m,2018-05-22T19:53:26.033066Z
,,3,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:33Z,k3,m,2018-05-22T18:00:00Z
"
t_to_time = (table=<-) => table
    |> range(start: 2018-05-22T19:52:00Z)
    |> toTime()

test _to = () => ({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_to_time})
