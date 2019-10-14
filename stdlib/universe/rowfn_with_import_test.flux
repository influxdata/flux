package universe_test

import "testing"
import "strings"
import "math"

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
#datatype,string,long,dateTime:RFC3339,string,string,string,string
#group,false,false,false,false,true,true,true
#default,_result,,,,,,
,result,table,_time,_value,_field,_measurement,host
,,0,2018-05-22T19:53:36Z,101I,load1,system,host.local
,,0,2018-05-22T19:53:46Z,102I,load1,system,host.local
"

t_row_fn = (table=<-) =>
	table
	  |> filter(fn: (r) => (float(v: r._value) > math.pow(x: 10.0, y: 2.0)))
		|> map(fn: (r) => ({r with _value: string(v: r._value) + "i"}))
		|> map(fn: (r) => ({r with _value: strings.toUpper(v: r._value)}))

test _map = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_row_fn})

