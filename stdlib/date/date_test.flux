package date_test


import "date"
import "testing"
import "array"
import "timezone"

testcase date_timeable {
    option now = () => 2021-03-01T00:00:00Z

    want = array.from(rows: [{_value: 2}])
    got = array.from(rows: [{_value: date.month(t: -1mo)}])

    testing.diff(want: want, got: got)
}

testcase time_location {
    option location = timezone.location(name: "Asia/Kolkata")

    want = array.from(rows: [{_time: 2021-03-01T05:30:00Z}])
    got = array.from(rows: [{_time: date.time(t: 2021-03-01T00:00:00Z)}])

    testing.diff(want: want, got: got)
}

testcase duration_location {
    option location = timezone.location(name: "Asia/Kolkata")
    option now = () => 2021-03-01T00:00:00Z

    want = array.from(rows: [{_time: 2021-03-01T06:30:00Z}, {_time: 2021-04-09T06:30:00Z}])
    got = array.from(rows: [{_time: date.time(t: 1h)}, {_time: date.time(t: 1mo1w1d1h)}])

    testing.diff(want: want, got: got)
}
