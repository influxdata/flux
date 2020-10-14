package experimental_test

import "experimental"
import "influxdata/influxdb/v1"
import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,string,string,string,string,double,dateTime:RFC3339
#group,false,false,true,true,true,true,false,false
#default,_result,,,,,,,
,result,table,_field,_measurement,cpu,host,_value,_time
,,0,usage_guest,cpu,cpu-total,ip-192-168-1-16.ec2.internal,0,2020-10-09T22:18:00Z
,,0,usage_guest,cpu,cpu-total,ip-192-168-1-16.ec2.internal,0,2020-10-09T22:19:00Z
,,0,usage_guest,cpu,cpu-total,ip-192-168-1-16.ec2.internal,0,2020-10-09T22:19:44.191958Z
,,1,usage_idle,cpu,cpu-total,ip-192-168-1-16.ec2.internal,94.62634341438049,2020-10-09T22:18:00Z
,,1,usage_idle,cpu,cpu-total,ip-192-168-1-16.ec2.internal,92.28242486302014,2020-10-09T22:19:00Z
,,1,usage_idle,cpu,cpu-total,ip-192-168-1-16.ec2.internal,91.15346397579125,2020-10-09T22:19:44.191958Z
,,2,usage_system,cpu,cpu-total,ip-192-168-1-16.ec2.internal,2.0994751312170705,2020-10-09T22:18:00Z
,,2,usage_system,cpu,cpu-total,ip-192-168-1-16.ec2.internal,2.5586762674700636,2020-10-09T22:19:00Z
,,2,usage_system,cpu,cpu-total,ip-192-168-1-16.ec2.internal,2.6547010580713986,2020-10-09T22:19:44.191958Z

#datatype,string,long,string,string,string,string,string,string,string,double,dateTime:RFC3339
#group,false,false,true,true,true,true,true,true,true,false,false
#default,_result,,,,,,,,,,
,result,table,_field,_measurement,device,fstype,host,mode,path,_value,_time
,,3,inodes_free,disk,disk1s1,apfs,ip-192-168-1-16.ec2.internal,rw,/System/Volumes/Data,4878333294,2020-10-09T22:18:00Z
,,3,inodes_free,disk,disk1s1,apfs,ip-192-168-1-16.ec2.internal,rw,/System/Volumes/Data,4878333286,2020-10-09T22:19:00Z
,,3,inodes_free,disk,disk1s1,apfs,ip-192-168-1-16.ec2.internal,rw,/System/Volumes/Data,4878333253.4,2020-10-09T22:19:44.191958Z
,,4,inodes_total,disk,disk1s1,apfs,ip-192-168-1-16.ec2.internal,rw,/System/Volumes/Data,4882452840,2020-10-09T22:18:00Z
,,4,inodes_total,disk,disk1s1,apfs,ip-192-168-1-16.ec2.internal,rw,/System/Volumes/Data,4882452840,2020-10-09T22:19:00Z
,,4,inodes_total,disk,disk1s1,apfs,ip-192-168-1-16.ec2.internal,rw,/System/Volumes/Data,4882452840,2020-10-09T22:19:44.191958Z
,,5,inodes_used,disk,disk1s1,apfs,ip-192-168-1-16.ec2.internal,rw,/System/Volumes/Data,4119546,2020-10-09T22:18:00Z
,,5,inodes_used,disk,disk1s1,apfs,ip-192-168-1-16.ec2.internal,rw,/System/Volumes/Data,4119554,2020-10-09T22:19:00Z
,,5,inodes_used,disk,disk1s1,apfs,ip-192-168-1-16.ec2.internal,rw,/System/Volumes/Data,4119586.6,2020-10-09T22:19:44.191958Z
"

outData = "
#group,false,false,true,false,false,false,false,false,false,false,false,false
#datatype,string,long,string,string,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,double,double,double,double
#default,want,,,,,,,,,,,
,result,table,host,_measurement,_start,_stop,_time,cpu,inodes_free,usage_guest,usage_idle,usage_system
,,0,ip-192-168-1-16.ec2.internal,cpu,2020-10-01T00:00:00Z,2030-01-01T00:00:00Z,2020-10-09T22:20:00Z,cpu-total,4878333253.4,0,91.15346397579125,2.6547010580713986
"

join_test_fn = (table=<-) => {
    bounded_stream = table |> range(start: 2020-10-01T00:00:00Z)
    a = bounded_stream
        |> filter(fn: (r) => r._measurement == "cpu")
        |> aggregateWindow(fn: last, every: 5m, createEmpty: false)
        |> v1.fieldsAsCols()
        |> group(columns: ["host"])

    b = bounded_stream
        |> filter(fn: (r) => r._measurement == "disk")
        |> aggregateWindow(fn: last, every: 5m, createEmpty: false)
        |> v1.fieldsAsCols()
        |> group(columns: ["host"])

    return experimental.join(left:a, right:b, fn:(left, right) => ({left with inodes_free: right.inodes_free}))
}

test experimental_join = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: join_test_fn})
