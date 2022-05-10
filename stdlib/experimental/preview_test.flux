package experimental_test


import "array"
import "experimental"
import "testing"

inData =
    array.from(
        rows: [
            {_time: 2022-05-09T00:00:00Z, t0: "a", _value: 1.0},
            {_time: 2022-05-09T00:00:00Z, t0: "a", _value: 2.0},
            {_time: 2022-05-09T00:00:00Z, t0: "a", _value: 3.0},
            {_time: 2022-05-09T00:00:00Z, t0: "b", _value: 4.0},
            {_time: 2022-05-09T00:00:00Z, t0: "b", _value: 5.0},
            {_time: 2022-05-09T00:00:00Z, t0: "b", _value: 6.0},
            {_time: 2022-05-09T00:00:00Z, t0: "c", _value: 7.0},
            {_time: 2022-05-09T00:00:00Z, t0: "c", _value: 8.0},
            {_time: 2022-05-09T00:00:00Z, t0: "c", _value: 9.0},
        ],
    )
        |> group(columns: ["t0"])

testcase basic {
    want =
        array.from(
            rows: [
                {_time: 2022-05-09T00:00:00Z, t0: "a", _value: 1.0},
                {_time: 2022-05-09T00:00:00Z, t0: "a", _value: 2.0},
                {_time: 2022-05-09T00:00:00Z, t0: "b", _value: 4.0},
                {_time: 2022-05-09T00:00:00Z, t0: "b", _value: 5.0},
            ],
        )
            |> group(columns: ["t0"])

    got =
        inData
            |> experimental.preview(nrows: 2, ntables: 2)

    testing.diff(got, want) |> yield()
}

testcase multi_buffer {
    want =
        array.from(
            rows: [
                {_time: 2022-05-09T00:00:00Z, t0: "a", _value: 1.0},
                {_time: 2022-05-09T00:00:00Z, t0: "a", _value: 2.0},
                {_time: 2022-05-09T00:00:00Z, t0: "a", _value: 3.0},
                {_time: 2022-05-09T00:00:00Z, t0: "b", _value: 4.0},
                {_time: 2022-05-09T00:00:00Z, t0: "b", _value: 5.0},
            ],
        )

    got =
        inData
            |> group()
            |> experimental.preview()

    testing.diff(got, want) |> yield()
}

testcase small {
    want =
        array.from(
            rows: [
                {_time: 2022-05-09T00:00:00Z, t0: "a", _value: 1.0},
                {_time: 2022-05-09T00:00:00Z, t0: "a", _value: 2.0},
                {_time: 2022-05-09T00:00:00Z, t0: "a", _value: 3.0},
                {_time: 2022-05-09T00:00:00Z, t0: "b", _value: 4.0},
                {_time: 2022-05-09T00:00:00Z, t0: "b", _value: 5.0},
                {_time: 2022-05-09T00:00:00Z, t0: "b", _value: 6.0},
                {_time: 2022-05-09T00:00:00Z, t0: "c", _value: 7.0},
                {_time: 2022-05-09T00:00:00Z, t0: "c", _value: 8.0},
                {_time: 2022-05-09T00:00:00Z, t0: "c", _value: 9.0},
            ],
        )
            |> group(columns: ["t0"])

    got =
        inData
            |> experimental.preview()

    testing.diff(got, want) |> yield()
}
