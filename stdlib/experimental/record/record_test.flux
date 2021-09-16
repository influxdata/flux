package record_test


import "testing"
import "array"
import "experimental/record"

// Currently in Flux the defaults values constrain the type of function parameters.
// This is normally good except when you want a polymorphic parameter with a default.
// The record.any value is a polymorphic record value which can be used to allow
// parameters to have a default empty record value and still remain polymorphic.
// Once https://github.com/influxdata/flux/issues/3461 is fixed this workaround will no longer be needed.
testcase polymorphic_default {
    // define function with polymorphic default value for r
    f = (r=record.any) => ({r with x: true})

    want = array.from(
        rows: [
            {pass: true},
            {pass: true},
            {pass: true},
        ],
    )

    got = array.from(
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
