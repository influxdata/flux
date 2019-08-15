package pandas_test

import "testing"
import "strings"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,string,string,string,string,string,string,string
#group,false,false,false,false,true,true,true,true,true,true
#default,_result,,,,,,,,,
,result,table,_time,_value,_field,_measurement,device,fstype,host,path
,,0,2018-05-22T19:53:26Z, a ,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:36Z,   k9n  gm ,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:46Z,  b  ,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:56Z,2COTDe   ,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:54:06Z, cLnSkNMI ,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:54:16Z,    13F2,used_percent,disk,disk1,apfs,host.local,/
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,string,string,string,string,string
#group,false,false,true,true,false,false,true,true,true,true,true,true
#default,_result,,,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,device,fstype,host,path
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:26Z,a,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:36Z,k9n  gm,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:46Z,b,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:53:56Z,2COTDe,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:06Z,cLnSkNMI,used_percent,disk,disk1,apfs,host.local,/
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,2018-05-22T19:54:16Z,13F2,used_percent,disk,disk1,apfs,host.local,/
"

t_string_lstrip = (table=<-) =>
	(table
		|> range(start: 2018-05-22T19:53:26Z)
		|> map(fn: (r) =>
        			({r with _value: strings.trimRight(v: strings.trimLeft(v: r._value, cutset: " "), cutset: " ")})))

test _string_lstrip = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_string_lstrip})