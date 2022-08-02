package universe_test


import "testing"
import "csv"

option now = () => 2030-01-01T00:00:00Z

inData =
    "
#datatype,string,long,string,string,dateTime:RFC3339,string
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_measurement,_field,_time,_value
,,0,Sgf,DlXwgrw,2018-12-18T22:11:05Z,word
,,0,Sgf,DlXwgrw,2018-12-18T22:11:15Z,glass
,,0,Sgf,DlXwgrw,2018-12-18T22:11:25Z,more
,,0,Sgf,DlXwgrw,2018-12-18T22:11:35Z,or
,,0,Sgf,DlXwgrw,2018-12-18T22:11:45Z,less
,,0,Sgf,DlXwgrw,2018-12-18T22:11:55Z,glass
"
outData =
    "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string
#group,false,false,true,true,true,true,false
#default,_result,,,,,,
,result,table,_start,_stop,_measurement,_field,_value
,,0,2018-12-01T00:00:00Z,2030-01-01T00:00:00Z,Sgf,DlXwgrw,glass
"

testcase mode_string {
    got =
        csv.from(csv: inData)
            |> testing.load()
            |> range(start: 2018-12-01T00:00:00Z)
            |> mode()
    want = csv.from(csv: outData)

    testing.diff(got, want)
}
