package anomalydetection 

import "math"
import "experimental"

mad = (table=<-, threshold=3.0) => {
    // MEDiXi = med(x)
    data = table |> group(columns: ["_time"], mode:"by")
    med = data |> median(column: "_value")
    // diff = |Xi - MEDiXi| = math.abs(xi-med(xi))
    diff = join(tables: {data: data, med: med}, on: ["_time"], method: "inner")
    |> map(fn: (r) => ({ r with _value: math.abs(x: r._value_data - r._value_med) }))
    |> drop(columns: ["_start", "_stop", "_value_med", "_value_data"])
    // The constant k is needed to make the estimator consistent for the parameter of interest.
    // In the case of the usual parameter a at Gaussian distributions k = 1.4826
    k = 1.4826
    // MAD =  k * MEDi * |Xi - MEDiXi| 
    diff_med =
    diff
        |> median(column: "_value")
        |> map(fn: (r) => ({ r with MAD: k * r._value}))
        |> filter(fn: (r) => r.MAD > 0.0)
    output = join(tables: {diff: diff, diff_med: diff_med}, on: ["_time"], method: "inner")
        |> map(fn: (r) => ({ r with _value: r._value_diff/r._value_diff_med}))
    |> map(fn: (r) => ({ r with
            level:
                if r._value >= threshold then "anomaly"
                else "normal"
        }))
return output
}


