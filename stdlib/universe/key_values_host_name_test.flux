package universe_test
 
import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,string,string,string,string
#group,false,false,false,true,false,false,true
#default,_result,,,,,,
,result,table,_time,_measurement,host,name,_field
,,0,2018-05-22T19:53:30Z,_m0,h0,n0,ff
,,0,2018-05-22T19:53:40Z,_m0,h0,n0,ff
,,0,2018-05-22T19:53:50Z,_m0,h0,n0,ff
,,0,2018-05-22T19:53:00Z,_m0,h0,n0,ff
,,0,2018-05-22T19:53:10Z,_m0,h1,n1,ff
,,0,2018-05-22T19:53:20Z,_m0,h1,n1,ff
,,1,2018-05-22T19:53:30Z,_m1,h1,n1,ff
,,1,2018-05-22T19:53:40Z,_m1,h1,n1,ff
,,1,2018-05-22T19:53:50Z,_m1,h2,n2,ff
,,1,2018-05-22T19:54:00Z,_m1,h1,n1,ff
,,1,2018-05-22T19:54:10Z,_m1,h1,n1,ff
,,1,2018-05-22T19:54:30Z,_m1,h1,n1,ff
,,1,2018-05-22T19:54:40Z,_m1,h3,n3,ff
,,1,2018-05-22T19:53:50Z,_m1,h3,n2,ff
,,1,2018-05-22T19:54:00Z,_m1,h3,n3,ff
,,2,2018-05-22T19:53:10Z,_m2,h3,n3,ff
,,2,2018-05-22T19:53:30Z,_m2,h5,n5,ff
,,2,2018-05-22T19:54:40Z,_m2,h5,n5,ff
,,2,2018-05-22T19:53:50Z,_m2,h5,n5,ff
,,3,2018-05-22T19:54:00Z,_m3,h5,n5,ff
,,3,2018-05-22T19:54:10Z,_m3,h4,n1,ff
,,3,2018-05-22T19:54:20Z,_m3,h4,n4,ff
"

outData = "
#datatype,string,long,string,string,string,string
#group,false,false,true,false,false,true
#default,_result,,,,,
,result,table,_measurement,_key,_value,_field
,,0,_m0,host,h0,ff
,,0,_m0,name,n0,ff
,,0,_m0,host,h1,ff
,,0,_m0,name,n1,ff
,,1,_m1,host,h1,ff
,,1,_m1,name,n1,ff
,,1,_m1,host,h2,ff
,,1,_m1,name,n2,ff
,,1,_m1,host,h3,ff
,,1,_m1,name,n3,ff
,,2,_m2,host,h3,ff
,,2,_m2,name,n3,ff
,,2,_m2,host,h5,ff
,,2,_m2,name,n5,ff
,,3,_m3,host,h5,ff
,,3,_m3,name,n5,ff
,,3,_m3,host,h4,ff
,,3,_m3,name,n1,ff
,,3,_m3,name,n4,ff
"

t_key_values_host_name = (table=<-) =>
	(table
		|> keyValues(keyColumns: ["host", "name"]))
		|> drop(columns: ["_start", "_stop"])

test _key_values_host_name = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_key_values_host_name})

