package testdata_test

import "testing"
import "promql"

option now = () => (2030-01-01T00:00:00Z)

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double
#group,false,true,true,true,false,false
#default,_result,0,1970-01-01T00:00:00Z,1970-01-01T00:00:00Z,,
,result,table,_start,_stop,_time,_value
"

t_emptyTable = (table=<-) => table

test _emptyTable = () =>
	({input: promql.emptyTable(), want: testing.loadMem(csv: outData), fn: t_emptyTable})
