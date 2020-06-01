package anomalyDetection 

import "math"
import "experimental"

mad = (table=<-) => {
    // _value_med = med(x)
    data = table |> group(columns: ["_time"], mode:"by")
    med = data |> median(column: "_value")
    
    // _value_diff = xi-med(xi)
    diff = experimental.join(
    left: data,
    right: med,
    fn: (left, right) => ({ right with _value: math.abs(x: left._value - right._value) })
    )
   
   //The constant b is needed to make the estimator consistent for the parameter of interest.
    /// In the case of the usual parameter a at Gaussian distributions
    b = 1.4826
    
    diff_med =
    diff
        |> median(column: "_value")
        |> map(fn: (r) => ({ r with MAD: b * r._value}))
        |> filter(fn: (r) => r.MAD > 0.0)
    
    threshold = 3.0
    
    output = union(tables: [diff, diff_med])
    |> filter(fn: (r) => exists r.MAD)
    |> map(fn: (r) => ({ r with _value: r._value / r.MAD }))
    |> map(fn: (r) => ({ r with
            level:
                if r._value >= threshold then "anomaly"
                else "normal"
        }))

return output
}


