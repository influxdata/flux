package universe_test


import "testing"
import "array"

inData = array.from(
    rows: [
        {_measurement: "command", _field: "id", _time: 2018-12-19T22:13:30.005Z, _value: 12},
        {_measurement: "command", _field: "id", _time: 2018-12-19T22:13:40.005Z, _value: 23},
        {_measurement: "command", _field: "id", _time: 2018-12-19T22:13:50.005Z, _value: 34},
        {_measurement: "command", _field: "guild", _time: 2018-12-19T22:13:30.005Z, _value: 12},
        {_measurement: "command", _field: "guild", _time: 2018-12-19T22:13:40.005Z, _value: 23},
        {_measurement: "command", _field: "guild", _time: 2018-12-19T22:13:50.005Z, _value: 34},
    ],
)
get_input = () => inData
outData = "
#group,false,false,false,false,false,false,true,true,true,true,false,false,false
#datatype,string,long,string,string,string,string,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,long,long
#default,_result,,,,,,,,,,,,
,result,table,_field_d1,_field_d2,_measurement_d1,_measurement_d2,_start_d1,_start_d2,_stop_d1,_stop_d2,_time,_value_d1,_value_d2
,,0,id,guild,command,command,2018-12-19T22:13:30Z,2018-12-19T22:13:30Z,2018-12-19T22:13:50Z,2018-12-19T22:13:50Z,2018-12-19T22:13:31Z,12,12
,,0,id,guild,command,command,2018-12-19T22:13:30Z,2018-12-19T22:13:30Z,2018-12-19T22:13:50Z,2018-12-19T22:13:50Z,2018-12-19T22:13:41Z,23,23
"
t_join_two_same_sources = (table=<-) => {
    data_1 = table
        |> range(start: 2018-12-19T22:13:30Z, stop: 2018-12-19T22:13:50Z)
        |> filter(fn: (r) => r._measurement == "command" and r._field == "id")
        |> aggregateWindow(every: 1s, fn: last)
    data_2 = table
        |> range(start: 2018-12-19T22:13:30Z, stop: 2018-12-19T22:13:50Z)
        |> filter(fn: (r) => r._measurement == "command" and r._field == "guild")
        |> aggregateWindow(every: 1s, fn: last)

    return join(tables: {d1: data_1, d2: data_2}, on: ["_time"])
}

test _join_two_same_sources = () => ({input: get_input(), want: testing.loadMem(csv: outData), fn: t_join_two_same_sources})
