# End-to-End Testing

End-to-end test must be included as a part of a PR for any contribution to Flux including: 
[Flux Functions](/Flux_Functions.md), 
[Standalone Scalar Functions](/Scalar_Functions.md), 
[Sink/Source Functions](/Source_Sink_Functions.md), 
[Stream Transformation Functions](/Stream_Transformation_Functions.md). 

## Required guidelines

Please help us make the contribution process easier by providing feedback about your experience and any technical hurdles you encountered here.

### Pure Flux Code Functions Guidelines

If you add a new [test](../stdlib/testing/testdata), our test framework automatically detects the file and tests it. To ensure your file passes the test, consider the correct form of simple_max.flux:

package testdata_test
 
import "testing"

-Set this option to a distant future so that queries based on ```now()``` will be consistent
```option now = () => (2030-01-01T00:00:00Z)```

-In the correct format, copy the data from another test, or run a query via HTTP using `curl` 
```
inData = "
#datatype,string,long,dateTime:RFC3339,string,string,double
#group,false,false,false,true,true,false
#default,_result,,,,,
,result,table,_time,_measurement,_field,_value
,,0,2018-04-17T00:00:00Z,m1,f1,42
,,0,2018-04-17T00:00:01Z,m1,f1,43
"
```

-To generate output data do one of the following
* Run the utility [here](../cmd/refactortests) to help generate the outData.  
* Build data manually.  
```
outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,string,string,dateTime:RFC3339,double
#group,false,false,true,true,true,true,false,false
#default,_result,,,,,,,
,result,table,_start,_stop,_measurement,_field,_time,_value
,,0,2018-04-17T00:00:00Z,2030-01-01T00:00:00Z,m1,f1,2018-04-17T00:00:01Z,43
"
```

-The following code is the actual test.  The query inputs tables with inData and outputs equivalent tables with outDtata.
```
// This query is source-independent. The test framework will gather the inData and outData into tables and then execute this function on them.  
simple_max = (table=<-) =>
	table
		|> range(start: 2018-04-17T00:00:00Z)
		|> max(column: "_value")`
```

-Register the test with our test platform.  
```test _simple_max = () =>
	({input: testing.loadStorage(csv: inData), want: testing.loadMem(csv: outData), fn: simple_max})
```

-After registering your test with our test platform, we load your data, run the test, and then check the output.
 A failed test will print a ```diff```

