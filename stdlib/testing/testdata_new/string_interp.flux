package main
import "testing"

option now = () =>
	(2030-01-01T00:00:00Z)

inData = "
#datatype,string,string
#group,true,true
#default,,
,error,reference
,failed to execute query: failed to initialize execute state: missing expected annotation group,
"
outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string,string,string
#partition,false,false,false,false,false,false,true,true,true,true,true,true
#default,_result,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,path
,,0,2018-05-22T19:53:24.421470485Z,2018-05-22T19:54:24.421470485Z,2018-05-22T19:53:56Z,34.982447293755506,field1,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T19:53:24.421470485Z,2018-05-22T19:54:24.421470485Z,2018-05-22T19:54:06Z,34.98204153981662,field1,disk,disk1s1,apfs,host.local,/
,,0,2018-05-22T19:53:24.421470485Z,2018-05-22T19:54:24.421470485Z,2018-05-22T19:54:16Z,34.982252364543626,field1,disk,disk1s1,apfs,host.local,/
"
n = 1
fieldSelect = "field{n}"
t_string_interp = (table=<-) =>
	(table
		|> range(start: 2018-05-22T19:53:26Z)
		|> filter(fn: (r) =>
			(r._field == fieldSelect)))

test string_interp = {input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_string_interp}