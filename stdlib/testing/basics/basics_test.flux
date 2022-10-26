package basics


import "testing"

testcase addition {
    got = 1 + 1
    want = 2

    testing.assertEqualValues(got, want)
}
