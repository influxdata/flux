package pkga_test


import "testing"
import "array"

testcase bar {
    option testing.tags = ["foo"]
    array.from(rows: [{}])
}

testcase untagged_extends extends "testdata/test_test.untagged" {
    option testing.tags = ["foo"]
    // Note this test is tagged with foo because of the package level
    // option statement
    super()
}

testcase duplicate {
    want = array.from(rows: [{_value: 1}])
    got = array.from(rows: [{_value: 1}])

    testing.diff(want, got)
}
