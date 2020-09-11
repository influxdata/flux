package universe_test

import "testing"
import "csv"
import "experimental"

option now = () => (2030-01-01T00:00:00Z)

// _value column in "times" field are timestamps, starting from
// 1970-01-02, 1970-01-03, etc
inData = "
#datatype,string,long,string,string,dateTime:RFC3339,long,string
#group,false,false,true,true,false,false,true
#default,_result,,,,,,
,result,table,_field,_measurement,_time,_value,city
,,0,times,events,2020-08-11T17:56:20Z,86400000000000,New York
,,1,times,events,2020-08-11T17:56:20Z,172800000000000,Chicago
,,2,times,events,2020-08-11T17:56:20Z,259200000000000,Los Angeles
,,3,times,events,2020-08-11T17:56:20Z,345600000000000,Boston

#datatype,string,long,string,string,dateTime:RFC3339,double,string
#group,false,false,true,true,false,false,true
#default,_result,,,,,,
,result,table,_field,_measurement,_time,_value,city
,,4,temp,city_data,1970-01-02T00:00:10Z,20.0,New York
,,4,temp,city_data,1970-01-02T00:00:20Z,21.0,New York
,,4,temp,city_data,1970-01-02T00:00:30Z,22.0,New York
,,4,temp,city_data,1970-01-02T00:00:40Z,23.0,New York
,,4,temp,city_data,1970-01-02T00:00:50Z,24.0,New York

,,5,temp,city_data,1970-01-03T00:00:10Z,18.0,Chicago
,,5,temp,city_data,1970-01-03T00:00:20Z,19.0,Chicago
,,5,temp,city_data,1970-01-03T00:00:30Z,20.0,Chicago
,,5,temp,city_data,1970-01-03T00:00:40Z,21.0,Chicago
,,5,temp,city_data,1970-01-03T00:00:50Z,22.0,Chicago

,,6,temp,city_data,1970-01-04T00:00:10Z,47.0,Los Angeles
,,6,temp,city_data,1970-01-04T00:00:20Z,48.0,Los Angeles
,,6,temp,city_data,1970-01-04T00:00:30Z,49.0,Los Angeles
,,6,temp,city_data,1970-01-04T00:00:40Z,50.0,Los Angeles
,,6,temp,city_data,1970-01-04T00:00:50Z,51.0,Los Angeles
,,6,temp,city_data,1970-01-04T01:01:00Z,52.0,Los Angeles
,,6,temp,city_data,1970-01-04T01:01:10Z,53.0,Los Angeles
,,6,temp,city_data,1970-01-04T01:01:20Z,54.0,Los Angeles
,,6,temp,city_data,1970-01-04T01:01:30Z,55.0,Los Angeles
,,6,temp,city_data,1970-01-04T01:01:40Z,56.0,Los Angeles

,,7,temp,city_data,1970-01-05T00:00:10Z,15.0,Boston
,,7,temp,city_data,1970-01-05T00:00:20Z,16.0,Boston
,,7,temp,city_data,1970-01-05T00:00:30Z,17.0,Boston
,,7,temp,city_data,1970-01-05T00:00:40Z,18.0,Boston
,,7,temp,city_data,1970-01-05T00:00:50Z,19.0,Boston

,,8,temp,city_data,1970-01-06T00:00:10Z,65.0,Austin
,,8,temp,city_data,1970-01-06T00:00:20Z,66.0,Austin
,,8,temp,city_data,1970-01-06T00:00:30Z,67.0,Austin
,,8,temp,city_data,1970-01-06T00:00:40Z,68.0,Austin
,,8,temp,city_data,1970-01-06T00:00:50Z,69.0,Austin
"

outData = "
#datatype,string,long,string,dateTime:RFC3339,double,string
#group,false,false,true,false,false,true
#default,_result,,,,,
,result,table,_field,_time,_value,city
,,0,event_temp_mean,1970-01-05T01:00:00Z,17,Boston
,,1,event_temp_mean,1970-01-03T01:00:00Z,20,Chicago
,,2,event_temp_mean,1970-01-04T01:00:00Z,49,Los Angeles
,,3,event_temp_mean,1970-01-02T01:00:00Z,22,New York
"

t_table_fns_findcolumn_map = (table=<-) =>
  table
  |> range(start: 2020-08-10T00:00:00Z)
  |> filter(fn: (r) => r._measurement == "events" and r._field == "times")
  |> map(fn: (r) => {
    start = time(v: r._value)
    stop = experimental.addDuration(to: start, d: 1h)
    city = r.city
    agg = csv.from(csv: inData)
      |> range(start, stop)
      |> filter(fn: (r) => r._measurement == "city_data" and r._field == "temp" and r.city == city)
      |> mean()
      |> findColumn(fn: (key) => true, column: "_value")
    return {city: r.city, _time: stop, _value: agg[0], _field: "event_temp_mean"}
  })

test _table_fns_findcolumn_map = () =>
  ({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_table_fns_findcolumn_map})
