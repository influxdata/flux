package dict_test


import "testing"
import "dict"
import "csv"

codes = [:]
inData =
    "
#datatype,string,long,dateTime:RFC3339,string,string,long
#group,false,false,false,true,true,false
#default,_result,,,,,
,result,table,_time,_measurement,_field,_value
,,0,2018-05-22T19:53:26Z,_m,_f,0
,,0,2018-05-22T19:53:36Z,_m,_f,0
"
outData =
    "
#datatype,string,long,dateTime:RFC3339,string,string,long,long
#group,false,false,false,true,true,false,false
#default,_result,,,,,,
,result,table,_time,_measurement,_field,_value,code
,,0,2018-05-22T19:53:26Z,_m,_f,0,2
,,0,2018-05-22T19:53:36Z,_m,_f,0,2
"

testcase dict_empty_lit {
    got =
        csv.from(csv: inData)
            |> testing.load()
            |> range(start: 2018-05-22T19:53:26Z)
            |> drop(columns: ["_start", "_stop"])
            |> map(fn: (r) => ({r with code: dict.get(dict: codes, key: 1, default: 2)}))
    want = csv.from(csv: outData)

    testing.diff(got, want)
}
