package sample_test


import "influxdata/influxdb/sample"
import "csv"
import "testing"

expected = "
#group,false,false,false
#datatype,string,long,boolean
#default,_result,,
,result,table,_value
,,0,true
"

t_data = () => {
    got = sample.data(set: "airSensor")
        |> group()
        |> count()
        |> map(fn: (r) => ({_value: r._value > 0}))
    want = csv.from(csv: expected)

    return testing.diff(got: got, want: want)
}

testcase _data {
    t_data()
}
