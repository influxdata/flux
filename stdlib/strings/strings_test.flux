package strings_test


import "array"
import "strings"
import "testing"

option now = () => 2030-01-01T00:00:00Z

// leading and trailing whitespace is removed from a string
testcase string_trim
{
        want = array.from(rows: [{_value: "trailing"}, {_value: "leading"}])
        got =
            array.from(rows: [{_value: "trailing  "}, {_value: "   leading"}])
                |> map(fn: (r) => ({_value: strings.trimSpace(v: r._value)}))

        testing.diff(got: got, want: want)
    }

// string is converted to uppercase
testcase string_toUpper
{
        want = array.from(rows: [{_value: "LOWERCASE"}, {_value: "LOLLERCASE"}])
        got =
            array.from(rows: [{_value: "lowercase"}, {_value: "LoLlErCaSe"}])
                |> map(fn: (r) => ({_value: strings.toUpper(v: r._value)}))

        testing.diff(got: got, want: want)
    }

// string is converted to lowercase
testcase string_toLower
{
        want = array.from(rows: [{_value: "uppercase"}, {_value: "lollercase"}])
        got =
            array.from(rows: [{_value: "uppercase"}, {_value: "LoLlErCaSe"}])
                |> map(fn: (r) => ({_value: strings.toLower(v: r._value)}))

        testing.diff(got: got, want: want)
    }

// string is converted to title case, e.g. This Is Title Case
testcase string_title
{
        want = array.from(rows: [{_value: "The Little Blue Truck"}])
        got =
            array.from(rows: [{_value: "the little blue truck"}])
                |> map(fn: (r) => ({_value: strings.title(v: r._value)}))

        testing.diff(got: got, want: want)
    }

// A substring is returned
testcase string_substring
{
        input =
            array.from(
                rows: [
                    {
                        _time: 2018-05-22T19:53:26Z,
                        _value: "What's the Story, Morning Glory",
                        _field: "_field",
                        _measurement: "_measurement",
                    },
                ],
            )

        // XXX: rockstar (31 Mar 2021) - Adding `map` here, otherwise `testing.diff`
        // gets a type error because it thinks it's missing a _value label.
        // See https://github.com/influxdata/flux/issues/3443
        want =
            array.from(
                rows: [
                    {
                        _time: 2018-05-22T19:53:26Z,
                        _value: "Morning Glory",
                        _field: "_field",
                        _measurement: "_measurement",
                    },
                ],
            )
                |> map(fn: (r) => ({r with _value: r._value}))
        got =
            input
                |> range(start: 2018-05-22T19:53:26Z)
                |> drop(columns: ["_start", "_stop"])
                |> map(fn: (r) => ({r with _value: strings.substring(v: r._value, start: 18, end: 31)}))

        testing.diff(got: got, want: want)
    }

// All instances of one string are replaced with another
testcase string_replaceAll
{
        want = array.from(rows: [{_value: "This is fine. Everything is fine."}])
        got =
            array.from(rows: [{_value: "This sucks. Everything sucks."}])
                |> map(fn: (r) => ({_value: strings.replaceAll(v: r._value, t: "sucks", u: "is fine")}))

        testing.diff(got: got, want: want)
    }

// A string is replaced by another the first specified number of times.
testcase string_replace
{
        want = array.from(rows: [{_value: "This is fine. Everything sucks."}])
        got =
            array.from(rows: [{_value: "This sucks. Everything sucks."}])
                |> map(fn: (r) => ({_value: strings.replace(v: r._value, t: "sucks", u: "is fine", i: 1)}))

        testing.diff(got: got, want: want)
    }

// Calling replace in two different maps works in the correct order.
testcase string_replace_twice
{
        want = array.from(rows: [{_value: "This is fine. Everything is a-okay."}])
        got =
            array.from(rows: [{_value: "This sucks. Everything sucks."}])
                |> map(fn: (r) => ({_value: strings.replace(v: r._value, t: "sucks", u: "is fine", i: 1)}))
                |> map(fn: (r) => ({_value: strings.replace(v: r._value, t: "sucks", u: "is a-okay", i: 1)}))

        testing.diff(got: got, want: want)
    }

