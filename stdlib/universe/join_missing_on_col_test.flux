package universe_test

import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,long,string,string,long
#group,false,false,false,false,true,true,true
#default,_result,,,,,,
,result,table,_time,_value,_field,_measurement,status
,_result,0,2019-09-12T20:00:00Z,510,resp_bytes,http_request,400

#datatype,string,long,dateTime:RFC3339,long,string,string,string,long
#group,false,false,false,false,true,true,true,true
#default,_result,,,,,,,
,result,table,_time,_value,_field,_measurement,org_id,status
,_result,1,2019-09-12T20:00:00Z,180654,resp_bytes,http_request,03b60149f6fb9000,200
,_result,2,2019-09-12T20:00:00Z,144,resp_bytes,http_request,03b60149f6fb9000,500
,_result,3,2019-09-12T20:00:00Z,152,resp_bytes,http_request,10603619cf2e665b,500

#datatype,string,long,dateTime:RFC3339,long,string,string,string,long
#group,false,false,false,false,true,true,true,true
#default,_result,,,,,,,
,result,table,_time,_value,_field,_measurement,org_id,status
,_result,0,2019-09-12T20:00:00Z,9654177,req_bytes,http_request,039e9eb802d3f000,204
,_result,1,2019-09-12T20:00:00Z,33784154,req_bytes,http_request,03a2c4456f0f5000,204
,_result,2,2019-09-12T20:00:00Z,870346480,req_bytes,http_request,03b60149f6fb9000,204
,_result,3,2019-09-12T20:00:00Z,48275,req_bytes,http_request,03c68af671227000,204
"

outData = "
#datatype,string,long,long,long,string
#group,false,false,false,false,true
#default,_result,,,,
,result,table,datain,dataout,org_id
,_result,0,9654177,,039e9eb802d3f000
,_result,1,33784154,,03a2c4456f0f5000
,_result,2,870346480,180798,03b60149f6fb9000
,_result,3,48275,,03c68af671227000
,_result,4,,152,10603619cf2e665b
"

t_join_missing_on_col = (tables=<-) => {
    lhs = tables
        |> range(start: 2019-01-01T00:00:00Z)
        |> filter(fn: (r) => r._field == "resp_bytes")
        |> keep(columns: ["_time", "org_id", "_value"])
        |> sum()
        |> rename(columns: {_value: "dataout"})
    rhs = tables
        |> range(start: 2019-01-01T00:00:00Z)
        |> filter(fn: (r) => r._field == "req_bytes")
        |> keep(columns: ["_time", "org_id", "_value"])
        |> sum()
        |> rename(columns: {_value: "datain"})
    return join(tables: {lhs: lhs, rhs: rhs}, on: ["org_id"])
}

test _join_missing_on_col = () =>
    ({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_join_missing_on_col})
