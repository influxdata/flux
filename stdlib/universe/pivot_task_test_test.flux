package universe_test
 
import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,string,string,string,string,string,string
#group,false,false,false,false,true,true,true,true,true
#default,_result,,,,,,,,
,result,table,_time,_value,_field,taskID,orgID,state,_measurement
,,0,2018-10-03T17:55:11.01114Z,1970-01-01T00:00:01Z,requestedAt,02bac3c8f0f37000,02bac3c8d6c5b000,started,records
,,1,2018-10-03T17:55:11.01114Z,02bac3c908f37000,runID,02bac3c8f0f37000,02bac3c8d6c5b000,started,records
,,1,2018-10-03T17:55:11.113222Z,02bac3c922737000,runID,02bac3c8f0f37000,02bac3c8d6c5b000,started,records
,,2,2018-10-03T17:55:10.920529Z,02bac3c8f1737000,runID,02bac3c8f0f37000,02bac3c8d6c5b000,success,records
,,2,2018-10-03T17:55:11.01435Z,02bac3c908f37000,runID,02bac3c8f0f37000,02bac3c8d6c5b000,success,records
,,2,2018-10-03T17:55:11.115415Z,02bac3c922737000,runID,02bac3c8f0f37000,02bac3c8d6c5b000,success,records
,,3,2018-10-03T17:55:11.01114Z,2018-10-03T17:55:12Z,scheduledFor,02bac3c8f0f37000,02bac3c8d6c5b000,started,records
,,3,2018-10-03T17:55:11.113222Z,2018-10-03T17:55:13Z,scheduledFor,02bac3c8f0f37000,02bac3c8d6c5b000,started,records
,,4,2018-10-03T17:55:10.920529Z,2018-10-03T17:55:11Z,scheduledFor,02bac3c8f0f37000,02bac3c8d6c5b000,success,records
,,4,2018-10-03T17:55:11.01435Z,2018-10-03T17:55:12Z,scheduledFor,02bac3c8f0f37000,02bac3c8d6c5b000,success,records
,,4,2018-10-03T17:55:11.115415Z,2018-10-03T17:55:13Z,scheduledFor,02bac3c8f0f37000,02bac3c8d6c5b000,success,records
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,string,string,string,string,string
#group,false,false,true,true,false,true,true,true,true,false,false,false
#default,_result,,,,,,,,,,,
,result,table,_start,_stop,_time,taskID,orgID,state,_measurement,requestedAt,runID,scheduledFor
,,0,2018-10-02T17:55:11.520461Z,2030-01-01T00:00:00Z,2018-10-03T17:55:11.01114Z,02bac3c8f0f37000,02bac3c8d6c5b000,started,records,1970-01-01T00:00:01Z,02bac3c908f37000,2018-10-03T17:55:12Z
,,0,2018-10-02T17:55:11.520461Z,2030-01-01T00:00:00Z,2018-10-03T17:55:11.113222Z,02bac3c8f0f37000,02bac3c8d6c5b000,started,records,,02bac3c922737000,2018-10-03T17:55:13Z
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,string,string,string,string
#group,false,false,true,true,false,true,true,true,true,false,false
#default,_result,,,,,,,,,,
,result,table,_start,_stop,_time,taskID,orgID,state,_measurement,runID,scheduledFor
,,1,2018-10-02T17:55:11.520461Z,2030-01-01T00:00:00Z,2018-10-03T17:55:10.920529Z,02bac3c8f0f37000,02bac3c8d6c5b000,success,records,02bac3c8f1737000,2018-10-03T17:55:11Z
,,1,2018-10-02T17:55:11.520461Z,2030-01-01T00:00:00Z,2018-10-03T17:55:11.01435Z,02bac3c8f0f37000,02bac3c8d6c5b000,success,records,02bac3c908f37000,2018-10-03T17:55:12Z
,,1,2018-10-02T17:55:11.520461Z,2030-01-01T00:00:00Z,2018-10-03T17:55:11.115415Z,02bac3c8f0f37000,02bac3c8d6c5b000,success,records,02bac3c922737000,2018-10-03T17:55:13Z
"

t_pivot_task_test = (table=<-) =>
	(table
		|> range(start: 2018-10-02T17:55:11.520461Z)
		|> filter(fn: (r) =>
			(r._measurement == "records" and r.taskID == "02bac3c8f0f37000"))
		|> pivot(rowKey: ["_time"], columnKey: ["_field"], valueColumn: "_value"))

test _pivot_task_test = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_pivot_task_test})

