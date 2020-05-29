package linearRegression_test

import (
	"context"
	"testing"

	_ "github.com/influxdata/flux/builtin"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/runtime"
)

indata = 
"#group,false,false,true,true,false,false,true,true,true,true
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,string,string,string
#default,_result,,,,,,,,,
,result,table,_start,_stop,_time,_value,_field,_measurement,shelter,type
,,0,2020-04-28T23:03:24.565187Z,2020-05-28T23:03:24.565187Z,2020-05-21T21:43:45Z,8,young,cats,B,tabby
,,0,2020-04-28T23:03:24.565187Z,2020-05-28T23:03:24.565187Z,2020-05-21T21:45:08Z,5,young,cats,B,tabby
,,0,2020-04-28T23:03:24.565187Z,2020-05-28T23:03:24.565187Z,2020-05-21T21:46:25Z,4,young,cats,B,tabby
,,0,2020-04-28T23:03:24.565187Z,2020-05-28T23:03:24.565187Z,2020-05-21T21:48:38Z,2,young,cats,B,tabby
"
outdata = 
"#group,false,false,false,true,true,true,true,false,false,true,false,false,false,false,false,true,false,false,false
#datatype,string,long,double,string,string,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,double,string,double,double,double,double,double,string,double,double,double
#default,_result,,,,,,,,,,,,,,,,,,
,result,table,N,_field,_measurement,_start,_stop,_time,errors,shelter,slope,sx,sxx,sxy,sy,type,x,y,y_hat
,,0,4,young,cats,2020-04-28T22:36:37.605243Z,2020-05-28T22:36:37.605243Z,2020-05-21T21:43:45Z,0.16000000000000028,B,-1.9,10,30,38,19,tabby,1,8,7.6
,,0,4,young,cats,2020-04-28T22:36:37.605243Z,2020-05-28T22:36:37.605243Z,2020-05-21T21:45:08Z,0.49000000000000027,B,-1.9,10,30,38,19,tabby,2,5,5.7
,,0,4,young,cats,2020-04-28T22:36:37.605243Z,2020-05-28T22:36:37.605243Z,2020-05-21T21:46:25Z,0.039999999999999716,B,-1.9,10,30,38,19,tabby,3,4,3.8000000000000007
,,0,4,young,cats,2020-04-28T22:36:37.605243Z,2020-05-28T22:36:37.605243Z,2020-05-21T21:48:38Z,0.009999999999999929,B,-1.9,10,30,38,19,tabby,4,2,1.9000000000000004
"

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

test linearRegression = () =>
({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: linearRegression})

© 2020 GitHub, Inc.
