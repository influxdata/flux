package universe_test
 
import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,string,string,string
#group,false,false,false,true,false,false
#default,_result,,,,,
,result,table,_time,_measurement,host,name
,,0,2018-05-22T19:53:30Z,_m0,h0,n0
,,0,2018-05-22T19:53:40Z,_m0,h0,n0
,,0,2018-05-22T19:53:50Z,_m0,h0,n0
,,0,2018-05-22T19:53:00Z,_m0,h0,n0
,,0,2018-05-22T19:53:10Z,_m0,h1,n1
,,0,2018-05-22T19:53:20Z,_m0,h1,n1
,,1,2018-05-22T19:53:30Z,_m1,h1,n1
,,1,2018-05-22T19:53:40Z,_m1,h1,n1
,,1,2018-05-22T19:53:50Z,_m1,h2,n2
,,1,2018-05-22T19:54:00Z,_m1,h1,n1
,,1,2018-05-22T19:54:10Z,_m1,h1,n1
,,1,2018-05-22T19:54:30Z,_m1,h1,n1
,,1,2018-05-22T19:54:40Z,_m1,h3,n3
,,1,2018-05-22T19:53:50Z,_m1,h3,n2
,,1,2018-05-22T19:54:00Z,_m1,h3,n3
,,2,2018-05-22T19:53:10Z,_m2,h3,n3
,,2,2018-05-22T19:53:30Z,_m2,h5,n5
,,2,2018-05-22T19:54:40Z,_m2,h5,n5
,,2,2018-05-22T19:53:50Z,_m2,h5,n5
,,3,2018-05-22T19:54:00Z,_m3,h5,n5
,,3,2018-05-22T19:54:10Z,_m3,h4,n1
,,3,2018-05-22T19:54:20Z,_m3,h4,n4
"

outData = "
#datatype,string,long,string,string,string
#group,false,false,true,false,false
#default,_result,,,,
,result,table,_measurement,_key,_value
,,0,_m0,host,h0
,,0,_m0,name,n0
,,0,_m0,host,h1
,,0,_m0,name,n1
,,1,_m1,host,h1
,,1,_m1,name,n1
,,1,_m1,host,h2
,,1,_m1,name,n2
,,1,_m1,host,h3
,,1,_m1,name,n3
,,2,_m2,host,h3
,,2,_m2,name,n3
,,2,_m2,host,h5
,,2,_m2,name,n5
,,3,_m3,host,h5
,,3,_m3,name,n5
,,3,_m3,host,h4
,,3,_m3,name,n1
,,3,_m3,name,n4
"

t_key_values_host_name = (table=<-) =>
	(table
		|> keyValues(keyColumns: ["host", "name"]))

test _key_values_host_name = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_key_values_host_name})

