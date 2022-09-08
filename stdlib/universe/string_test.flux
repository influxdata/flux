package universe_test


import "testing"
import "csv"

testcase label_to_string {
    testing.assertEqualValues(got: string(v: .myLabel), want: "myLabel")
}
