package testdata_test

import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,string,string
#group,false,false,false,false,true
#default,_result,,,,
,result,table,_time,_value,_field
,,0,2018-05-22T19:53:26Z,2018-05-22T19:53:26Z,k
,,0,2018-05-22T19:53:26Z,2018-05-22T19:53:26.033Z,k
,,0,2018-05-22T19:53:26Z,2018-05-22T19:53:26.033066Z,k
,,0,2018-05-22T19:53:26Z,2018-05-22T20:00:00+01:00,k
,,0,2018-05-22T19:53:26Z,2018-05-22T20:00:00.000+01:00,k
"

// NOTE: This test will fail with differences in the last two rows when time zone support arrives.
outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,dateTime:RFC3339
#group,false,false,true,true,true,false
#default,got,,,,,
,result,table,_start,_stop,_field,_value
,,0,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,k,2018-05-22T19:53:26Z
,,0,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,k,2018-05-22T19:53:26.033Z
,,0,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,k,2018-05-22T19:53:26.033066Z
,,0,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,k,2018-05-22T19:00:00Z
,,0,2018-05-22T19:52:00Z,2030-01-01T00:00:00Z,k,2018-05-22T19:00:00Z
"

t_toTime = (table=<-) =>
	(table
		|> range(start: 2018-05-22T19:52:00Z)
		|> toTime()

test _toTime = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_toTime})

