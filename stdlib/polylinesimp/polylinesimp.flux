package polylinesimp

// Package polylinesimp provides methods for polyline simplication which is an efficient way of downsampling curves and lines without losing much of its variation
// throughout the path. This enables efficient rendering of graphs and visualizations without having to load the entire corpus of data points into memory.
// This is done by reducing the number of vertices used in a set of polylines while keeping the overall shape as much as possible.


// RDP is an algorithm that decimates a curve composed of line segments to a similar curve with fewer points.
//
// ## Parameters
// - column: The column that corresponds to the Y axis of the given curve. (optional) (default column: _value)
// - timeColumn: The time column that corresponds to the X axis of the given curve. (optional) (default column: _time)
// - epsilon: The user defined maximum tolerance value that determines the amount of compression. (optional) (epsilon should be greater than 0.0)
// - retention: The user defined retention rate, which indicates the percentage of points to be retained after downsampling. (optional) (Retention rate should be between 0.0 and 100.0)
//
// ## Examples
//
// ### Downsample the data from abcd.csv using the epsilon value 1.5
// ```
// # import "polylinesimp"
// # import "csv"
// #
// # data =
// #     csv.from(file : "abcd.csv")
// #          |> polylinesimp.rdp(column: "_value",timeColumn: "_time", epsilon:1.5) 
// #
// ```
//
// ### Downsample the data from abcd.csv using the retention rate of 90%
// ```
// # import "polylinesimp"
// # import "csv"
// #
// # data =
// #     csv.from(file : "abcd.csv")
// #          |> polylinesimp.rdp(column: "_value",timeColumn: "_time", retention:90.0) 
// #
// ```
// ## Metadata
// introduced: 0.7.0
// tags: transformations
//

builtin rdp : (
        <-tables: stream[A],
        ?column: string,
        ?timeColumn: string,
        ?epsilon: float,
        ?retention: float,
    ) => stream[B]
    where
    A: Record,
    B: Record
