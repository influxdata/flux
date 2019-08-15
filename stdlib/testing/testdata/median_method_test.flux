package testdata_test

import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,string,string,dateTime:RFC3339,double
#group,false,false,true,true,false,false
#default,_result,,,,,
,result,table,_measurement,_field,_time,_value
,,0,SOYcRk,NC7N,2018-12-18T21:12:45Z,55
,,0,SOYcRk,NC7N,2018-12-18T21:12:55Z,15
,,0,SOYcRk,NC7N,2018-12-18T21:13:05Z,25
,,0,SOYcRk,NC7N,2018-12-18T21:13:15Z,5
,,0,SOYcRk,NC7N,2018-12-18T21:13:25Z,105
,,0,SOYcRk,NC7N,2018-12-18T21:13:35Z,45
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,dateTime:RFC3339,double
#group,false,false,true,true,true,true,false,false
#default,_result,,,,,,,
,result,table,_start,_stop,_measurement,_field,_time,_value
,,0,2018-12-01T00:00:00Z,2030-01-01T00:00:00Z,SOYcRk,NC7N,2018-12-18T21:13:05Z,25
"


test _median = () => ({
        input: testing.loadStorage(csv: inData),
        want: testing.loadMem(csv: outData),
        fn: (tables=<-) =>
            tables
                |> range(start: 2018-12-01T00:00:00Z)
                |> median(method:"exact_selector")
    })
