package testdata_test
 
import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,string,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,tag0
,,0,2018-05-22T19:54:16Z,20,f0,a
,,0,2018-05-22T19:53:56Z,55,f0,a
,,0,2018-05-22T19:54:06Z,20,f0,a
,,1,2018-05-22T19:53:26Z,35,f0,b
,,1,2018-05-22T19:53:46Z,70,f0,b
,,2,2018-05-22T19:53:36Z,15,f0,c
,,2,2018-05-22T19:54:16Z,11,f1,c
,,2,2018-05-22T19:53:56Z,99,f1,c
,,2,2018-05-22T19:54:06Z,85,f1,c
,,3,2018-05-22T19:53:26Z,23,f1,d
,,4,2018-05-22T19:53:46Z,37,f1,e
,,4,2018-05-22T19:53:36Z,69,f1,e
"

outData = "
#datatype,string,long,dateTime:RFC3339,string,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,tag0
,,0,2018-05-22T19:54:16Z,20,f0,a
,,1,2018-05-22T19:53:26Z,35,f0,b
,,2,2018-05-22T19:53:36Z,15,f0,c
,,3,2018-05-22T19:53:26Z,23,f1,d
,,4,2018-05-22T19:53:46Z,37,f1,e
"

t_unique = (table=<-) =>
	(table
		|> unique(column: "tag0"))

test _unique = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_unique})

