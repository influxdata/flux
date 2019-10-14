package regexp_test

import "testing"
import "regexp"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,long,string,string,string,string
#group,false,false,false,false,true,true,true,true
#default,_result,,,,,,,
,result,table,_time,_value,_field,_measurement,host,name
,,0,2018-05-22T19:53:26Z,15204688,io_time,diskio,host.local,disk0
,,0,2018-05-22T19:53:36Z,15204894,io_time,diskio,host.local,disk0
,,0,2018-05-22T19:53:46Z,15205102,io_time,diskio,host.local,disk0
,,0,2018-05-22T19:53:56Z,15205226,io_time,diskio,host.local,disk0
,,0,2018-05-22T19:54:06Z,15205499,io_time,diskio,host.local,disk0
,,0,2018-05-22T19:54:16Z,15205755,io_time,diskio,host.local,disk0
,,1,2018-05-22T19:53:26Z,648,io_time,diskio,host.local,disk2
,,1,2018-05-22T19:53:36Z,648,io_time,diskio,host.local,disk2
,,1,2018-05-22T19:53:46Z,648,io_time,diskio,host.local,disk2
,,1,2018-05-22T19:53:56Z,648,io_time,diskio,host.local,disk2
,,1,2018-05-22T19:54:06Z,648,io_time,diskio,host.local,disk2
,,1,2018-05-22T19:54:16Z,648,io_time,diskio,host.local,disk2
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,dateTime:RFC3339,long,string
#group,false,false,true,true,true,true,true,false,false,true
#default,_result,,,,,,,,,
,result,table,_start,_stop,_measurement,_field,host,_time,_value,name
,,0,2018-05-20T19:53:26Z,2030-01-01T00:00:00Z,diskio,io_time,host.local,2018-05-22T19:53:26Z,15204688,disk9
,,0,2018-05-20T19:53:26Z,2030-01-01T00:00:00Z,diskio,io_time,host.local,2018-05-22T19:53:36Z,15204894,disk9
,,0,2018-05-20T19:53:26Z,2030-01-01T00:00:00Z,diskio,io_time,host.local,2018-05-22T19:53:46Z,15205102,disk9
,,0,2018-05-20T19:53:26Z,2030-01-01T00:00:00Z,diskio,io_time,host.local,2018-05-22T19:53:56Z,15205226,disk9
,,0,2018-05-20T19:53:26Z,2030-01-01T00:00:00Z,diskio,io_time,host.local,2018-05-22T19:54:06Z,15205499,disk9
,,0,2018-05-20T19:53:26Z,2030-01-01T00:00:00Z,diskio,io_time,host.local,2018-05-22T19:54:16Z,15205755,disk9
"

re = regexp.compile(v: ".*0")

t_filter_by_regex = (table=<-) =>
table
  |> range(start: 2018-05-20T19:53:26Z)
  |> filter(fn: (r) => r["name"] =~ /.*0/)
  |> map(fn: (r) =>
                ({r with name: regexp.replaceAllString(r: re, v: r.name, t: "disk9")}))

test _filter_by_regex = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_filter_by_regex})
