package universe_test


import "testing"

inData = "
#datatype,string,long,dateTime:RFC3339,long,string,string,string
#group,false,false,false,false,true,true,true
#default,_result,,,,,,
,result,table,_time,_value,_field,_measurement,k
,,0,1970-01-01T00:00:00Z,0,f,m0,k0
,,0,1970-01-01T00:00:01Z,1,f,m0,k0
,,0,1970-01-01T00:00:02Z,2,f,m0,k0
,,0,1970-01-01T00:00:03Z,3,f,m0,k0
,,0,1970-01-01T00:00:04Z,4,f,m0,k0
,,0,1970-01-01T00:00:05Z,5,f,m0,k0
,,0,1970-01-01T00:00:06Z,6,f,m0,k0
,,0,1970-01-01T00:00:07Z,5,f,m0,k0
,,0,1970-01-01T00:00:08Z,0,f,m0,k0
,,0,1970-01-01T00:00:09Z,6,f,m0,k0
,,0,1970-01-01T00:00:10Z,6,f,m0,k0
,,0,1970-01-01T00:00:11Z,7,f,m0,k0
,,0,1970-01-01T00:00:12Z,5,f,m0,k0
,,0,1970-01-01T00:00:13Z,8,f,m0,k0
,,0,1970-01-01T00:00:14Z,9,f,m0,k0
,,0,1970-01-01T00:00:15Z,5,f,m0,k0
"
outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string
#group,false,false,true,true,false,true,true,true
#default,_result,,,,,,,
,result,table,_start,_stop,_value,_field,_measurement,k
,,0,1970-01-01T00:00:00Z,1970-01-01T00:00:02Z,2,f,m0,k0
,,1,1970-01-01T00:00:02Z,1970-01-01T00:00:07Z,5,f,m0,k0
,,2,1970-01-01T00:00:07Z,1970-01-01T00:00:12Z,5,f,m0,k0
,,3,1970-01-01T00:00:12Z,1970-01-01T00:00:15Z,3,f,m0,k0
"
t_window = (table=<-) => table
    |> range(start: 1970-01-01T00:00:00Z, stop: 1970-01-01T00:00:15Z)
    |> window(every: 5s, offset: 2s)
    |> count()

test _window = () => ({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_window})
