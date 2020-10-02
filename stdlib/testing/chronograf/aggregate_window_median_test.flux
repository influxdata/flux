package chronograf_test
 
import "testing"

inData = "
#datatype,string,long,dateTime:RFC3339,string,string,string,string,string,double
#group,false,false,false,true,true,true,true,true,false
#default,_result,,,,,,,,
,result,table,_time,_measurement,_field,device,fstype,host,_value
,,0,2018-05-22T00:00:00Z,disk,percentage,disk1s1,apfs,host.local,67.1
,,0,2018-05-22T00:00:10Z,disk,percentage,disk1s1,apfs,host.local,67.4
,,0,2018-05-22T00:00:20Z,disk,percentage,disk1s1,apfs,host.local,67.5
,,0,2018-05-22T00:00:30Z,disk,percentage,disk1s1,apfs,host.local,67.6
,,0,2018-05-22T00:00:40Z,disk,percentage,disk1s1,apfs,host.local,67.9
,,0,2018-05-22T00:00:50Z,disk,percentage,disk1s1,apfs,host.local,67.9
,,1,2018-05-22T00:00:00Z,disk,percentage,disk2s1,apfs,host.local,92.2
,,1,2018-05-22T00:00:10Z,disk,percentage,disk2s1,apfs,host.local,92.2
,,1,2018-05-22T00:00:20Z,disk,percentage,disk2s1,apfs,host.local,92.2
,,1,2018-05-22T00:00:30Z,disk,percentage,disk2s1,apfs,host.local,92.2
,,1,2018-05-22T00:00:40Z,disk,percentage,disk2s1,apfs,host.local,92.2
,,1,2018-05-22T00:00:50Z,disk,percentage,disk2s1,apfs,host.local,92.2


#datatype,string,long,dateTime:RFC3339,string,string,string,string,string,double
#group,false,false,false,true,true,true,true,true,false
#default,_result,,,,,,,,
,result,table,_time,_measurement,_field,device,fstype,host,_value
,,2,2018-05-22T00:00:00Z,disk,percentage,disk1s1,apfs,host.remote,30
,,2,2018-05-22T00:00:10Z,disk,percentage,disk1s1,apfs,host.remote,30
,,2,2018-05-22T00:00:20Z,disk,percentage,disk1s1,apfs,host.remote,30
,,2,2018-05-22T00:00:30Z,disk,percentage,disk1s1,apfs,host.remote,30
,,2,2018-05-22T00:00:40Z,disk,percentage,disk1s1,apfs,host.remote,30
,,2,2018-05-22T00:00:50Z,disk,percentage,disk1s1,apfs,host.remote,30
,,3,2018-05-22T00:00:00Z,disk,percentage,disk2s1,apfs,host.remote,35
,,3,2018-05-22T00:00:10Z,disk,percentage,disk2s1,apfs,host.remote,35
,,3,2018-05-22T00:00:20Z,disk,percentage,disk2s1,apfs,host.remote,35
,,3,2018-05-22T00:00:30Z,disk,percentage,disk2s1,apfs,host.remote,35
,,3,2018-05-22T00:00:40Z,disk,percentage,disk2s1,apfs,host.remote,35
,,3,2018-05-22T00:00:50Z,disk,percentage,disk2s1,apfs,host.remote,35

#datatype,string,long,dateTime:RFC3339,string,string,string,string,double
#group,false,false,false,true,true,true,true,false
#default,_result,,,,,,,
,result,table,_time,_measurement,_field,device,host,_value
,,0,2018-05-22T00:00:00Z,cpu,percentage,core1,host.local,89.7
,,0,2018-05-22T00:00:10Z,cpu,percentage,core1,host.local,73.4
,,0,2018-05-22T00:00:20Z,cpu,percentage,core1,host.local,88.8
,,0,2018-05-22T00:00:30Z,cpu,percentage,core1,host.local,91.0
,,0,2018-05-22T00:00:40Z,cpu,percentage,core1,host.local,81.1
,,0,2018-05-22T00:00:50Z,cpu,percentage,core1,host.local,87.8
,,1,2018-05-22T00:00:00Z,cpu,percentage,core2,host.local,70.3
,,1,2018-05-22T00:00:10Z,cpu,percentage,core2,host.local,80.4
,,1,2018-05-22T00:00:20Z,cpu,percentage,core2,host.local,95.6
,,1,2018-05-22T00:00:30Z,cpu,percentage,core2,host.local,94.4
,,1,2018-05-22T00:00:40Z,cpu,percentage,core2,host.local,91.2
,,1,2018-05-22T00:00:50Z,cpu,percentage,core2,host.local,90.6


#datatype,string,long,dateTime:RFC3339,string,string,string,double
#group,false,false,false,true,true,true,false
#default,_result,,,,,,
,result,table,_time,_measurement,_field,host,_value
,,0,2018-05-22T00:00:00Z,mem,percentage,host.local,82.5
,,0,2018-05-22T00:00:10Z,mem,percentage,host.local,82.5
,,0,2018-05-22T00:00:20Z,mem,percentage,host.local,82.6
,,0,2018-05-22T00:00:30Z,mem,percentage,host.local,82.6
,,0,2018-05-22T00:00:40Z,mem,percentage,host.local,82.6
,,0,2018-05-22T00:00:50Z,mem,percentage,host.local,82.5
,,1,2018-05-22T00:00:00Z,mem,percentage,host.remote,35
,,1,2018-05-22T00:00:10Z,mem,percentage,host.remote,35
,,1,2018-05-22T00:00:20Z,mem,percentage,host.remote,35
,,1,2018-05-22T00:00:30Z,mem,percentage,host.remote,35
,,1,2018-05-22T00:00:40Z,mem,percentage,host.remote,35
,,1,2018-05-22T00:00:50Z,mem,percentage,host.remote,35
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,string,string,double
#group,false,false,false,false,true,true,true,true,true,false
#default,_result,,,,,,,,,
,result,table,_start,_stop,_field,_measurement,device,fstype,host,_value
,,0,2018-05-22T00:00:00Z,2018-05-22T00:00:30Z,percentage,disk,disk1s1,apfs,host.remote,30
,,0,2018-05-22T00:00:30Z,2018-05-22T00:01:00Z,percentage,disk,disk1s1,apfs,host.remote,30
,,1,2018-05-22T00:00:00Z,2018-05-22T00:00:30Z,percentage,disk,disk2s1,apfs,host.remote,35
,,1,2018-05-22T00:00:30Z,2018-05-22T00:01:00Z,percentage,disk,disk2s1,apfs,host.remote,35


"

agg_window_median_fn = (table=<-) => table
    |> range(start: 2018-05-22T00:00:00Z, stop: 2018-05-22T00:01:00Z)
    |> filter(fn: (r) => r._measurement == "disk" or r._measurement == "mem")
    |> filter(fn: (r) => r.host == "host.remote")
    |> window(period: 30s)
    |> median()
    |> group(columns: ["_value", "_time", "_start", "_stop"], mode: "except")

test agg_window_median = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: agg_window_median_fn})
