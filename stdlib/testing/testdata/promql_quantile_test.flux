package testdata_test

import "testing"
import "internal/promql" 

option now = () => (2030-01-01T00:00:00Z)

inData = "
#datatype,string,long,string,string,string,dateTime:RFC3339,double
#group,false,false,true,true,true,false,false
#default,_result,,,,,,
,result,table,_measurement,_field,t0,_time,_value
,,2,Reiva,OAOJWe7,BnR,2019-01-09T19:44:58Z,7.940387008821781
,,2,Reiva,OAOJWe7,BnR,2019-01-09T19:45:08Z,49.460104214779086
,,2,Reiva,OAOJWe7,BnR,2019-01-09T19:45:18Z,-36.564150808873954
,,2,Reiva,OAOJWe7,BnR,2019-01-09T19:45:28Z,34.319039251798635
,,2,Reiva,OAOJWe7,BnR,2019-01-09T19:45:38Z,79.27019811403116
,,2,Reiva,OAOJWe7,BnR,2019-01-09T19:45:48Z,41.91029522104053
,,3,Reiva,OAOJWe7,qCnJDC,2019-01-09T19:44:58Z,-61.68790887989735
,,3,Reiva,OAOJWe7,qCnJDC,2019-01-09T19:45:08Z,-6.3173755351186465
,,3,Reiva,OAOJWe7,qCnJDC,2019-01-09T19:45:18Z,-26.049728557657513
,,3,Reiva,OAOJWe7,qCnJDC,2019-01-09T19:45:28Z,114.285955884979
,,3,Reiva,OAOJWe7,qCnJDC,2019-01-09T19:45:38Z,16.140262630578995
,,3,Reiva,OAOJWe7,qCnJDC,2019-01-09T19:45:48Z,29.50336437998469
"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,string,double
#group,false,false,true,true,true,true,true,false
#default,_result,,,,,,,
,result,table,_start,_stop,_measurement,_field,t0,_value
,,2,2019-01-01T00:00:00Z,2030-01-01T00:00:00Z,Reiva,OAOJWe7,BnR,47.57265196634445
,,3,2019-01-01T00:00:00Z,2030-01-01T00:00:00Z,Reiva,OAOJWe7,qCnJDC,26.162588942633263
"

// testing normal range value
t_quantile = (tables=<-) =>
    tables
        |> range(start: 2019-01-01T00:00:00Z)
        |> promql.quantile(q: 0.75)

test _quantile = () => 
    ({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: t_quantile})