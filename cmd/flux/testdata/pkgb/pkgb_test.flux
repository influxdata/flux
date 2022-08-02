package pkgb_test


import "testing"
import "array"

testcase duplicate {
    want = array.from(rows: [{_value: 1}])
    got = array.from(rows: [{_value: 1}])

    testing.diff(want, got)
}
