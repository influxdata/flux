package print_test

import "testing"
import "array"
import "contrib/anaisdg/print"

testcase printIntTest {
    got =
        print(2, "Int")
    want =
        array.from(
            rows: [
                {_value: 2}}
            ]
        )
        |> yield(name: "Int")
    
    testing.diff(got, want)
}