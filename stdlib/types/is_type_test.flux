package types_test


import "testing"
import "types"

testcase isType {
    testing.assertEqualValues(want: true, got: types.isType(v: "a", type: "string"))
}
testcase isType2 {
    testing.assertEqualValues(want: false, got: types.isType(v: "a", type: "strin"))
}
testcase isType3 {
    testing.assertEqualValues(want: false, got: types.isType(v: "a", type: "int"))
}
testcase isType4 {
    testing.assertEqualValues(want: true, got: types.isType(v: 1, type: "int"))
}
testcase isType5 {
    testing.assertEqualValues(want: true, got: types.isType(v: 2030-01-01T00:00:00Z, type: "time"))
}
testcase isType6 {
    testing.assertEqualValues(want: false, got: types.isType(v: 2030-01-01T00:00:00Z, type: "int"))
}
