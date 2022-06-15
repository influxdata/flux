package testing_test


import "testing"
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

testcase test_should_error {

    testing.shouldError(fn: () => json.encode(v: array.from(rows: [{}])), want: "error calling function \"encode\" @23:35-23:73: got table stream instead of array. Try using tableFind() or findRecord() to extract data from stream")
}
