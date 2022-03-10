// Package math provides implementations of aggregate functions.
package math


// minIndex returns the index of the minimum value within the array.
//
// ## Parameters
// - values: Array of values.
builtin minIndex : (values: [A]) => int where A: Numeric

// min returns the minimum value within the array.
//
// ## Parameters
// - values: Array of values.
min = (values) => {
    index = minIndex(values)

    return values[index]
}

// maxIndex returns the index of the maximum value within the array.
//
// ## Parameters
// - values: Array of values.
builtin maxIndex : (values: [A]) => int where A: Numeric

// max returns the maximum value within the array.
//
// ## Parameters
// - values: Array of values.
max = (values) => {
    index = maxIndex(values)

    return values[index]
}

// sum returns the sum of all values within the array.
//
// ## Parameters
// - values: Array of values.
builtin sum : (values: [A]) => A where A: Numeric
