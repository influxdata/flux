package testing


import "array"
import "experimental"
import "regexp"
import "testing"
import internalTesting "internal/testing"

testcase test_assert_matches {
    internalTesting.assertMatches(got: "44444", want: /4+/)
}

testcase test_assert_matches_should_error {
    testing.shouldError(
        fn: () => internalTesting.assertMatches(got: "", want: /4+/),
        want: /Regex `4\+` does not match ``/,
    )
}
