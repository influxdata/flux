package experimental_test


import "testing"
import "experimental"

data = "
#datatype,string,long,string,string,string,dateTime:RFC3339,boolean
#group,false,false,true,true,true,false,false
#default,_result,,,,,,
,result,table,_measurement,_field,t0,_time,_value
,,0,m0,f0,tagvalue,2018-12-19T22:13:30Z,false
,,0,m0,f0,tagvalue,2018-12-19T22:13:40Z,true
,,0,m0,f0,tagvalue,2018-12-19T22:13:50Z,false
,,0,m0,f0,tagvalue,2018-12-19T22:14:00Z,false
,,0,m0,f0,tagvalue,2018-12-19T22:14:10Z,true
,,0,m0,f0,tagvalue,2018-12-19T22:14:20Z,true
"
input = experimental.table(rows: [{
	_measurement: "m0",
	_field: "f0",
	t0: "tagvalue",
	_time: 2018-12-19T22:13:30Z,
	_value: false,
}, {
	_measurement: "m0",
	_field: "f0",
	t0: "tagvalue",
	_time: 2018-12-19T22:13:40Z,
	_value: true,
}, {
	_measurement: "m0",
	_field: "f0",
	t0: "tagvalue",
	_time: 2018-12-19T22:13:50Z,
	_value: false,
}, {
	_measurement: "m0",
	_field: "f0",
	t0: "tagvalue",
	_time: 2018-12-19T22:14:00Z,
	_value: false,
}, {
	_measurement: "m0",
	_field: "f0",
	t0: "tagvalue",
	_time: 2018-12-19T22:14:10Z,
	_value: true,
}, {
	_measurement: "m0",
	_field: "f0",
	t0: "tagvalue",
	_time: 2018-12-19T22:14:20Z,
	_value: true,
}])
pass = (table=<-) =>
	(table)

test _set = () =>
	({input: input, want: testing.loadMem(csv: data), fn: pass})
