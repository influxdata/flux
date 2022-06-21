package universe_test


import "testing"
import "csv"
import "timezone"

option now = () => 2030-01-01T00:00:00Z

inData =
    "
#datatype,string,long,string,string,dateTime:RFC3339,unsignedLong
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_measurement,_field,_time,_value
,,0,Sgf,DlXwgrw,2018-12-18T06:11:05Z,70
,,0,Sgf,DlXwgrw,2018-12-18T16:11:15Z,48
,,0,Sgf,DlXwgrw,2018-12-18T17:11:25Z,33
,,0,Sgf,DlXwgrw,2018-12-19T18:11:35Z,63
,,0,Sgf,DlXwgrw,2018-12-19T19:11:45Z,48
,,0,Sgf,DlXwgrw,2018-12-19T22:11:55Z,63
"

testcase hour_selection {
    got =
        csv.from(csv: inData)
            |> range(start: 2018-12-01T00:00:00Z)
            |> hourSelection(start: 15, stop: 19, timeColumn: "_time")

    want =
        csv.from(
            csv:
                "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,dateTime:RFC3339,unsignedLong
#group,false,false,true,true,true,true,false,false
#default,_result,,,,,,,
,result,table,_start,_stop,_measurement,_field,_time,_value
,,0,2018-12-01T00:00:00Z,2030-01-01T00:00:00Z,Sgf,DlXwgrw,2018-12-18T16:11:15Z,48
,,0,2018-12-01T00:00:00Z,2030-01-01T00:00:00Z,Sgf,DlXwgrw,2018-12-18T17:11:25Z,33
,,0,2018-12-01T00:00:00Z,2030-01-01T00:00:00Z,Sgf,DlXwgrw,2018-12-19T18:11:35Z,63
,,0,2018-12-01T00:00:00Z,2030-01-01T00:00:00Z,Sgf,DlXwgrw,2018-12-19T19:11:45Z,48
",
        )

    testing.diff(got: got, want: want)
}

testcase hour_selection_overnight_range {
    got =
        csv.from(csv: inData)
            |> range(start: 2018-12-01T00:00:00Z)
            |> hourSelection(start: 22, stop: 6, timeColumn: "_time")

    want =
        csv.from(
            csv:
                "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,dateTime:RFC3339,unsignedLong
#group,false,false,true,true,true,true,false,false
#default,_result,,,,,,,
,result,table,_start,_stop,_measurement,_field,_time,_value
,,0,2018-12-01T00:00:00Z,2030-01-01T00:00:00Z,Sgf,DlXwgrw,2018-12-18T06:11:05Z,70
,,0,2018-12-01T00:00:00Z,2030-01-01T00:00:00Z,Sgf,DlXwgrw,2018-12-19T22:11:55Z,63
",
        )

    testing.diff(got: got, want: want)
}

testcase hour_selection_location {
    option location = timezone.location(name: "America/Los_Angeles")

    got =
        csv.from(csv: inData)
            |> range(start: 2018-12-01T00:00:00Z)
            |> hourSelection(start: 9, stop: 10, timeColumn: "_time")

    want =
        csv.from(
            csv:
                "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,dateTime:RFC3339,unsignedLong
#group,false,false,true,true,true,true,false,false
#default,_result,,,,,,,
,result,table,_start,_stop,_measurement,_field,_time,_value
,,0,2018-12-01T00:00:00Z,2030-01-01T00:00:00Z,Sgf,DlXwgrw,2018-12-18T17:11:25Z,33
,,0,2018-12-01T00:00:00Z,2030-01-01T00:00:00Z,Sgf,DlXwgrw,2018-12-19T18:11:35Z,63
",
        )

    testing.diff(got: got, want: want)
}
