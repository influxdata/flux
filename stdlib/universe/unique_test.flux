package universe_test
 
import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,string,string,string,string
#group,false,false,false,false,true,true,true
#default,_result,,,,,,
,result,table,_time,_value,_field,tag0,_measurement
,,0,2018-05-22T19:54:16Z,20,f0,a,aa
,,0,2018-05-22T19:53:56Z,55,f0,a,aa
,,0,2018-05-22T19:54:06Z,20,f0,a,aa
,,1,2018-05-22T19:53:26Z,35,f0,b,aa
,,1,2018-05-22T19:53:46Z,70,f0,b,aa
,,2,2018-05-22T19:53:36Z,15,f1,c,aa
,,2,2018-05-22T19:54:16Z,11,f1,c,aa
,,2,2018-05-22T19:53:56Z,99,f1,c,aa
,,2,2018-05-22T19:54:06Z,85,f1,c,aa
,,3,2018-05-22T19:53:26Z,23,f1,d,aa
,,4,2018-05-22T19:53:46Z,37,f1,e,aa
,,4,2018-05-22T19:53:36Z,69,f1,e,aa
"

outData = "
#datatype,string,long,dateTime:RFC3339,string,string,string,string
#group,false,false,false,false,true,true,true
#default,_result,,,,,,
,result,table,_time,_value,_field,tag0,_measurement
,,0,2018-05-22T19:54:16Z,20,f0,a,aa
,,1,2018-05-22T19:53:26Z,35,f0,b,aa
,,2,2018-05-22T19:53:36Z,15,f1,c,aa
,,3,2018-05-22T19:53:26Z,23,f1,d,aa
,,4,2018-05-22T19:53:46Z,37,f1,e,aa
"

t_unique = (table=<-) =>
	(table
		|> unique(column: "tag0")
		|> drop(columns: ["_start", "_stop"])
	)

test _unique = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_unique})
