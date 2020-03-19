package geo_test

import "experimental/geo"
import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#group,false,false,true,true,false,true,true,true,false,false,false,false,false
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,string,long,double,double,double,double
#default,_result,,,,,,,,,,,,
,result,table,_start,_stop,_time,s2_cell_id,_measurement,_pt,tid,dist,lon,lat,tip
,,0,2019-02-18T04:17:43.176943164Z,2020-02-18T10:17:43.176943164Z,2019-11-01T00:07:18.082153551Z,89c2594,taxi,end,1572566409947779410,1.2,-73.951332,40.7122,0

#group,false,false,true,true,false,true,true,true,false,false,false,false,false
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,string,double,double,long,double,double
#default,_result,,,,,,,,,,,,
,result,table,_start,_stop,_time,s2_cell_id,_measurement,_pt,lat,lon,tid,tip,dist
,,1,2019-02-18T04:17:43.176943164Z,2020-02-18T10:17:43.176943164Z,2019-11-01T00:07:41.235010051Z,89c25b4,taxi,end,40.671928,-73.962692,1572566426666145821,1.75,1.3

#group,false,false,true,true,false,true,true,true,false,false,false,false,false
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,string,long,double,double,double,double
#default,_result,,,,,,,,,,,,
,result,table,_start,_stop,_time,s2_cell_id,_measurement,_pt,tid,lon,tip,lat,dist
,,2,2019-02-18T04:17:43.176943164Z,2020-02-18T10:17:43.176943164Z,2019-11-01T13:41:59.331776148Z,89c261c,taxi,end,1572615165632969758,-73.752579,0,40.725647,1.3

#group,false,false,true,true,false,true,true,true,false,false,false,false,false
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,string,double,double,double,long,double
#default,_result,,,,,,,,,,,,
,result,table,_start,_stop,_time,s2_cell_id,_measurement,_pt,lat,dist,tip,tid,lon
,,3,2019-02-18T04:17:43.176943164Z,2020-02-18T10:17:43.176943164Z,2019-11-01T00:33:07.54916732Z,89c2624,taxi,end,40.728077,9.7,5,1572567458287113937,-73.716583

#group,false,false,true,true,false,true,true,true,false,false,false
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,string,double,double,long
#default,_result,,,,,,,,,,
,result,table,_start,_stop,_time,s2_cell_id,_measurement,_pt,lat,lon,tid
,,4,2019-02-18T04:17:43.176943164Z,2020-02-18T10:17:43.176943164Z,2019-11-01T00:00:09.94777941Z,89c25bc,taxi,start,40.712173,-73.963913,1572566409947779410
,,4,2019-02-18T04:17:43.176943164Z,2020-02-18T10:17:43.176943164Z,2019-11-01T00:00:26.666145821Z,89c25bc,taxi,start,40.688564,-73.965881,1572566426666145821
,,5,2019-02-18T04:17:43.176943164Z,2020-02-18T10:17:43.176943164Z,2019-11-01T13:32:45.632969758Z,89c2624,taxi,start,40.733585,-73.737175,1572615165632969758

#group,false,false,true,true,false,true,true,true,false,false,false
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,string,long,double,double
#default,_result,,,,,,,,,,
,result,table,_start,_stop,_time,s2_cell_id,_measurement,_pt,tid,lon,lat
,,6,2019-02-18T04:17:43.176943164Z,2020-02-18T10:17:43.176943164Z,2019-11-01T00:17:38.287113937Z,89c2664,taxi,start,1572567458287113937,-73.776665,40.645245
"

outData = "
#group,false,false,false,true,true,true,false,false,false,false,false
#datatype,string,long,dateTime:RFC3339,string,string,string,long,double,double,double,double
#default,_result,,,,,,,,,,
,result,table,_time,s2_cell_id,_measurement,_pt,tid,lon,tip,lat,dist
,,0,2019-11-01T13:41:59.331776148Z,89c261c,taxi,end,1572615165632969758,-73.752579,0,40.725647,1.3

#group,false,false,false,true,true,true,false,false,false,false,false
#datatype,string,long,dateTime:RFC3339,string,string,string,double,double,double,long,double
#default,_result,,,,,,,,,,
,result,table,_time,s2_cell_id,_measurement,_pt,lat,dist,tip,tid,lon
,,1,2019-11-01T00:33:07.54916732Z,89c2624,taxi,end,40.728077,9.7,5,1572567458287113937,-73.716583

#group,false,false,false,true,true,true,false,false,false
#datatype,string,long,dateTime:RFC3339,string,string,string,double,double,long
#default,_result,,,,,,,,
,result,table,_time,s2_cell_id,_measurement,_pt,lat,lon,tid
,,2,2019-11-01T13:32:45.632969758Z,89c2624,taxi,start,40.733585,-73.737175,1572615165632969758
"

t_filterRowsPivoted = (table=<-) =>
  table
    |> range(start: 2019-11-01T00:00:00Z)
    |> geo.filterRows(region: {lat: 40.7090214, lon: -73.61846, radius: 15.0}, strict: true)
    |> drop(columns: ["_start", "_stop"])
test _filterRowsPivoted = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_filterRowsPivoted})
