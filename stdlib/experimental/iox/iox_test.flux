package iox_test


import "array"
import "testing"
import "experimental/iox"

testcase iox_sql_interval_testcase {
    got =
        array.from(
            rows: [
                {_value: "12y3mo4w"},
                {_value: "12d34h5m"},
                {_value: "123s56ms"},
                {_value: "123us56ns"},
            ],
        )
            |> map(fn: (r) => ({_value: iox.sqlInterval(d: duration(v: r._value))}))

    want =
        array.from(
            rows: [
                {_value: "12 years 3 months 4 weeks"},
                {_value: "1 weeks 6 days 10 hours 5 minutes"},
                {_value: "2 minutes 3 seconds 56 milliseconds"},
                {_value: "123 microseconds 56 nanoseconds"},
            ],
        )

    testing.diff(got: got, want: want)
}
