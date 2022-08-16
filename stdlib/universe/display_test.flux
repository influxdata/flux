package universe_test


import "array"
import "testing"

testcase display_record {
    want = array.from(rows: [{display: "{a: 1, b: 2, c: 3}"}])
    d = display(v: {a: 1, b: 2, c: 3})
    got = array.from(rows: [{display: d}])

    testing.diff(want: want, got: got)
}
testcase display_array {
    want = array.from(rows: [{display: "[1, 2, 3]"}])
    d = display(v: [1, 2, 3])
    got = array.from(rows: [{display: d}])

    testing.diff(want: want, got: got)
}
testcase display_dictionary {
    want = array.from(rows: [{display: "[a: 1, b: 2, c: 3]"}])
    d = display(v: ["a": 1, "b": 2, "c": 3])
    got = array.from(rows: [{display: d}])

    testing.diff(want: want, got: got)
}
testcase display_bytes {
    want = array.from(rows: [{display: "0x616263"}])
    d = display(v: bytes(v: "abc"))
    got = array.from(rows: [{display: d}])

    testing.diff(want: want, got: got)
}
testcase display_composite {
    want =
        array.from(
            rows:
                [
                    {
                        display:
                            "{
    array: [1, 2, 3],
    bytes: 0x616263,
    dict: [a: 1, b: 2, c: 3],
    string: str
}",
                    },
                ],
        )

    d =
        display(
            v: {
                bytes: bytes(v: "abc"),
                string: "str",
                array: [1, 2, 3],
                dict: ["a": 1, "b": 2, "c": 3],
            },
        )

    got = array.from(rows: [{display: d}])

    testing.diff(want: want, got: got)
}
