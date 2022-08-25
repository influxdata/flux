// Package polyline provides methods for polyline simplication, an efficient way of downsampling curves while retaining moments of variation throughout the path.
//
// This class of algorithms enable efficient rendering of graphs and visualizations without having to load all data into memory.
// This is done by reducing the number of vertices that do not contribute significantly to the convexity and concavity of the shape.
//
// ## Metadata
// introduced: NEXT
//
package polyline


// rdp applies the Ramer Douglas Peucker (RDP) algorithm to input data to downsample curves composed
// of line segments into visually indistinguishable curves with fewer points.
//
// ## Parameters
// - valColumn: Column with Y axis values of the given curve. Default is `_value`.
// - timeColumn: Column with X axis values of the given curve. Default is `_time`.
// - epsilon: Maximum tolerance value that determines the amount of compression.
//
//   Epsilon should be greater than `0.0`.
//
// - retention: Percentage of points to retain after downsampling.
//
//   Retention rate should be between `0.0` and `100.0`.
//
// - tables: Input data. Default is piped-forward data (`<-`).
//
// ## Examples
//
// ### Downsample the data using the epsilon value 1.5
// ```
// # import "internal/gen"
// import "experimental/polyline"
//
// # data = gen.tables(n: 16, seed: 1234)
// #
// < data
// >     |> polyline.rdp(epsilon: 1.5)
// ```
//
// ### Downsample the data using a retention rate of 90%
// ```
// # import "internal/gen"
// import "experimental/polyline"
//
// # data = gen.tables(n: 16, seed: 1234)
// #
// < data
// >     |> polyline.rdp(retention: 90.0)
// ```
//
// ### Downsample the data by automatically calculating the maximum tolerance beyond which producing a visually indistingushable curve will not be possible. This can be used when both epsilon and the retention rate are tricky to be judged.
// ```
// # import "internal/gen"
// import "experimental/polyline"
//
// # data = gen.tables(n: 16, seed: 1234)
// #
// < data
// >     |> polyline.rdp()
// ```
//
// ## Metadata
// tags: transformations
//
builtin rdp : (
        <-tables: stream[A],
        ?valColumn: string,
        ?timeColumn: string,
        ?epsilon: float,
        ?retention: float,
    ) => stream[B]
    where
    A: Record,
    B: Record
