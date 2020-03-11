package aggregate_test

import "experimental"
import "experimental/aggregate"
import "testing"

inData = "
#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,in-data,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,host,interface
,,0,2020-02-20T23:00:00Z,2020-02-20T23:01:00Z,2020-02-20T23:00:00Z,1,bytes_recv,net,host.local,en7
,,0,2020-02-20T23:00:00Z,2020-02-20T23:01:00Z,2020-02-20T23:00:10Z,2,bytes_recv,net,host.local,en7
,,0,2020-02-20T23:00:00Z,2020-02-20T23:01:00Z,2020-02-20T23:00:20Z,3,bytes_recv,net,host.local,en7
,,0,2020-02-20T23:00:00Z,2020-02-20T23:01:00Z,2020-02-20T23:00:30Z,4,bytes_recv,net,host.local,en7
,,0,2020-02-20T23:00:00Z,2020-02-20T23:01:00Z,2020-02-20T23:00:40Z,5,bytes_recv,net,host.local,en7
,,0,2020-02-20T23:00:00Z,2020-02-20T23:01:00Z,2020-02-20T23:00:50Z,6,bytes_recv,net,host.local,en7
,,1,2020-02-20T23:00:00Z,2020-02-20T23:01:00Z,2020-02-20T23:00:00Z,10,bytes_recv,net,host.local,utun2
,,1,2020-02-20T23:00:00Z,2020-02-20T23:01:00Z,2020-02-20T23:00:10Z,20,bytes_recv,net,host.local,utun2
,,1,2020-02-20T23:00:00Z,2020-02-20T23:01:00Z,2020-02-20T23:00:20Z,30,bytes_recv,net,host.local,utun2
,,1,2020-02-20T23:00:00Z,2020-02-20T23:01:00Z,2020-02-20T23:00:30Z,40,bytes_recv,net,host.local,utun2
,,1,2020-02-20T23:00:00Z,2020-02-20T23:01:00Z,2020-02-20T23:00:40Z,50,bytes_recv,net,host.local,utun2
,,1,2020-02-20T23:00:00Z,2020-02-20T23:01:00Z,2020-02-20T23:00:50Z,60,bytes_recv,net,host.local,utun2
"

outData = "
#group,false,false,true,true,true,true,false,false
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,double,dateTime:RFC3339
#default,out-data,,,,,,,
,result,table,_start,_stop,host,interface,_value,_time
,,0,2020-02-20T23:00:00Z,2020-02-20T23:01:00Z,host.local,en7,0.1,2020-02-20T23:00:20Z
,,0,2020-02-20T23:00:00Z,2020-02-20T23:01:00Z,host.local,en7,0.1,2020-02-20T23:00:40Z
,,0,2020-02-20T23:00:00Z,2020-02-20T23:01:00Z,host.local,en7,0.1,2020-02-20T23:01:00Z
,,1,2020-02-20T23:00:00Z,2020-02-20T23:01:00Z,host.local,utun2,1,2020-02-20T23:00:20Z
,,1,2020-02-20T23:00:00Z,2020-02-20T23:01:00Z,host.local,utun2,1,2020-02-20T23:00:40Z
,,1,2020-02-20T23:00:00Z,2020-02-20T23:01:00Z,host.local,utun2,1,2020-02-20T23:01:00Z
"

t_rate = (table=<-) =>
    table
        |> range(start: 2020-02-20T23:00:00Z, stop: 2020-02-20T23:01:00Z)
        |> filter(fn: (r) => r._measurement == "net" and r._field == "bytes_recv")
        |> aggregate.rate(every: 20s, groupColumns: ["host", "interface"], unit: 1s)


test rate = () => ({
        input: testing.loadStorage(csv: inData),
        want: testing.loadMem(csv: outData),
        fn: t_rate
})
