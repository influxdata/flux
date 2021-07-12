package sample_test


import "influxdata/influxdb/sample"
import "testing"

testData = sample.list()
    |> count()
    |> map(fn: (r) => ({_value: r._value > 0}))

outData = "
#group,false,false,false
#datatype,string,long,boolean
#default,_result,,
,result,table,_value
,,0,true
"

test _list = () => ({input: testing.load(tables: testData), want: testing.loadMem(csv: outData)})
