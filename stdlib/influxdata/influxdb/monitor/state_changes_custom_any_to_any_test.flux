package monitor_test

import "influxdata/influxdb/monitor"
import "influxdata/influxdb/v1"
import "testing"
import "experimental"

option now = () => 2018-05-22T19:54:40Z

option monitor.log = (tables=<-) => tables |> drop(columns:["_start", "_stop"])

// These statuses were produce by custom check query
inData = "
#group,false,false,false,true,true,true,true,true,true,true,false,false,false,false
#datatype,string,long,dateTime:RFC3339,string,string,string,string,string,string,string,string,long,double,double
#default,_result,,,,,,,,,,,,,
,result,table,_time,_check_id,_check_name,_level,_measurement,_source_measurement,_type,id,_message,_source_timestamp,lat,lon
,,0,2020-04-01T13:19:00.734009512Z,000000000000000a,LLIR,ok,statuses,mta,custom,GO506_20_8813,GO506_20_8813 is in,1585747116000000000,40.676922,-73.76787
,,0,2020-04-01T13:19:00.734068665Z,000000000000000a,LLIR,ok,statuses,mta,custom,GO506_20_8813,GO506_20_8813 is in,1585747134000000000,40.676922,-73.76787
,,0,2020-04-01T13:20:01.055589553Z,000000000000000a,LLIR,ok,statuses,mta,custom,GO506_20_8813,GO506_20_8813 is in,1585747116000000000,40.676922,-73.76787
,,0,2020-04-01T13:20:01.055681722Z,000000000000000a,LLIR,ok,statuses,mta,custom,GO506_20_8813,GO506_20_8813 is in,1585747134000000000,40.676922,-73.76787
,,0,2020-04-01T13:20:01.055731206Z,000000000000000a,LLIR,ok,statuses,mta,custom,GO506_20_8813,GO506_20_8813 is in,1585747151000000000,40.676922,-73.76787
,,0,2020-04-01T13:20:01.055757119Z,000000000000000a,LLIR,ok,statuses,mta,custom,GO506_20_8813,GO506_20_8813 is in,1585747168000000000,40.676922,-73.76787
,,0,2020-04-01T13:20:01.055841776Z,000000000000000a,LLIR,ok,statuses,mta,custom,GO506_20_8813,GO506_20_8813 is in,1585747185000000000,40.676922,-73.76787
,,0,2020-04-01T13:25:01.120735321Z,000000000000000a,LLIR,ok,statuses,mta,custom,GO506_20_8813,GO506_20_8813 is in,1585747203000000000,40.676922,-73.76787
,,0,2020-04-01T13:25:01.120827394Z,000000000000000a,LLIR,ok,statuses,mta,custom,GO506_20_8813,GO506_20_8813 is in,1585747219000000000,40.676922,-73.76787
,,1,2020-04-01T13:25:01.119696459Z,000000000000000a,LLIR,warn,statuses,mta,custom,GO506_20_8813,GO506_20_8813 is out,1585747271000000000,40.699608,-73.80853
,,1,2020-04-01T13:25:01.119812609Z,000000000000000a,LLIR,warn,statuses,mta,custom,GO506_20_8813,GO506_20_8813 is out,1585747288000000000,40.699608,-73.80853
,,1,2020-04-01T13:25:01.119843339Z,000000000000000a,LLIR,warn,statuses,mta,custom,GO506_20_8813,GO506_20_8813 is out,1585747304000000000,40.699608,-73.80853
,,1,2020-04-01T13:25:01.119944446Z,000000000000000a,LLIR,warn,statuses,mta,custom,GO506_20_8813,GO506_20_8813 is out,1585747322000000000,40.699608,-73.80853
,,1,2020-04-01T13:25:01.119986133Z,000000000000000a,LLIR,warn,statuses,mta,custom,GO506_20_8813,GO506_20_8813 is out,1585747339000000000,40.699608,-73.80853
,,1,2020-04-01T13:25:01.12003354Z,000000000000000a,LLIR,warn,statuses,mta,custom,GO506_20_8813,GO506_20_8813 is out,1585747356000000000,40.699608,-73.80853
,,1,2020-04-01T13:25:01.120075771Z,000000000000000a,LLIR,warn,statuses,mta,custom,GO506_20_8813,GO506_20_8813 is out,1585747374000000000,40.699608,-73.80853
,,1,2020-04-01T13:25:01.120119872Z,000000000000000a,LLIR,warn,statuses,mta,custom,GO506_20_8813,GO506_20_8813 is out,1585747390000000000,40.699608,-73.80853
,,1,2020-04-01T13:25:01.120620071Z,000000000000000a,LLIR,warn,statuses,mta,custom,GO506_20_8813,GO506_20_8813 is out,1585747237000000000,40.70075,-73.804858
,,1,2020-04-01T13:25:01.120707032Z,000000000000000a,LLIR,warn,statuses,mta,custom,GO506_20_8813,GO506_20_8813 is out,1585747254000000000,40.70075,-73.804858"

outData = "
#group,false,false,true,true,true,false,true,false,false,true,true,false,false,true
#datatype,string,long,string,string,string,string,string,long,dateTime:RFC3339,string,string,double,double,string
#default,_result,,,,,,,,,,,,,
,result,table,_check_id,_check_name,_measurement,_message,_source_measurement,_source_timestamp,_time,_type,id,lat,lon,_level
,,0,000000000000000a,LLIR,statuses,GO506_20_8813 is out,mta,1585747237000000000,2020-04-01T13:25:01.120620071Z,custom,GO506_20_8813,40.70075,-73.804858,warn
"

t_state_changes_custom_any_to_any = (table=<-) => table
    |> monitor.stateChanges(
        fromLevel: "any",
        toLevel: "any",
    )
    |> drop(columns: ["_start","_stop"])

test monitor_state_changes_custom_any_to_any = () =>
    ({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_state_changes_custom_any_to_any})
