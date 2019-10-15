package universe_test
 
import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,string,string,dateTime:RFC3339,unsignedLong
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_measurement,_field,_time,_value
,,0,Sgf,DlXwgrw,2018-12-18T22:11:05Z,70
,,0,Sgf,DlXwgrw,2018-12-18T22:11:15Z,48
,,0,Sgf,DlXwgrw,2018-12-18T22:11:25Z,33
,,0,Sgf,DlXwgrw,2018-12-18T22:11:35Z,63
,,0,Sgf,DlXwgrw,2018-12-18T22:11:45Z,48
,,0,Sgf,DlXwgrw,2018-12-18T22:11:55Z,63
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,unsignedLong
#group,false,false,true,true,true,true,false
#default,_result,,,,,,
,result,table,_start,_stop,_measurement,_field,_value
,,0,2018-12-01T00:00:00Z,2030-01-01T00:00:00Z,Sgf,DlXwgrw,48
,,0,2018-12-01T00:00:00Z,2030-01-01T00:00:00Z,Sgf,DlXwgrw,63
"

t_mode = (table=<-) =>
	(table
		|> range(start: 2018-12-01T00:00:00Z)
		|> mode())

test _mode = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_mode})
