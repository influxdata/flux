package universe_test


import "testing"
import "csv"

option now = () => 2030-01-01T00:00:00Z

inData =
    "
#datatype,string,long,dateTime:RFC3339,double,string,string,string,string,string,string
#group,false,false,false,false,true,true,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_time,_value,_field,_measurement,device,fstype,host,path
,,0,2018-05-22T00:00:00Z,30,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:10Z,30,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:20Z,30,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:30Z,40,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:40Z,40,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:50Z,40,used_percent,disk,disk1s1,apfs,host.local,/
"
outData =
    "
#group,false,false,true,true,false,false,true,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string,string,string
#default,_result,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,path
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:00:20Z,30,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:00:30Z,35,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:00:40Z,37.5,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:00:50Z,38.75,used_percent,disk,disk1s1,apfs,host.local,/
"

testcase exponential_moving_average {
    got =
        csv.from(csv: inData)
            |> testing.load()
            |> range(start: 2018-05-22T00:00:00Z)
            |> exponentialMovingAverage(n: 3)
    want = csv.from(csv: outData)

    testing.diff(got, want)
}
