package promql_test

import "testing"
import "internal/promql"
import c "csv"

option now = () =>
	(2030-01-01T00:00:00Z)
// todo(faith): remove overload https://github.com/influxdata/flux/issues/3155
option testing.loadStorage = (csv) => c.from(csv: csv)
    |> map(fn: (r) => ({r with
    _field: if exists r._field then r._field else die(msg: "test input table does not have _field column"),
    _measurement: if exists r._measurement then r._measurement else die(msg: "test input table does not have _measurement column"),
    _time: if exists r._time then r._time else die(msg: "test input table does not have _time column")
    }))

inData = "
#datatype,string,long,dateTime:RFC3339,string,double,string
#group,false,false,false,true,false,true
#default,inData,,,,,
,result,table,_time,_field,_value,_measurement
,,0,2018-12-18T20:52:33Z,metric_name1,1,prometheus
,,0,2018-12-18T20:52:43Z,metric_name1,1,prometheus
,,0,2018-12-18T20:52:53Z,metric_name1,1,prometheus
,,0,2018-12-18T20:53:03Z,metric_name1,1,prometheus
,,0,2018-12-18T20:53:13Z,metric_name1,1,prometheus
,,0,2018-12-18T20:53:23Z,metric_name1,1,prometheus
,,1,2018-12-18T20:52:33Z,metric_name2,1,prometheus
,,1,2018-12-18T20:52:43Z,metric_name2,1,prometheus
,,1,2018-12-18T20:52:53Z,metric_name2,1,prometheus
,,1,2018-12-18T20:53:03Z,metric_name2,100,prometheus
,,1,2018-12-18T20:53:13Z,metric_name2,100,prometheus
,,1,2018-12-18T20:53:23Z,metric_name2,100,prometheus
,,2,2018-12-18T20:52:33Z,metric_name3,100,prometheus
,,2,2018-12-18T20:52:43Z,metric_name3,200,prometheus
,,2,2018-12-18T20:52:53Z,metric_name3,300,prometheus
,,2,2018-12-18T20:53:03Z,metric_name3,200,prometheus
,,2,2018-12-18T20:53:13Z,metric_name3,300,prometheus
,,2,2018-12-18T20:53:23Z,metric_name3,400,prometheus
"
outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,double,string
#group,false,false,true,true,true,false,true
#default,outData,,,,,,
,result,table,_start,_stop,_field,_value,_measurement
,,0,2018-12-18T20:50:00Z,2018-12-18T20:55:00Z,metric_name1,0,prometheus
,,1,2018-12-18T20:50:00Z,2018-12-18T20:55:00Z,metric_name2,0.39599999999999996,prometheus
,,2,2018-12-18T20:50:00Z,2018-12-18T20:55:00Z,metric_name3,1.2,prometheus
"
t_extrapolatedRate = (table=<-) =>
	(table
		|> range(start: 2018-12-18T20:50:00Z, stop: 2018-12-18T20:55:00Z)
		|> promql.extrapolatedRate(isCounter: false, isRate: true))

test _extrapolatedRate = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_extrapolatedRate})
