package monitor_test

import "influxdata/influxdb/monitor"
import "influxdata/influxdb/v1"
import "testing"
import "experimental"

option monitor.log = (tables=<-) => tables |> drop(columns:["_start", "_stop"])

// These statuses were produce by custom check query
inData = "
#group,false,false,false,false,true,true,true,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,string,string,string,string,string,string,string,string,string
#default,_result,,,,,,,,,,,
,result,table,_time,_value,_check_id,_check_name,_field,_level,_measurement,_source_measurement,_type,id
,,0,2020-04-01T13:20:01.055501743Z,GO506_20_8813 is in,000000000000000a,LLIR,_message,ok,statuses,mta,custom,GO506_20_8813
,,0,2020-04-01T13:20:01.055589553Z,GO506_20_8813 is in,000000000000000a,LLIR,_message,ok,statuses,mta,custom,GO506_20_8813
,,0,2020-04-01T13:20:01.055681722Z,GO506_20_8813 is in,000000000000000a,LLIR,_message,ok,statuses,mta,custom,GO506_20_8813
,,0,2020-04-01T13:20:01.055731206Z,GO506_20_8813 is in,000000000000000a,LLIR,_message,ok,statuses,mta,custom,GO506_20_8813
,,0,2020-04-01T13:20:01.055757119Z,GO506_20_8813 is in,000000000000000a,LLIR,_message,ok,statuses,mta,custom,GO506_20_8813
,,0,2020-04-01T13:20:01.055841776Z,GO506_20_8813 is in,000000000000000a,LLIR,_message,ok,statuses,mta,custom,GO506_20_8813
,,0,2020-04-01T13:20:01.055893004Z,GO506_20_8813 is in,000000000000000a,LLIR,_message,ok,statuses,mta,custom,GO506_20_8813
,,0,2020-04-01T13:20:01.05593662Z,GO506_20_8813 is in,000000000000000a,LLIR,_message,ok,statuses,mta,custom,GO506_20_8813
,,0,2020-04-01T13:25:01.120735321Z,GO506_20_8813 is in,000000000000000a,LLIR,_message,ok,statuses,mta,custom,GO506_20_8813
,,0,2020-04-01T13:25:01.120827394Z,GO506_20_8813 is in,000000000000000a,LLIR,_message,ok,statuses,mta,custom,GO506_20_8813
,,1,2020-04-01T13:25:01.119696459Z,GO506_20_8813 is out,000000000000000a,LLIR,_message,warn,statuses,mta,custom,GO506_20_8813
,,1,2020-04-01T13:25:01.119812609Z,GO506_20_8813 is out,000000000000000a,LLIR,_message,warn,statuses,mta,custom,GO506_20_8813
,,1,2020-04-01T13:25:01.119843339Z,GO506_20_8813 is out,000000000000000a,LLIR,_message,warn,statuses,mta,custom,GO506_20_8813
,,1,2020-04-01T13:25:01.119944446Z,GO506_20_8813 is out,000000000000000a,LLIR,_message,warn,statuses,mta,custom,GO506_20_8813
,,1,2020-04-01T13:25:01.119986133Z,GO506_20_8813 is out,000000000000000a,LLIR,_message,warn,statuses,mta,custom,GO506_20_8813
,,1,2020-04-01T13:25:01.12003354Z,GO506_20_8813 is out,000000000000000a,LLIR,_message,warn,statuses,mta,custom,GO506_20_8813
,,1,2020-04-01T13:25:01.120075771Z,GO506_20_8813 is out,000000000000000a,LLIR,_message,warn,statuses,mta,custom,GO506_20_8813
,,1,2020-04-01T13:25:01.120119872Z,GO506_20_8813 is out,000000000000000a,LLIR,_message,warn,statuses,mta,custom,GO506_20_8813
,,1,2020-04-01T13:25:01.120162813Z,GO506_20_8813 is out,000000000000000a,LLIR,_message,warn,statuses,mta,custom,GO506_20_8813
,,1,2020-04-01T13:25:01.120177679Z,GO506_20_8813 is out,000000000000000a,LLIR,_message,warn,statuses,mta,custom,GO506_20_8813
,,1,2020-04-01T13:25:01.12024583Z,GO506_20_8813 is out,000000000000000a,LLIR,_message,warn,statuses,mta,custom,GO506_20_8813
,,1,2020-04-01T13:25:01.120285437Z,GO506_20_8813 is out,000000000000000a,LLIR,_message,warn,statuses,mta,custom,GO506_20_8813
,,1,2020-04-01T13:25:01.120315321Z,GO506_20_8813 is out,000000000000000a,LLIR,_message,warn,statuses,mta,custom,GO506_20_8813
,,1,2020-04-01T13:25:01.120341734Z,GO506_20_8813 is out,000000000000000a,LLIR,_message,warn,statuses,mta,custom,GO506_20_8813
,,1,2020-04-01T13:25:01.120620071Z,GO506_20_8813 is out,000000000000000a,LLIR,_message,warn,statuses,mta,custom,GO506_20_8813
,,1,2020-04-01T13:25:01.120707032Z,GO506_20_8813 is out,000000000000000a,LLIR,_message,warn,statuses,mta,custom,GO506_20_8813

#group,false,false,false,false,true,true,true,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,long,string,string,string,string,string,string,string,string
#default,_result,,,,,,,,,,,
,result,table,_time,_value,_check_id,_check_name,_field,_level,_measurement,_source_measurement,_type,id
,,2,2020-04-01T13:20:01.055501743Z,1585747099000000000,000000000000000a,LLIR,_source_timestamp,ok,statuses,mta,custom,GO506_20_8813
,,2,2020-04-01T13:20:01.055589553Z,1585747116000000000,000000000000000a,LLIR,_source_timestamp,ok,statuses,mta,custom,GO506_20_8813
,,2,2020-04-01T13:20:01.055681722Z,1585747134000000000,000000000000000a,LLIR,_source_timestamp,ok,statuses,mta,custom,GO506_20_8813
,,2,2020-04-01T13:20:01.055731206Z,1585747151000000000,000000000000000a,LLIR,_source_timestamp,ok,statuses,mta,custom,GO506_20_8813
,,2,2020-04-01T13:20:01.055757119Z,1585747168000000000,000000000000000a,LLIR,_source_timestamp,ok,statuses,mta,custom,GO506_20_8813
,,2,2020-04-01T13:20:01.055841776Z,1585747185000000000,000000000000000a,LLIR,_source_timestamp,ok,statuses,mta,custom,GO506_20_8813
,,2,2020-04-01T13:20:01.055893004Z,1585746980000000000,000000000000000a,LLIR,_source_timestamp,ok,statuses,mta,custom,GO506_20_8813
,,2,2020-04-01T13:20:01.05593662Z,1585746997000000000,000000000000000a,LLIR,_source_timestamp,ok,statuses,mta,custom,GO506_20_8813
,,2,2020-04-01T13:25:01.120735321Z,1585747203000000000,000000000000000a,LLIR,_source_timestamp,ok,statuses,mta,custom,GO506_20_8813
,,2,2020-04-01T13:25:01.120827394Z,1585747219000000000,000000000000000a,LLIR,_source_timestamp,ok,statuses,mta,custom,GO506_20_8813
,,3,2020-04-01T13:25:01.119696459Z,1585747271000000000,000000000000000a,LLIR,_source_timestamp,warn,statuses,mta,custom,GO506_20_8813
,,3,2020-04-01T13:25:01.119812609Z,1585747288000000000,000000000000000a,LLIR,_source_timestamp,warn,statuses,mta,custom,GO506_20_8813
,,3,2020-04-01T13:25:01.119843339Z,1585747304000000000,000000000000000a,LLIR,_source_timestamp,warn,statuses,mta,custom,GO506_20_8813
,,3,2020-04-01T13:25:01.119944446Z,1585747322000000000,000000000000000a,LLIR,_source_timestamp,warn,statuses,mta,custom,GO506_20_8813
,,3,2020-04-01T13:25:01.119986133Z,1585747339000000000,000000000000000a,LLIR,_source_timestamp,warn,statuses,mta,custom,GO506_20_8813
,,3,2020-04-01T13:25:01.12003354Z,1585747356000000000,000000000000000a,LLIR,_source_timestamp,warn,statuses,mta,custom,GO506_20_8813
,,3,2020-04-01T13:25:01.120075771Z,1585747374000000000,000000000000000a,LLIR,_source_timestamp,warn,statuses,mta,custom,GO506_20_8813
,,3,2020-04-01T13:25:01.120119872Z,1585747390000000000,000000000000000a,LLIR,_source_timestamp,warn,statuses,mta,custom,GO506_20_8813
,,3,2020-04-01T13:25:01.120162813Z,1585747407000000000,000000000000000a,LLIR,_source_timestamp,warn,statuses,mta,custom,GO506_20_8813
,,3,2020-04-01T13:25:01.120177679Z,1585747424000000000,000000000000000a,LLIR,_source_timestamp,warn,statuses,mta,custom,GO506_20_8813
,,3,2020-04-01T13:25:01.12024583Z,1585747442000000000,000000000000000a,LLIR,_source_timestamp,warn,statuses,mta,custom,GO506_20_8813
,,3,2020-04-01T13:25:01.120285437Z,1585747459000000000,000000000000000a,LLIR,_source_timestamp,warn,statuses,mta,custom,GO506_20_8813
,,3,2020-04-01T13:25:01.120315321Z,1585747476000000000,000000000000000a,LLIR,_source_timestamp,warn,statuses,mta,custom,GO506_20_8813
,,3,2020-04-01T13:25:01.120341734Z,1585747493000000000,000000000000000a,LLIR,_source_timestamp,warn,statuses,mta,custom,GO506_20_8813
,,3,2020-04-01T13:25:01.120620071Z,1585747237000000000,000000000000000a,LLIR,_source_timestamp,warn,statuses,mta,custom,GO506_20_8813
,,3,2020-04-01T13:25:01.120707032Z,1585747254000000000,000000000000000a,LLIR,_source_timestamp,warn,statuses,mta,custom,GO506_20_8813

#group,false,false,false,false,true,true,true,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,double,string,string,string,string,string,string,string,string
#default,_result,,,,,,,,,,,
,result,table,_time,_value,_check_id,_check_name,_field,_level,_measurement,_source_measurement,_type,id
,,4,2020-04-01T13:20:01.055501743Z,40.676922,000000000000000a,LLIR,lat,ok,statuses,mta,custom,GO506_20_8813
,,4,2020-04-01T13:20:01.055589553Z,40.676922,000000000000000a,LLIR,lat,ok,statuses,mta,custom,GO506_20_8813
,,4,2020-04-01T13:20:01.055681722Z,40.676922,000000000000000a,LLIR,lat,ok,statuses,mta,custom,GO506_20_8813
,,4,2020-04-01T13:20:01.055731206Z,40.676922,000000000000000a,LLIR,lat,ok,statuses,mta,custom,GO506_20_8813
,,4,2020-04-01T13:20:01.055757119Z,40.676922,000000000000000a,LLIR,lat,ok,statuses,mta,custom,GO506_20_8813
,,4,2020-04-01T13:20:01.055841776Z,40.676922,000000000000000a,LLIR,lat,ok,statuses,mta,custom,GO506_20_8813
,,4,2020-04-01T13:20:01.055893004Z,40.672562,000000000000000a,LLIR,lat,ok,statuses,mta,custom,GO506_20_8813
,,4,2020-04-01T13:20:01.05593662Z,40.672562,000000000000000a,LLIR,lat,ok,statuses,mta,custom,GO506_20_8813
,,4,2020-04-01T13:25:01.120735321Z,40.676922,000000000000000a,LLIR,lat,ok,statuses,mta,custom,GO506_20_8813
,,4,2020-04-01T13:25:01.120827394Z,40.676922,000000000000000a,LLIR,lat,ok,statuses,mta,custom,GO506_20_8813
,,5,2020-04-01T13:25:01.119696459Z,40.699608,000000000000000a,LLIR,lat,warn,statuses,mta,custom,GO506_20_8813
,,5,2020-04-01T13:25:01.119812609Z,40.699608,000000000000000a,LLIR,lat,warn,statuses,mta,custom,GO506_20_8813
,,5,2020-04-01T13:25:01.119843339Z,40.699608,000000000000000a,LLIR,lat,warn,statuses,mta,custom,GO506_20_8813
,,5,2020-04-01T13:25:01.119944446Z,40.699608,000000000000000a,LLIR,lat,warn,statuses,mta,custom,GO506_20_8813
,,5,2020-04-01T13:25:01.119986133Z,40.699608,000000000000000a,LLIR,lat,warn,statuses,mta,custom,GO506_20_8813
,,5,2020-04-01T13:25:01.12003354Z,40.699608,000000000000000a,LLIR,lat,warn,statuses,mta,custom,GO506_20_8813
,,5,2020-04-01T13:25:01.120075771Z,40.699608,000000000000000a,LLIR,lat,warn,statuses,mta,custom,GO506_20_8813
,,5,2020-04-01T13:25:01.120119872Z,40.699608,000000000000000a,LLIR,lat,warn,statuses,mta,custom,GO506_20_8813
,,5,2020-04-01T13:25:01.120162813Z,40.699608,000000000000000a,LLIR,lat,warn,statuses,mta,custom,GO506_20_8813
,,5,2020-04-01T13:25:01.120177679Z,40.699608,000000000000000a,LLIR,lat,warn,statuses,mta,custom,GO506_20_8813
,,5,2020-04-01T13:25:01.12024583Z,40.699608,000000000000000a,LLIR,lat,warn,statuses,mta,custom,GO506_20_8813
,,5,2020-04-01T13:25:01.120285437Z,40.699608,000000000000000a,LLIR,lat,warn,statuses,mta,custom,GO506_20_8813
,,5,2020-04-01T13:25:01.120315321Z,40.699608,000000000000000a,LLIR,lat,warn,statuses,mta,custom,GO506_20_8813
,,5,2020-04-01T13:25:01.120341734Z,40.699608,000000000000000a,LLIR,lat,warn,statuses,mta,custom,GO506_20_8813
,,5,2020-04-01T13:25:01.120620071Z,40.70075,000000000000000a,LLIR,lat,warn,statuses,mta,custom,GO506_20_8813
,,5,2020-04-01T13:25:01.120707032Z,40.70075,000000000000000a,LLIR,lat,warn,statuses,mta,custom,GO506_20_8813
,,6,2020-04-01T13:20:01.055501743Z,-73.76787,000000000000000a,LLIR,lon,ok,statuses,mta,custom,GO506_20_8813
,,6,2020-04-01T13:20:01.055589553Z,-73.76787,000000000000000a,LLIR,lon,ok,statuses,mta,custom,GO506_20_8813
,,6,2020-04-01T13:20:01.055681722Z,-73.76787,000000000000000a,LLIR,lon,ok,statuses,mta,custom,GO506_20_8813
,,6,2020-04-01T13:20:01.055731206Z,-73.76787,000000000000000a,LLIR,lon,ok,statuses,mta,custom,GO506_20_8813
,,6,2020-04-01T13:20:01.055757119Z,-73.76787,000000000000000a,LLIR,lon,ok,statuses,mta,custom,GO506_20_8813
,,6,2020-04-01T13:20:01.055841776Z,-73.76787,000000000000000a,LLIR,lon,ok,statuses,mta,custom,GO506_20_8813
,,6,2020-04-01T13:20:01.055893004Z,-73.760456,000000000000000a,LLIR,lon,ok,statuses,mta,custom,GO506_20_8813
,,6,2020-04-01T13:20:01.05593662Z,-73.760456,000000000000000a,LLIR,lon,ok,statuses,mta,custom,GO506_20_8813
,,6,2020-04-01T13:25:01.120735321Z,-73.76787,000000000000000a,LLIR,lon,ok,statuses,mta,custom,GO506_20_8813
,,6,2020-04-01T13:25:01.120827394Z,-73.76787,000000000000000a,LLIR,lon,ok,statuses,mta,custom,GO506_20_8813
,,7,2020-04-01T13:25:01.119696459Z,-73.80853,000000000000000a,LLIR,lon,warn,statuses,mta,custom,GO506_20_8813
,,7,2020-04-01T13:25:01.119812609Z,-73.80853,000000000000000a,LLIR,lon,warn,statuses,mta,custom,GO506_20_8813
,,7,2020-04-01T13:25:01.119843339Z,-73.80853,000000000000000a,LLIR,lon,warn,statuses,mta,custom,GO506_20_8813
,,7,2020-04-01T13:25:01.119944446Z,-73.80853,000000000000000a,LLIR,lon,warn,statuses,mta,custom,GO506_20_8813
,,7,2020-04-01T13:25:01.119986133Z,-73.80853,000000000000000a,LLIR,lon,warn,statuses,mta,custom,GO506_20_8813
,,7,2020-04-01T13:25:01.12003354Z,-73.80853,000000000000000a,LLIR,lon,warn,statuses,mta,custom,GO506_20_8813
,,7,2020-04-01T13:25:01.120075771Z,-73.80853,000000000000000a,LLIR,lon,warn,statuses,mta,custom,GO506_20_8813
,,7,2020-04-01T13:25:01.120119872Z,-73.80853,000000000000000a,LLIR,lon,warn,statuses,mta,custom,GO506_20_8813
,,7,2020-04-01T13:25:01.120162813Z,-73.80853,000000000000000a,LLIR,lon,warn,statuses,mta,custom,GO506_20_8813
,,7,2020-04-01T13:25:01.120177679Z,-73.80853,000000000000000a,LLIR,lon,warn,statuses,mta,custom,GO506_20_8813
,,7,2020-04-01T13:25:01.12024583Z,-73.80853,000000000000000a,LLIR,lon,warn,statuses,mta,custom,GO506_20_8813
,,7,2020-04-01T13:25:01.120285437Z,-73.80853,000000000000000a,LLIR,lon,warn,statuses,mta,custom,GO506_20_8813
,,7,2020-04-01T13:25:01.120315321Z,-73.80853,000000000000000a,LLIR,lon,warn,statuses,mta,custom,GO506_20_8813
,,7,2020-04-01T13:25:01.120341734Z,-73.80853,000000000000000a,LLIR,lon,warn,statuses,mta,custom,GO506_20_8813
,,7,2020-04-01T13:25:01.120620071Z,-73.804858,000000000000000a,LLIR,lon,warn,statuses,mta,custom,GO506_20_8813
,,7,2020-04-01T13:25:01.120707032Z,-73.804858,000000000000000a,LLIR,lon,warn,statuses,mta,custom,GO506_20_8813
"

outData = "
#group,false,false,true,true,true,false,true,false,false,true,true,false,false,true
#datatype,string,long,string,string,string,string,string,long,dateTime:RFC3339,string,string,double,double,string
#default,_result,,,,,,,,,,,,,
,result,table,_check_id,_check_name,_measurement,_message,_source_measurement,_source_timestamp,_time,_type,id,lat,lon,_level
,,0,000000000000000a,LLIR,statuses,GO506_20_8813 is out,mta,1585747237000000000,2020-04-01T13:25:01.120620071Z,custom,GO506_20_8813,40.70075,-73.804858,warn
"

t_state_changes_custom_any_to_any = (table=<-) => table
    |> range(start: 2018-05-22T19:54:40Z)
    |> v1.fieldsAsCols()
    |> monitor.stateChanges(
        fromLevel: "any",
        toLevel: "any",
    )
    |> drop(columns: ["_start","_stop"])

test monitor_state_changes_custom_any_to_any = () =>
    ({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_state_changes_custom_any_to_any})
