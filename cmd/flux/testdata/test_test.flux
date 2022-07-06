package test_test


import "array"
import "testing"

testcase a {
    option testing.tags = ["a"]

    array.from(rows: [{}])
}
testcase ab {
    option testing.tags = ["a", "b"]

    array.from(rows: [{}])
}
testcase abc {
    option testing.tags = ["a", "b", "c"]

    array.from(rows: [{}])
}

testcase fails {
    option testing.tags = ["fail"]

    want = array.from(rows: [{_value: 1}])
    got = array.from(rows: [{_value: 0}])

    testing.diff(want, got)
}
testcase untagged {
    array.from(rows: [{}])
}
