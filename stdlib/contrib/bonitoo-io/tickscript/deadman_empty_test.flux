package tickscript_test


import "testing"
import "csv"
import "contrib/bonitoo-io/tickscript"
import "influxdata/influxdb/monitor"
import "influxdata/influxdb/schema"

option now = () => 2020-11-25T14:05:30Z

// overwrite as buckets are not avail in Flux tests
option monitor.write = (tables=<-) => tables
option monitor.log = (tables=<-) => tables

inData = "
#group,false,false,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,
,result,table,_time,_value,_field,_measurement,host,realm
,,0,2020-11-25T14:05:03.477635916Z,1.819231109049999,kafka_message_in_rate,testm,kafka07,ft
,,0,2020-11-25T14:05:04.541635074Z,1.635878190200181,kafka_message_in_rate,testm,kafka07,ft
,,0,2020-11-25T14:05:05.623191313Z,39.33716449678206,kafka_message_in_rate,testm,kafka07,ft
,,0,2020-11-25T14:05:06.696061106Z,26.33716449678206,kafka_message_in_rate,testm,kafka07,ft
,,0,2020-11-25T14:05:07.768317097Z,8.33716449678206,kafka_message_in_rate,testm,kafka07,ft
,,0,2020-11-25T14:05:08.868317091Z,1.33716449678206,kafka_message_in_rate,testm,kafka07,ft
"
outData = "
#group,false,false,true,true,true,true,false,true,false,true,false,false
#datatype,string,long,string,string,string,string,string,string,long,string,boolean,string
#default,_result,,,,,,,,,,,
,result,table,_check_id,_check_name,_level,_measurement,_message,_source_measurement,_source_timestamp,_type,dead,id
,,0,rate-check,Rate Check,crit,statuses,Deadman Check: Rate Check is: dead,testm,1606313130000000000,deadman,true,Realm: ft - Hostname: unknown / Metric: kafka_message_in_rate deadman alert
"
check = {
    _check_id: "rate-check",
    _check_name: "Rate Check",
    // tickscript?
    _type: "deadman",
    tags: {},
}
metric_type = "kafka_message_in_rate"
tier = "ft"
tickscript_deadman = (table=<-) => table
    |> range(start: 2020-11-25T14:05:15Z)
    |> filter(fn: (r) => r._measurement == "testm" and r._field == metric_type and r.realm == tier)
    |> schema.fieldsAsCols()
    |> tickscript.groupBy(columns: ["host", "realm"])
    |> tickscript.deadman(
        check: check,
        measurement: "testm",
        threshold: 10,
        id: (r) => "Realm: ${tier} - Hostname: unknown / Metric: ${metric_type} deadman alert",
    )
    // to avoid issue with validation
    |> drop(columns: ["details"])
    |> drop(columns: ["_time"])

test _tickscript_deadman = () => ({
    input: testing.loadStorage(csv: inData),
    want: testing.loadMem(csv: outData),
    fn: tickscript_deadman,
})
