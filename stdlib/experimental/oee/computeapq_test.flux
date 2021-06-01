package oee_test


import "experimental/oee"
import "testing"

option now = () => 2030-01-01T00:00:00Z

// not used
inData = "
#group,false,false,false,false,false,false
#datatype,string,long,dateTime:RFC3339,string,long,long
#default,_result,,,,,
,result,table,_time,state,partCount,badCount
"
productionData = "
#group,false,false,false,false
#datatype,string,long,dateTime:RFC3339,string
#default,_result,,,
,result,table,_time,state
,,0,2021-03-22T00:00:00Z,running
,,0,2021-03-22T01:00:00Z,running
,,0,2021-03-22T02:00:00Z,stopped
,,0,2021-03-22T03:00:00Z,running
"
partData = "
#group,false,false,false,false,false
#datatype,string,long,dateTime:RFC3339,long,long
#default,_result,,,,
,result,table,_time,partCount,badCount
,,0,2021-03-22T00:00:00Z,1200,10
,,0,2021-03-22T00:15:00Z,1225,10
,,0,2021-03-22T00:30:00Z,1250,11
,,0,2021-03-22T00:45:00Z,1275,11
,,0,2021-03-22T01:00:00Z,1300,11
,,0,2021-03-22T01:15:00Z,1310,11
,,0,2021-03-22T01:30:00Z,1340,11
,,0,2021-03-22T01:45:00Z,1380,11
,,0,2021-03-22T02:00:00Z,1400,11
,,0,2021-03-22T03:15:00Z,1425,12
,,0,2021-03-22T03:30:00Z,1440,14
"
outData = "
#group,false,false,false,false,false,false,false,false
#datatype,string,long,dateTime:RFC3339,double,double,double,double,long
#default,_result,,,,,,,
,result,table,_time,availability,oee,performance,quality,runTime
,,0,2021-03-22T04:00:00Z,0.375,0.24583333333333332,0.6666666666666666,0.9833333333333333,10800000000000
"
productionEvents = testing.loadMem(csv: productionData)
    |> range(start: 2021-03-22T00:00:00Z, stop: 2021-03-22T04:00:00Z)
partEvents = testing.loadMem(csv: partData)
    |> range(start: 2021-03-22T00:00:00Z, stop: 2021-03-22T04:00:00Z)
t_computeAPQ = (table=<-) => {
    return oee.computeAPQ(
        productionEvents: productionEvents,
        partEvents: partEvents,
        runningState: "running",
        plannedTime: 8h,
        idealCycleTime: 30s,
    )
        |> drop(columns: ["_start", "_stop"])
}

test _computeAPQ = () => ({input: testing.loadMem(csv: inData), want: testing.loadMem(csv: outData), fn: t_computeAPQ})
