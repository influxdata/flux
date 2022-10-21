package json_test


import "array"
import "experimental/json"
import "testing"
import "csv"

option now = () => 2030-01-01T00:00:00Z

inData =
    "
#datatype,string,long,dateTime:RFC3339,string,string,string
#group,false,false,false,true,true,false
#default,_result,,,,,
,result,table,_time,_field,_measurement,_value
,,0,2018-05-22T19:53:26Z,json,m,\"{\"\"a\"\":1,\"\"b\"\":2,\"\"c\"\":3}\"
,,0,2018-05-22T19:53:36Z,json,m,\"{\"\"a\"\":2,\"\"b\"\":4,\"\"c\"\":6}\"
,,0,2018-05-22T19:53:46Z,json,m,\"{\"\"a\"\":3,\"\"b\"\":5,\"\"c\"\":7}\"
"
outData =
    "
#datatype,string,long,dateTime:RFC3339,double,double,double
#group,false,false,false,false,false,false
#default,_result,,,,,
,result,table,_time,a,b,c
,,0,2018-05-22T19:53:26Z,1,2,3
,,0,2018-05-22T19:53:36Z,2,4,6
,,0,2018-05-22T19:53:46Z,3,5,7
"

testcase parse {
    got =
        csv.from(csv: inData)
            |> testing.load()
            |> range(start: 2018-05-22T19:53:26Z)
            |> map(
                fn: (r) => {
                    data = json.parse(data: bytes(v: r._value))

                    return {_time: r._time, a: data.a, b: data.b, c: data.c}
                },
            )
    want = csv.from(csv: outData)

    testing.diff(got, want)
}

testcase parse_to_array_from {
    data = json.parse(data: bytes(v: "[{\"_value\": 2}]"))
    got = array.from(rows: data)

    want = array.from(rows: [{_value: 2.0}])

    testing.diff(got, want)
}
