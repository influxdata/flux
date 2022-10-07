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

testcase dynamic_member_access_undefined_deep {
    // is an object, but f1 does not exist, nor does f2 or f3
    a = dynamic.dynamic(v: {f0: 123}).f1.f2.f3
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

testcase dynamic_kitchen_sink {
    // The stuff this is aiming to hit:
    // - homogenous arrays
    // - heterogeneous arrays
    // - full range of types (baring vector which cannot be constructed trivially)
    // - nested combinations
    // Note: display() will order the fields in its output.
    // Also, the whitespace is tricky -- would be nice to have an easier way to compare these.
    want =
        "dynamic({
    arr1: dynamic([dynamic(1), dynamic(2)]),
    arr2: dynamic([dynamic(1), dynamic(2.3)]),
    arr3: dynamic([dynamic({x: dynamic(1), y: dynamic(2)}), dynamic({x: dynamic(1), y: dynamic(<null>), z: dynamic(3)})]),
    arr4: dynamic([dynamic([dynamic(1), dynamic(2)]), dynamic([dynamic(3), dynamic(4)])]),
    arr5: dynamic([dynamic({a: dynamic(1), b: dynamic(2)}), dynamic({a: dynamic(3), b: dynamic(4)})]),
    bfalse: dynamic(false),
    btrue: dynamic(true),
    bytes: dynamic(0x616263),
    dict: dynamic([a: 1]),
    dur: dynamic(1y),
    func: dynamic(() => bool),
    n: dynamic(<null>),
    num: dynamic(0),
    obj: dynamic({bar: dynamic(100)}),
    re: dynamic(abc\\d),
    str: dynamic(foo),
    stream: dynamic(<stream>),
    time: dynamic(2022-10-05T00:00:00.000000000Z)
})"
    got =
        display(
            v:
                dynamic.dynamic(
                    v: {
                        arr1: [1, 2],
                        arr2: [dynamic.dynamic(v: 1), dynamic.dynamic(v: 2.3)],
                        arr3: [
                            dynamic.dynamic(v: {x: 1, y: 2}),
                            dynamic.dynamic(v: {x: 1, y: debug.null(), z: 3}),
                        ],
                        arr4: [[1, 2], [3, 4]],
                        // n.b. uint renders just like int - just the number.
                        arr5: [{a: 1, b: uint(v: 2)}, {a: 3, b: uint(v: 4)}],
                        bfalse: false,
                        btrue: true,
                        // -> 0x616263
                        bytes: bytes(v: "abc"),
                        dict: ["a": 1],
                        dur: 1y,
                        func: () => true,
                        n: debug.null(),
                        num: 0,
                        obj: {bar: 100},
                        re: /abc\d/,
                        str: "foo",
                        stream: array.from(rows: [{x: 1}]),
                        time: 2022-10-05T00:00:00Z,
                    },
                ),
        )

    testing.assertEqualValues(got, want)
}

testcase dynamic_isType_string {
    testing.assertEqualValues(
        want: true,
        got: dynamic.isType(v: dynamic.dynamic(v: ""), type: "string"),
    )
}

testcase dynamic_isType_bytes {
    testing.assertEqualValues(
        want: true,
        got: dynamic.isType(v: dynamic.dynamic(v: bytes(v: "foo")), type: "bytes"),
    )
}

testcase dynamic_isType_int {
    testing.assertEqualValues(
        want: true,
        got: dynamic.isType(v: dynamic.dynamic(v: 123), type: "int"),
    )
}

testcase dynamic_isType_uint {
    testing.assertEqualValues(
        want: true,
        got: dynamic.isType(v: dynamic.dynamic(v: uint(v: 123)), type: "uint"),
    )
}

testcase dynamic_isType_float {
    testing.assertEqualValues(
        want: true,
        got: dynamic.isType(v: dynamic.dynamic(v: 1.23), type: "float"),
    )
}

testcase dynamic_isType_bool {
    testing.assertEqualValues(
        want: true,
        got: dynamic.isType(v: dynamic.dynamic(v: false), type: "bool"),
    )
}

testcase dynamic_isType_time {
    testing.assertEqualValues(
        want: true,
        got: dynamic.isType(v: dynamic.dynamic(v: 2022-10-06T00:00:00Z), type: "time"),
    )
}

testcase dynamic_isType_duration {
    testing.assertEqualValues(
        want: true,
        got: dynamic.isType(v: dynamic.dynamic(v: 1d), type: "duration"),
    )
}

testcase dynamic_isType_regexp {
    testing.assertEqualValues(
        want: true,
        got: dynamic.isType(v: dynamic.dynamic(v: /abc\d/), type: "regexp"),
    )
}

testcase dynamic_isType_array {
    testing.assertEqualValues(
        want: true,
        got: dynamic.isType(v: dynamic.dynamic(v: [1, 2, 3]), type: "array"),
    )
}

testcase dynamic_isType_object {
    testing.assertEqualValues(
        want: true,
        got: dynamic.isType(v: dynamic.dynamic(v: {foo: "bar"}), type: "object"),
    )
}

testcase dynamic_isType_function {
    testing.assertEqualValues(
        want: true,
        got: dynamic.isType(v: dynamic.dynamic(v: () => true), type: "function"),
    )
}

testcase dynamic_isType_dictionary {
    testing.assertEqualValues(
        want: true,
        got: dynamic.isType(v: dynamic.dynamic(v: [1: "one"]), type: "dictionary"),
    )
}

// isType won't let you ask if a dynamic contains a null (callers can use
// `exists` for that), but there is a case where we risk a panic if we don't
// short circuit on null inputs.
testcase dynamic_isType_null_should_not_panic {
    testing.assertEqualValues(want: false, got: dynamic.isType(v: debug.null(), type: "int"))
}

testcase dynamic_cast_bool_to_bool {
    testing.assertEqualValues(want: true, got: bool(v: dynamic.dynamic(v: true)))
}

testcase dynamic_cast_float_to_bool_error {
    testing.shouldError(fn: () => bool(v: dynamic.dynamic(v: 1.1)), want: /cannot convert float/)
}

testcase dynamic_cast_string_to_bytes {
    // bytes can't be used as column data so we need to use `display()` to
    // convert to string before we compare.
    testing.assertEqualValues(
        want: "0x616263",
        got: display(v: bytes(v: dynamic.dynamic(v: "abc"))),
    )
}

testcase dynamic_cast_int_to_bytes_error {
    testing.shouldError(fn: () => bytes(v: dynamic.dynamic(v: 1)), want: /cannot convert int/)
}

testcase dynamic_cast_int_to_duration {
    // duration can't be used as column data so we need to use `display()` to
    // convert to string before we compare.
    testing.assertEqualValues(
        want: "18ms",
        got: display(v: duration(v: dynamic.dynamic(v: 18000000))),
    )
}

testcase dynamic_cast_string_to_duration_error {
    testing.shouldError(
        fn: () => duration(v: dynamic.dynamic(v: "not a duration")),
        want: /cannot convert string/,
    )
}

testcase dynamic_cast_int_to_float {
    testing.assertEqualValues(want: 123.0, got: float(v: dynamic.dynamic(v: 123)))
}

testcase dynamic_cast_string_to_float_error {
    testing.shouldError(
        fn: () => float(v: dynamic.dynamic(v: "not a float")),
        want: /cannot convert string/,
    )
}

testcase dynamic_cast_float_to_int {
    testing.assertEqualValues(want: 123, got: int(v: dynamic.dynamic(v: 123.0)))
}

testcase dynamic_cast_string_to_int_error {
    testing.shouldError(
        fn: () => int(v: dynamic.dynamic(v: "not an int")),
        want: /cannot convert string/,
    )
}

testcase dynamic_cast_int_to_string {
    testing.assertEqualValues(want: "123", got: string(v: dynamic.dynamic(v: 123)))
}

testcase dynamic_cast_function_to_string_error {
    testing.shouldError(
        fn: () => string(v: dynamic.dynamic(v: () => true)),
        want: /cannot convert \(\) => bool to string/,
    )
}

testcase dynamic_cast_int_to_time {
    testing.assertEqualValues(want: 1970-01-01T00:00:00Z, got: time(v: dynamic.dynamic(v: 0)))
}

testcase dynamic_cast_string_to_time_error {
    testing.shouldError(
        fn: () => time(v: dynamic.dynamic(v: "not a time")),
        want: /cannot convert string/,
    )
}

testcase dynamic_cast_int_to_uint {
    testing.assertEqualValues(want: uint(v: 123), got: uint(v: dynamic.dynamic(v: 123)))
}

testcase dynamic_cast_string_to_uint_error {
    testing.shouldError(
        fn: () => uint(v: dynamic.dynamic(v: "not a uint")),
        want: /cannot convert string/,
    )
}

testcase dynamic_cast_from_json {
    want = {number: 123, msg: "flux dynamic support", isWorking: true}
    data = "{\"number\": 123,\"msg\": \"flux dynamic support\",\"isWorking\":true}"
    raw = dynamic.jsonParse(data: bytes(v: data))
    got = {number: int(v: raw.number), msg: string(v: raw.msg), isWorking: bool(v: raw.isWorking)}

    testing.assertEqualValues(want: display(v: want), got: display(v: got))
}

testcase dynamic_cast_from_json_deep {
    data =
        "
    {
      \"items\":  [
          {\"name\": \"a\", \"pos\": {\"x\": 10, \"y\": 10, \"z\": 20}},
          {\"name\": \"b\", \"pos\": {\"x\": 30, \"y\": 99}},
          {\"name\": \"c\", \"pos\": {}},
          {\"name\": \"d\"}
        ]
    }
    "
    raw = dynamic.jsonParse(data: bytes(v: data))
    converted =
        raw.items
            |> dynamic.asArray()
            |> array.map(
                fn: (x) =>
                    ({
                        name: string(v: x.name),
                        posX: int(v: x.pos.x),
                        posY: int(v: x.pos.y),
                        posZ: int(v: x.pos.z),
                    }),
            )

    want =
        array.from(
            rows: [
                {name: "a", posX: 10, posY: 10, posZ: 20},
                {name: "b", posX: 30, posY: 99, posZ: debug.null(type: "int")},
                {
                    name: "c",
                    posX: debug.null(type: "int"),
                    posY: debug.null(type: "int"),
                    posZ: debug.null(type: "int"),
                },
                {
                    name: "d",
                    posX: debug.null(type: "int"),
                    posY: debug.null(type: "int"),
                    posZ: debug.null(type: "int"),
                },
            ],
        )
    got = array.from(rows: converted)

    testing.diff(want, got)
}
