package statsmodels_test


import "internal/debug"
import "testing"
import "array"
import "contrib/anaisdg/statsmodels"

testcase linearRegressionGrouped {
    data =
        array.from(
            rows: [
                {t: "a", _value: 7.0},
                {t: "a", _value: 5.0},
                {t: "a", _value: 4.0},
                {t: "a", _value: 3.0},
                {t: "b", _value: 6.0},
                {t: "b", _value: 5.0},
                {t: "b", _value: 4.0},
                {t: "b", _value: 3.0},
            ],
        )
            |> group(columns: ["t"])

    got =
        data
            |> statsmodels.linearRegressionGrouped()

    want =
        array.from(
            rows: [
                {
                    t: "a",
                    _time: now(),
                    x: 1.0,
                    y: 7.0,
                    y_hat: 6.7,
                    slope: -1.3,
                    intercept: 8.0,
                    errors: 0.0899999999999999,
                },
                {
                    t: "a",
                    _time: now(),
                    x: 2.0,
                    y: 5.0,
                    y_hat: 5.4,
                    slope: -1.3,
                    intercept: 8.0,
                    errors: 0.16000000000000028,
                },
                {
                    t: "a",
                    _time: now(),
                    x: 3.0,
                    y: 4.0,
                    y_hat: 4.1,
                    slope: -1.3,
                    intercept: 8.0,
                    errors: 0.009999999999999929,
                },
                {
                    t: "a",
                    _time: now(),
                    x: 4.0,
                    y: 3.0,
                    y_hat: 2.8,
                    slope: -1.3,
                    intercept: 8.0,
                    errors: 0.04000000000000007,
                },
                {
                    t: "b",
                    _time: now(),
                    x: 1.0,
                    y: 6.0,
                    y_hat: 6.0,
                    slope: -1.0,
                    intercept: 7.0,
                    errors: 0.0,
                },
                {
                    t: "b",
                    _time: now(),
                    x: 2.0,
                    y: 5.0,
                    y_hat: 5.0,
                    slope: -1.0,
                    intercept: 7.0,
                    errors: 0.0,
                },
                {
                    t: "b",
                    _time: now(),
                    x: 3.0,
                    y: 4.0,
                    y_hat: 4.0,
                    slope: -1.0,
                    intercept: 7.0,
                    errors: 0.0,
                },
                {
                    t: "b",
                    _time: now(),
                    x: 4.0,
                    y: 3.0,
                    y_hat: 3.0,
                    slope: -1.0,
                    intercept: 7.0,
                    errors: 0.0,
                },
            ],
        )
            |> group(columns: ["t"])

    testing.diff(want: want, got: got)
}
