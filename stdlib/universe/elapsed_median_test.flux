package universe_test

import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,double,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,0,2018-05-22T19:53:26Z,34.98234271799806,active,mem
,,0,2018-05-22T19:53:36Z,34.98234941084654,active,mem
,,0,2018-05-22T19:53:46Z,34.982447293755506,active,mem
,,0,2018-05-22T19:53:56Z,34.982447293755506,active,mem
,,0,2018-05-22T19:54:06Z,34.98204153981662,active,mem
,,0,2018-05-22T19:54:16Z,34.982252364543626,active,mem
,,1,2018-05-22T19:53:26Z,34.98234271799806,f,m2
,,1,2018-05-22T19:53:36Z,34.98234941084654,f,m2
,,1,2018-05-22T19:53:46Z,34.982447293755506,f,m2
,,1,2018-05-22T19:53:56Z,34.982447293755506,f,m2
,,1,2018-05-22T19:54:06Z,34.98204153981662,f,m2
,,1,2018-05-22T19:54:16Z,34.982252364543626,f,m2
"

outData = "
#datatype,string,long,double
#group,false,false,false
#default,_result,,
,result,table,elapsed
,,0,10
"

t_elapsed = (table=<-) =>
	(table
		|> range(start: 2018-05-22T19:53:26Z)
        |> filter(fn: (r) => r._measurement == "mem" and r._field == "active")
        |> elapsed()
        |> group()
        |> map(fn: (r) => ({r with elapsed: float(v: r.elapsed)}))
        |> median(column: "elapsed")
    )

test _elapsed_median = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_elapsed})

