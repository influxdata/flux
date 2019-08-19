package http_test

import "testing"
import "http"
import "json"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,double,string,string,string,string,string,string
#group,false,false,false,false,true,true,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_time,_value,_field,_measurement,device,fstype,host,path
,,0,2018-05-22T00:00:00Z,1,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:10Z,2,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:20Z,3,used_percent,disk,disk1s1,apfs,host.local,/
"

outData = "
#datatype,string,long,dateTime:RFC3339,double,string,string,string,string,string,string,string
#group,false,false,false,false,true,true,true,true,true,true,true
#default,_result,,,,,,,,,,
,result,table,_time,_value,_field,_measurement,device,fstype,host,path,_sent
,,0,2018-05-22T00:00:00Z,1,used_percent,disk,disk1s1,apfs,host.local,/,true
,,0,2018-05-22T00:00:10Z,2,used_percent,disk,disk1s1,apfs,host.local,/,true
,,0,2018-05-22T00:00:20Z,3,used_percent,disk,disk1s1,apfs,host.local,/,true
"

endpoint = http.endpoint(url:"http://localhost:7777")

post = (table=<-) =>
    table
        |> range(start:2018-05-22T00:00:00Z)
        |> drop(columns: ["_start", "_stop"])
        |> endpoint(mapFn:(r) => ({data: json.encode(v: r)}))()

test _post = () =>
    ({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: post})
