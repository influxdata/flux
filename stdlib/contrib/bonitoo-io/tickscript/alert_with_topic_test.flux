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
#group,false,false,false,true,true,true,true,false,true,false,true,false,true,false,true,true
#datatype,string,long,double,string,string,string,string,string,string,long,string,string,string,string,string,string
#default,_result,,,,,,,,,,,,,,,
,result,table,KafkaMsgRate,_check_id,_check_name,_level,_measurement,_message,_source_measurement,_source_timestamp,_type,details,host,id,realm,_topic
,,0,1.819231109049999,rate-check,Rate Check,ok,statuses,Realm: ft - Hostname: kafka07 / Metric: kafka_message_in_rate threshold alert: ok - 1.819231109049999,testm,1606313103477635916,custom,some detail: myrealm=ft,kafka07,Realm: ft - Hostname: kafka07 / Metric: kafka_message_in_rate threshold alert,ft,TESTING
,,0,1.635878190200181,rate-check,Rate Check,ok,statuses,Realm: ft - Hostname: kafka07 / Metric: kafka_message_in_rate threshold alert: ok - 1.635878190200181,testm,1606313104541635074,custom,some detail: myrealm=ft,kafka07,Realm: ft - Hostname: kafka07 / Metric: kafka_message_in_rate threshold alert,ft,TESTING
,,0,1.33716449678206,rate-check,Rate Check,ok,statuses,Realm: ft - Hostname: kafka07 / Metric: kafka_message_in_rate threshold alert: ok - 1.33716449678206,testm,1606313108868317091,custom,some detail: myrealm=ft,kafka07,Realm: ft - Hostname: kafka07 / Metric: kafka_message_in_rate threshold alert,ft,TESTING
,,1,39.33716449678206,rate-check,Rate Check,crit,statuses,Realm: ft - Hostname: kafka07 / Metric: kafka_message_in_rate threshold alert: crit - 39.33716449678206,testm,1606313105623191313,custom,some detail: myrealm=ft,kafka07,Realm: ft - Hostname: kafka07 / Metric: kafka_message_in_rate threshold alert,ft,TESTING
,,1,26.33716449678206,rate-check,Rate Check,crit,statuses,Realm: ft - Hostname: kafka07 / Metric: kafka_message_in_rate threshold alert: crit - 26.33716449678206,testm,1606313106696061106,custom,some detail: myrealm=ft,kafka07,Realm: ft - Hostname: kafka07 / Metric: kafka_message_in_rate threshold alert,ft,TESTING
,,2,8.33716449678206,rate-check,Rate Check,warn,statuses,Realm: ft - Hostname: kafka07 / Metric: kafka_message_in_rate threshold alert: warn - 8.33716449678206,testm,1606313107768317097,custom,some detail: myrealm=ft,kafka07,Realm: ft - Hostname: kafka07 / Metric: kafka_message_in_rate threshold alert,ft,TESTING
"
check = {
    _check_id: "rate-check",
    _check_name: "Rate Check",
    // tickscript?
    _type: "custom",
    tags: {},
}
metric_type = "kafka_message_in_rate"
tier = "ft"
h_threshold = 10
w_threshold = 5
l_threshold = 0.002
tickscript_alert = (table=<-) => table
    |> range(start: 2020-11-25T14:05:00Z)
    |> filter(fn: (r) => r._field == metric_type and r.realm == tier)
    |> schema.fieldsAsCols()
    |> tickscript.select(column: metric_type, as: "KafkaMsgRate")
    |> tickscript.groupBy(columns: ["host", "realm"])
    |> tickscript.alert(
        check: check,
        id: (r) => "Realm: ${r.realm} - Hostname: ${r.host} / Metric: ${metric_type} threshold alert",
        message: (r) => "${r.id}: ${r._level} - ${string(v: r.KafkaMsgRate)}",
        details: (r) => "some detail: myrealm=${r.realm}",
        crit: (r) => r.KafkaMsgRate > h_threshold or r.KafkaMsgRate < l_threshold,
        warn: (r) => r.KafkaMsgRate > w_threshold or r.KafkaMsgRate < l_threshold,
        topic: "TESTING",
    )
    |> drop(columns: ["_time"])

test _tickscript_alert = () => ({
    input: testing.loadStorage(csv: inData),
    want: testing.loadMem(csv: outData),
    fn: tickscript_alert,
})
