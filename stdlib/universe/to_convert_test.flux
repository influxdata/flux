package universe_test


import "testing"
import "csv"

option now = () => 2030-01-01T00:00:00Z

testcase to_float {
    inData =
        "
#datatype,string,long,dateTime:RFC3339,boolean,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,0,2018-05-22T19:53:00Z,true,k0,m
,,0,2018-05-22T19:53:01Z,false,k0,m
#datatype,string,long,dateTime:RFC3339,double,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,1,2018-05-22T19:53:10Z,0,k1,m
,,1,2018-05-22T19:53:11Z,1,k1,m
,,1,2018-05-22T19:53:12Z,1.0,k1,m
,,1,2018-05-22T19:53:13Z,1.1e5,k1,m
,,1,2018-05-22T19:53:14Z,1.1e-5,k1,m
,,1,2018-05-22T19:53:15Z,0.0000024,k1,m
,,1,2018-05-22T19:53:16Z,-23.123456,k1,m
,,1,2018-05-22T19:53:17Z,-922337203.6854775808,k1,m
#datatype,string,long,dateTime:RFC3339,long,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,2,2018-05-22T19:53:20Z,1,k2,m
,,2,2018-05-22T19:53:21Z,-1,k2,m
,,2,2018-05-22T19:53:22Z,0,k2,m
,,2,2018-05-22T19:53:23Z,2147483647,k2,m
,,2,2018-05-22T19:53:24Z,-2147483648,k2,m
,,2,2018-05-22T19:53:25Z,9223372036854775807,k2,m
,,2,2018-05-22T19:53:26Z,-9223372036854775808,k2,m
#datatype,string,long,dateTime:RFC3339,string,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,3,2018-05-22T19:53:30Z,0,k3,m
,,3,2018-05-22T19:53:31Z,1,k3,m
,,3,2018-05-22T19:53:32Z,1.0,k3,m
,,3,2018-05-22T19:53:33Z,1.1e5,k3,m
,,3,2018-05-22T19:53:34Z,1.1e-5,k3,m
,,3,2018-05-22T19:53:35Z,0.0000024,k3,m
,,3,2018-05-22T19:53:36Z,-23.123456,k3,m
,,3,2018-05-22T19:53:37Z,-922337203.6854775808,k3,m
,,3,2018-05-22T19:53:37Z,+Inf,k3,m
,,3,2018-05-22T19:53:37Z,-Inf,k3,m
#datatype,string,long,dateTime:RFC3339,unsignedLong,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,4,2018-05-22T19:53:40Z,0,k4,m
,,4,2018-05-22T19:53:41Z,1,k4,m
,,4,2018-05-22T19:53:42Z,18446744073709551615,k4,m
,,4,2018-05-22T19:53:43Z,4294967296,k4,m
,,4,2018-05-22T19:53:44Z,9223372036854775808,k4,m
"
    outData =
        "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,double
#group,false,false,true,true,false,true,true,false
#default,want,,,,,,,
,result,table,_start,_stop,_time,_field,_measurement,_value
,,0,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:00Z,k0,m,1.0
,,0,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:01Z,k0`,m,0.0
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:10Z,k1,m,0
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:11Z,k1,m,1
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:12Z,k1,m,1.0
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:13Z,k1,m,110000
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:14Z,k1,m,0.000011
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:15Z,k1,m,0.0000024
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:16Z,k1,m,-23.123456
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:17Z,k1,m,-922337203.6854775808
,,2,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:20Z,k2,m,1
,,2,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:21Z,k2,m,-1
,,2,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:22Z,k2,m,0
,,2,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:23Z,k2,m,2147483647
,,2,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:24Z,k2,m,-2147483648
,,2,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:25Z,k2,m,9223372036854775807
,,2,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:26Z,k2,m,-9223372036854775808
,,3,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:30Z,k3,m,0
,,3,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:31Z,k3,m,1
,,3,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:32Z,k3,m,1.0
,,3,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:33Z,k3,m,110000
,,3,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:34Z,k3,m,0.000011
,,3,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:35Z,k3,m,0.0000024
,,3,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:36Z,k3,m,-23.123456
,,3,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:37Z,k3,m,-922337203.6854775808
,,3,2018-05-22T19:53:37Z,2030-01-01T00:00:00Z,2018-05-22T19:53:37Z,k3,m,+Inf
,,3,2018-05-22T19:53:37Z,2030-01-01T00:00:00Z,2018-05-22T19:53:37Z,k3,m,-Inf
,,4,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:40Z,k4,m,0
,,4,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:41Z,k4,m,1
,,4,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:42Z,k4,m,18446744073709551615
,,4,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:43Z,k4,m,4294967296
,,4,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:44Z,k4,m,9223372036854775808
"
    got =
        csv.from(csv: inData)
            |> range(start: 2018-05-22T19:52:00Z)
            |> toFloat()
    want = csv.from(csv: outData)

    testing.diff(got: got, want: want)
}
testcase to_bool {
    inData =
        "
#datatype,string,long,dateTime:RFC3339,boolean,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,0,2018-05-22T19:53:00Z,true,k0,m
,,0,2018-05-22T19:53:01Z,false,k0,m
#datatype,string,long,dateTime:RFC3339,double,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,1,2018-05-22T19:53:10Z,1e0,k1,m
,,1,2018-05-22T19:53:11Z,1.00,k1,m
,,1,2018-05-22T19:53:12Z,1,k1,m
,,1,2018-05-22T19:53:13Z,1.0,k1,m
,,1,2018-05-22T19:53:14Z,0.0,k1,m
,,1,2018-05-22T19:53:15Z,0.00,k1,m
,,1,2018-05-22T19:53:16Z,0,k1,m
#datatype,string,long,dateTime:RFC3339,long,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,2,2018-05-22T19:53:20Z,1,k2,m
,,2,2018-05-22T19:53:21Z,0,k2,m
#datatype,string,long,dateTime:RFC3339,string,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,3,2018-05-22T19:53:30Z,true,k3,m
,,3,2018-05-22T19:53:31Z,false,k3,m
#datatype,string,long,dateTime:RFC3339,unsignedLong,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,4,2018-05-22T19:53:40Z,1,k4,m
,,4,2018-05-22T19:53:41Z,0,k4,m
"
    outData =
        "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,boolean
#group,false,false,true,true,false,true,true,false
#default,want,,,,,,,
,result,table,_start,_stop,_time,_field,_measurement,_value
,,0,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:00Z,k0,m,true
,,0,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:01Z,k0,m,false
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:10Z,k1,m,true
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:11Z,k1,m,true
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:12Z,k1,m,true
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:13Z,k1,m,true
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:14Z,k1,m,false
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:15Z,k1,m,false
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:16Z,k1,m,false
,,2,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:20Z,k2,m,true
,,2,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:21Z,k2,m,false
,,3,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:30Z,k3,m,true
,,3,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:31Z,k3,m,false
,,4,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:40Z,k4,m,true
,,4,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:41Z,k4,m,false
"
    got =
        csv.from(csv: inData)
            |> range(start: 2018-05-22T19:52:00Z)
            |> toBool()
    want = csv.from(csv: outData)

    testing.diff(got: got, want: want)
}
testcase to_string {
    inData =
        "
#datatype,string,long,dateTime:RFC3339,boolean,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,0,2018-05-22T19:53:00Z,true,k0,m
,,0,2018-05-22T19:53:01Z,false,k0,m
#datatype,string,long,dateTime:RFC3339,double,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,1,2018-05-22T19:53:10Z,0,k1,m
,,1,2018-05-22T19:53:11Z,1,k1,m
,,1,2018-05-22T19:53:12Z,1.0,k1,m
,,1,2018-05-22T19:53:13Z,110000,k1,m
,,1,2018-05-22T19:53:14Z,0.000011,k1,m
,,1,2018-05-22T19:53:15Z,0.0000024,k1,m
,,1,2018-05-22T19:53:16Z,-23.123456,k1,m
,,1,2018-05-22T19:53:17Z,-922337203.6854775808,k1,m
#datatype,string,long,dateTime:RFC3339,long,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,2,2018-05-22T19:53:20Z,1,k2,m
,,2,2018-05-22T19:53:21Z,-1,k2,m
,,2,2018-05-22T19:53:22Z,0,k2,m
,,2,2018-05-22T19:53:23Z,2147483647,k2,m
,,2,2018-05-22T19:53:24Z,-2147483648,k2,m
,,2,2018-05-22T19:53:25Z,9223372036854775807,k2,m
,,2,2018-05-22T19:53:26Z,-9223372036854775808,k2,m
#datatype,string,long,dateTime:RFC3339,string,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,3,2018-05-22T19:53:30Z,one,k3,m
,,3,2018-05-22T19:53:31Z,one's,k3,m
,,3,2018-05-22T19:53:32Z,one_two,k3,m
,,3,2018-05-22T19:53:33Z,one two,k3,m
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,4,2018-05-22T19:53:40Z,2018-05-22T19:53:26Z,k4,m
,,4,2018-05-22T19:53:41Z,2018-05-22T19:53:26.033Z,k4,m
,,4,2018-05-22T19:53:42Z,2018-05-22T19:53:26.033066Z,k4,m
,,4,2018-05-22T19:53:43Z,2018-05-22T19:00:00+01:00,k4,m
#datatype,string,long,dateTime:RFC3339,unsignedLong,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,5,2018-05-22T19:53:50Z,0,k5,m
,,5,2018-05-22T19:53:51Z,1,k5,m
,,5,2018-05-22T19:53:52Z,18446744073709551615,k5,m
,,5,2018-05-22T19:53:53Z,4294967296,k5,m
,,5,2018-05-22T19:53:54Z,9223372036854775808,k5,m
"
    outData =
        "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,string
#group,false,false,true,true,false,true,true,false
#default,want,,,,,,,
,result,table,_start,_stop,_time,_field,_measurement,_value
,,0,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:00Z,k0,m,true
,,0,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:01Z,k0,m,false
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:10Z,k1,m,0
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:11Z,k1,m,1
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:12Z,k1,m,1
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:13Z,k1,m,110000
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:14Z,k1,m,0.000011
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:15Z,k1,m,0.0000024
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:16Z,k1,m,-23.123456
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:17Z,k1,m,-922337203.6854776
,,2,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:20Z,k2,m,1
,,2,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:21Z,k2,m,-1
,,2,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:22Z,k2,m,0
,,2,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:23Z,k2,m,2147483647
,,2,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:24Z,k2,m,-2147483648
,,2,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:25Z,k2,m,9223372036854775807
,,2,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:26Z,k2,m,-9223372036854775808
,,3,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:30Z,k3,m,one
,,3,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:31Z,k3,m,one's
,,3,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:32Z,k3,m,one_two
,,3,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:33Z,k3,m,one two
,,4,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:40Z,k4,m,2018-05-22T19:53:26.000000000Z
,,4,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:41Z,k4,m,2018-05-22T19:53:26.033000000Z
,,4,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:42Z,k4,m,2018-05-22T19:53:26.033066000Z
,,4,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:43Z,k4,m,2018-05-22T18:00:00.000000000Z
,,5,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:50Z,k5,m,0
,,5,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:51Z,k5,m,1
,,5,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:52Z,k5,m,18446744073709551615
,,5,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:53Z,k5,m,4294967296
,,5,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:54Z,k5,m,9223372036854775808
"
    got =
        csv.from(csv: inData)
            |> range(start: 2018-05-22T19:52:00Z)
            |> toString()
    want = csv.from(csv: outData)

    testing.diff(got: got, want: want)
}
testcase to_time {
    inData =
        "
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
    outData =
        "
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
    got =
        csv.from(csv: inData)
            |> range(start: 2018-05-22T19:52:00Z)
            |> toTime()
    want = csv.from(csv: outData)

    testing.diff(got: got, want: want)
}
testcase to_int {
    inData =
        "
#datatype,string,long,dateTime:RFC3339,boolean,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,0,2018-05-22T19:53:00Z,true,k0,m
,,0,2018-05-22T19:53:01Z,false,k0,m
#datatype,string,long,dateTime:RFC3339,double,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,1,2018-05-22T19:53:10Z,0.0,k1,m
,,1,2018-05-22T19:53:11Z,1.1,k1,m
,,1,2018-05-22T19:53:12Z,1.8,k1,m
,,1,2018-05-22T19:53:13Z,110000.00011,k1,m
,,1,2018-05-22T19:53:14Z,0.000011,k1,m
,,1,2018-05-22T19:53:15Z,2036854775807.123,k1,m
,,1,2018-05-22T19:53:16Z,-23.123456,k1,m
,,1,2018-05-22T19:53:17Z,-922337203.6854775808,k1,m
#datatype,string,long,dateTime:RFC3339,long,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,2,2018-05-22T19:53:20Z,1,k2,m
,,2,2018-05-22T19:53:21Z,-1,k2,m
,,2,2018-05-22T19:53:22Z,0,k2,m
,,2,2018-05-22T19:53:23Z,2147483647,k2,m
,,2,2018-05-22T19:53:24Z,-2147483648,k2,m
,,2,2018-05-22T19:53:25Z,9223372036854775807,k2,m
,,2,2018-05-22T19:53:26Z,-9223372036854775808,k2,m
#datatype,string,long,dateTime:RFC3339,string,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,3,2018-05-22T19:53:30Z,1,k3,m
,,3,2018-05-22T19:53:31Z,-1,k3,m
,,3,2018-05-22T19:53:32Z,0,k3,m
,,3,2018-05-22T19:53:33Z,2147483647,k3,m
,,3,2018-05-22T19:53:34Z,-2147483648,k3,m
,,3,2018-05-22T19:53:35Z,9223372036854775807,k3,m
,,3,2018-05-22T19:53:36Z,-9223372036854775808,k3,m
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,4,2018-05-22T19:53:40Z,1970-01-01T00:00:00Z,k4,m
,,4,2018-05-22T19:53:41Z,2018-05-22T19:53:26.033Z,k4,m
,,4,2018-05-22T19:53:42Z,2018-05-22T19:53:26.033066Z,k4,m
,,4,2018-05-22T19:53:43Z,2018-05-22T19:00:00+01:00,k4,m
#datatype,string,long,dateTime:RFC3339,unsignedLong,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,5,2018-05-22T19:53:50Z,0,k5,m
,,5,2018-05-22T19:53:51Z,1,k5,m
,,5,2018-05-22T19:53:52Z,18446744073709551615,k5,m
,,5,2018-05-22T19:53:53Z,4294967296,k5,m
,,5,2018-05-22T19:53:54Z,9223372036854775807,k5,m
"
    outData =
        "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,long
#group,false,false,true,true,false,true,true,false
#default,want,,,,,,,
,result,table,_start,_stop,_time,_field,_measurement,_value
,,0,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:00Z,k0,m,1
,,0,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:01Z,k0,m,0
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:10Z,k1,m,0
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:11Z,k1,m,1
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:12Z,k1,m,1
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:13Z,k1,m,110000
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:14Z,k1,m,0
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:15Z,k1,m,2036854775807
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:16Z,k1,m,-23
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:17Z,k1,m,-922337203
,,2,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:20Z,k2,m,1
,,2,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:21Z,k2,m,-1
,,2,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:22Z,k2,m,0
,,2,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:23Z,k2,m,2147483647
,,2,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:24Z,k2,m,-2147483648
,,2,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:25Z,k2,m,9223372036854775807
,,2,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:26Z,k2,m,-9223372036854775808
,,3,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:30Z,k3,m,1
,,3,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:31Z,k3,m,-1
,,3,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:32Z,k3,m,0
,,3,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:33Z,k3,m,2147483647
,,3,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:34Z,k3,m,-2147483648
,,3,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:35Z,k3,m,9223372036854775807
,,3,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:36Z,k3,m,-9223372036854775808
,,4,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:40Z,k4,m,0
,,4,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:41Z,k4,m,1527018806033000000
,,4,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:42Z,k4,m,1527018806033066000
,,4,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:43Z,k4,m,1527012000000000000
,,5,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:50Z,k5,m,0
,,5,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:51Z,k5,m,1
,,5,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:52Z,k5,m,-1
,,5,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:53Z,k5,m,4294967296
,,5,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:54Z,k5,m,9223372036854775807
"
    got =
        csv.from(csv: inData)
            |> range(start: 2018-05-22T19:52:00Z)
            |> toInt()
    want = csv.from(csv: outData)

    testing.diff(got: got, want: want)
}
testcase to_uint {
    inData =
        "
#datatype,string,long,dateTime:RFC3339,boolean,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,0,2018-05-22T19:53:00Z,true,k0,m
,,0,2018-05-22T19:53:01Z,false,k0,m
#datatype,string,long,dateTime:RFC3339,double,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,1,2018-05-22T19:53:10Z,0.0,k1,m
,,1,2018-05-22T19:53:11Z,1.1,k1,m
,,1,2018-05-22T19:53:12Z,1.8,k1,m
,,1,2018-05-22T19:53:13Z,110000.00011,k1,m
,,1,2018-05-22T19:53:14Z,0.000011,k1,m
,,1,2018-05-22T19:53:15Z,2036854775807.123,k1,m
#datatype,string,long,dateTime:RFC3339,long,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,2,2018-05-22T19:53:20Z,0,k2,m
,,2,2018-05-22T19:53:21Z,1,k2,m
,,2,2018-05-22T19:53:22Z,-1,k2,m
,,2,2018-05-22T19:53:23Z,4294967296,k2,m
,,2,2018-05-22T19:53:24Z,-9223372036854775808,k2,m
#datatype,string,long,dateTime:RFC3339,string,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,3,2018-05-22T19:53:30Z,0,k3,m
,,3,2018-05-22T19:53:31Z,1,k3,m
,,3,2018-05-22T19:53:32Z,4294967296,k3,m
,,3,2018-05-22T19:53:33Z,18446744073709551615,k3,m
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,4,2018-05-22T19:53:40Z,1970-01-01T00:00:00Z,k4,m
,,4,2018-05-22T19:53:41Z,2554-07-21T23:34:33.709551615Z,k4,m
,,4,2018-05-22T19:53:42Z,2018-05-22T19:53:26.033Z,k4,m
,,4,2018-05-22T19:53:43Z,2018-05-22T19:53:26.033066Z,k4,m
,,4,2018-05-22T19:53:44Z,2018-05-22T19:00:00+01:00,k4,m
#datatype,string,long,dateTime:RFC3339,unsignedLong,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,5,2018-05-22T19:53:50Z,0,k5,m
,,5,2018-05-22T19:53:51Z,1,k5,m
,,5,2018-05-22T19:53:52Z,18446744073709551615,k5,m
,,5,2018-05-22T19:53:53Z,4294967296,k5,m
,,5,2018-05-22T19:53:54Z,9223372036854775808,k5,m
"
    outData =
        "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,unsignedLong
#group,false,false,true,true,false,true,true,false
#default,want,,,,,,,
,result,table,_start,_stop,_time,_field,_measurement,_value
,,0,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:00Z,k0,m,1
,,0,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:01Z,k0,m,0
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:10Z,k1,m,0
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:11Z,k1,m,1
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:12Z,k1,m,1
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:13Z,k1,m,110000
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:14Z,k1,m,0
,,1,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:15Z,k1,m,2036854775807
,,2,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:20Z,k2,m,0
,,2,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:21Z,k2,m,1
,,2,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:22Z,k2,m,18446744073709551615
,,2,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:23Z,k2,m,4294967296
,,2,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:24Z,k2,m,9223372036854775808
,,3,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:30Z,k3,m,0
,,3,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:31Z,k3,m,1
,,3,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:32Z,k3,m,4294967296
,,3,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:33Z,k3,m,18446744073709551615
,,4,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:40Z,k4,m,0
,,4,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:41Z,k4,m,18446744073709551615
,,4,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:42Z,k4,m,1527018806033000000
,,4,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:43Z,k4,m,1527018806033066000
,,4,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:44Z,k4,m,1527012000000000000
,,5,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:50Z,k5,m,0
,,5,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:51Z,k5,m,1
,,5,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:52Z,k5,m,18446744073709551615
,,5,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:53Z,k5,m,4294967296
,,5,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,2018-05-22T19:53:54Z,k5,m,9223372036854775808
"
    got =
        csv.from(csv: inData)
            |> range(start: 2018-05-22T19:52:00Z)
            |> toUInt()
    want = csv.from(csv: outData)

    testing.diff(got: got, want: want)
}
