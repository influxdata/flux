package statsmodels

import "math"
import "generate"

// performs linear regression, calculates y_hat, and residuals squared (rse) 

linearRegression = (tables=<-) => {
  renameAndSum = tables
    |> rename(columns: {_value: "y"})
    |> map(fn: (r) => ({r with x: 1.0}))
    |> cumulativeSum(columns: ["x"])

  t = renameAndSum 
    |> reduce(
      fn: (r, accumulator) => ({
        sx: r.x + accumulator.sx,
        sy: r.y + accumulator.sy,
        N: accumulator.N + 1.0,  
        sxy: r.x * r.y + accumulator.sxy, 
        sxx: r.x * r.x + accumulator.sxx
      }), 
      identity: {sxy: 0.0, sx:0.0, sy:0.0, sxx:0.0, N:0.0})
    |> tableFind(fn: (key) => true)
    |> getRecord(idx: 0)

  xbar = t.sx/t.N 

  ybar = t.sy/t.N

  slope = (t.sxy - xbar*ybar*t.N)/(t.sxx - t.N*xbar*xbar)

  intercept = (ybar - slope * xbar)

  y_hat = (r) => ({r with y_hat: slope * r.x + intercept, slope:slope, sx: t.sx, sxy: t.sxy, sxx: t.sxx, N: t.N, sy: t.sy})

  rse = (r) => ({r with errors: (r.y - r.y_hat)^2.0})

  output = renameAndSum
    |> map(fn: y_hat)
    |> map(fn: rse)
    
  return output
}