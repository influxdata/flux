package polyline_test


import "array"
import "testing"
import "internal/gen"
import "internal/debug"
import "csv"
import "experimental/polyline"

option now = () => 2030-01-01T00:00:00Z

testcase polyline_rdp_with_epsilon_testcase {
    got =
        gen.tables(n: 16, seed: 1234)
            |> polyline.rdp(valColumn: "_value", timeColumn: "_time", epsilon: 55.0)
            |> drop(columns: ["_time"])
    want =
        array.from(
            rows: [
                {_value: 10.56555566168836},
                {_value: -47.25865245658065},
                {_value: 66.16082461651365},
                {_value: -56.89169240573004},
                {_value: 28.71147881415803},
                {_value: -44.44668391211515},
            ],
        )

    testing.diff(got: got, want: want)
}
testcase polyline_rdp_with_retentionrate_testcase {
    got =
        gen.tables(n: 16, seed: 1234)
            |> polyline.rdp(valColumn: "_value", timeColumn: "_time", retention: 30.0)
            |> drop(columns: ["_time"])
    want =
        array.from(
            rows: [
                {_value: 10.56555566168836},
                {_value: -47.25865245658065},
                {_value: -56.89169240573004},
                {_value: -44.44668391211515},
            ],
        )

    testing.diff(got: got, want: want)
}
testcase polyline_rdp_with_automatic_learning_testcase {
    got =
        gen.tables(n: 4, seed: 1234)
            |> polyline.rdp(valColumn: "_value", timeColumn: "_time")
            |> drop(columns: ["_time"])
    want =
        array.from(
            rows: [
                {_value: 10.56555566168836},
                {_value: -29.76098586714259},
                {_value: -67.50435038579738},
                {_value: -16.758669047964453},
            ],
        )

    testing.diff(got: got, want: want)
}
