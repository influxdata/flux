package universe_test


import "array"
import "universe"
import "testing"

testcase us_pacific_daily {
    option location = "America/Los_Angeles"

    got = array.from(
        rows: [
            {_time: 2017-02-24T12:00:00-08:00},
            {_time: 2017-09-03T12:00:00-07:00},
            {_time: 2017-03-12T03:00:00-07:00},
            {_time: 2017-11-05T01:30:00-08:00},
        ],
    )
        |> window(every: 1d)

    want = array.from(
        rows: [
            {_time: 2017-02-24T12:00:00-08:00, _start: 2017-02-24T00:00:00-08:00, _stop: 2017-02-25T00:00:00-08:00},
            {_time: 2017-09-03T12:00:00-07:00, _start: 2017-09-03T00:00:00-07:00, _stop: 2017-09-04T00:00:00-07:00},
            {_time: 2017-03-12T03:00:00-07:00, _start: 2017-03-12T00:00:00-08:00, _stop: 2017-03-13T00:00:00-07:00},
            {_time: 2017-11-05T01:30:00-08:00, _start: 2017-11-05T00:00:00-07:00, _stop: 2017-11-06T00:00:00-08:00},
        ],
    )
        |> group(columns: ["_start", "_stop"])

    testing.diff(got: got, want: want) |> yield()
}

testcase us_pacific_offset {
    option location = "America/Los_Angeles"

    got = array.from(
        rows: [
            {_time: 2017-03-12T01:45:00-08:00},
            {_time: 2017-11-05T01:45:00-08:00},
        ],
    )
        |> window(every: 1h, offset: 30m)

    want = array.from(
        rows: [
            {_time: 2017-03-12T01:45:00-08:00, _start: 2017-03-12T01:30:00-08:00, _stop: 2017-03-12T03:00:00-07:00},
            {_time: 2017-11-05T01:45:00-08:00, _start: 2017-11-05T01:30:00-07:00, _stop: 2017-11-05T02:30:00-08:00},
        ],
    )
        |> group(columns: ["_start", "_stop"])

    testing.diff(got: got, want: want) |> yield()
}

testcase australia_east_daily {
    option location = "Australia/Sydney"

    got = array.from(
        rows: [
            {_time: 2017-09-03T12:00:00+10:00},
            {_time: 2017-02-24T12:00:00+11:00},
            {_time: 2017-10-01T03:00:00+11:00},
            {_time: 2017-04-02T02:30:00+10:00},
        ],
    )
        |> window(every: 1d)

    want = array.from(
        rows: [
            {_time: 2017-09-03T12:00:00+10:00, _start: 2017-09-03T00:00:00+10:00, _stop: 2017-09-04T00:00:00+10:00},
            {_time: 2017-02-24T12:00:00+11:00, _start: 2017-02-24T00:00:00+11:00, _stop: 2017-02-25T00:00:00+11:00},
            {_time: 2017-10-01T03:00:00+11:00, _start: 2017-10-01T00:00:00+10:00, _stop: 2017-10-02T00:00:00+11:00},
            {_time: 2017-04-02T02:30:00+10:00, _start: 2017-04-02T00:00:00+11:00, _stop: 2017-04-03T00:00:00+10:00},
        ],
    )
        |> group(columns: ["_start", "_stop"])

    testing.diff(got: got, want: want) |> yield()
}

testcase australia_east_offset {
    option location = "Australia/Sydney"

    got = array.from(
        rows: [
            {_time: 2017-10-01T01:45:00+10:00},
            {_time: 2017-04-02T02:45:00+10:00},
        ],
    )
        |> window(every: 1h, offset: 30m)

    want = array.from(
        rows: [
            {_time: 2017-10-01T01:45:00+10:00, _start: 2017-10-01T01:30:00+10:00, _stop: 2017-10-01T03:00:00+11:00},
            {_time: 2017-04-02T02:45:00+10:00, _start: 2017-04-02T02:30:00+11:00, _stop: 2017-04-02T03:30:00+10:00},
        ],
    )
        |> group(columns: ["_start", "_stop"])

    testing.diff(got: got, want: want) |> yield()
}

testcase american_samoa_day_skip {
    option location = "Pacific/Apia"

    got = array.from(
        rows: [
            {_time: 2011-12-29T16:00:00-10:00},
            {_time: 2011-12-31T04:00:00+14:00},
        ],
    )
        |> window(every: 1d, offset: 12h)

    want = array.from(
        rows: [
            {_time: 2011-12-29T16:00:00-10:00, _start: 2011-12-29T12:00:00-10:00, _stop: 2011-12-31T00:00:00+14:00},
            {_time: 2011-12-31T04:00:00+14:00, _start: 2011-12-31T00:00:00+14:00, _stop: 2011-12-31T12:00:00+14:00},
        ],
    )
        |> group(columns: ["_start", "_stop"])

    testing.diff(got: got, want: want) |> yield()
}
