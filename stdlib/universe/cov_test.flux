package universe_test
 
import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,double,string,string,string
#group,false,false,false,false,true,true,true
#default,_result,,,,,,
,result,table,_time,_value,_measurement,user,_field
,,0,2018-05-22T19:53:26Z,0,CPU,user1,f1
,,0,2018-05-22T19:53:36Z,1,CPU,user1,f1
,,1,2018-05-22T19:53:26Z,4,CPU,user2,f1
,,1,2018-05-22T19:53:36Z,20,CPU,user2,f1
,,1,2018-05-22T19:53:46Z,7,CPU,user2,f1
,,1,2018-05-22T19:53:56Z,10,CPU,user2,f1
,,2,2018-05-22T19:53:26Z,1,RAM,user1,f1
,,2,2018-05-22T19:53:36Z,2,RAM,user1,f1
,,2,2018-05-22T19:53:46Z,3,RAM,user1,f1
,,2,2018-05-22T19:53:56Z,5,RAM,user1,f1
,,3,2018-05-22T19:53:26Z,2,RAM,user2,f1
,,3,2018-05-22T19:53:36Z,4,RAM,user2,f1
,,3,2018-05-22T19:53:46Z,4,RAM,user2,f1
,,3,2018-05-22T19:53:56Z,0,RAM,user2,f1
,,3,2018-05-22T19:54:06Z,2,RAM,user2,f1
,,3,2018-05-22T19:54:16Z,10,RAM,user2,f1
"

outData = "
#datatype,string,long,string,double
#group,false,false,true,false
#default,_result,,,
,result,table,_measurement,_value
,,0,CPU,8
,,1,RAM,-1.8333333333333333
"

t_cov = (table=<-) => {
    t1 = table
		|> range(start: 2018-05-22T19:53:00Z, stop: 2018-05-22T19:55:00Z)
		|> drop(columns: ["_start", "_stop"])
		|> filter(fn: (r) => r.user == "user1")
		|> group(columns: ["_measurement"])

    t2 = table
		|> range(start: 2018-05-22T19:53:00Z, stop: 2018-05-22T19:55:00Z)
		|> drop(columns: ["_start", "_stop"])
		|> filter(fn: (r) => r.user == "user2")
		|> group(columns: ["_measurement"])

    return cov(x: t1, y: t2, on: ["_time", "_measurement"])
}

test _cov = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_cov})
