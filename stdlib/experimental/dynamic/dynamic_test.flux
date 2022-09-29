package dynamic_test


import "testing"
import "experimental/dynamic"

testcase dynamic_not_comparable {
    testing.shouldError(
        fn: () => dynamic.dynamic(v: 123) == dynamic.dynamic(v: 123),
        want: /unsupported/,
    )
}
