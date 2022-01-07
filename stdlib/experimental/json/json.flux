// Package json provides tools for working with JSON.
//
// introduced: 0.69.0
// tags: json
//
package json


// parse takes JSON data as bytes and returns a value.
//
// The function can return lists, records, strings, booleans, and float values.
// All numeric values are returned as floats.
//
// ## Parameters
// - data: JSON data (as bytes) to parse.
//
// ## Examples
// Parse and use JSON data to restructure tables
// ```
// # import "array"
// import "experimental/json"
// #
// # data = array.from(
// #     rows: [
// #         {_time: 2021-01-01T00:00:00Z, _field: "foo", _value: "{\"a\":1,\"b\":2,\"c\":3}"},
// #         {_time: 2021-01-01T01:00:00Z, _field: "foo", _value: "{\"a\":4,\"b\":5,\"c\":6}"},
// #         {_time: 2021-01-01T02:00:00Z, _field: "foo", _value: "{\"a\":7,\"b\":8,\"c\":9}"},
// #         {_time: 2021-01-01T00:00:00Z, _field: "bar", _value: "{\"a\":10,\"b\":9,\"c\":8}"},
// #         {_time: 2021-01-01T01:00:00Z, _field: "bar", _value: "{\"a\":7,\"b\":6,\"c\":5}"},
// #         {_time: 2021-01-01T02:00:00Z, _field: "bar", _value: "{\"a\":4,\"b\":3,\"c\":2}"},
// #     ],
// # )
//
// < data
//     |> map(
//         fn: (r) => {
//             jsonData = json.parse(data: bytes(v: r._value))
//
//             return {
//                 _time: r._time,
//                 _field: r._field,
//                 a: jsonData.a,
//                 b: jsonData.b,
//                 c: jsonData.c,
//             }
//         },
// >     )
// ```
//
// tags: type-conversions
//
builtin parse : (data: bytes) => A
