package date_test


import "array"
import "date"
import "testing"

option now = () => 2022-05-10T00:00:00Z

scaleTime = (d, n) => date.add(to: now(), d: date.scale(d: d, n: n))

testcase scale {
    got =
        array.from(
            rows: [
                {_value: scaleTime(d: 1h, n: 1)},
                {_value: scaleTime(d: 1h, n: 2)},
                {_value: scaleTime(d: 1h, n: 3)},
                {_value: scaleTime(d: 1h, n: 4)},
            ],
        )

    want =
        array.from(
            rows: [
                {_value: 2022-05-10T01:00:00Z},
                {_value: 2022-05-10T02:00:00Z},
                {_value: 2022-05-10T03:00:00Z},
                {_value: 2022-05-10T04:00:00Z},
            ],
        )

    testing.diff(want, got) |> yield()
}
