package tasks_test

import "testing"
import "experimental/array"
import "influxdata/influxdb/tasks"

option now = () => 2020-09-08T09:00:00Z
option tasks.lastSuccessTime = 2020-09-08T08:00:00Z

outData = "
#datatype,string,long,dateTime:RFC3339
#group,false,false,false
#default,_result,,
,result,table,_time
,,0,2020-09-08T08:00:00Z
"

t_last_success = () =>
	array.from(rows: [
		{_time: tasks.lastSuccess(orTime: -1d)},
	])

test _last_success = () => ({
	input: t_last_success(),
	want: testing.loadMem(csv: outData),
	fn: (tables=<-) => tables,
})
