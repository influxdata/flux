package testdata_test

import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,string,string,long,long
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_measurement,host,port,bytes
,,0,m,A,80,3524
,,0,m,A,443,7253
,,0,m,B,443,9082
"

outData = "
#datatype,string,long,string,long,long,long
#group,false,false,true,false,false,false
#default,_result,,,,,
,result,table,_measurement,A_80,A_443,B_443
,,0,m,3524,7253,9082
"

t_pivot = (table=<-) =>
	(table
		|> pivot(rowKey: ["_measurement"], columnKey: ["host", "port"], valueColumn: "bytes"))

test _pivot = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_pivot})

// Equivalent TICKscript query:
// stream
//   |flatten()
//     .on('host', 'port')