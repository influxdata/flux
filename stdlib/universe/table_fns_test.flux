package universe_test

import "csv"

option now = () => 2020-02-22T18:00:00Z

csvdata ="
#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,location,state
,,0,2018-04-06T10:49:41.565Z,2020-04-06T11:49:41.564Z,2020-02-22T15:01:00Z,50,bottom_degrees,h2o_temperature,santa_monica,CA
,,0,2018-04-06T10:49:41.565Z,2020-04-06T11:49:41.564Z,2020-02-22T15:31:00Z,49,bottom_degrees,h2o_temperature,santa_monica,CA
,,0,2018-04-06T10:49:41.565Z,2020-04-06T11:49:41.564Z,2020-02-22T16:01:00Z,49,bottom_degrees,h2o_temperature,santa_monica,CA
,,0,2018-04-06T10:49:41.565Z,2020-04-06T11:49:41.564Z,2020-02-22T16:31:00Z,49,bottom_degrees,h2o_temperature,santa_monica,CA
,,0,2018-04-06T10:49:41.565Z,2020-04-06T11:49:41.564Z,2020-02-22T17:01:00Z,48,bottom_degrees,h2o_temperature,santa_monica,CA
,,0,2018-04-06T10:49:41.565Z,2020-04-06T11:49:41.564Z,2020-02-22T17:31:00Z,48,bottom_degrees,h2o_temperature,santa_monica,CA
,,0,2018-04-06T10:49:41.565Z,2020-04-06T11:49:41.564Z,2020-02-22T17:46:00Z,48,bottom_degrees,h2o_temperature,santa_monica,CA
"

data = csv.from( csv: csvdata )
	|> range( start: -3h )

col = data
	|> tableFind( fn: (key) => (true) )
	|> getColumn( column: "_value" )

t_now = (table=<-) =>
	(table
		|> filter( fn: (r) => ( contains( value: r._value, set: col ) ) ) ) 

test _sum = () =>
	({input: data, want: data, fn: t_now})
