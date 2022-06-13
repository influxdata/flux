package gen


import "array"
import "testing"
import "internal/gen"
import "internal/debug"

option now = () => 2030-01-01T00:00:00Z

testcase gen_tables_seed {
    got =
        gen.tables(n: 5, seed: 123)
            |> drop(columns: ["_time"])

    want =
        array.from(
            rows: [
                {_value: -39.7289264835779},
                {_value: 3.561659508431711},
                {_value: 34.956531511845476},
                {_value: -53.72013420347799},
                {_value: 49.20701082669753},
            ],
        )

    testing.diff(got: got, want: want)
}
