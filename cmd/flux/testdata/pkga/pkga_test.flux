package pkga_test


import "testing"
import "array"

option testing.tags = ["foo"]

testcase bar {
    array.from(rows: [{}])
}

testcase untagged_extends extends "testdata/test_test.untagged" {
    // Note this test is tagged with foo because of the package level
    // option statement
    super()
}
