package universe_test


import "testing"
import "math"
import "csv"

option now = () => 2018-05-23T20:00:00Z

inData =
    "
#datatype,string,long,dateTime:RFC3339,double,string,string
#group,false,false,false,false,true,true
#default,_result,,,,,
,result,table,_time,_value,_field,_measurement
,,0,2018-05-23T19:53:30Z,1.0,diameter,turbine
,,0,2018-05-23T19:53:40Z,2.0,diameter,turbine
"
outData =
    "
#datatype,string,long,dateTime:RFC3339,double,double,string,string
#group,false,false,false,false,false,true,true
#default,_result,,,,,,
,result,table,_time,diameter,circumference,_field,_measurement
,,0,2018-05-23T19:53:30Z,1.0,3.141592653589793,diameter,turbine
,,0,2018-05-23T19:53:40Z,2.0,6.283185307179586,diameter,turbine
"

testcase math_pi_test {
    got =
        csv.from(csv: inData)
            |> testing.load()
            |> range(start: -10m)
            |> filter(fn: (r) => r._measurement == "turbine" and r._field == "diameter")
            |> rename(columns: {_value: "diameter"})
            |> map(fn: (r) => ({r with circumference: r.diameter * math.pi}))
            |> drop(columns: ["_value", "_start", "_stop"])
    want = csv.from(csv: outData)

    testing.diff(got, want)
}
