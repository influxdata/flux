package universe_test
//
import "testing"

option now = () =>
	(2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,long,string,string,string
#group,false,false,false,false,true,true,true
#default,_result,,,,,,
,result,table,_time,_value,_field,_measurement,host
,,0,2018-05-22T19:53:26Z,100,load1,system,host.local
,,0,2018-05-22T19:53:36Z,101,load1,system,host.local
,,0,2018-05-22T19:53:46Z,102,load1,system,host.local
"
outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,dateTime:RFC3339,long,long,long,long,long,string,dateTime:RFC3339,long
#group,false,false,true,true,true,true,true,false,false,false,false,false,false,false,false,false
#default,got,,,,,,,,,,,,,,,
,result,table,_start,_stop,_field,_measurement,host,_time,_value,array,boolAdd,floatAdd,intAdd,string,time,uintAdd
,,0,1947-11-13T00:00:00Z,2030-01-01T00:00:00Z,load1,system,host.local,2018-05-22T19:53:26Z,100,101,101,101,99,1,2018-05-22T19:53:26Z,101
,,0,1947-11-13T00:00:00Z,2030-01-01T00:00:00Z,load1,system,host.local,2018-05-22T19:53:36Z,101,102,102,102,100,1,2018-05-22T19:53:26Z,102
,,0,1947-11-13T00:00:00Z,2030-01-01T00:00:00Z,load1,system,host.local,2018-05-22T19:53:46Z,102,103,103,103,101,1,2018-05-22T19:53:26Z,103
"
obj = {
	b: true,
	i: -1,
	d: 1.0,
	u: 1,
	s: "1",
	t: 2018-05-22T19:53:26Z,
	r: -30000d,
}
arr = [1, 2, 3, 4]
t_map = (table=<-) =>
	(table
		|> range(start: obj.r)
		|> map(fn: (r) =>
			({r with
				boolAdd: int(v: obj.b) + r._value,
				intAdd: obj.i + r._value,
				floatAdd: int(v: obj.d) + r._value,
				uintAdd: int(v: obj.u) + r._value,
				string: obj.s,
				time: obj.t,
				array: arr[0] + r._value,
			})))


test _map = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_map})
