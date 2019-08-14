package testdata_test

import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,double,string,string,string,string,string,string
#group,false,false,false,false,true,true,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_time,_value,_field,_measurement,device,fstype,host,path
,,0,2018-05-22T00:00:00Z,20,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:10Z,21,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:20Z,22,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:30Z,23,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:40Z,22,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:50Z,21,used_percent,disk,disk1s1,apfs,host.local,/
"

outData = "
#datatype,string,long,dateTime:RFC3339,double,string,string,string,string,string,string
#group,false,false,false,false,true,true,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_time,_value,_field,_measurement,device,fstype,host,path
,,0,2018-05-22T00:00:30Z,1,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:40Z,0.33333333333333337,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:50Z,0.33333333333333337,used_percent,disk,disk1s1,apfs,host.local,/
"

ker = (table=<-) =>
    (table
        |> range(start:2018-05-22T00:00:00Z)
        |> drop(columns: ["_start", "_stop"])
        |> kaufmansER(n: 3)
    )

test _ker = () =>
    ({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: ker})