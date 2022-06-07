package promql_test


import "testing"
import "internal/promql"
import "csv"

option now = () => 2030-01-01T00:00:00Z

outData =
    "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double
#group,false,true,true,true,false,false
#default,_result,0,1970-01-01T00:00:00Z,1970-01-01T00:00:00Z,,
,result,table,_start,_stop,_time,_value
"

testcase emptyTable {
    table = promql.emptyTable()
    got = table
    want = csv.from(csv: outData)

    testing.diff(got, want)
}
