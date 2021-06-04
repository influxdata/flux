package strings_test


import "array"
import "strings"
import "testing"

option now = () => 2030-01-01T00:00:00Z

// leading and trailing whitespace is removed from a string
testcase string_trim {
    input = array.from(
        rows: [
            {_time: 2018-05-22T19:53:26Z, _value: "trailing  ", _field: "_field", _measurement: "_measurement"},
            {_time: 2018-05-22T19:53:36Z, _value: "   leading", _field: "_field", _measurement: "_measurement"},
        ],
    )

    // XXX: rockstar (31 Mar 2021) - Adding `map` here, otherwise `testing.diff`
    // gets a type error because it thinks it's missing a _value label.
    // See https://github.com/influxdata/flux/issues/3443
    want = array.from(
        rows: [
            {_time: 2018-05-22T19:53:26Z, _value: "trailing", _field: "_field", _measurement: "_measurement"},
            {_time: 2018-05-22T19:53:36Z, _value: "leading", _field: "_field", _measurement: "_measurement"},
        ],
    )
        |> map(fn: (r) => ({r with _value: r._value}))
    got = input
        |> range(start: 2018-05-22T19:53:26Z)
        |> drop(columns: ["_start", "_stop"])
        |> map(fn: (r) => ({r with _value: strings.trimSpace(v: r._value)}))

    testing.diff(got: got, want: want)
}

// string is converted to uppercase
testcase string_toUpper {
    input = array.from(
        rows: [
            {_time: 2018-05-22T19:53:26Z, _value: "lowercase", _field: "_field", _measurement: "_measurement"},
            {_time: 2018-05-22T19:53:36Z, _value: "LoLlErCaSe", _field: "_field", _measurement: "_measurement"},
        ],
    )

    // XXX: rockstar (31 Mar 2021) - Adding `map` here, otherwise `testing.diff`
    // gets a type error because it thinks it's missing a _value label.
    // See https://github.com/influxdata/flux/issues/3443
    want = array.from(
        rows: [
            {_time: 2018-05-22T19:53:26Z, _value: "LOWERCASE", _field: "_field", _measurement: "_measurement"},
            {_time: 2018-05-22T19:53:36Z, _value: "LOLLERCASE", _field: "_field", _measurement: "_measurement"},
        ],
    )
        |> map(fn: (r) => ({r with _value: r._value}))
    got = input
        |> range(start: 2018-05-22T19:53:26Z)
        |> drop(columns: ["_start", "_stop"])
        |> map(fn: (r) => ({r with _value: strings.toUpper(v: r._value)}))

    testing.diff(got: got, want: want)
}

// string is converted to lowercase
testcase string_toLower {
    input = array.from(
        rows: [
            {_time: 2018-05-22T19:53:26Z, _value: "uppercase", _field: "_field", _measurement: "_measurement"},
            {_time: 2018-05-22T19:53:36Z, _value: "LoLlErCaSe", _field: "_field", _measurement: "_measurement"},
        ],
    )

    // XXX: rockstar (31 Mar 2021) - Adding `map` here, otherwise `testing.diff`
    // gets a type error because it thinks it's missing a _value label.
    // See https://github.com/influxdata/flux/issues/3443
    want = array.from(
        rows: [
            {_time: 2018-05-22T19:53:26Z, _value: "uppercase", _field: "_field", _measurement: "_measurement"},
            {_time: 2018-05-22T19:53:36Z, _value: "lollercase", _field: "_field", _measurement: "_measurement"},
        ],
    )
        |> map(fn: (r) => ({r with _value: r._value}))
    got = input
        |> range(start: 2018-05-22T19:53:26Z)
        |> drop(columns: ["_start", "_stop"])
        |> map(fn: (r) => ({r with _value: strings.toLower(v: r._value)}))

    testing.diff(got: got, want: want)
}

// string is converted to title case, e.g. This Is Title Case
testcase string_title {
    input = array.from(
        rows: [
            {_time: 2018-05-22T19:53:26Z, _value: "the little blue truck", _field: "_field", _measurement: "_measurement"},
        ],
    )

    // XXX: rockstar (31 Mar 2021) - Adding `map` here, otherwise `testing.diff`
    // gets a type error because it thinks it's missing a _value label.
    // See https://github.com/influxdata/flux/issues/3443
    want = array.from(
        rows: [
            {_time: 2018-05-22T19:53:26Z, _value: "The Little Blue Truck", _field: "_field", _measurement: "_measurement"},
        ],
    )
        |> map(fn: (r) => ({r with _value: r._value}))
    got = input
        |> range(start: 2018-05-22T19:53:26Z)
        |> drop(columns: ["_start", "_stop"])
        |> map(fn: (r) => ({r with _value: strings.title(v: r._value)}))

    testing.diff(got: got, want: want)
}

// A substring is returned
testcase string_substring {
    input = array.from(
        rows: [
            {_time: 2018-05-22T19:53:26Z, _value: "What's the Story, Morning Glory", _field: "_field", _measurement: "_measurement"},
        ],
    )

    // XXX: rockstar (31 Mar 2021) - Adding `map` here, otherwise `testing.diff`
    // gets a type error because it thinks it's missing a _value label.
    // See https://github.com/influxdata/flux/issues/3443
    want = array.from(
        rows: [
            {_time: 2018-05-22T19:53:26Z, _value: "Morning Glory", _field: "_field", _measurement: "_measurement"},
        ],
    )
        |> map(fn: (r) => ({r with _value: r._value}))
    got = input
        |> range(start: 2018-05-22T19:53:26Z)
        |> drop(columns: ["_start", "_stop"])
        |> map(fn: (r) => ({r with _value: strings.substring(v: r._value, start: 18, end: 31)}))

    testing.diff(got: got, want: want)
}

// All instances of one string are replaced with another
testcase string_replaceAll {
    input = array.from(
        rows: [
            {_time: 2018-05-22T19:53:26Z, _value: "This sucks. Everything sucks.", _field: "_field", _measurement: "_measurement"},
        ],
    )

    // XXX: rockstar (31 Mar 2021) - Adding `map` here, otherwise `testing.diff`
    // gets a type error because it thinks it's missing a _value label.
    // See https://github.com/influxdata/flux/issues/3443
    want = array.from(
        rows: [
            {_time: 2018-05-22T19:53:26Z, _value: "This is fine. Everything is fine.", _field: "_field", _measurement: "_measurement"},
        ],
    )
        |> map(fn: (r) => ({r with _value: r._value}))
    got = input
        |> range(start: 2018-05-22T19:53:26Z)
        |> drop(columns: ["_start", "_stop"])
        |> map(fn: (r) => ({r with _value: strings.replaceAll(v: r._value, t: "sucks", u: "is fine")}))

    testing.diff(got: got, want: want)
}

// A string is replaced by another the first specified number of times.
testcase string_replace {
    input = array.from(
        rows: [
            {_time: 2018-05-22T19:53:26Z, _value: "This sucks. Everything sucks.", _field: "_field", _measurement: "_measurement"},
        ],
    )

    // XXX: rockstar (31 Mar 2021) - Adding `map` here, otherwise `testing.diff`
    // gets a type error because it thinks it's missing a _value label.
    // See https://github.com/influxdata/flux/issues/3443
    want = array.from(
        rows: [
            {_time: 2018-05-22T19:53:26Z, _value: "This is fine. Everything sucks.", _field: "_field", _measurement: "_measurement"},
        ],
    )
        |> map(fn: (r) => ({r with _value: r._value}))
    got = input
        |> range(start: 2018-05-22T19:53:26Z)
        |> drop(columns: ["_start", "_stop"])
        |> map(fn: (r) => ({r with _value: strings.replace(v: r._value, t: "sucks", u: "is fine", i: 1)}))

    testing.diff(got: got, want: want)
}

// Calling replace in two different maps works in the correct order.
testcase string_replace_twice {
    input = array.from(
        rows: [
            {_time: 2018-05-22T19:53:26Z, _value: "This sucks. Everything sucks.", _field: "_field", _measurement: "_measurement"},
        ],
    )

    // XXX: rockstar (31 Mar 2021) - Adding `map` here, otherwise `testing.diff`
    // gets a type error because it thinks it's missing a _value label.
    // See https://github.com/influxdata/flux/issues/3443
    want = array.from(
        rows: [
            {_time: 2018-05-22T19:53:26Z, _value: "This is fine. Everything is a-okay.", _field: "_field", _measurement: "_measurement"},
        ],
    )
        |> map(fn: (r) => ({r with _value: r._value}))
        |> map(fn: (r) => ({r with _value: r._value}))
    got = input
        |> range(start: 2018-05-22T19:53:26Z)
        |> drop(columns: ["_start", "_stop"])
        |> map(fn: (r) => ({r with _value: strings.replace(v: r._value, t: "sucks", u: "is fine", i: 1)}))
        |> map(fn: (r) => ({r with _value: strings.replace(v: r._value, t: "sucks", u: "is a-okay", i: 1)}))

    testing.diff(got: got, want: want)
}

// A string's length is returned
testcase string_length {
    input = array.from(
        rows: [
            {_time: 2018-05-22T19:53:26Z, _value: "字", _field: "_field", _measurement: "_measurement"},
            {_time: 2018-05-22T19:53:26Z, _value: "Supercalifragilisticexpialidocious", _field: "_field", _measurement: "_measurement"},
        ],
    )
    want = array.from(
        rows: [
            {_time: 2018-05-22T19:53:26Z, _value: "字", _field: "_field", _measurement: "_measurement", len: 1},
            {_time: 2018-05-22T19:53:26Z, _value: "Supercalifragilisticexpialidocious", _field: "_field", _measurement: "_measurement", len: 34},
        ],
    )
    got = input
        |> range(start: 2018-05-22T19:53:26Z)
        |> drop(columns: ["_start", "_stop"])
        |> map(fn: (r) => ({r with len: strings.strlen(v: r._value)}))

    testing.diff(got: got, want: want)
}
