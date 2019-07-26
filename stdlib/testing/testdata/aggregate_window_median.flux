package testdata_test

import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,string,string,string,dateTime:RFC3339,double
#group,false,false,true,true,true,false,false
#default,_result,,,,,,
,result,table,_measurement,_field,t0,_time,_value
,,0,Reiva,OAOJWe7,BnR,2019-01-09T19:44:58Z,7.940387008821781
,,0,Reiva,OAOJWe7,BnR,2019-01-09T19:45:08Z,49.460104214779086
,,0,Reiva,OAOJWe7,BnR,2019-01-09T19:45:18Z,-36.564150808873954
,,0,Reiva,OAOJWe7,BnR,2019-01-09T19:45:28Z,34.319039251798635
,,0,Reiva,OAOJWe7,BnR,2019-01-09T19:45:38Z,79.27019811403116
,,0,Reiva,OAOJWe7,BnR,2019-01-09T19:45:48Z,41.91029522104053
,,1,Reiva,OAOJWe7,qCnJDC,2019-01-09T19:44:58Z,-61.68790887989735
,,1,Reiva,OAOJWe7,qCnJDC,2019-01-09T19:45:08Z,-6.3173755351186465
,,1,Reiva,OAOJWe7,qCnJDC,2019-01-09T19:45:18Z,-26.049728557657513
,,1,Reiva,OAOJWe7,qCnJDC,2019-01-09T19:45:28Z,114.285955884979
,,1,Reiva,OAOJWe7,qCnJDC,2019-01-09T19:45:38Z,16.140262630578995
,,1,Reiva,OAOJWe7,qCnJDC,2019-01-09T19:45:48Z,29.50336437998469
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,dateTime:RFC3339,double
#group,false,false,true,true,true,true,true,false,false
#default,_result,,,,,,,,
,result,table,_start,_stop,_measurement,_field,t0,_time,_value
,,0,2019-01-09T19:44:30Z,2019-01-09T19:46:00Z,Reiva,OAOJWe7,BnR,2019-01-09T19:45:00Z,7.940387008821781
,,0,2019-01-09T19:44:30Z,2019-01-09T19:46:00Z,Reiva,OAOJWe7,BnR,2019-01-09T19:45:30Z,34.319039251798635
,,0,2019-01-09T19:44:30Z,2019-01-09T19:46:00Z,Reiva,OAOJWe7,BnR,2019-01-09T19:46:00Z,60.590246667535845
,,1,2019-01-09T19:44:30Z,2019-01-09T19:46:00Z,Reiva,OAOJWe7,qCnJDC,2019-01-09T19:45:00Z,-61.68790887989735
,,1,2019-01-09T19:44:30Z,2019-01-09T19:46:00Z,Reiva,OAOJWe7,qCnJDC,2019-01-09T19:45:30Z,-6.3173755351186465
,,1,2019-01-09T19:44:30Z,2019-01-09T19:46:00Z,Reiva,OAOJWe7,qCnJDC,2019-01-09T19:46:00Z,22.821813505281843
"

t_aggregate_window_median = (table=<-) =>
	table
		|> range(start: 2019-01-09T19:44:30Z, stop: 2019-01-09T19:46:00Z)
    |> aggregateWindow(every: 30s, fn: median)

test _aggregate_window_median = () => ({
	input: testing.loadStorage(csv: inData),
	want: testing.loadMem(csv: outData),
	fn: t_aggregate_window_median,
})
