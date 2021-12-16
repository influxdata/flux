package date_test


import "date"
import "testing"
import "array"

testcase date_timeable {
    option now = () => 2021-03-01T00:00:00Z

    want = array.from(rows: [{_value: 2}])
    got = array.from(rows: [{_value: date.month(t: -1mo)}])

    testing.diff(want: want, got: got)
}
