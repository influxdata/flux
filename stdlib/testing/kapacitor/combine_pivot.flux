package testdata_test

import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,string,dateTime:RFC3339,double,string
#group,false,false,true,false,false,true
#default,_result,,,,,
,result,table,_measurement,_time,_value,user
,,0,memory,2018-12-19T22:13:30Z,1,user1
,,0,memory,2018-12-19T22:13:40Z,5,user1
,,0,memory,2018-12-19T22:13:50Z,3,user1
,,0,memory,2018-12-19T22:14:00Z,6,user1
,,0,memory,2018-12-19T22:14:10Z,6,user1
,,0,memory,2018-12-19T22:14:20Z,3,user1
,,1,memory,2018-12-19T22:13:30Z,6,user2
,,1,memory,2018-12-19T22:13:40Z,7,user2
,,1,memory,2018-12-19T22:13:50Z,3,user2
,,1,memory,2018-12-19T22:14:00Z,4,user2
,,1,memory,2018-12-19T22:14:10Z,9,user2
,,1,memory,2018-12-19T22:14:20Z,8,user2
"

outData = "
#datatype,string,long,string,dateTime:RFC3339,double,double
#group,false,false,true,false,false,false
#default,_result,,,,,
,result,table,_measurement,_time,user1,user2
,,0,memory,2018-12-19T22:13:30Z,1,6
,,0,memory,2018-12-19T22:13:40Z,5,7
,,0,memory,2018-12-19T22:13:50Z,3,3
,,0,memory,2018-12-19T22:14:00Z,6,4
,,0,memory,2018-12-19T22:14:10Z,6,9
,,0,memory,2018-12-19T22:14:20Z,3,8
"

t_combine_join = (table=<-) =>
    (table
		|> pivot(rowKey:["_time", "_measurement"], columnKey: ["user"], valueColumn: "_value"))

test _combine_join = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_combine_join})


// Equivalent TICKscript query:
//
// stream
//    |from()
//      .measurement('request_latency')
//    |combine(lambda: "user"=='user1;, lambda: "user"=='f2')
//      .as('user1', 'user2')