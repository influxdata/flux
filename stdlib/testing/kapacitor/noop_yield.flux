package testdata_test

import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,double,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_measurement,_field
,,0,2018-05-22T19:53:26Z,0,CPU,user1
,,0,2018-05-22T19:53:36Z,1,CPU,user1
,,1,2018-05-22T19:53:26Z,4,CPU,user2
,,1,2018-05-22T19:53:36Z,20,CPU,user2
,,1,2018-05-22T19:53:46Z,7,CPU,user2
,,1,2018-05-22T19:53:56Z,10,CPU,user2
,,2,2018-05-22T19:53:26Z,1,RAM,user1
,,2,2018-05-22T19:53:36Z,2,RAM,user1
,,2,2018-05-22T19:53:46Z,3,RAM,user1
,,2,2018-05-22T19:53:56Z,5,RAM,user1
,,3,2018-05-22T19:53:26Z,2,RAM,user2
,,3,2018-05-22T19:53:36Z,4,RAM,user2
,,3,2018-05-22T19:53:46Z,4,RAM,user2
,,3,2018-05-22T19:53:56Z,0,RAM,user2
,,3,2018-05-22T19:54:06Z,2,RAM,user2
,,3,2018-05-22T19:54:16Z,10,RAM,user2
"

outData = "
#datatype,string,long,dateTime:RFC3339,double,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_measurement,_field
,_results,0,2018-05-22T19:53:26Z,0,CPU,user1
,_results,0,2018-05-22T19:53:36Z,1,CPU,user1
,_results,1,2018-05-22T19:53:26Z,4,CPU,user2
,_results,1,2018-05-22T19:53:36Z,20,CPU,user2
,_results,1,2018-05-22T19:53:46Z,7,CPU,user2
,_results,1,2018-05-22T19:53:56Z,10,CPU,user2
,_results,2,2018-05-22T19:53:26Z,1,RAM,user1
,_results,2,2018-05-22T19:53:36Z,2,RAM,user1
,_results,2,2018-05-22T19:53:46Z,3,RAM,user1
,_results,2,2018-05-22T19:53:56Z,5,RAM,user1
,_results,3,2018-05-22T19:53:26Z,2,RAM,user2
,_results,3,2018-05-22T19:53:36Z,4,RAM,user2
,_results,3,2018-05-22T19:53:46Z,4,RAM,user2
,_results,3,2018-05-22T19:53:56Z,0,RAM,user2
,_results,3,2018-05-22T19:54:06Z,2,RAM,user2
,_results,3,2018-05-22T19:54:16Z,10,RAM,user2
"

t_noop_yield = (table=<-) =>
    (table
        |> range(start: 2018-05-15T00:00:00Z)
        |> drop(columns: ["_start", "_stop"]))
// yield() is implicit here

test _noop_yield = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_noop_yield})

// In TICKscript, noOp is implicit (NoOpNode is automatically appended to any node that is a source for a StatsNode)