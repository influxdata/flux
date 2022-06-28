// Package json provides tools for working with JSON.
//
// ## Metadata
// introduced: 0.69.0
// tags: json
//
package json


// parse takes JSON data as bytes and returns a value.
//
// JSON types are converted to Flux types as follows:
//
// | JSON type | Flux type |
// | --------- | --------- |
// | boolean   | boolean   |
// | number    | float     |
// | string    | string    |
// | array     | array     |
// | object    | record    |
//
//
// ## Parameters
// - data: JSON data (as bytes) to parse.
//
// ## Examples
//
// ### Parse and use JSON data to restructure tables
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
// ### Parse JSON and use array functions to manipulate into a table
//
// ```
// import "experimental/json"
// import "experimental/array"
//
// jsonStr = bytes(v:"{
//      \"node\": {
//          \"items\": [
//              {
//                  \"id\": \"15612462\",
//                  \"color\": \"red\",
//                  \"states\": [
//                      {
//                          \"name\": \"ready\",
//                          \"duration\": 10
//                      },
//                      {
//                          \"name\": \"closed\",
//                          \"duration\": 13
//                      },
//                      {
//                          \"name\": \"pending\",
//                          \"duration\": 3
//                      }
//                  ]
//              },
//              {
//                  \"id\": \"15612462\",
//                  \"color\": \"blue\",
//                  \"states\": [
//                      {
//                          \"name\": \"ready\",
//                          \"duration\": 5
//                      },
//                      {
//                          \"name\": \"closed\",
//                          \"duration\": 0
//                      },
//                      {
//                          \"name\": \"pending\",
//                          \"duration\": 16
//                      }
//                  ]
//              }
//          ]
//      }
// }")
//
//  data = json.parse(data: jsonStr)
//
//  // Map over all items in the JSON extracting
//  // the id, color and pending duration of each.
//  // Construct a table from the final records.
//  array.from(rows:
//      data.node.items
//          |> array.map(fn:(x) => {
//              pendingState = x.states
//                  |> array.filter(fn: (x) => x.name == "pending")
//              pendingDur = if length(arr: pendingState) == 1
//                  then
//                      pendingState[0].duration
//                  else
//                      0.0
//              return {
//                  id: x.id,
//                  color: x.color,
//                  pendingDuration: pendingDur,
//              }
//          })
// > )
// ```
//
// ## Metadata
// tags: type-conversions
//
builtin parse : (data: bytes) => A
