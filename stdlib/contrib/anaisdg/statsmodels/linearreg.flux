// Package statsmodels provides functions for calculating statistical models.
//
// ## Metadata
// introduced: 0.90.0
// contributors: **GitHub**: [@anaisdg](https://github.com/anaisdg) | **InfluxDB Slack**: [@Anais](https://influxdata.com/slack)
//
package statsmodels


import "math"
import "generate"

// linearRegression performs a linear regression.
//
// It calculates and returns [*ŷ*](https://en.wikipedia.org/wiki/Hat_operator#Estimated_value) (`y_hat`),
// and [residual sum of errors](https://en.wikipedia.org/wiki/Residual_sum_of_squares) (`rse`).
// Output data includes the following columns:
//
// - **N**: Number of points in the calculation.
// - **slope**: Slope of the calculated regression.
// - **sx**: Sum of x.
// - **sxx**: Sum of x squared.
// - **sxy**: Sum of x*y.
// - **sy**: Sum of y.
// - **errors**: Residual sum of squares.
//   Defined by `(r.y - r.y_hat) ^ 2` in this context
// - **x**: An index [1,2,3,4...n], with the assumption that the timestamps are regularly spaced.
// - **y**: Field value
// - **y\_hat**: Linear regression values
//
// ## Parameters
// - tables: Input data. Default is piped-forward data (`<-`).
//
//
// ## Examples
//
// ### Perform a linear regression on a dataset
// ```no_run
// import "contrib/anaisdg/statsmodels"
// import "sampledata"
//
// < sampledata.float()
// >     |> statsmodels.linearRegression()
// ```
linearRegression = (tables=<-) => {
    renameAndSum =
        tables
            |> rename(columns: {_value: "y"})
            |> map(fn: (r) => ({r with x: 1.0}))
            |> cumulativeSum(columns: ["x"])
    t =
        renameAndSum
            |> reduce(
                fn: (r, accumulator) =>
                    ({
                        sx: r.x + accumulator.sx,
                        sy: r.y + accumulator.sy,
                        N: accumulator.N + 1.0,
                        sxy: r.x * r.y + accumulator.sxy,
                        sxx: r.x * r.x + accumulator.sxx,
                    }),
                identity: {
                    sxy: 0.0,
                    sx: 0.0,
                    sy: 0.0,
                    sxx: 0.0,
                    N: 0.0,
                },
            )
            |> tableFind(fn: (key) => true)
            |> getRecord(idx: 0)
    xbar = t.sx / t.N
    ybar = t.sy / t.N
    slope = (t.sxy - xbar * ybar * t.N) / (t.sxx - t.N * xbar * xbar)
    intercept = ybar - slope * xbar
    y_hat = (r) =>
        ({r with
            y_hat: slope * r.x + intercept,
            slope: slope,
            sx: t.sx,
            sxy: t.sxy,
            sxx: t.sxx,
            N: t.N,
            sy: t.sy,
        })
    rse = (r) => ({r with errors: (r.y - r.y_hat) ^ 2.0})
    output =
        renameAndSum
            |> map(fn: y_hat)
            |> map(fn: rse)

    return output
}
