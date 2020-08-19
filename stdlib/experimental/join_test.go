package experimental_test

import (
	"testing"

	"github.com/influxdata/flux/querytest"
)

func TestExperimentalJoin_Errors(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name: "experimental group extend",
			Raw: `import "csv"
import "experimental"
import "testing"

outData = "
#datatype,string,long,dateTime:RFC3339,dateTime:RFC3339,dateTime:RFC3339,string,string,double
#group,false,false,true,true,false,true,true,false
#default,_result,,,,,,,
,result,table,_start,_stop,_time,_measurement,tag0,value_a
,,0,2018-12-19T00:00:00Z,2018-12-20T00:00:00Z,2018-12-19T22:13:30Z,_m,g,2
,,0,2018-12-19T00:00:00Z,2018-12-20T00:00:00Z,2018-12-19T22:13:40Z,_m,g,3
,,0,2018-12-19T00:00:00Z,2018-12-20T00:00:00Z,2018-12-19T22:13:50Z,_m,g,4
,,0,2018-12-19T00:00:00Z,2018-12-20T00:00:00Z,2018-12-19T22:14:00Z,_m,g,5
,,0,2018-12-19T00:00:00Z,2018-12-20T00:00:00Z,2018-12-19T22:14:10Z,_m,g,6
,,0,2018-12-19T00:00:00Z,2018-12-20T00:00:00Z,2018-12-19T22:14:20Z,_m,g,7
"
outTable = csv.from(csv: outData)

inData = "
#datatype,string,long,dateTime:RFC3339,string,string,string,double
#group,false,false,false,true,true,true,false
#default,_result,,,,,,
,result,table,_time,_measurement,_field,tag0,_value
,,0,2018-12-19T22:13:30Z,_m,c,t,1

#datatype,string,long,dateTime:RFC3339,string,string,string,double
#group,false,false,false,true,true,true,false
#default,_result,,,,,,
,result,table,_time,_measurement,_field,tag0,_value
,,4,2018-12-19T22:13:30Z,_m,d,s,1
"
table = csv.from(csv: inData)

a = table
	|> range(start: 2018-12-19T00:00:00Z, stop: 2018-12-20T00:00:00Z)
	|> filter(fn: (r) => r._field == "c")
	|> drop(columns: ["_field"])
	|> rename(columns: {_value: "value_c"})

b = table
	|> range(start: 2018-12-19T00:00:00Z, stop: 2018-12-20T00:00:00Z)
	|> filter(fn: (r) => r._field == "d")
	|> drop(columns: ["_field"])
	|> rename(columns: {_value: "value_d"})

c = experimental.join(left:a, right:b, fn:(left, right) => ({left with value_c: right.value_c}))
testing.diff(got: c, want: outData)
`,
			WantErr: true,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			querytest.NewQueryTestHelper(t, tc)
		})
	}
}
