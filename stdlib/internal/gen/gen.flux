// Package gen provides methods for generating data.
//
// ## Metadata
// introduced: 0.50.0
package gen


// tables generates a stream of table data.
//
// ## Parameters
// - n: Number of rows to generate.
// - nulls: Percentage chance that a null value will be used in the input. Valid value range is `[0.0 - 1.0]`.
// - tags: Set of tags with their cardinality to generate.
// - seed: Pass seed to tables generator to get the very same sequence each time.
builtin tables : (
        n: int,
        ?nulls: float,
        ?tags: [{name: string, cardinality: int}],
        ?seed: int,
    ) => stream[{A with _time: time, _value: float}]
