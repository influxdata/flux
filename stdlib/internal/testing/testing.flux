// Package testing provides functions for testing Flux operations.
//
// ## Metadata
// introduced: 0.182.0
//
package testing


import "array"
import "experimental"
import "testing"

// shouldErrorWithCode calls a function that catches any error and checks that the error matches the expected value.
//
// ## Parameters
// - fn: Function to call.
// - want: Regular expression to match the expected error.
// - code: Which flux error code to expect
//
// ## Examples
//
// ### Test die function errors
//
// ```no_run
// import "testing"
//
// testing.shouldErrorWithCode(fn: () => die(msg: "error message"), want: /error message/, code: 3)
// ```
//
// ## Metadata
// introduced: 0.182.0
// tags: tests
//
shouldErrorWithCode = (fn, want, code) => {
    got = experimental.catch(fn)

    return
        testing.diff(
            got: array.from(rows: [{code: got.code, match: got.msg =~ want}]),
            want: array.from(rows: [{code: code, match: true}]),
        )
}
