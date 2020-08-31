package secrets_test

import "testing"
import "influxdata/influxdb/secrets"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,double,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,0,2018-05-22T19:53:26Z,1.83,load1,system
"

outData = "
#datatype,string,long,dateTime:RFC3339,double,string,string,string
#group,false,false,false,false,true,true,false
#default,_result,,,,,,
,result,table,_time,_value,_field,_measurement,token
,,0,2018-05-22T19:53:26Z,1.83,load1,system,mysecrettoken
"

token = secrets.get(key: "token")
t_get_secret = (table=<-) =>
	table
    |> set(key: "token", value: token)
    |> drop(columns: ["_start", "_stop"])

test _get_secret = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_get_secret})

