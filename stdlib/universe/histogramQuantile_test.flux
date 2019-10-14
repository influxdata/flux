package testdata_test

import "experimental/prometheus"
import "testing"

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,dateTime:RFC3339,string,string,string,string,double
#group,false,false,false,true,true,false,true,false
#default,_result,,,,,,,
,result,table,_time,_measurement,_field,url,le,_value
,,0,2018-05-22T13:00:00Z,prometheus,prometheus_test_metric,http://prometheus.test,100,0.1

#datatype,string,long,dateTime:RFC3339,string,string,string,string,double
#group,false,false,false,true,true,false,true,false
#default,_result,,,,,,,
,result,table,_time,_measurement,_field,url,le,_value
,,1,2018-05-22T13:00:00Z,prometheus,prometheus_test_metric,http://prometheus.test,150,0.2

#datatype,string,long,dateTime:RFC3339,string,string,string,string,double
#group,false,false,false,true,true,false,true,false
#default,_result,,,,,,,
,result,table,_time,_measurement,_field,url,le,_value
,,2,2018-05-22T13:00:00Z,prometheus,prometheus_test_metric,http://prometheus.test,200,0.25

#datatype,string,long,dateTime:RFC3339,string,string,string,string,double
#group,false,false,false,true,true,false,true,false
#default,_result,,,,,,,
,result,table,_time,_measurement,_field,url,le,_value
,,3,2018-05-22T13:00:00Z,prometheus,prometheus_test_metric,http://prometheus.test,250,0.267

#datatype,string,long,dateTime:RFC3339,string,string,string,string,double
#group,false,false,false,true,true,false,true,false
#default,_result,,,,,,,
,result,table,_time,_measurement,_field,url,le,_value
,,4,2018-05-22T13:00:00Z,prometheus,prometheus_test_metric,http://prometheus.test,300,0.267
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,string,double
#group,false,false,true,true,true,true,true,true,false
#default,_result,,,,,,,,
,result,table,_start,_stop,_time,_measurement,_field,url,_value
,,0,2018-05-22T13:00:00Z,2030-01-01T00:00:00Z,2018-05-22T13:00:00Z,prometheus,prometheus_test_metric,http://prometheus.test,116.75
"

t_histogramQuantile = (table=<-) =>
    (table
        |> range(start: 2018-05-22T13:00:00Z))
        |> prometheus.histogramQuantile(quantile: 0.5, field: "prometheus_test_metric")

test _histogramQuantile = () => 
({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_histogramQuantile})