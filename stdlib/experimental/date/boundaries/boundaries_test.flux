package boundaries_test


import "array"
import "testing"
import "timezone"
import "experimental/date/boundaries"

testcase yesterday_test {
    option now = () => 2022-06-01T12:20:11Z

    ret = boundaries.yesterday()

    want = array.from(rows: [{_value_a: 2022-05-31T00:00:00Z, _value_b: 2022-06-01T00:00:00Z}])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want: want, got: got)
}

testcase yesterday_test_t {
    option now = () => 2018-10-12T14:20:11Z

    ret = boundaries.yesterday()

    want = array.from(rows: [{_value_a: 2018-10-11T00:00:00Z, _value_b: 2018-10-12T00:00:00Z}])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want: want, got: got)
}

testcase week_start_default_sunday_test {
    option now = () => 2022-06-06T10:02:10Z

    ret = boundaries.week(start_sunday: true)

    want = array.from(rows: [{_value_a: 2022-06-05T00:00:00Z, _value_b: 2022-06-12T00:00:00Z}])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want: want, got: got)
}

testcase week_start_default_monday_test {
    option now = () => 2022-06-08T14:20:11Z

    ret = boundaries.week(start_sunday: false)

    want = array.from(rows: [{_value_a: 2022-06-06T00:00:00Z, _value_b: 2022-06-13T00:00:00Z}])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want: want, got: got)
}

testcase week_start_default_monday_offset_test {
    option now = () => 2022-06-08T14:20:11Z

    ret = boundaries.week(week_offset: 1, start_sunday: false)

    want = array.from(rows: [{_value_a: 2022-06-13T00:00:00Z, _value_b: 2022-06-20T00:00:00Z}])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want: want, got: got)
}

testcase week_start_default_monday_two_test {
    option now = () => 2022-06-08T14:20:11Z

    ret = boundaries.week(week_offset: -1, start_sunday: false)

    want = array.from(rows: [{_value_a: 2022-05-30T00:00:00Z, _value_b: 2022-06-06T00:00:00Z}])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want: want, got: got)
}

testcase week_current_week_for_monday {
    option now = () => 2022-07-25T14:20:11Z

    ret = boundaries.week()
    want = array.from(rows: [{_value_a: 2022-07-25T00:00:00Z, _value_b: 2022-08-01T00:00:00Z}])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want: want, got: got)
}

testcase week_current_week_for_sunday {
    option now = () => 2022-07-24T14:20:11Z

    ret = boundaries.week(start_sunday: true)
    want = array.from(rows: [{_value_a: 2022-07-24T00:00:00Z, _value_b: 2022-07-31T00:00:00Z}])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want: want, got: got)
}

testcase month_start_one_test {
    option now = () => 2021-03-10T22:10:00Z

    ret = boundaries.month()

    want = array.from(rows: [{_value_a: 2021-03-01T00:00:00Z, _value_b: 2021-04-01T00:00:00Z}])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want: want, got: got)
}

testcase month_start_two_test {
    option now = () => 2020-12-10T22:10:00Z

    ret = boundaries.month()

    want = array.from(rows: [{_value_a: 2020-12-01T00:00:00Z, _value_b: 2021-01-01T00:00:00Z}])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want: want, got: got)
}

testcase month_start_two_offset_test {
    option now = () => 2020-12-10T22:10:00Z

    ret = boundaries.month(month_offset: 1)

    want = array.from(rows: [{_value_a: 2021-01-01T00:00:00Z, _value_b: 2021-02-01T00:00:00Z}])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want: want, got: got)
}

testcase month_start_two_offset_test_t {
    option now = () => 2020-12-10T22:10:00Z

    ret = boundaries.month(month_offset: -2)

    want = array.from(rows: [{_value_a: 2020-10-01T00:00:00Z, _value_b: 2020-11-01T00:00:00Z}])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want: want, got: got)
}

testcase monday_test_timeable {
    option now = () => 2021-03-05T12:10:11Z

    ret = boundaries.monday()

    want = array.from(rows: [{_value_a: 2021-03-01T00:00:00Z, _value_b: 2021-03-02T00:00:00Z}])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want: want, got: got)
}

testcase monday_test_two_timeable {
    option now = () => 2021-12-30T00:40:44Z

    ret = boundaries.monday()

    want = array.from(rows: [{_value_a: 2021-12-27T00:00:00Z, _value_b: 2021-12-28T00:00:00Z}])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want: want, got: got)
}

testcase monday_test_tz_timeable {
    option now = () => 2021-12-30T00:40:44Z

    //pdt
    option location = timezone.fixed(offset: -8h)

    ret = boundaries.monday()

    want = array.from(rows: [{_value_a: 2021-12-27T08:00:00Z, _value_b: 2021-12-28T08:00:00Z}])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want: want, got: got)
}

testcase tuesday_test_timeable {
    option now = () => 2021-03-05T12:10:11Z

    ret = boundaries.tuesday()

    want = array.from(rows: [{_value_a: 2021-03-02T00:00:00Z, _value_b: 2021-03-03T00:00:00Z}])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want: want, got: got)
}

testcase tuesday_test_two_timeable {
    option now = () => 2021-12-30T00:40:44Z

    ret = boundaries.tuesday()

    want = array.from(rows: [{_value_a: 2021-12-28T00:00:00Z, _value_b: 2021-12-29T00:00:00Z}])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want: want, got: got)
}

testcase tuesday_test_two_timeable_t {
    option now = () => 2021-12-30T00:40:44Z
    option location = timezone.fixed(offset: 6h)

    ret = boundaries.tuesday()

    want = array.from(rows: [{_value_a: 2021-12-27T18:00:00Z, _value_b: 2021-12-28T18:00:00Z}])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want: want, got: got)
}

testcase wednesday_test_timeable {
    option now = () => 2022-01-01T00:40:44Z

    ret = boundaries.wednesday()

    want = array.from(rows: [{_value_a: 2021-12-29T00:00:00Z, _value_b: 2021-12-30T00:00:00Z}])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want: want, got: got)
}

testcase wednesday_test_two_timeable {
    option now = () => 2021-12-05T12:10:11Z

    ret = boundaries.wednesday()

    want = array.from(rows: [{_value_a: 2021-12-01T00:00:00Z, _value_b: 2021-12-02T00:00:00Z}])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want: want, got: got)
}

testcase wednesdady_tz_test_timeable {
    option now = () => 2021-12-05T12:10:11Z

    //tehran IRST at this time of year
    option location = timezone.fixed(offset: 3h30m)

    ret = boundaries.wednesday()

    want = array.from(rows: [{_value_a: 2021-11-30T20:30:00Z, _value_b: 2021-12-01T20:30:00Z}])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want: want, got: got)
}

testcase thursday_test_timeable {
    option now = () => 2022-01-10T00:40:44Z

    ret = boundaries.thursday()

    want = array.from(rows: [{_value_a: 2022-01-06T00:00:00Z, _value_b: 2022-01-07T00:00:00Z}])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want: want, got: got)
}

testcase thursday_test_two_timeable {
    option now = () => 2022-01-21T12:10:11Z

    ret = boundaries.thursday()

    want = array.from(rows: [{_value_a: 2022-01-20T00:00:00Z, _value_b: 2022-01-21T00:00:00Z}])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want: want, got: got)
}

testcase thursday_test_tzt_timeable {
    option now = () => 2022-01-21T12:10:11Z
    option location = timezone.fixed(offset: 1h)

    ret = boundaries.thursday()

    want = array.from(rows: [{_value_a: 2022-01-19T23:00:00Z, _value_b: 2022-01-20T23:00:00Z}])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want: want, got: got)
}

testcase thursday_timeable_test_africa {
    option now = () => 2022-01-21T12:10:11Z
    option location = timezone.location(name: "Africa/Algiers")

    ret = boundaries.thursday()

    want = array.from(rows: [{_value_a: 2022-01-19T23:00:00Z, _value_b: 2022-01-20T23:00:00Z}])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want: want, got: got)
}

testcase friday_test_one_timeable {
    option now = () => 2022-01-23T12:10:11Z

    ret = boundaries.friday()

    want = array.from(rows: [{_value_a: 2022-01-21T00:00:00Z, _value_b: 2022-01-22T00:00:00Z}])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want: want, got: got)
}

testcase friday_test_two_timeable {
    option now = () => 2022-01-03T12:10:11Z

    ret = boundaries.friday()

    want = array.from(rows: [{_value_a: 2021-12-31T00:00:00Z, _value_b: 2022-01-01T00:00:00Z}])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want: want, got: got)
}

testcase friday_test_three_timeable {
    option now = () => 2022-01-25T12:10:11Z

    ret = boundaries.friday()

    want = array.from(rows: [{_value_a: 2022-01-21T00:00:00Z, _value_b: 2022-01-22T00:00:00Z}])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want: want, got: got)
}

testcase friday_tz_test_timeable {
    option now = () => 2022-01-25T12:10:11Z
    option location = timezone.fixed(offset: -10h)

    ret = boundaries.friday()

    want = array.from(rows: [{_value_a: 2022-01-21T10:00:00Z, _value_b: 2022-01-22T10:00:00Z}])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want: want, got: got)
}

testcase saturday_test_timeable {
    option now = () => 2022-01-25T12:10:11Z

    ret = boundaries.saturday()

    want = array.from(rows: [{_value_a: 2022-01-22T00:00:00Z, _value_b: 2022-01-23T00:00:00Z}])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want: want, got: got)
}

testcase saturday_test_two_timeable {
    option now = () => 2022-01-15T12:10:11Z

    ret = boundaries.saturday()

    want = array.from(rows: [{_value_a: 2022-01-08T00:00:00Z, _value_b: 2022-01-09T00:00:00Z}])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want: want, got: got)
}

testcase saturday_savings_test {
    option now = () => 2022-11-08T12:10:11Z
    option location = timezone.location(name: "America/Los_Angeles")

    //goes -8 to -7
    ret = boundaries.sunday()

    want = array.from(rows: [{_value_a: 2022-11-06T07:00:00Z, _value_b: 2022-11-07T08:00:00Z}])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want: want, got: got)
}

testcase sunday_test_one_timeable {
    option now = () => 2022-01-22T12:10:11Z

    ret = boundaries.sunday()

    want = array.from(rows: [{_value_a: 2022-01-16T00:00:00Z, _value_b: 2022-01-17T00:00:00Z}])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want: want, got: got)
}

testcase sunday_test_two_timeable {
    option now = () => 2022-01-24T12:10:11Z

    ret = boundaries.sunday()

    want = array.from(rows: [{_value_a: 2022-01-23T00:00:00Z, _value_b: 2022-01-24T00:00:00Z}])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want: want, got: got)
}

testcase monday_test_tz_named_st {
    // DST begins March 14, ends Nov 7.
    option now = () => 2021-12-30T00:40:44Z
    option location = timezone.location(name: "America/Los_Angeles")

    ret = boundaries.monday()

    want = array.from(rows: [{_value_a: 2021-12-27T08:00:00Z, _value_b: 2021-12-28T08:00:00Z}])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want: want, got: got)
}

testcase monday_test_tz_named_dst {
    // DST begins March 14, ends Nov 7.
    option now = () => 2021-10-29T00:40:44Z
    option location = timezone.location(name: "America/Los_Angeles")

    ret = boundaries.monday()

    want = array.from(rows: [{_value_a: 2021-10-25T07:00:00Z, _value_b: 2021-10-26T07:00:00Z}])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want: want, got: got)
}

testcase month_straddling_offset_change {
    option now = () => 2021-11-21T00:40:44Z
    option location = timezone.location(name: "America/Los_Angeles")

    ret = boundaries.month()

    want = array.from(rows: [{_value_a: 2021-11-01T07:00:00Z, _value_b: 2021-12-01T08:00:00Z}])
    got = array.from(rows: [{_value_a: ret.start, _value_b: ret.stop}])

    testing.diff(want: want, got: got)
}
