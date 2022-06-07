package universe_test


import "testing"
import "csv"

option now = () => 2030-01-01T00:00:00Z

inData =
    "
#datatype,string,long,string,string,dateTime:RFC3339,double
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_measurement,_field,_time,_value
,,0,SOYcRk,NC7N,2018-12-18T21:12:45Z,-61.68790887989735
,,0,SOYcRk,NC7N,2018-12-18T21:12:55Z,-6.3173755351186465
,,0,SOYcRk,NC7N,2018-12-18T21:13:05Z,-26.049728557657513
,,0,SOYcRk,NC7N,2018-12-18T21:13:15Z,114.285955884979
,,0,SOYcRk,NC7N,2018-12-18T21:13:25Z,16.140262630578995
,,0,SOYcRk,NC7N,2018-12-18T21:13:35Z,29.50336437998469

#datatype,string,long,string,string,dateTime:RFC3339,long
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_measurement,_field,_time,_value
,,1,Reiva,dGpnr,2019-01-09T19:44:58Z,17
,,1,Reiva,dGpnr,2019-01-09T19:45:08Z,-44
,,1,Reiva,dGpnr,2019-01-09T19:45:18Z,-99
,,1,Reiva,dGpnr,2019-01-09T19:45:28Z,-85
,,1,Reiva,dGpnr,2019-01-09T19:45:38Z,18
,,1,Reiva,dGpnr,2019-01-09T19:45:48Z,99

#datatype,string,long,string,string,dateTime:RFC3339,unsignedLong
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_measurement,_field,_time,_value
,,2,Reiva,rc2iOD1,2019-01-09T19:44:58Z,79
,,2,Reiva,rc2iOD1,2019-01-09T19:45:08Z,33
,,2,Reiva,rc2iOD1,2019-01-09T19:45:18Z,97
,,2,Reiva,rc2iOD1,2019-01-09T19:45:28Z,90
,,2,Reiva,rc2iOD1,2019-01-09T19:45:38Z,96
,,2,Reiva,rc2iOD1,2019-01-09T19:45:48Z,10
"
outData =
    "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,double
#group,false,false,true,true,true,true,false
#default,_result,,,,,,
,result,table,_start,_stop,_measurement,_field,_value
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,SOYcRk,NC7N,29.50336437998469
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,Reiva,dGpnr,18
,,2,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,Reiva,rc2iOD1,96
"

testcase quantile_tdigest {
    got =
        csv.from(csv: inData)
            |> testing.load()
            |> range(start: 2018-01-01T00:00:00Z)
            |> quantile(q: 0.75, method: "estimate_tdigest")
    want = csv.from(csv: outData)

    testing.diff(got, want)
}
