package experimental_test

import "experimental"
import "testing"

inData = "
#datatype,string,long,dateTime:RFC3339,string,string,string,double
#group,false,false,false,true,true,true,false
#default,_result,,,,,,
,result,table,_time,_measurement,_field,tag0,_value
,,0,2018-12-19T22:13:30Z,_m,a,t,1
,,0,2018-12-19T22:13:40Z,_m,a,t,2
,,0,2018-12-19T22:13:50Z,_m,a,t,3
,,0,2018-12-19T22:14:00Z,_m,a,t,4
,,0,2018-12-19T22:14:10Z,_m,a,t,5
,,0,2018-12-19T22:14:20Z,_m,a,t,6
,,1,2018-12-19T22:13:30Z,_m,a,g,2
,,1,2018-12-19T22:13:40Z,_m,a,g,3
,,1,2018-12-19T22:13:50Z,_m,a,g,4
,,1,2018-12-19T22:14:00Z,_m,a,g,5
,,1,2018-12-19T22:14:10Z,_m,a,g,6
,,1,2018-12-19T22:14:20Z,_m,a,g,7

#datatype,string,long,dateTime:RFC3339,string,string,string,double
#group,false,false,false,true,true,true,false
#default,_result,,,,,,
,result,table,_time,_measurement,_field,tag1,_value
,,2,2018-12-19T22:13:30Z,_m,a,t,1
,,2,2018-12-19T22:13:40Z,_m,a,t,2
,,2,2018-12-19T22:13:50Z,_m,a,t,3
,,2,2018-12-19T22:14:00Z,_m,a,t,4
,,2,2018-12-19T22:14:10Z,_m,a,t,5
,,2,2018-12-19T22:14:20Z,_m,a,t,6
,,3,2018-12-19T22:13:30Z,_m,a,g,1
,,3,2018-12-19T22:13:40Z,_m,a,g,2
,,3,2018-12-19T22:13:50Z,_m,a,g,3
,,3,2018-12-19T22:14:00Z,_m,a,g,4
,,3,2018-12-19T22:14:10Z,_m,a,g,5
,,3,2018-12-19T22:14:20Z,_m,a,g,6

#datatype,string,long,dateTime:RFC3339,string,string,string,double
#group,false,false,false,true,true,true,false
#default,_result,,,,,,
,result,table,_time,_measurement,_field,tag0,_value
,,4,2018-12-19T22:13:30Z,_m,b,s,1
,,4,2018-12-19T22:13:40Z,_m,b,s,2
,,4,2018-12-19T22:13:50Z,_m,b,s,3
,,4,2018-12-19T22:14:00Z,_m,b,s,4
,,4,2018-12-19T22:14:10Z,_m,b,s,5
,,4,2018-12-19T22:14:20Z,_m,b,s,6
,,5,2018-12-19T22:13:30Z,_m,b,g,1
,,5,2018-12-19T22:13:40Z,_m,b,g,2
,,5,2018-12-19T22:13:50Z,_m,b,g,3
,,5,2018-12-19T22:14:00Z,_m,b,g,4
,,5,2018-12-19T22:14:10Z,_m,b,g,5
,,5,2018-12-19T22:14:20Z,_m,b,g,6

#datatype,string,long,dateTime:RFC3339,string,string,string,double
#group,false,false,false,true,true,true,false
#default,_result,,,,,,
,result,table,_time,_measurement,_field,tag1,_value
,,6,2018-12-19T22:13:30Z,_m,b,s,1
,,6,2018-12-19T22:13:40Z,_m,b,s,2
,,6,2018-12-19T22:13:50Z,_m,b,s,3
,,6,2018-12-19T22:14:00Z,_m,b,s,4
,,6,2018-12-19T22:14:10Z,_m,b,s,5
,,6,2018-12-19T22:14:20Z,_m,b,s,6
,,7,2018-12-19T22:13:30Z,_m,b,p,1
,,7,2018-12-19T22:13:40Z,_m,b,p,2
,,7,2018-12-19T22:13:50Z,_m,b,p,3
,,7,2018-12-19T22:14:00Z,_m,b,p,4
,,7,2018-12-19T22:14:10Z,_m,b,p,5
,,7,2018-12-19T22:14:20Z,_m,b,p,6
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,double,double
#group,false,false,true,true,false,true,true,false,false
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_measurement,tag0,value_a,value_b
,,0,2018-12-19T00:00:00Z,2018-12-20T00:00:00Z,2018-12-19T22:13:30Z,_m,g,2,1
,,0,2018-12-19T00:00:00Z,2018-12-20T00:00:00Z,2018-12-19T22:13:40Z,_m,g,3,2
,,0,2018-12-19T00:00:00Z,2018-12-20T00:00:00Z,2018-12-19T22:13:50Z,_m,g,4,3
,,0,2018-12-19T00:00:00Z,2018-12-20T00:00:00Z,2018-12-19T22:14:00Z,_m,g,5,4
,,0,2018-12-19T00:00:00Z,2018-12-20T00:00:00Z,2018-12-19T22:14:10Z,_m,g,6,5
,,0,2018-12-19T00:00:00Z,2018-12-20T00:00:00Z,2018-12-19T22:14:20Z,_m,g,7,6
"

join_test_fn = (table=<-) => {
    a = table
        |> range(start: 2018-12-19T00:00:00Z, stop: 2018-12-20T00:00:00Z)
        |> filter(fn: (r) => r._field == "a")
        |> drop(columns: ["_field"])
        |> rename(columns: {_value: "value_a"})

    b = table
        |> range(start: 2018-12-19T00:00:00Z, stop: 2018-12-20T00:00:00Z)
        |> filter(fn: (r) => r._field == "b")
        |> drop(columns: ["_field"])
        |> rename(columns: {_value: "value_b"})

    return experimental.join(left:a, right:b, fn:(left, right) => ({left with value_b: right.value_b}))
}

test experimental_join = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: join_test_fn})
