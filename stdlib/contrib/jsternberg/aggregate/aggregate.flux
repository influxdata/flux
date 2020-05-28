package aggregate

import "contrib/jsternberg/math"

// table will aggregate columns and create tables with a single
// row containing the aggregated value.
//
// This function takes a single parameter of `columns`. The parameter
// is an array of objects. Each object must have three attributes:
//     column = string
//         The column name for the input.
//     with = {init, reduce, compute}
//         The aggregate as defined below.
//     as = string
//         The name of the output column.
//
// The `with` attribute is also an object. It contains at least the
// following attributes:
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
//
// An example of usage is:
//     tables |> aggregate.table(columns: [
//         {column: "bottom_degrees", with: aggregate.min, as: "min_bottom_degrees"},
//     ])
builtin table

_make_selector = (fn) => ({
	init: fn,
    reduce: (values, state) => {
    	v = fn(values)
    	return fn(values: [state, v])
    },
    compute: (state) => state,
})

min = _make_selector(fn: math.min)
max = _make_selector(fn: math.max)

sum = {
	init: (values) => math.sum(values),
	reduce: (values, state) => {
		return state + math.sum(values)
	},
	compute: (state) => state,
}

count = {
	init: (values) => length(arr: values),
	reduce: (values, state) => {
		return state + length(arr: values)
	},
	compute: (state) => state,
}
