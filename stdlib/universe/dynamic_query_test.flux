package universe_test


import "testing"
import "csv"

option now = () => 2030-01-01T00:00:00Z

inData =
    "
#datatype,string,long,dateTime:RFC3339,double,string,string,string
#group,false,false,false,false,true,true,true
#default,_result,,,,,,
,result,table,_time,_value,_measurement,user,_field
,,0,2018-05-22T19:53:26Z,0,CPU,user1,f1
,,0,2018-05-22T19:53:36Z,1,CPU,user1,f1
,,1,2018-05-22T19:53:26Z,4,CPU,user2,f1
,,1,2018-05-22T19:53:36Z,20,CPU,user2,f1
,,1,2018-05-22T19:53:46Z,7,CPU,user2,f1
,,2,2018-05-22T19:53:26Z,1,RAM,user1,f1
"
outData =
    "
#datatype,string,long,dateTime:RFC3339,double,string,string,string
#group,false,false,false,false,true,true,true
#default,_result,,,,,,
,result,table,_time,_value,_measurement,user,_field
,,0,2018-05-22T19:53:26Z,0,CPU,user1,f1
,,0,2018-05-22T19:53:36Z,1,CPU,user1,f1
,,1,2018-05-22T19:53:26Z,1,RAM,user1,f1
"

testcase dynamic_query {
    table = csv.from(csv: inData) |> testing.load()

    r = table |> range(start: 2018-05-22T19:53:26Z) |> drop(columns: ["_start", "_stop"])
    t = r |> tableFind(fn: (key) => key._measurement == "CPU")
    users = t |> getColumn(column: "user")
    got = r |> filter(fn: (r) => contains(value: r.user, set: users))
    want = csv.from(csv: outData)

    testing.diff(got, want)
}
