package date_test

import "testing"
import "date"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,string,string,double
#group,false,false,false,true,true,false
#default,_result,,,,,
,result,table,_time,_measurement,_field,_value
,,0,2018-01-22T19:53:00Z,_m,FF,1
,,0,2019-02-22T19:53:10Z,_m,FF,1
,,0,2020-03-22T19:53:20Z,_m,FF,1
,,0,2021-04-22T19:53:30Z,_m,FF,1
,,0,2022-05-22T19:53:40Z,_m,FF,1
,,0,2023-06-22T19:53:50Z,_m,FF,1
,,1,2024-07-22T19:53:00Z,_m,QQ,1
,,1,2025-08-22T19:53:10Z,_m,QQ,1
,,1,2026-09-22T19:53:20Z,_m,QQ,1
,,1,2027-10-22T19:53:30Z,_m,QQ,1
,,1,2028-11-22T19:53:40Z,_m,QQ,1
,,1,2029-12-22T19:53:50Z,_m,QQ,1
"

outData = "
#group,false,false,true,true,true,true,false,false
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,dateTime:RFC3339,long
#default,_result,,,,,,,
,result,table,_start,_stop,_field,_measurement,_time,_value
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2018-01-22T19:53:00Z,2018
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2019-02-22T19:53:10Z,2019
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2020-03-22T19:53:20Z,2020
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2021-04-22T19:53:30Z,2021
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2022-05-22T19:53:40Z,2022
,,0,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,FF,_m,2023-06-22T19:53:50Z,2023
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,QQ,_m,2024-07-22T19:53:00Z,2024
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,QQ,_m,2025-08-22T19:53:10Z,2025
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,QQ,_m,2026-09-22T19:53:20Z,2026
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,QQ,_m,2027-10-22T19:53:30Z,2027
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,QQ,_m,2028-11-22T19:53:40Z,2028
,,1,2018-01-01T00:00:00Z,2030-01-01T00:00:00Z,QQ,_m,2029-12-22T19:53:50Z,2029
"

t_time_year = (table=<-) =>
	(table
 	    |> range(start: 2018-01-01T00:00:00Z)
 		|> map(fn: (r) => ({r with _value: date.year(t: r._time)})))

  test _time_year = () =>
 	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_time_year})
