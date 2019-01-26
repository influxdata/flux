package main
 
import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,double,string,string,string,string
#group,false,false,false,false,true,true,true,true
#default,_result,,,,,,,
,result,table,_time,_value,_field,_measurement,cpu,host
,,0,2018-05-22T19:53:26Z,0,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:36Z,0,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:46Z,0,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:53:56Z,0,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:54:06Z,0,usage_guest,cpu,cpu-total,host.local
,,0,2018-05-22T19:54:16Z,0,usage_guest,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:26Z,0,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:36Z,0,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:46Z,0,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:53:56Z,0,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:54:06Z,0,usage_guest_nice,cpu,cpu-total,host.local
,,1,2018-05-22T19:54:16Z,0,usage_guest_nice,cpu,cpu-total,host.local
,,2,2018-05-22T19:53:26Z,91.7364670583823,usage_idle,cpu,cpu-total,host.local
,,2,2018-05-22T19:53:36Z,89.51118889861233,usage_idle,cpu,cpu-total,host.local
,,2,2018-05-22T19:53:46Z,91.0977744436109,usage_idle,cpu,cpu-total,host.local
,,2,2018-05-22T19:53:56Z,91.02836436336374,usage_idle,cpu,cpu-total,host.local
,,2,2018-05-22T19:54:06Z,68.304576144036,usage_idle,cpu,cpu-total,host.local
,,2,2018-05-22T19:54:16Z,87.88598574821853,usage_idle,cpu,cpu-total,host.local
,,3,2018-05-22T19:53:26Z,0,usage_iowait,cpu,cpu-total,host.local
,,3,2018-05-22T19:53:36Z,0,usage_iowait,cpu,cpu-total,host.local
,,3,2018-05-22T19:53:46Z,0,usage_iowait,cpu,cpu-total,host.local
,,3,2018-05-22T19:53:56Z,0,usage_iowait,cpu,cpu-total,host.local
,,3,2018-05-22T19:54:06Z,0,usage_iowait,cpu,cpu-total,host.local
,,3,2018-05-22T19:54:16Z,0,usage_iowait,cpu,cpu-total,host.local
,,4,2018-05-22T19:53:26Z,0,usage_irq,cpu,cpu-total,host.local
,,4,2018-05-22T19:53:36Z,0,usage_irq,cpu,cpu-total,host.local
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,string,string
#group,false,false,true,true,true,true,true,true,false
#default,0,,,,,,,,
,result,table,_start,_stop,_field,_measurement,cpu,host,_value
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_guest,cpu,cpu-total,host.local,_start
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_guest,cpu,cpu-total,host.local,_stop
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_guest,cpu,cpu-total,host.local,_field
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_guest,cpu,cpu-total,host.local,_measurement
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_guest,cpu,cpu-total,host.local,cpu
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_guest,cpu,cpu-total,host.local,host
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_guest_nice,cpu,cpu-total,host.local,_start
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_guest_nice,cpu,cpu-total,host.local,_stop
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_guest_nice,cpu,cpu-total,host.local,_field
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_guest_nice,cpu,cpu-total,host.local,_measurement
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_guest_nice,cpu,cpu-total,host.local,cpu
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_guest_nice,cpu,cpu-total,host.local,host
,,2,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_idle,cpu,cpu-total,host.local,_start
,,2,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_idle,cpu,cpu-total,host.local,_stop
,,2,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_idle,cpu,cpu-total,host.local,_field
,,2,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_idle,cpu,cpu-total,host.local,_measurement
,,2,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_idle,cpu,cpu-total,host.local,cpu
,,2,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_idle,cpu,cpu-total,host.local,host
,,3,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_iowait,cpu,cpu-total,host.local,_start
,,3,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_iowait,cpu,cpu-total,host.local,_stop
,,3,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_iowait,cpu,cpu-total,host.local,_field
,,3,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_iowait,cpu,cpu-total,host.local,_measurement
,,3,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_iowait,cpu,cpu-total,host.local,cpu
,,3,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_iowait,cpu,cpu-total,host.local,host
,,4,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_irq,cpu,cpu-total,host.local,_start
,,4,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_irq,cpu,cpu-total,host.local,_stop
,,4,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_irq,cpu,cpu-total,host.local,_field
,,4,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_irq,cpu,cpu-total,host.local,_measurement
,,4,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_irq,cpu,cpu-total,host.local,cpu
,,4,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_irq,cpu,cpu-total,host.local,host
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,string,string
#group,false,false,true,true,true,true,true,true,false
#default,_result,,,,,,,,
,result,table,_start,_stop,_field,_measurement,cpu,host,_value
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_guest,cpu,cpu-total,host.local,_start
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_guest,cpu,cpu-total,host.local,_stop
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_guest,cpu,cpu-total,host.local,_field
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_guest,cpu,cpu-total,host.local,_measurement
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_guest,cpu,cpu-total,host.local,cpu
,,0,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_guest,cpu,cpu-total,host.local,host
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_guest_nice,cpu,cpu-total,host.local,_start
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_guest_nice,cpu,cpu-total,host.local,_stop
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_guest_nice,cpu,cpu-total,host.local,_field
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_guest_nice,cpu,cpu-total,host.local,_measurement
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_guest_nice,cpu,cpu-total,host.local,cpu
,,1,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_guest_nice,cpu,cpu-total,host.local,host
,,2,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_idle,cpu,cpu-total,host.local,_start
,,2,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_idle,cpu,cpu-total,host.local,_stop
,,2,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_idle,cpu,cpu-total,host.local,_field
,,2,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_idle,cpu,cpu-total,host.local,_measurement
,,2,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_idle,cpu,cpu-total,host.local,cpu
,,2,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_idle,cpu,cpu-total,host.local,host
,,3,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_iowait,cpu,cpu-total,host.local,_start
,,3,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_iowait,cpu,cpu-total,host.local,_stop
,,3,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_iowait,cpu,cpu-total,host.local,_field
,,3,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_iowait,cpu,cpu-total,host.local,_measurement
,,3,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_iowait,cpu,cpu-total,host.local,cpu
,,3,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_iowait,cpu,cpu-total,host.local,host
,,4,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_irq,cpu,cpu-total,host.local,_start
,,4,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_irq,cpu,cpu-total,host.local,_stop
,,4,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_irq,cpu,cpu-total,host.local,_field
,,4,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_irq,cpu,cpu-total,host.local,_measurement
,,4,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_irq,cpu,cpu-total,host.local,cpu
,,4,2018-05-22T19:53:26Z,2030-01-01T00:00:00Z,usage_irq,cpu,cpu-total,host.local,host
"

t_meta_query_keys = (table=<-) => {
	zero = table
		|> range(start: 2018-05-22T19:53:26Z)
		|> filter(fn: (r) =>
			(r._measurement == "cpu"))
		|> keys()
		|> yield(name: "0")
	one = table
		|> range(start: 2018-05-22T19:53:26Z)
		|> filter(fn: (r) =>
			(r._measurement == "cpu"))
		|> group(columns: ["host"])
		|> distinct(column: "host")
		|> group()
		|> yield(name: "1")

	return union(tables: [zero, one])
}

test _meta_query_keys = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_meta_query_keys})

testing.run(case: _meta_query_keys)