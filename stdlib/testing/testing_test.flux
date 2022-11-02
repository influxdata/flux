package testing_test


import internalTesting "internal/testing"
import "testing"
import "experimental"
import "array"
import "csv"
import "json"

testcase test_option {
    // Option value in parent test case
    // See extend_test.flux that contains a test case
    // that extends this test case
    option x = 1

    want = array.from(rows: [{_value: 1}])
    got = array.from(rows: [{_value: x}])

    testing.diff(want: want, got: got)
}

testcase succeed_on_non_empty_result {
    // Just to make sure we yield something to `error`
    testing.assertEqualValues(got: 0, want: 0)

    // non-empty result
    array.from(rows: [{}])
}

testcase test_should_error {
    testing.shouldError(fn: () => die(msg: "error message"), want: /error message$/)
}

testcase test_should_error_should_error_when_no_error {
    got = experimental.catch(fn: () => testing.shouldError(fn: () => "abc", want: /error message$/))

    internalTesting.assertMatches(got: got.msg, want: /shouldError expected an error/)
}
