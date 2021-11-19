package array_test


import "testing"
import "array"

fromElement = (v) => array.from(rows: [{v: v}])

testcase fromElementTest {
    want = fromElement(v: 123)
    got = fromElement(v: 123)

    testing.diff(want, got)
        |> yield()
}
