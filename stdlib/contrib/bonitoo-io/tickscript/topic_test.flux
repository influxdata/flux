package tickscript_test

import "testing"
import "csv"
import "contrib/bonitoo-io/tickscript"
import "influxdata/influxdb/monitor"

option now = () => (2020-11-25T14:05:30Z)

// overwrite as buckets are not avail in Flux tests
option monitor.write = (tables=<-) => tables
option monitor.log = (tables=<-) => tables

inData = "
#group,false,false,false,true,true,true,true,false,true,false,false,true,false,true,false,true
#datatype,string,long,double,string,string,string,string,string,string,long,dateTime:RFC3339,string,string,string,string,string
#default,_result,,,,,,,,,,,,,,,
,result,table,KafkaMsgRate,_check_id,_check_name,_level,_measurement,_message,_source_measurement,_source_timestamp,_time,_type,details,host,id,realm
,,0,1.819231109049999,rate-check,Rate Check,ok,statuses,Realm: ft - Hostname: kafka07 / Metric: kafka_message_in_rate threshold alert: ok - 1.819231109049999,testm,1606313103477635916,2020-11-25T14:05:25.856319359Z,custom,some detail: myrealm=ft,kafka07,Realm: ft - Hostname: kafka07 / Metric: kafka_message_in_rate threshold alert,ft
,,0,1.635878190200181,rate-check,Rate Check,ok,statuses,Realm: ft - Hostname: kafka07 / Metric: kafka_message_in_rate threshold alert: ok - 1.635878190200181,testm,1606313104541635074,2020-11-25T14:05:25.856444173Z,custom,some detail: myrealm=ft,kafka07,Realm: ft - Hostname: kafka07 / Metric: kafka_message_in_rate threshold alert,ft
,,1,39.33716449678206,rate-check,Rate Check,crit,statuses,Realm: ft - Hostname: kafka07 / Metric: kafka_message_in_rate threshold alert: crit - 39.33716449678206,testm,1606313105623191313,2020-11-25T14:05:25.856485929Z,custom,some detail: myrealm=ft,kafka07,Realm: ft - Hostname: kafka07 / Metric: kafka_message_in_rate threshold alert,ft
,,1,26.33716449678206,rate-check,Rate Check,crit,statuses,Realm: ft - Hostname: kafka07 / Metric: kafka_message_in_rate threshold alert: crit - 26.33716449678206,testm,1606313106696061106,2020-11-25T14:05:25.856575405Z,custom,some detail: myrealm=ft,kafka07,Realm: ft - Hostname: kafka07 / Metric: kafka_message_in_rate threshold alert,ft
,,2,8.33716449678206,rate-check,Rate Check,warn,statuses,Realm: ft - Hostname: kafka07 / Metric: kafka_message_in_rate threshold alert: warn - 8.33716449678206,testm,1606313107768317097,2020-11-25T14:05:25.856624989Z,custom,some detail: myrealm=ft,kafka07,Realm: ft - Hostname: kafka07 / Metric: kafka_message_in_rate threshold alert,ft
"

outData = "
#group,false,false,false,true,true,true,true,false,true,false,false,true,false,true,false,true,true
#datatype,string,long,double,string,string,string,string,string,string,long,dateTime:RFC3339,string,string,string,string,string,string
#default,_result,,,,,,,,,,,,,,,,
,result,table,KafkaMsgRate,_check_id,_check_name,_level,_measurement,_message,_source_measurement,_source_timestamp,_time,_type,details,host,id,realm,_topic
,,0,1.819231109049999,rate-check,Rate Check,ok,statuses,Realm: ft - Hostname: kafka07 / Metric: kafka_message_in_rate threshold alert: ok - 1.819231109049999,testm,1606313103477635916,2020-11-25T14:05:25.856319359Z,custom,some detail: myrealm=ft,kafka07,Realm: ft - Hostname: kafka07 / Metric: kafka_message_in_rate threshold alert,ft,TESTING
,,0,1.635878190200181,rate-check,Rate Check,ok,statuses,Realm: ft - Hostname: kafka07 / Metric: kafka_message_in_rate threshold alert: ok - 1.635878190200181,testm,1606313104541635074,2020-11-25T14:05:25.856444173Z,custom,some detail: myrealm=ft,kafka07,Realm: ft - Hostname: kafka07 / Metric: kafka_message_in_rate threshold alert,ft,TESTING
,,1,39.33716449678206,rate-check,Rate Check,crit,statuses,Realm: ft - Hostname: kafka07 / Metric: kafka_message_in_rate threshold alert: crit - 39.33716449678206,testm,1606313105623191313,2020-11-25T14:05:25.856485929Z,custom,some detail: myrealm=ft,kafka07,Realm: ft - Hostname: kafka07 / Metric: kafka_message_in_rate threshold alert,ft,TESTING
,,1,26.33716449678206,rate-check,Rate Check,crit,statuses,Realm: ft - Hostname: kafka07 / Metric: kafka_message_in_rate threshold alert: crit - 26.33716449678206,testm,1606313106696061106,2020-11-25T14:05:25.856575405Z,custom,some detail: myrealm=ft,kafka07,Realm: ft - Hostname: kafka07 / Metric: kafka_message_in_rate threshold alert,ft,TESTING
,,2,8.33716449678206,rate-check,Rate Check,warn,statuses,Realm: ft - Hostname: kafka07 / Metric: kafka_message_in_rate threshold alert: warn - 8.33716449678206,testm,1606313107768317097,2020-11-25T14:05:25.856624989Z,custom,some detail: myrealm=ft,kafka07,Realm: ft - Hostname: kafka07 / Metric: kafka_message_in_rate threshold alert,ft,TESTING
"

tickscript_topic = (table=<-) => table
    |> tickscript.topic(name: "TESTING")

test _tickscript_topic = () => ({
	input: testing.loadMem(csv: inData), // use loadMem because inData is pivoted (it is output of alert()) ie. without _field
	want: testing.loadMem(csv: outData),
	fn: tickscript_topic,
})
