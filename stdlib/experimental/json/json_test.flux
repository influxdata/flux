package json_test


import "experimental/json"
import "testing"

option now = () =>
	(2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,string,string,string
#group,false,false,false,true,true,false
#default,_result,,,,,
,result,table,_time,_field,_measurement,_value
,,0,2018-05-22T19:53:26Z,json,m,\"{\"\"a\"\":1,\"\"b\"\":2,\"\"c\"\":3}\"
,,0,2018-05-22T19:53:36Z,json,m,\"{\"\"a\"\":2,\"\"b\"\":4,\"\"c\"\":6}\"
,,0,2018-05-22T19:53:46Z,json,m,\"{\"\"a\"\":3,\"\"b\"\":5,\"\"c\"\":7}\"
"
outData = "
#datatype,string,long,dateTime:RFC3339,double,double,double
#group,false,false,false,false,false,false
#default,_result,,,,,
,result,table,_time,a,b,c
,,0,2018-05-22T19:53:26Z,1,2,3
,,0,2018-05-22T19:53:36Z,2,4,6
,,0,2018-05-22T19:53:46Z,3,5,7
"
_json = (table=<-) =>
	(table
		|> range(start: 2018-05-22T19:53:26Z)
		|> map(fn: (r) => {
			data = json.parse(data: bytes(v: r._value))
			return {
				_time: r._time,
				a: data.a,
				b: data.b,
				c: data.c,
			}
		}))

test parse = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: _json})
