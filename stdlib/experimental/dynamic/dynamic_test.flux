package dynamic_test


import "array"
import "testing"
import "experimental/dynamic"
import "internal/debug"

testcase dynamic_not_comparable {
    testing.shouldError(
        fn: () => dynamic.dynamic(v: 123) == dynamic.dynamic(v: 123),
        want: /unsupported/,
    )
}

testcase dynamic_does_not_rewrap {
    a = dynamic.dynamic(v: 123)
    b = dynamic.dynamic(v: a)

    testing.assertEqualValues(want: true, got: dynamic._equal(a, b))
}

// asArray should blow up if you feed it a dynamic that doesn't have an array in it.
testcase asArray_errors_on_nonarray {
    testing.shouldError(
        fn: () => dynamic.dynamic(v: 123) |> dynamic.asArray(),
        want: /unable to convert/,
    )
}

testcase asArray_errors_on_null {
    testing.shouldError(fn: () => debug.null() |> dynamic.asArray(), want: /unable to convert/)
}

// Verify we can pass an array of dynamic elements into dynamic() to wrap it, then unwrap it with asArray.
testcase asArray_accepts_actual_array {
    arr = dynamic.dynamic(v: [dynamic.dynamic(v: 123)]) |> dynamic.asArray()

    testing.assertEqualValues(
        want: true,
        got: dynamic._equal(a: dynamic.dynamic(v: 123), b: arr[0]),
    )
}

// This is similar to the "actual array" test except that the elements in the
// array are not wrapped in dynamic ahead of time. asArray therefore needs to
// wrap the elements prior to producing the `[dynamic]` it'll return.
testcase asArray_converts_non_dynamic_homogeneous_array_elements {
    input = [123, 456]
    converted = dynamic.dynamic(v: input) |> dynamic.asArray()

    got = {
        elm0: dynamic._equal(a: dynamic.dynamic(v: input[0]), b: converted[0]),
        elm1: dynamic._equal(a: dynamic.dynamic(v: input[1]), b: converted[1]),
    }

    testing.diff(want: array.from(rows: [{elm0: true, elm1: true}]), got: array.from(rows: [got]))
}

// Similar to the "actual array" test but using a heterogeneous array as input.
testcase dynamic_heterogeneous_array_roundtrip {
    input = [dynamic.dynamic(v: 123), dynamic.dynamic(v: 4.56)]
    converted = dynamic.dynamic(v: input) |> dynamic.asArray()

    got = {
        elm0: dynamic._equal(a: input[0], b: converted[0]),
        elm1: dynamic._equal(a: input[1], b: converted[1]),
    }

    testing.diff(want: array.from(rows: [{elm0: true, elm1: true}]), got: array.from(rows: [got]))
}
