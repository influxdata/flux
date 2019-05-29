package testdata_test

import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,double,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_measurement,user
,,0,2018-05-22T19:53:26Z,0,CPU,user1
,,0,2018-05-22T19:53:36Z,1,CPU,user1
,,1,2018-05-22T19:53:26Z,4,CPU,user2
,,1,2018-05-22T19:53:36Z,20,CPU,user2
,,1,2018-05-22T19:53:46Z,7,CPU,user2
,,2,2018-05-22T19:53:26Z,1,RAM,user1"

outData = "
#datatype,string,long,dateTime:RFC3339,double,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_measurement,user
,,0,2018-05-22T19:53:26Z,0,CPU,user1
,,0,2018-05-22T19:53:36Z,1,CPU,user1
,,1,2018-05-22T19:53:26Z,1,RAM,user1
"

t_dynamic = (table=<-) => {
    r = table |> range(start: 2018-05-22T19:53:26Z) |> drop(columns: ["_start", "_stop"])
    t = r |> tableFind(fn: (key) => key._measurement == "CPU")
    users = t |> getColumn(column: "user")

    // This is a dynamic query, because this query uses the results of another one.
    return r |> filter(fn: (r) => contains(value: r.user, set: users))
}

test _dynamic_query = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_dynamic})