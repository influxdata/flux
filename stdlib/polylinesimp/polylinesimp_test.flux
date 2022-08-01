package polylinesimp_test


import "array"
import "testing"
import "internal/gen"
import "internal/debug"
import "csv"
import "polylinesimp"

option now = () => 2030-01-01T00:00:00Z

testcase polylinesimp_rdp_with_epsilon_testcase {
    got =
        gen.tables(n: 4, seed: 1234)
            |> polylinesimp.rdp(column: "_value", timeColumn: "_time", epsilon: 2.5)
            |> drop(columns: ["_time"])
    want = array.from(rows: [{_value: 10.56555566168836}, {_value: -67.50435038579738}, {_value: -16.758669047964453}])

    testing.diff(got: got, want: want)
}

testcase polylinesimp_rdp_with_retentionrate_testcase {
    got =
        gen.tables(n: 4, seed: 1234)
            |> polylinesimp.rdp(column: "_value", timeColumn: "_time", retention: 50.0)
            |> drop(columns: ["_time"])
    want = array.from(rows: [{_value: 10.56555566168836}, {_value: -16.758669047964453}])

    testing.diff(got: got, want: want)
}

testcase polylinesimp_rdp_with_automatic_learning_testcase {
    got =
        gen.tables(n: 4, seed: 1234)
            |> polylinesimp.rdp(column: "_value", timeColumn: "_time")
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
