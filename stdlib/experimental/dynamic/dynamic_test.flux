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

    testing.assertEqualValues(want: true, got: dynamic._isNotDistinct(a, b))
}

testcase dynamic_member_access_valid {
    a = dynamic.dynamic(v: {f0: 123}).f0
    b = dynamic.dynamic(v: 123)

    testing.assertEqualValues(want: true, got: dynamic._isNotDistinct(a, b))
}

testcase dynamic_member_access_invalid {
    // not an object
    a = dynamic.dynamic(v: 123).f1
    b = dynamic.dynamic(v: debug.null())

    testing.assertEqualValues(want: true, got: dynamic._isNotDistinct(a, b))
}
testcase dynamic_member_access_undefined {
    // is an object, but f1 does not exist
    a = dynamic.dynamic(v: {f0: 123}).f1
    b = dynamic.dynamic(v: debug.null())

    testing.assertEqualValues(want: true, got: dynamic._isNotDistinct(a, b))
}

testcase dynamic_index_access_valid {
    a = dynamic.dynamic(v: [123])[0]
    b = dynamic.dynamic(v: 123)

    testing.assertEqualValues(want: true, got: dynamic._isNotDistinct(a, b))
}

testcase dynamic_index_access_invalid {
    // not an array
    a = dynamic.dynamic(v: 123)[0]
    b = dynamic.dynamic(v: debug.null())

    testing.assertEqualValues(want: true, got: dynamic._isNotDistinct(a, b))
}

testcase dynamic_index_access_oob {
    // is an array, but index is out of bounds
    a = dynamic.dynamic(v: [123])[4]
    b = dynamic.dynamic(v: debug.null())

    testing.assertEqualValues(want: true, got: dynamic._isNotDistinct(a, b))
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
        got: dynamic._isNotDistinct(a: dynamic.dynamic(v: 123), b: arr[0]),
    )
}

// This is similar to the "actual array" test except that the elements in the
// array are not wrapped in dynamic ahead of time. asArray therefore needs to
// wrap the elements prior to producing the `[dynamic]` it'll return.
testcase asArray_converts_non_dynamic_homogeneous_array_elements {
    input = [123, 456]
    converted = dynamic.dynamic(v: input) |> dynamic.asArray()

    got = {
        elm0: dynamic._isNotDistinct(a: dynamic.dynamic(v: input[0]), b: converted[0]),
        elm1: dynamic._isNotDistinct(a: dynamic.dynamic(v: input[1]), b: converted[1]),
    }

    testing.diff(want: array.from(rows: [{elm0: true, elm1: true}]), got: array.from(rows: [got]))
}

// Similar to the "actual array" test but using a heterogeneous array as input.
testcase dynamic_heterogeneous_array_roundtrip {
    input = [dynamic.dynamic(v: 123), dynamic.dynamic(v: 4.56)]
    converted = dynamic.dynamic(v: input) |> dynamic.asArray()

    got = {
        elm0: dynamic._isNotDistinct(a: input[0], b: converted[0]),
        elm1: dynamic._isNotDistinct(a: input[1], b: converted[1]),
    }

    testing.diff(want: array.from(rows: [{elm0: true, elm1: true}]), got: array.from(rows: [got]))
}

jsonArray = "[0,\"foo\",true,false,{\"bar\":100},[1,2],null]"
jsonObject =
    "{\"arr\":[1,2],\"bfalse\":false,\"btrue\":true,\"n\":null,\"num\":0,\"obj\":{\"bar\":100},\"str\":\"foo\"}"

testcase dynamic_json_parse_array {
    want =
        "dynamic([
    dynamic(0),
    dynamic(foo),
    dynamic(true),
    dynamic(false),
    dynamic({bar: dynamic(100)}),
    dynamic([dynamic(1), dynamic(2)]),
    dynamic(<null>)
])"
    got = display(v: dynamic.jsonParse(data: bytes(v: jsonArray)))

    testing.assertEqualValues(got, want)
}

testcase dynamic_json_parse_object {
    want =
        "dynamic({
    arr: dynamic([dynamic(1), dynamic(2)]),
    bfalse: dynamic(false),
    btrue: dynamic(true),
    n: dynamic(<null>),
    num: dynamic(0),
    obj: dynamic({bar: dynamic(100)}),
    str: dynamic(foo)
})"
    got = display(v: dynamic.jsonParse(data: bytes(v: jsonObject)))

    testing.assertEqualValues(got, want)
}

testcase dynamic_json_encode {
    want = array.from(rows: [{name: "array", data: jsonArray}, {name: "object", data: jsonObject}])

    got =
        want
            |> map(
                fn: (r) => {
                    roundtrip = dynamic.jsonEncode(v: dynamic.jsonParse(data: bytes(v: r.data)))

                    return {name: r.name, data: string(v: roundtrip)}
                },
            )

    testing.diff(got, want)
}
