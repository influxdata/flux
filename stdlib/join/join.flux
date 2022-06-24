// Package join provides functions that join two table streams together.
//
// ## Outer joins
//
// The join transformation supports left, right, and full outer joins.
//
// - **Left outer joins** generate at least one output row for each record in the left input stream.
//   If a record in the left input stream does not have a match in the right input stream,
//   `r` is substituted with a default record in the `as` function.
// - **Right outer joins** generate at least one output row for each record in the right input stream.
//   If a record in the right input stream does not have a match in the left input stream,
//   `l` is substituted with a default record in the `as` function.
// - **Full outer joins** generate at least one output row for each record in both input streams.
//   If a record in either input stream doesn't have a match in the other input stream,
//   one of the arguments to the `as` function is substituted with a default record
//   (either `l` or `r`, depending on which one is missing the matching record)
//
// A default record has the same columns as the records in the corresponding input
// table, but only group key columns are populated with a value. All other columns
// are null.
//
// ## Inner joins
//
// Inner joins drop any records that don't have a match in the other input stream. There is no
// need to account for default or unmatched records when performing an inner join.
//
// ## Metadata
// introduced: 0.172.0
// tags: transformations
package join


// tables joins two input streams together using a specified method, predicate, and a function to join two corresponding records, one from each input stream.
//
// `join.tables()` only compares records with the same group key. Output tables have the same grouping as the input tables.
//
// ## Parameters
// - left: Left input stream. Default is piped-forward data (`<-`).
// - right: Right input stream.
// - on: Function that takes a left and right record (`l`, and `r` respectively), and returns a boolean.
//
//   The body of the function must be a single boolean expression, consisting of one
//   or more equality comparisons between a property of `l` and a property of `r`,
//   each chained together by the `and` operator.
//
// - as: Function that takes a left and a right record (`l` and `r` respectively), and returns a record.
//   The returned record is included in the final output.
// - method: String that specifies the join method.
//
//   **Supported methods:**
//
//   - inner
//   - left
//   - right
//   - full
//
// ## Examples
//
// ### Perform an inner join
// ```
// import "sampledata"
// import "join"
//
// ints = sampledata.int()
// strings = sampledata.string()
//
// join.tables(
//     method: "inner",
//     left: ints,
//     right: strings,
//     on: (l, r) => l._time == r._time,
//     as: (l, r) => ({l with label: r._value}),
// > )
// ```
//
// ### Perform a left outer join
//
// If the join method is anything other than `inner`, pay special attention to how
// the output record is constructed in the `as` function.
//
// Because of how flux handles outer joins, it's possible for either `l` or `r` to be a
// default record. This means any value in a non-group-key column could be null.
//
// For more information about the behavior of outer joins, see the [Outer joins](https://docs.influxdata.com/flux/v0.x/stdlib/join/#outer-joins)
// section in the `join` package documentation.
//
// In the case of a left outer join, `l` is guaranteed to not be a default record. To
// ensure that the output record has non-null values for any columns that aren't part
// of the group key, use values from `l`. Using a non-group-key value from `r` risks
// that value being null.
//
// The example below constructs the output record almost entirely from properties of `l`.
// The only exception is the `v_right` column which gets its value from `r._value`.
// In this case, understand and expect that `v_right` will sometimes be null.
//
// ```
// import "array"
// import "join"
//
// left =
//     array.from(
//         rows: [
//             {_time: 2022-01-01T00:00:00Z, _value: 1, label: "a"},
//             {_time: 2022-01-01T00:00:00Z, _value: 2, label: "b"},
//             {_time: 2022-01-01T00:00:00Z, _value: 3, label: "d"},
//         ],
//     )
// right =
//     array.from(
//         rows: [
//             {_time: 2022-01-01T00:00:00Z, _value: 0.4, id: "a"},
//             {_time: 2022-01-01T00:00:00Z, _value: 0.5, id: "c"},
//             {_time: 2022-01-01T00:00:00Z, _value: 0.6, id: "d"},
//         ],
//     )
//
// join.tables(
//     method: "left",
//     left: left,
//     right: right,
//     on: (l, r) => l.label == r.id and l._time == r._time,
//     as: (l, r) => ({_time: l._time, label: l.label, v_left: l._value, v_right: r._value}),
// > )
// ```
//
// ### Perform a right outer join
//
// The next example is nearly identical to the [previous example](#perform-a-left-outer-join),
// but uses the `right` join method. With this method, `r` is guaranteed to not be a default
// record, but `l` may be a default record. Because `l` is more likely to contain null values,
// the output record is built almost entirely from proprties of `r`, with the exception of
// `v_left`, which we expect to sometimes be null.
//
// ```
// import "array"
// import "join"
//
// left =
//     array.from(
//         rows: [
//             {_time: 2022-01-01T00:00:00Z, _value: 1, label: "a"},
//             {_time: 2022-01-01T00:00:00Z, _value: 2, label: "b"},
//             {_time: 2022-01-01T00:00:00Z, _value: 3, label: "d"},
//         ],
//     )
// right =
//     array.from(
//         rows: [
//             {_time: 2022-01-01T00:00:00Z, _value: 0.4, id: "a"},
//             {_time: 2022-01-01T00:00:00Z, _value: 0.5, id: "c"},
//             {_time: 2022-01-01T00:00:00Z, _value: 0.6, id: "d"},
//         ],
//     )
//
// join.tables(
//     method: "right",
//     left: left,
//     right: right,
//     on: (l, r) => l.label == r.id and l._time == r._time,
//     as: (l, r) => ({_time: r._time, label: r.id, v_left: l._value, v_right: r._value}),
// > )
// ```
//
// ### Perform a full outer join
//
// In a full outer join, there are no guarantees about `l` or `r`. Either one of them could
// be a default record, but they will never both be a default record at the same time.
//
// To get non-null values for the output record, check both `l` and `r` to see which contains
// the desired values.
//
// The example below defines a function for the `as` parameter that appropriately handles
// the uncertainty of a full outer join.
//
// `v_left` and `v_right` still use values from `l` and `r` directly, because we expect
// them to sometimes be null in the output table.
//
// ```
// import "array"
// import "join"
//
// left =
//     array.from(
//         rows: [
//             {_time: 2022-01-01T00:00:00Z, _value: 1, label: "a"},
//             {_time: 2022-01-01T00:00:00Z, _value: 2, label: "b"},
//             {_time: 2022-01-01T00:00:00Z, _value: 3, label: "d"},
//         ],
//     )
// right =
//     array.from(
//         rows: [
//             {_time: 2022-01-01T00:00:00Z, _value: 0.4, id: "a"},
//             {_time: 2022-01-01T00:00:00Z, _value: 0.5, id: "c"},
//             {_time: 2022-01-01T00:00:00Z, _value: 0.6, id: "d"},
//         ],
//     )
//
// join.tables(
//     method: "full",
//     left: left,
//     right: right,
//     on: (l, r) => l.label == r.id and l._time == r._time,
//     as: (l, r) => {
//         time = if exists l._time then l._time else r._time
//         label = if exists l.label then l.label else r.id
//
//         return {_time: time, label: label, v_left: l._value, v_right: r._value}
//     },
// > )
// ```
//
// ## Metadata
// introduced: 0.172.0
// tags: transformations
builtin tables : (
        <-left: stream[L],
        right: stream[R],
        on: (l: L, r: R) => bool,
        as: (l: L, r: R) => A,
        method: string,
    ) => stream[A]
    where
    A: Record,
    L: Record,
    R: Record

// time joins two table streams together exclusively on the `_time` column.
//
// This function calls `join.tables()` with the `on` parameter set to `(l, r) => l._time == r._time`.
//
// ## Parameters
// - left: Left input stream. Default is piped-forward data (<-).
// - right: Right input stream.
// - as: Function that takes a left and a right record (`l` and `r` respectively), and returns a record.
//   The returned record is included in the final output.
// - method: String that specifies the join method. Default is `inner`.
//
//   **Supported methods:**
//
//   - inner
//   - left
//   - right
//   - full
//
// ## Examples
//
// ### Join two tables by timestamp
// ```
// import "sampledata"
// import "join"
//
// ints = sampledata.int()
// strings = sampledata.string()
//
// join.time(
//     left: ints,
//     right: strings,
//     as: (l, r) => ({l with label: r._value}),
// > )
// ```
// ## Metadata
// introduced: 0.172.0
// tags: transformations
time = (left=<-, right, as, method="inner") =>
    tables(
        left: left,
        right: right,
        on: (l, r) => l._time == r._time,
        as: as,
        method: method,
    )

// inner performs an inner join on two table streams.
//
// The function calls `join.tables()` with the `method` parameter set to `"inner"`.
//
// ## Parameters
// - left: Left input stream. Default is piped-forward data (<-).
// - right: Right input stream.
// - on: Function that takes a left and right record (`l`, and `r` respectively), and returns a boolean.
//
//   The body of the function must be a single boolean expression, consisting of one
//   or more equality comparisons between a property of `l` and a property of `r`,
//   each chained together by the `and` operator.
//
// - as: Function that takes a left and a right record (`l` and `r` respectively), and returns a record.
//   The returned record is included in the final output.
//
// ## Examples
//
// ### Perform an inner join
// ```
// import "sampledata"
// import "join"
//
// ints = sampledata.int()
// strings = sampledata.string()
//
// join.inner(
//     left: ints,
//     right: strings,
//     on: (l, r) => l._time == r._time,
//     as: (l, r) => ({l with label: r._value}),
// > )
// ```
//
// ## Metadata
// introduced: 0.172.0
// tags: transformations
inner = (left=<-, right, on, as) =>
    tables(
        left: left,
        right: right,
        on: on,
        as: as,
        method: "inner",
    )

// full performs a full outer join on two table streams.
//
// The function calls `join.tables()` with the `method` parameter set to `"full"`.
//
// ## Parameters
// - left: Left input stream. Default is piped-forward data (<-).
// - right: Right input stream.
// - on: Function that takes a left and right record (`l`, and `r` respectively), and returns a boolean.
//
//   The body of the function must be a single boolean expression, consisting of one
//   or more equality comparisons between a property of `l` and a property of `r`,
//   each chained together by the `and` operator.
//
// - as: Function that takes a left and a right record (`l` and `r` respectively), and returns a record.
//   The returned record is included in the final output.
//
// ## Examples
//
// ### Perform a full outer join
//
// In a full outer join, either `l` or `r` could be a default record, but they will
// never both be a default record at the same time.
//
// To get non-null values for the output record, check both `l` and `r` to see which contains
// the desired values.
//
// The example below defines a function for the `as` parameter that appropriately handles
// the uncertainty of a full outer join.
//
// `v_left` and `v_right` still use values from `l` and `r` directly, because we expect
// them to sometimes be null in the output table.
//
// For more information about the behavior of outer joins, see the [Outer joins](https://docs.influxdata.com/flux/v0.x/stdlib/join/#outer-joins)
// section in the `join` package documentation.
//
// ```
// import "array"
// import "join"
//
// left =
//     array.from(
//         rows: [
//             {_time: 2022-01-01T00:00:00Z, _value: 1, label: "a"},
//             {_time: 2022-01-01T00:00:00Z, _value: 2, label: "b"},
//             {_time: 2022-01-01T00:00:00Z, _value: 3, label: "d"},
//         ],
//     )
// right =
//     array.from(
//         rows: [
//             {_time: 2022-01-01T00:00:00Z, _value: 0.4, id: "a"},
//             {_time: 2022-01-01T00:00:00Z, _value: 0.5, id: "c"},
//             {_time: 2022-01-01T00:00:00Z, _value: 0.6, id: "d"},
//         ],
//     )
//
// join.full(
//     left: left,
//     right: right,
//     on: (l, r) => l.label == r.id and l._time == r._time,
//     as: (l, r) => {
//         time = if exists l._time then l._time else r._time
//         label = if exists l.label then l.label else r.id
//
//         return {_time: time, label: label, v_left: l._value, v_right: r._value}
//     },
// > )
// ```
//
// ## Metadata
// introduced: 0.172.0
// tags: transformations
full = (left=<-, right, on, as) =>
    tables(
        left: left,
        right: right,
        on: on,
        as: as,
        method: "full",
    )

// left performs a left outer join on two table streams.
//
// The function calls `join.tables()` with the `method` parameter set to `"left"`.
//
// ## Parameters
// - left: Left input stream. Default is piped-forward data (<-).
// - right: Right input stream.
// - on: Function that takes a left and right record (`l`, and `r` respectively), and returns a boolean.
//
//   The body of the function must be a single boolean expression, consisting of one
//   or more equality comparisons between a property of `l` and a property of `r`,
//   each chained together by the `and` operator.
//
// - as: Function that takes a left and a right record (`l` and `r` respectively), and returns a record.
//   The returned record is included in the final output.
//
// ## Examples
//
// ### Perform a left outer join
//
// In a left outer join, `l` is guaranteed to not be a default record, but `r` may be a
// default record. Because `r` is more likely to contain null values, the output record
// is built almost entirely from proprties of `l`, with the exception of `v_right`, which
// we expect to sometimes be null.
//
// For more information about the behavior of outer joins, see the [Outer joins](https://docs.influxdata.com/flux/v0.x/stdlib/join/#outer-joins)
// section in the `join` package documentation.
//
// ```
// import "array"
// import "join"
//
// left =
//     array.from(
//         rows: [
//             {_time: 2022-01-01T00:00:00Z, _value: 1, label: "a"},
//             {_time: 2022-01-01T00:00:00Z, _value: 2, label: "b"},
//             {_time: 2022-01-01T00:00:00Z, _value: 3, label: "d"},
//         ],
//     )
// right =
//     array.from(
//         rows: [
//             {_time: 2022-01-01T00:00:00Z, _value: 0.4, id: "a"},
//             {_time: 2022-01-01T00:00:00Z, _value: 0.5, id: "c"},
//             {_time: 2022-01-01T00:00:00Z, _value: 0.6, id: "d"},
//         ],
//     )
//
// join.left(
//     left: left,
//     right: right,
//     on: (l, r) => l.label == r.id and l._time == r._time,
//     as: (l, r) => ({_time: l._time, label: l.label, v_left: l._value, v_right: r._value}),
// > )
// ```
// ## Metadata
// introduced: 0.172.0
// tags: transformations
left = (left=<-, right, on, as) =>
    tables(
        left: left,
        right: right,
        on: on,
        as: as,
        method: "left",
    )

// right performs a right outer join on two table streams.
//
// The function calls `join.tables()` with the `method` parameter set to `"right"`.
//
// ## Parameters
// - left: Left input stream. Default is piped-forward data (<-).
// - right: Right input stream.
// - on: Function that takes a left and right record (`l`, and `r` respectively), and returns a boolean.
//
//   The body of the function must be a single boolean expression, consisting of one
//   or more equality comparisons between a property of `l` and a property of `r`,
//   each chained together by the `and` operator.
//
// - as: Function that takes a left and a right record (`l` and `r` respectively), and returns a record.
//   The returned record is included in the final output.
//
// ## Examples
//
// ### Perform a right outer join
//
// In a right outer join, `r` is guaranteed to not be a default record, but `l` may be a
// default record. Because `l` is more likely to contain null values, the output record
// is built almost entirely from proprties of `r`, with the exception of `v_left`, which
// we expect to sometimes be null.
//
// For more information about the behavior of outer joins, see the [Outer joins](https://docs.influxdata.com/flux/v0.x/stdlib/join/#outer-joins)
// section in the `join` package documentation.
//
// ```
// import "array"
// import "join"
//
// left =
//     array.from(
//         rows: [
//             {_time: 2022-01-01T00:00:00Z, _value: 1, label: "a"},
//             {_time: 2022-01-01T00:00:00Z, _value: 2, label: "b"},
//             {_time: 2022-01-01T00:00:00Z, _value: 3, label: "d"},
//         ],
//     )
// right =
//     array.from(
//         rows: [
//             {_time: 2022-01-01T00:00:00Z, _value: 0.4, id: "a"},
//             {_time: 2022-01-01T00:00:00Z, _value: 0.5, id: "c"},
//             {_time: 2022-01-01T00:00:00Z, _value: 0.6, id: "d"},
//         ],
//     )
//
// join.right(
//     left: left,
//     right: right,
//     on: (l, r) => l.label == r.id and l._time == r._time,
//     as: (l, r) => ({_time: r._time, label: r.id, v_left: l._value, v_right: r._value}),
// > )
// ```
// ## Metadata
// introduced: 0.172.0
// tags: transformations
right = (left=<-, right, on, as) =>
    tables(
        left: left,
        right: right,
        on: on,
        as: as,
        method: "right",
    )
