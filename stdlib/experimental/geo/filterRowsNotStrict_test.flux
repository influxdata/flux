package geo_test

import "experimental/geo"
import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_ci,_field,_measurement,_pt
,,0,2019-02-18T04:17:43.176943164Z,2020-02-18T10:17:43.176943164Z,2019-11-01T00:33:07.54916732Z,40.728077,89c2624,lat,taxi,end

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_ci,_field,_measurement,_pt
,,1,2019-02-18T04:17:43.176943164Z,2020-02-18T10:17:43.176943164Z,2019-11-01T13:41:59.331776148Z,1572615165632969758,89c261c,tid,taxi,end

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_ci,_field,_measurement,_pt
,,2,2019-02-18T04:17:43.176943164Z,2020-02-18T10:17:43.176943164Z,2019-11-01T00:07:18.082153551Z,1572566409947779410,89c2594,tid,taxi,end

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_ci,_field,_measurement,_pt
,,3,2019-02-18T04:17:43.176943164Z,2020-02-18T10:17:43.176943164Z,2019-11-01T00:07:41.235010051Z,40.671928,89c25b4,lat,taxi,end

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_ci,_field,_measurement,_pt
,,4,2019-02-18T04:17:43.176943164Z,2020-02-18T10:17:43.176943164Z,2019-11-01T00:07:41.235010051Z,-73.962692,89c25b4,lon,taxi,end

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_ci,_field,_measurement,_pt
,,5,2019-02-18T04:17:43.176943164Z,2020-02-18T10:17:43.176943164Z,2019-11-01T00:07:41.235010051Z,1572566426666145821,89c25b4,tid,taxi,end

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_ci,_field,_measurement,_pt
,,6,2019-02-18T04:17:43.176943164Z,2020-02-18T10:17:43.176943164Z,2019-11-01T00:17:38.287113937Z,1572567458287113937,89c2664,tid,taxi,start

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_ci,_field,_measurement,_pt
,,7,2019-02-18T04:17:43.176943164Z,2020-02-18T10:17:43.176943164Z,2019-11-01T13:41:59.331776148Z,-73.752579,89c261c,lon,taxi,end

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_ci,_field,_measurement,_pt
,,8,2019-02-18T04:17:43.176943164Z,2020-02-18T10:17:43.176943164Z,2019-11-01T13:41:59.331776148Z,0,89c261c,tip,taxi,end

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_ci,_field,_measurement,_pt
,,9,2019-02-18T04:17:43.176943164Z,2020-02-18T10:17:43.176943164Z,2019-11-01T00:07:18.082153551Z,1.2,89c2594,dist,taxi,end

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_ci,_field,_measurement,_pt
,,10,2019-02-18T04:17:43.176943164Z,2020-02-18T10:17:43.176943164Z,2019-11-01T00:17:38.287113937Z,-73.776665,89c2664,lon,taxi,start

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_ci,_field,_measurement,_pt
,,11,2019-02-18T04:17:43.176943164Z,2020-02-18T10:17:43.176943164Z,2019-11-01T13:32:45.632969758Z,40.733585,89c2624,lat,taxi,start

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_ci,_field,_measurement,_pt
,,12,2019-02-18T04:17:43.176943164Z,2020-02-18T10:17:43.176943164Z,2019-11-01T00:07:18.082153551Z,-73.951332,89c2594,lon,taxi,end

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_ci,_field,_measurement,_pt
,,13,2019-02-18T04:17:43.176943164Z,2020-02-18T10:17:43.176943164Z,2019-11-01T00:07:41.235010051Z,1.75,89c25b4,tip,taxi,end

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_ci,_field,_measurement,_pt
,,14,2019-02-18T04:17:43.176943164Z,2020-02-18T10:17:43.176943164Z,2019-11-01T00:17:38.287113937Z,40.645245,89c2664,lat,taxi,start

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_ci,_field,_measurement,_pt
,,15,2019-02-18T04:17:43.176943164Z,2020-02-18T10:17:43.176943164Z,2019-11-01T00:33:07.54916732Z,9.7,89c2624,dist,taxi,end

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_ci,_field,_measurement,_pt
,,16,2019-02-18T04:17:43.176943164Z,2020-02-18T10:17:43.176943164Z,2019-11-01T13:41:59.331776148Z,40.725647,89c261c,lat,taxi,end

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_ci,_field,_measurement,_pt
,,17,2019-02-18T04:17:43.176943164Z,2020-02-18T10:17:43.176943164Z,2019-11-01T00:33:07.54916732Z,5,89c2624,tip,taxi,end

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_ci,_field,_measurement,_pt
,,18,2019-02-18T04:17:43.176943164Z,2020-02-18T10:17:43.176943164Z,2019-11-01T00:33:07.54916732Z,1572567458287113937,89c2624,tid,taxi,end

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_ci,_field,_measurement,_pt
,,19,2019-02-18T04:17:43.176943164Z,2020-02-18T10:17:43.176943164Z,2019-11-01T00:00:09.94777941Z,40.712173,89c25bc,lat,taxi,start
,,19,2019-02-18T04:17:43.176943164Z,2020-02-18T10:17:43.176943164Z,2019-11-01T00:00:26.666145821Z,40.688564,89c25bc,lat,taxi,start

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_ci,_field,_measurement,_pt
,,20,2019-02-18T04:17:43.176943164Z,2020-02-18T10:17:43.176943164Z,2019-11-01T00:07:18.082153551Z,40.7122,89c2594,lat,taxi,end

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_ci,_field,_measurement,_pt
,,21,2019-02-18T04:17:43.176943164Z,2020-02-18T10:17:43.176943164Z,2019-11-01T00:07:18.082153551Z,0,89c2594,tip,taxi,end

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_ci,_field,_measurement,_pt
,,22,2019-02-18T04:17:43.176943164Z,2020-02-18T10:17:43.176943164Z,2019-11-01T00:07:41.235010051Z,1.3,89c25b4,dist,taxi,end

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_ci,_field,_measurement,_pt
,,23,2019-02-18T04:17:43.176943164Z,2020-02-18T10:17:43.176943164Z,2019-11-01T13:32:45.632969758Z,-73.737175,89c2624,lon,taxi,start

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_ci,_field,_measurement,_pt
,,24,2019-02-18T04:17:43.176943164Z,2020-02-18T10:17:43.176943164Z,2019-11-01T13:32:45.632969758Z,1572615165632969758,89c2624,tid,taxi,start

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_ci,_field,_measurement,_pt
,,25,2019-02-18T04:17:43.176943164Z,2020-02-18T10:17:43.176943164Z,2019-11-01T13:41:59.331776148Z,1.3,89c261c,dist,taxi,end

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_ci,_field,_measurement,_pt
,,26,2019-02-18T04:17:43.176943164Z,2020-02-18T10:17:43.176943164Z,2019-11-01T00:00:09.94777941Z,-73.963913,89c25bc,lon,taxi,start
,,26,2019-02-18T04:17:43.176943164Z,2020-02-18T10:17:43.176943164Z,2019-11-01T00:00:26.666145821Z,-73.965881,89c25bc,lon,taxi,start

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_ci,_field,_measurement,_pt
,,27,2019-02-18T04:17:43.176943164Z,2020-02-18T10:17:43.176943164Z,2019-11-01T00:00:09.94777941Z,1572566409947779410,89c25bc,tid,taxi,start
,,27,2019-02-18T04:17:43.176943164Z,2020-02-18T10:17:43.176943164Z,2019-11-01T00:00:26.666145821Z,1572566426666145821,89c25bc,tid,taxi,start

#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_ci,_field,_measurement,_pt
,,28,2019-02-18T04:17:43.176943164Z,2020-02-18T10:17:43.176943164Z,2019-11-01T00:33:07.54916732Z,-73.716583,89c2624,lon,taxi,end
"

outData = "
#group,false,false,false,true,true,true,false,false,false,false,false
#datatype,string,long,dateTime:RFC3339,string,string,string,long,double,double,double,double
#default,_result,,,,,,,,,,
,result,table,_time,_ci,_measurement,_pt,tid,lon,tip,lat,dist
,,0,2019-11-01T13:41:59.331776148Z,89c261c,taxi,end,1572615165632969758,-73.752579,0,40.725647,1.3

#group,false,false,false,true,true,true,false,false,false,false,false
#datatype,string,long,dateTime:RFC3339,string,string,string,double,double,double,long,double
#default,_result,,,,,,,,,,
,result,table,_time,_ci,_measurement,_pt,lat,dist,tip,tid,lon
,,1,2019-11-01T00:33:07.54916732Z,89c2624,taxi,end,40.728077,9.7,5,1572567458287113937,-73.716583

#group,false,false,false,true,true,true,false,false,false
#datatype,string,long,dateTime:RFC3339,string,string,string,double,double,long
#default,_result,,,,,,,,
,result,table,_time,_ci,_measurement,_pt,lat,lon,tid
,,2,2019-11-01T13:32:45.632969758Z,89c2624,taxi,start,40.733585,-73.737175,1572615165632969758

#group,false,false,false,true,true,true,false,false,false
#datatype,string,long,dateTime:RFC3339,string,string,string,long,double,double
#default,_result,,,,,,,,
,result,table,_time,_ci,_measurement,_pt,tid,lon,lat
,,3,2019-11-01T00:17:38.287113937Z,89c2664,taxi,start,1572567458287113937,-73.776665,40.645245
"

t_filterRowsNotStrict = (table=<-) =>
  table
    |> range(start: 2019-11-01T00:00:00Z)
    |> geo.filterRows(circle: {lat: 40.7090214, lon: -73.61846, radius: 15.0}, strict: false)
    |> drop(columns: ["_start", "_stop"])
test _filterRowsNotStrict = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_filterRowsNotStrict})
