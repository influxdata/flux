package universe_test


import "csv"
import "testing"

option now = () => 2030-01-01T00:00:00Z

testcase increase_basic {
    got =
        csv.from(
            csv:
                "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#group,false,false,false,false,false,false,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,counter,_field,_measurement,cpu,host
,,0,2018-05-22T19:53:24.421470485Z,2018-05-22T19:55:00Z,2018-05-22T19:53:26Z,1,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:24.421470485Z,2018-05-22T19:55:00Z,2018-05-22T19:53:36Z,2,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:24.421470485Z,2018-05-22T19:55:00Z,2018-05-22T19:53:46Z,3,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:24.421470485Z,2018-05-22T19:55:00Z,2018-05-22T19:53:56Z,5,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:24.421470485Z,2018-05-22T19:55:00Z,2018-05-22T19:54:06Z,0,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:24.421470485Z,2018-05-22T19:55:00Z,2018-05-22T19:54:16Z,1,usage_guest,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:24.421470485Z,2018-05-22T19:55:00Z,2018-05-22T19:53:26Z,2,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:24.421470485Z,2018-05-22T19:55:00Z,2018-05-22T19:53:36Z,4,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:24.421470485Z,2018-05-22T19:55:00Z,2018-05-22T19:53:46Z,4,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:24.421470485Z,2018-05-22T19:55:00Z,2018-05-22T19:53:56Z,0,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:24.421470485Z,2018-05-22T19:55:00Z,2018-05-22T19:54:06Z,2,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:24.421470485Z,2018-05-22T19:55:00Z,2018-05-22T19:54:16Z,10,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:24.421470485Z,2018-05-22T19:55:00Z,2018-05-22T19:54:26Z,4,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:24.421470485Z,2018-05-22T19:55:00Z,2018-05-22T19:54:36Z,20,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:24.421470485Z,2018-05-22T19:55:00Z,2018-05-22T19:54:46Z,7,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:24.421470485Z,2018-05-22T19:55:00Z,2018-05-22T19:54:56Z,10,usage_guest_nice,cpu,cpu-total,host.local
",
        )
            |> range(start: 2018-05-22T19:53:26Z)
            |> increase(columns: ["counter"])

    want =
        csv.from(
            csv:
                "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,counter,_field,_measurement,cpu,host
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:26Z,0,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:36Z,1,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:46Z,2,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:56Z,4,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:06Z,4,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:16Z,5,usage_guest,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:26Z,0,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:36Z,2,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:46Z,2,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:56Z,2,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:06Z,4,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:16Z,12,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:26Z,16,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:36Z,32,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:46Z,39,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:56Z,42,usage_guest_nice,cpu,cpu-total,host.local
",
        )

    testing.diff(got: got, want: want)
}

// Negative difference case -
// The difference between two non-null values is their algebraic difference;
// OR current value, if the result is negative and the current value is greater than equal to zero and nonNegative: true
// OR zero, if the result is negative and the current value is less than zero and nonNegative: true
testcase increase_with_negative_difference {
    got =
        csv.from(
            csv:
                "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#group,false,false,false,false,false,false,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,counter,_field,_measurement,cpu,host
,,0,2018-05-22T19:53:24.421470485Z,2018-05-22T19:55:00Z,2018-05-22T19:53:26Z,4302,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:24.421470485Z,2018-05-22T19:55:00Z,2018-05-22T19:53:36Z,4844,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:24.421470485Z,2018-05-22T19:55:00Z,2018-05-22T19:53:46Z,5091,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:24.421470485Z,2018-05-22T19:55:00Z,2018-05-22T19:53:56Z,13,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:24.421470485Z,2018-05-22T19:55:00Z,2018-05-22T19:54:06Z,215,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:24.421470485Z,2018-05-22T19:55:00Z,2018-05-22T19:54:16Z,762,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:24.421470485Z,2018-05-22T19:55:00Z,2018-05-22T19:54:26Z,1108,usage_guest,cpu,cpu-total,host.local
",
        )
            |> range(start: 2018-05-22T19:53:26Z)
            |> increase(columns: ["counter"])

    want =
        csv.from(
            csv:
                "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,counter,_field,_measurement,cpu,host
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:26Z,0,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:36Z,542,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:46Z,789,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:56Z,802,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:06Z,1004,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:16Z,1551,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:26Z,1897,usage_guest,cpu,cpu-total,host.local
",
        )

    testing.diff(got: got, want: want)
}
