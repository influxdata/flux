package testdata_test

import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,string,string,string,long
#group,false,false,false,true,true,false,false
#default,_result,,,,,,
,result,table,_time,_measurement,_field,port,_value
,,0,2018-05-22T19:53:26Z,m,A,80,3524
,,0,2018-05-22T19:53:26Z,m,A,443,7253
,,0,2018-05-22T19:53:26Z,m,B,443,9082
"

outData = "
#datatype,string,long,dateTime:RFC3339,string,long,long,long
#group,false,false,false,true,false,false,false
#default,_result,,,,,,
,result,table,_time,_measurement,A_80,A_443,B_443
,,0,2018-05-22T19:53:26Z,m,3524,7253,9082
"

t_pivot = (table=<-) =>
	(table
	    |> range(start: 2018-05-15T00:00:00Z)
	    |> drop(columns: ["_start", "_stop"])
		|> pivot(rowKey: ["_time", "_measurement"], columnKey: ["_field", "port"], valueColumn: "_value"))

test _pivot = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_pivot})

// Equivalent TICKscript query:
// stream
//   |flatten()
//     .on('_field', 'port')