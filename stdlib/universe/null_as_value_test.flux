package universe_test


import "testing"
import "csv"

option now = () => 2030-01-01T00:00:00Z

inData =
    "
#datatype,string,string
#group,true,true
#default,,
,error,reference
,failed to execute query: failed to initialize execute state: EOF,
"
outData = "err: error calling function "

filter
": name "
null
" does not exist in scope
"

testcase null_as_value {
    got =
        csv.from(csv: inData)
            |> testing.load()
            |> range(start: 2018-05-22T19:53:26Z)
            |> filter(fn: (r) => r._value == null)
    want = csv.from(csv: outData)

    testing.diff(got, want)
}
