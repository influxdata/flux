#End-to-End Testing

End-to-end test must be included as a part of a PR for any contribution to Flux including: [Flux Functions](https://github.com/Anaisdg/flux/blob/contributing/docs/Flux_Functions.md), [Standalone Scalar Functions](https://github.com/Anaisdg/flux/blob/contributing/docs/Scalar_Functions.md), [Sink/Source Functions](https://github.com/Anaisdg/flux/blob/contributing/docs/Source_Sink_Functions.md), or [Stream Transformation Functions](https://github.com/Anaisdg/flux/blob/contributing/docs/Stream_Transformation_Functions.md). 

## Required guidelines

Please help us make the contribution process easier by providing feedback about your experience and any technical hurdles you encountered here.

### **Pure Flux Code Functions Guidelines**

If you add a new ***.flux file to stdlib/testing/testdata, our test framework will automatically find it and run the test, if you've written it correctly. Consider the anatomy of simple_max.flux:

package testdata_test
 
import "testing"

-Set this option to a distant future so that queries based on `now()` will be consistent
`option now = () => (2030-01-01T00:00:00Z)`

-To get sample data in the right format copy from another test, or else run a query via HTTP using `curl` 
`inData = "
#datatype,string,long,dateTime:RFC3339,string,string,double
#group,false,false,false,true,true,false
#default,_result,,,,,
,result,table,_time,_measurement,_field,_value
,,0,2018-04-17T00:00:00Z,m1,f1,42
,,0,2018-04-17T00:00:01Z,m1,f1,43
"`

-Run the utility in https://github.com/influxdata/flux/tree/master/internal/cmd/refactortests to help generate the outData.  Or just build it up by hand.  
`outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,dateTime:RFC3339,double
#group,false,false,true,true,true,true,false,false
#default,_result,,,,,,,
,result,table,_start,_stop,_measurement,_field,_time,_value
,,0,2018-04-17T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,2018-04-17T00:00:01Z,43
"`

-This is the actual test.  It should take tables representing inData and produce tables equivalent to outData
// the query is written to be source-independent.  the test framework will gather the data above into tables and then call this function on them.  
`simple_max = (table=<-) =>
	table
		|> range(start: 2018-04-17T00:00:00Z)
		|> max(column: "_value")`

-Register the test with our test platform.  We will load the data, run the test, and check the output. A failed test will print a `diff`
`test _simple_max = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: simple_max})`

