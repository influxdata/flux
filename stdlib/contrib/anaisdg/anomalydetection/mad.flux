// Package anomalydetection detects anomalies in time series data.
//
// ## Metadata
// introduced: 0.90.0
// contributors: **GitHub**: [@anaisdg](https://github.com/anaisdg) | **InfluxDB Slack**: [@Anais](https://influxdata.com/slack)
//
package anomalydetection


import "math"
import "experimental"

// mad uses the median absolute deviation (MAD) algorithm to detect anomalies in a data set.
//
// Input data requires `_time` and `_value` columns.
// Output data is grouped by `_time` and includes the following columns of interest:
//
// - **\_value**: difference between of the original `_value` from the computed MAD
//   divided by the median difference.
// - **MAD**: median absolute deviation of the group.
// - **level**: anomaly indicator set to either `anomaly` or `normal`.
//
// ## Parameters
// - threshold: Deviation threshold for anomalies.
//
// - table: Input data. Default is piped-forward data (`<-`).
//
//
// ## Examples
//
// ### Use the MAD algorithm to detect anomalies
// ```
// import "contrib/anaisdg/anomalydetection"
// import "sampledata"
//
// < sampledata.float()
// >     |> anomalydetection.mad(threshold: 1.0)
mad = (table=<-, threshold=3.0) => {
    // MEDiXi = med(x)
    data = table |> group(columns: ["_time"], mode: "by")
    med = data |> median(column: "_value")

    // diff = |Xi - MEDiXi| = math.abs(xi-med(xi))
    diff =
        join(tables: {data: data, med: med}, on: ["_time"], method: "inner")
            |> map(fn: (r) => ({r with _value: math.abs(x: r._value_data - r._value_med)}))
            |> drop(columns: ["_start", "_stop", "_value_med", "_value_data"])

    // The constant k is needed to make the estimator consistent for the parameter of interest.
    // In the case of the usual parameter at Gaussian distributions k = 1.4826
    k = 1.4826

    // MAD =  k * MEDi * |Xi - MEDiXi|
    diff_med =
        diff
            |> median(column: "_value")
            |> map(fn: (r) => ({r with MAD: k * r._value}))
            |> filter(fn: (r) => r.MAD > 0.0)
    output =
        join(tables: {diff: diff, diff_med: diff_med}, on: ["_time"], method: "inner")
            |> map(fn: (r) => ({r with _value: r._value_diff / r._value_diff_med}))
            |> map(
                fn: (r) =>
                    ({r with level:
                            if r._value >= threshold then
                                "anomaly"
                            else
                                "normal",
                    }),
            )

    return output
}
