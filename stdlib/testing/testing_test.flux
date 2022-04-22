package testing_test


import "testing"
import "array"
import "csv"

testcase test_option {
    // Option value in parent test case
    // See extend_test.flux that contains a test case
    // that extends this test case
    option x = 1

    want = array.from(rows: [{_value: 1}])
    got = array.from(rows: [{_value: x}])

    testing.diff(want: want, got: got)
}
