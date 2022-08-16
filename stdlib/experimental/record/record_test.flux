package record_test


import "testing"
import "array"
import "experimental/record"
import "json"

// Currently in Flux the defaults values constrain the type of function parameters.
// This is normally good except when you want a polymorphic parameter with a default.
// The record.any value is a polymorphic record value which can be used to allow
// parameters to have a default empty record value and still remain polymorphic.
// Once https://github.com/influxdata/flux/issues/3461 is fixed this workaround will no longer be needed.
testcase polymorphic_default {
    // define function with polymorphic default value for r
    f = (r=record.any) => ({r with x: true})

    want = array.from(rows: [{pass: true}, {pass: true}, {pass: true}])

    got =
        array.from(
            rows: [
                // call f with an empty record
                {pass: f(r: {}).x},
                // call f with a non-empty record
                {pass: f(r: {foo: 5.0}).x},
                // call f with a non-empty record shadowing x
                {pass: f(r: {x: 5.0}).x},
            ],
        )

    // TODO(nathanielc): When we have the ability to test using assertions change this code to use them.
    // Instead of using tables with testing.diff
    testing.diff(got: got, want: want)
}

testcase record_get_primitive {
    obj = {x: 1}

    want = array.from(rows: [{x: 1}, {x: 0}])

    got =
        array.from(
            rows: [
                {x: record.get(r: obj, key: "x", default: 0)},
                {x: record.get(r: obj, key: "y", default: 0)},
            ],
        )

    testing.diff(got: got, want: want)
}

testcase record_get_record {
    obj = {details: {x: 1}}

    want = array.from(rows: [{details: "{\"x\":1}"}, {details: "{}"}])

    got =
        array.from(
            rows: [
                {
                    details:
                        string(
                            v:
                                json.encode(
                                    v: record.get(r: obj, key: "details", default: record.any),
                                ),
                        ),
                },
                {
                    details:
                        string(
                            v:
                                json.encode(
                                    v: record.get(r: obj, key: "nosuchfield", default: record.any),
                                ),
                        ),
                },
            ],
        )

    testing.diff(got: got, want: want)
}
