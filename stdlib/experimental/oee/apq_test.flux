package oee_test


import "experimental/oee"
import "testing"

option now = () => 2030-01-01T00:00:00Z

inData = "
#group,false,false,false,false,false,false
#datatype,string,long,dateTime:RFC3339,string,long,long
#default,_result,,,,,
,result,table,_time,state,partCount,badCount
,,0,2021-03-22T00:00:00Z,running,1200,10
,,0,2021-03-22T01:00:00Z,running,1300,11
,,0,2021-03-22T02:00:00Z,stopped,1400,11
,,0,2021-03-22T03:00:00Z,running,1400,11
,,0,2021-03-22T03:30:00Z,running,1440,14
"
outData = "
#group,false,false,false,false,false,false,false,false
#datatype,string,long,dateTime:RFC3339,double,double,double,double,long
#default,_result,,,,,,,
,result,table,_time,availability,oee,performance,quality,runTime
,,0,2021-03-22T04:00:00Z,0.375,0.24583333333333332,0.6666666666666666,0.9833333333333333,10800000000000
"
t_APQ = (table=<-) => table
    |> range(start: 2021-03-22T00:00:00Z, stop: 2021-03-22T04:00:00Z)
    |> oee.APQ(runningState: "running", plannedTime: 8h, idealCycleTime: 30s)
    |> drop(columns: ["_start", "_stop"])

test _APQ = () => ({input: testing.loadMem(csv: inData), want: testing.loadMem(csv: outData), fn: t_APQ})
