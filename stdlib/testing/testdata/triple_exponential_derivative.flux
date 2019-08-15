package testdata_test

import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,double,string,string,string,string,string,string
#group,false,false,false,false,true,true,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_time,_value,_field,_measurement,device,fstype,host,path
,,0,2018-05-22T00:00:00Z,1,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:10Z,2,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:20Z,3,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:30Z,4,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:40Z,5,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:50Z,6,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:01:00Z,7,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:01:10Z,8,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:01:20Z,9,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:01:30Z,10,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:01:40Z,11,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:01:50Z,12,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:02:00Z,13,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:02:10Z,14,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:02:20Z,15,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:02:30Z,14,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:02:40Z,13,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:02:50Z,12,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:03:00Z,11,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:03:10Z,10,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:03:20Z,9,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:03:30Z,8,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:03:40Z,7,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:03:50Z,6,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:04:00Z,5,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:04:10Z,4,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:04:20Z,3,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:04:30Z,2,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:04:40Z,1,used_percent,disk,disk1s1,apfs,host.local,/
"

outData = "
#group,false,false,true,true,false,false,true,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string,string,string
#default,_result,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,path
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:01:40Z,18.181818181818187,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:01:50Z,15.384615384615374,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:02:00Z,13.33333333333333,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:02:10Z,11.764705882352944,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:02:20Z,10.526315789473696,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:02:30Z,8.304761904761904,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:02:40Z,5.641927541329594,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:02:50Z,3.0392222148231784,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:03:00Z,0.716067574030288,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:03:10Z,-1.2848911076603242,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:03:20Z,-2.9999661985600445,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:03:30Z,-4.493448741755913,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:03:40Z,-5.836238000516913,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:03:50Z,-7.099092024379772,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:04:00Z,-8.352897627933453,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:04:10Z,-9.673028502435233,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:04:20Z,-11.147601363985949,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:04:30Z,-12.891818138458877,used_percent,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T00:00:00Z,2030-01-01T00:00:00.000000000Z,2018-05-22T00:04:40Z,-15.074463280730022,used_percent,disk,disk1s1,apfs,host.local,/
"

triple_exponential_derivative = (table=<-) =>
    (table
       |> range(start:2018-05-22T00:00:00Z)
       |> tripleExponentialDerivative(n:4))

test _triple_exponential_derivative = () =>
    ({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: triple_exponential_derivative})