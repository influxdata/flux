package universe_test

import "testing"

option now = () => (2030-01-01T00:00:00Z)

input = "
#datatype,string,long,string,string,dateTime:RFC3339,double
#group,false,false,true,true,false,false
#default,_result,0,Wx8gaJ,U2FstZ,,
,result,table,foo,bar,_time,_value

#datatype,string,long,string,string,dateTime:RFC3339,double
#group,false,false,true,true,false,false
#default,_result,1,U37BFq3,KYxg,,
,result,table,foo,bar,_time,_value

#datatype,string,long,string,string,dateTime:RFC3339,double
#group,false,false,true,true,false,false
#default,_result,2,U37BFq3,U2FstZ,,
,result,table,foo,bar,_time,_value

#datatype,string,long,string,string,dateTime:RFC3339,double
#group,false,false,true,true,false,false
#default,_result,3,Wx8gaJ,KYxg,,
,result,table,foo,bar,_time,_value
"

output = "
#datatype,string,long,string,string,dateTime:RFC3339,double
#group,false,false,true,true,false,false
#default,_result,0,Wx8gaJ,U2FstZ,,
,result,table,foo,bar,_time,_value

#datatype,string,long,string,string,dateTime:RFC3339,double
#group,false,false,true,true,false,false
#default,_result,1,U37BFq3,KYxg,,
,result,table,foo,bar,_time,_value

#datatype,string,long,string,string,dateTime:RFC3339,double
#group,false,false,true,true,false,false
#default,_result,2,U37BFq3,U2FstZ,,
,result,table,foo,bar,_time,_value

#datatype,string,long,string,string,dateTime:RFC3339,double
#group,false,false,true,true,false,false
#default,_result,3,Wx8gaJ,KYxg,,
,result,table,foo,bar,_time,_value
"

order_fn = (tables=<-) => tables
    |> range(start: 2018-05-22T19:53:26Z)
    |> group(columns: ["_time"])
    |> filter(fn: (r) => r["_field"] == "load4", onEmpty: "drop")

test order_evaluate = () =>
    ({input: testing.loadStorage(csv: input), want: testing.loadMem(csv: output), fn: order_fn})