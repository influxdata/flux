package tasks_test


import "testing"
import "array"
import "influxdata/influxdb/tasks"
import "csv"

option now = () => 2020-09-08T09:00:00Z

outData =
    "
#datatype,string,long,dateTime:RFC3339
#group,false,false,false
#default,_result,,
,result,table,_time
,,0,2020-09-07T09:00:00Z
"
t_last_success = () => array.from(rows: [{_time: tasks.lastSuccess(orTime: -1d)}])

testcase last_success_duration_no_option {
    tables = t_last_success()
    got = tables
    want = csv.from(csv: outData)

    testing.diff(got, want)
}
