package sampledata_test


import "csv"
import "sampledata"
import "testing"

option sampledata.start = 2021-01-01T00:00:00Z
option sampledata.stop = 2021-01-01T00:01:00Z

// return sample data with float values
testcase sampledata_float {
    want = csv.from(
        csv: "#group,false,false,true,true,false,true,false
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,double
#default,_result,,,,,,
,result,table,_start,_stop,_time,tag,_value
,,0,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:00Z,t1,-2.18
,,0,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:10Z,t1,10.92
,,0,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:20Z,t1,7.35
,,0,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:30Z,t1,17.53
,,0,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:40Z,t1,15.23
,,0,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:50Z,t1,4.43
,,1,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:00Z,t2,19.85
,,1,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:10Z,t2,4.97
,,1,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:20Z,t2,-3.75
,,1,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:30Z,t2,19.77
,,1,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:40Z,t2,13.86
,,1,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:50Z,t2,1.86",
    )

    got = sampledata.float()
        |> range(start: sampledata.start, stop: sampledata.stop)

    testing.diff(got: got, want: want)
}

// return sample data with integer and null values
testcase sampledata_int_null {
    want = csv.from(
        csv: "#group,false,false,true,true,false,true,false
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,long
#default,_result,,,,,,
,result,table,_start,_stop,_time,tag,_value
,,0,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:00Z,t1,-2
,,0,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:10Z,t1,
,,0,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:20Z,t1,7
,,0,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:30Z,t1,
,,0,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:40Z,t1,
,,0,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:50Z,t1,4
,,1,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:00Z,t2,
,,1,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:10Z,t2,4
,,1,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:20Z,t2,-3
,,1,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:30Z,t2,19
,,1,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:40Z,t2,
,,1,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:50Z,t2,1",
    )

    got = sampledata.int(includeNull: true)
        |> range(start: sampledata.start, stop: sampledata.stop)
    
    testing.diff(got: got, want: want)
}

// return sample data with string values
testcase sampledata_string {
    want = csv.from(
        csv: "#group,false,false,true,true,false,true,false
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string
#default,_result,,,,,,
,result,table,_start,_stop,_time,tag,_value
,,0,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:00Z,t1,smpl_g9qczs
,,0,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:10Z,t1,smpl_0mgv9n
,,0,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:20Z,t1,smpl_phw664
,,0,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:30Z,t1,smpl_guvzy4
,,0,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:40Z,t1,smpl_5v3cce
,,0,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:50Z,t1,smpl_s9fmgy
,,1,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:00Z,t2,smpl_b5eida
,,1,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:10Z,t2,smpl_eu4oxp
,,1,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:20Z,t2,smpl_5g7tz4
,,1,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:30Z,t2,smpl_sox1ut
,,1,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:40Z,t2,smpl_wfm757
,,1,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:50Z,t2,smpl_dtn2bv",
    )

    got = sampledata.string()
        |> range(start: sampledata.start, stop: sampledata.stop)

    testing.diff(got: got, want: want)
}

// return sample data with numeric string values
testcase sampledata_numeric_string {
    want = csv.from(
        csv: "#group,false,false,true,true,false,false,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string
#default,_result,,,,,,
,result,table,_start,_stop,_time,_value,tag
,,0,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:00Z,-2.18,t1
,,0,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:10Z,10.92,t1
,,0,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:20Z,7.35,t1
,,0,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:30Z,17.53,t1
,,0,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:40Z,15.23,t1
,,0,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:50Z,4.43,t1
,,1,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:00Z,19.85,t2
,,1,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:10Z,4.97,t2
,,1,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:20Z,-3.75,t2
,,1,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:30Z,19.77,t2
,,1,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:40Z,13.86,t2
,,1,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:50Z,1.86,t2",
    )

    got = sampledata.numericString()
        |> range(start: sampledata.start, stop: sampledata.stop)

    testing.diff(got: got, want: want)
}

// return sample data with boolean values
testcase sampledata_bool {
    want = csv.from(
        csv: "#group,false,false,true,true,false,true,false
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,boolean
#default,_result,,,,,,
,result,table,_start,_stop,_time,tag,_value
,,0,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:00Z,t1,true
,,0,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:10Z,t1,true
,,0,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:20Z,t1,false
,,0,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:30Z,t1,true
,,0,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:40Z,t1,false
,,0,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:50Z,t1,false
,,1,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:00Z,t2,false
,,1,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:10Z,t2,true
,,1,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:20Z,t2,false
,,1,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:30Z,t2,true
,,1,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:40Z,t2,true
,,1,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:50Z,t2,false",
    )

    got = sampledata.bool()
        |> range(start: sampledata.start, stop: sampledata.stop)

    testing.diff(got: got, want: want)
}

// return sample data with numeric boolean values
testcase sampledata_numeric_bool {
    want = csv.from(
        csv: "#group,false,false,true,true,false,false,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string
#default,_result,,,,,,
,result,table,_start,_stop,_time,_value,tag
,,0,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:00Z,1,t1
,,0,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:10Z,1,t1
,,0,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:20Z,0,t1
,,0,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:30Z,1,t1
,,0,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:40Z,0,t1
,,0,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:50Z,0,t1
,,1,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:00Z,0,t2
,,1,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:10Z,1,t2
,,1,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:20Z,0,t2
,,1,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:30Z,1,t2
,,1,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:40Z,1,t2
,,1,2021-01-01T00:00:00Z,2021-01-01T00:01:00Z,2021-01-01T00:00:50Z,0,t2",
    )

    got = sampledata.numericBool()
        |> range(start: sampledata.start, stop: sampledata.stop)

    testing.diff(got: got, want: want)
}
