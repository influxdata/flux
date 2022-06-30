package universe_test


import "array"
import "testing"
import "csv"

testcase string_interpolation {
    string = "a"
    int = 1
    float = 0.1
    bool = true
    duration = 1h
    time = 2020-01-01T01:01:01Z

    got =
        array.from(
            rows: [
                {_value: "string ${string}"},
                {_value: "int ${int}"},
                {_value: "float ${float}"},
                {_value: "bool ${bool}"},
                {_value: "duration ${duration}"},
                {_value: "time ${time}"},
            ],
        )
    want =
        array.from(
            rows: [
                {_value: "string a"},
                {_value: "int 1"},
                {_value: "float 0.1"},
                {_value: "bool true"},
                {_value: "duration 1h"},
                {_value: "time 2020-01-01T01:01:01.000000000Z"},
            ],
        )

    testing.diff(got: got, want: want)
}
