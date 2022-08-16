package strings_test


import "internal/debug"
import "experimental/array"
import "strings"
import "testing"

option now = () => 2030-01-01T00:00:00Z

// leading and trailing whitespace is removed from a string
testcase string_trim {
    want = array.from(rows: [{_value: "trailing"}, {_value: "leading"}])
    got =
        array.from(rows: [{_value: "trailing  "}, {_value: "   leading"}])
            |> map(fn: (r) => ({_value: strings.trimSpace(v: r._value)}))

    testing.diff(got: got, want: want)
}

// string is converted to uppercase
testcase string_toUpper {
    want = array.from(rows: [{_value: "LOWERCASE"}, {_value: "LOLLERCASE"}])
    got =
        array.from(rows: [{_value: "lowercase"}, {_value: "LoLlErCaSe"}])
            |> map(fn: (r) => ({_value: strings.toUpper(v: r._value)}))

    testing.diff(got: got, want: want)
}

// string is converted to lowercase
testcase string_toLower {
    want = array.from(rows: [{_value: "uppercase"}, {_value: "lollercase"}])
    got =
        array.from(rows: [{_value: "uppercase"}, {_value: "LoLlErCaSe"}])
            |> map(fn: (r) => ({_value: strings.toLower(v: r._value)}))

    testing.diff(got: got, want: want)
}

// string is converted to title case, e.g. This Is Title Case
testcase string_title {
    want = array.from(rows: [{_value: "The Little Blue Truck"}])
    got =
        array.from(rows: [{_value: "the little blue truck"}])
            |> map(fn: (r) => ({_value: strings.title(v: r._value)}))

    testing.diff(got: got, want: want)
}

testcase string_substring {
    str = "What's the Story, Morning Glory ab£"
    bounds = [
        {start: 18, end: 31, str: "What's the Story, Morning Glory", want: "Morning Glory"},
        {start: 6, end: 11, str: "hello world", want: "world"},
        {start: 8, end: 14, str: "convert £ to €", want: "£ to €"},
        {start: 8, end: 13, str: "convert £ to €", want: "£ to "},
        {start: -1, end: 5, str: "start out of bounds", want: "start"},
        {start: 11, end: 100, str: "end out of bounds", want: "bounds"},
        {start: 5, end: -1, str: "end <= start", want: ""},
        // note end <= start here because start is past the actual end of the string
        {start: 30, end: 100, str: "end <= start", want: ""},
    ]
    want = array.from(rows: bounds |> array.map(fn: (x) => ({_value: x.want})))
    got =
        array.from(
            rows:
                bounds
                    |> array.map(
                        fn: (x) =>
                            ({_value: strings.substring(v: x.str, start: x.start, end: x.end)}),
                    ),
        )

    testing.diff(got: got, want: want)
}

testcase string_substring_nbsp {
    // XXX: Inputs of a certain sizes with trailing nbsp caused a panic
    // <https://github.com/influxdata/EAR/issues/3494>
    input = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxx  "
    output = strings.substring(v: input, start: 0, end: 100)
    want = array.from(rows: [{v: input}])
    got = array.from(rows: [{v: output}])

    testing.diff(want, got)
}

// All instances of one string are replaced with another
testcase string_replaceAll {
    want = array.from(rows: [{_value: "This is fine. Everything is fine."}])
    got =
        array.from(rows: [{_value: "This sucks. Everything sucks."}])
            |> map(fn: (r) => ({_value: strings.replaceAll(v: r._value, t: "sucks", u: "is fine")}))

    testing.diff(got: got, want: want)
}

// A string is replaced by another the first specified number of times.
testcase string_replace {
    want = array.from(rows: [{_value: "This is fine. Everything sucks."}])
    got =
        array.from(rows: [{_value: "This sucks. Everything sucks."}])
            |> map(
                fn: (r) => ({_value: strings.replace(v: r._value, t: "sucks", u: "is fine", i: 1)}),
            )

    testing.diff(got: got, want: want)
}

// Calling replace in two different maps works in the correct order.
testcase string_replace_twice {
    want = array.from(rows: [{_value: "This is fine. Everything is a-okay."}])
    got =
        array.from(rows: [{_value: "This sucks. Everything sucks."}])
            |> map(
                fn: (r) => ({_value: strings.replace(v: r._value, t: "sucks", u: "is fine", i: 1)}),
            )
            |> map(
                fn: (r) =>
                    ({_value: strings.replace(v: r._value, t: "sucks", u: "is a-okay", i: 1)}),
            )

    testing.diff(got: got, want: want)
}

// A string's length is returned
testcase string_length {
    want =
        array.from(
            rows: [{_value: "字", len: 1}, {_value: "Supercalifragilisticexpialidocious", len: 34}],
        )
    got =
        array.from(rows: [{_value: "字"}, {_value: "Supercalifragilisticexpialidocious"}])
            |> map(fn: (r) => ({_value: r._value, len: strings.strlen(v: r._value)}))

    testing.diff(got: got, want: want)
}
