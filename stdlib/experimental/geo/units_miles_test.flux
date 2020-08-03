package geo_test

import "experimental/geo"
import "influxdata/influxdb/v1"
import "testing"

option now = () => (2030-01-01T00:00:00Z)

// change units
option geo.units = {distance:"mile"}

inData = "
#group,false,false,false,false,true,true,true,true,true,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,double,string,string,string,string,string,string,string,string,string,string
#default,_result,,,,,,,,,,,,,
,result,table,_time,_value,_field,_measurement,_pt,area,id,s2_cell_id,seq_idx,status,stop_id,trip_id
,,0,2020-04-08T15:44:58Z,40.820317,lat,mta,start,LLIR,GO506_20_6431,89c288c54,1,STOPPED_AT,171,GO506_20_6431
,,1,2020-04-08T16:19:27Z,40.745249,lat,mta,via,LLIR,GO506_20_6431,89c2592bc,13,IN_TRANSIT_TO,237,GO506_20_6431
,,2,2020-04-08T16:16:50Z,40.751085,lat,mta,via,LLIR,GO506_20_6431,89c25f18c,13,IN_TRANSIT_TO,237,GO506_20_6431
,,3,2020-04-08T15:44:58Z,-73.68691,lon,mta,start,LLIR,GO506_20_6431,89c288c54,1,STOPPED_AT,171,GO506_20_6431
,,4,2020-04-08T16:19:27Z,-73.940563,lon,mta,via,LLIR,GO506_20_6431,89c2592bc,13,IN_TRANSIT_TO,237,GO506_20_6431
,,5,2020-04-08T16:16:50Z,-73.912119,lon,mta,via,LLIR,GO506_20_6431,89c25f18c,13,IN_TRANSIT_TO,237,GO506_20_6431

#group,false,false,false,false,true,true,true,true,true,true,true,true,true,true
#datatype,string,long,dateTime:RFC3339,long,string,string,string,string,string,string,string,string,string,string
#default,_result,,,,,,,,,,,,,
,result,table,_time,_value,_field,_measurement,_pt,area,id,s2_cell_id,seq_idx,status,stop_id,trip_id
,,6,2020-04-08T15:44:58Z,1586304000,tid,mta,start,LLIR,GO506_20_6431,89c288c54,1,STOPPED_AT,171,GO506_20_6431
,,7,2020-04-08T16:19:27Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c2592bc,13,IN_TRANSIT_TO,237,GO506_20_6431
,,8,2020-04-08T16:16:50Z,1586304000,tid,mta,via,LLIR,GO506_20_6431,89c25f18c,13,IN_TRANSIT_TO,237,GO506_20_6431
"

outData = "
#group,false,false,true,true,false,false,true,true,false,false,true,true,true,true,false,true
#datatype,string,long,string,string,double,dateTime:RFC3339,string,string,double,double,string,string,string,string,long,string
#default,_result,,,,,,,,,,,,,,,
,result,table,_measurement,_pt,_st_distance,_time,area,id,lat,lon,s2_cell_id,seq_idx,status,stop_id,tid,trip_id
,,0,mta,start,20.793,2020-04-08T15:44:58Z,LLIR,GO506_20_6431,40.820317,-73.68691,89c288c54,1,STOPPED_AT,171,1586304000,GO506_20_6431
,,1,mta,via,6.68,2020-04-08T16:19:27Z,LLIR,GO506_20_6431,40.745249,-73.940563,89c2592bc,13,IN_TRANSIT_TO,237,1586304000,GO506_20_6431
,,2,mta,via,8.144,2020-04-08T16:16:50Z,LLIR,GO506_20_6431,40.751085,-73.912119,89c25f18c,13,IN_TRANSIT_TO,237,1586304000,GO506_20_6431
"

// reference point (Statue of Liberty)
refPoint = {lat: 40.6892, lon: -74.0445}

// limit float to 3 decimal places
limitFloat = (value) =>
  (float(v: int(v: value * 1000.0)) / 1000.0)

t_stDistanceInMiles = (table=<-) =>
  table
    |> v1.fieldsAsCols()
    |> map(fn: (r) => ({
      r with _st_distance: limitFloat(value: geo.ST_Distance(region: refPoint, geometry: {lat: r.lat, lon: r.lon}))
    }))
    |> drop(columns: ["_start", "_stop"])
test _stDistanceInMiles = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_stDistanceInMiles})
