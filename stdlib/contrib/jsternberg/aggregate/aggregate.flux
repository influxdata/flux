package aggregate

import "contrib/jsternberg/math"

// table will aggregate columns and create tables with a single
// row containing the aggregated value.
//
// This function takes a single parameter of `columns`. The parameter
// is an object with the output column name as the key and the aggregate
// object as the value.
//
// The aggregate object is composed of at least the following required attributes:
//     column = string
//         The column name for the input.
//     init = (values) -> state
//         An initial function to compute the initial state of the
//         output. This can return either the final aggregate or a
//         temporary state object that can be used to compute the
//         final aggregate. The values parameter will always be a
//         non-empty array of values from the specified column.
//     reduce = (values, state) -> state
//         A function that takes in another buffer of values
//         and the current state of the aggregate and computes
//         the updated state.
//     compute = (state) -> value
//         A function that takes the state and computes the final
//         aggregate.
//     fill = value
//         The value passed to fill, if present, will determine what
//         the aggregate does when there are no values.
//         This can either be a value or one of the predefined
//         identifiers of null or none.
//         This value must be the same type as the value return from
//         compute.
//
// An example of usage is:
//     tables |> aggregate.table(columns: {
//         "min_bottom_degrees": aggregate.min(column: "bottom_degrees"),
//     ])
builtin table : (<-tables: [A], columns: C) => [B] where A: Record, B: Record, C: Record

// null is a sentinel value for fill that will fill
// in a null value if there were no values for an interval.
builtin null : A

// none is a sentinel value for fill that will skip
// emitting a row if there are no values for an interval.
builtin none : A

// define will define an aggregate function.
define = (init, reduce, compute, fill=null) => (column, fill=fill) => ({
	column: column,
	init: init,
	reduce: reduce,
	compute: compute,
	fill: fill,
})

_make_selector = (fn) => define(
	init: (values) => fn(values),
	reduce: (values, state) => {
		v = fn(values)
		return fn(values: [state, v])
	},
	compute: (state) => state,
)

// min constructs a min aggregate or selector for the column.
min = _make_selector(fn: math.min)

// max constructs a max aggregate or selector for the column.
max = _make_selector(fn: math.max)

// sum constructs a sum aggregate for the column.
sum = define(
	init: (values) => math.sum(values),
	reduce: (values, state) => {
		return state + math.sum(values)
	},
	compute: (state) => state,
)

// count constructs a count aggregate for the column.
count = define(
	init: (values) => length(arr: values),
	reduce: (values, state) => {
		return state + length(arr: values)
	},
	compute: (state) => state,
	fill: 0,
)

// mean constructs a mean aggregate for the column.
mean = define(
	init: (values) => ({
		sum: math.sum(values),
		count: length(arr: values),
	}),
	reduce: (values, state) => ({
		sum: state.sum + math.sum(values),
		count: state.count + length(arr: values),
	}),
	compute: (state) => float(v: state.sum) / float(v: state.count),
)