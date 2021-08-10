package hex_test

import (
	"context"
	"testing"
	"time"

	"github.com/influxdata/flux"
	_ "github.com/influxdata/flux/csv"
	_ "github.com/influxdata/flux/fluxinit/static"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/runtime"
)

func Test_ToString(t *testing.T) {
	testCases := []struct {
		inputType   string
		inputValue  string
		outputValue string
	}{
		{
			inputType:   "long",
			inputValue:  "-521",
			outputValue: "-209",
		},
		{
			inputType:   "long",
			inputValue:  "521",
			outputValue: "209",
		},
		{
			inputType:   "unsignedLong",
			inputValue:  "4294967294",
			outputValue: "fffffffe",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.inputType+"/"+tc.inputValue, func(t *testing.T) {
			fluxString := `import "csv"
import "contrib/bonitoo-io/hex"

data = "
#datatype,string,string,` + tc.inputType + `
#group,false,false,false
#default,_result,,
,result,,_value
,,,` + tc.inputValue + `"

csv.from(csv: data) |> hex.toString() `
			prog, err := lang.Compile(fluxString, runtime.Default, time.Now())
			if err != nil {
				t.Fatal(err)
			}
			ctx := flux.NewDefaultDependencies().Inject(context.Background())
			query, err := prog.Start(ctx, &memory.Allocator{})

			if err != nil {
				t.Fatal(err)
			}
			res := <-query.Results()
			_ = res
			values := 0
			err = res.Tables().Do(func(table flux.Table) error {
				return table.Do(func(reader flux.ColReader) error {
					if reader == nil {
						return nil
					}
					for i, meta := range reader.Cols() {
						if meta.Label == "_value" {
							values++
							if v := reader.Strings(i).Value(0); v != tc.outputValue {
								t.Fatalf("expecting _value=%v but got _value=%v", tc.outputValue, string(v))
							}
						}
					}
					return nil
				})
			})
			if values != 1 {
				t.Fatalf("expected one _value but get %v", values)
			}
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func Test_ToInt(t *testing.T) {
	testCases := []struct {
		inputValue  string
		outputValue int64
	}{
		{
			inputValue:  "-209",
			outputValue: -521,
		},
		{
			inputValue:  "+209",
			outputValue: 521,
		},
		{
			inputValue:  "-0x209",
			outputValue: -521,
		},
		{
			inputValue:  "+0x209",
			outputValue: 521,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.inputValue, func(t *testing.T) {
			fluxString := `import "csv"
import "contrib/bonitoo-io/hex"

data = "
#datatype,string,string,string
#group,false,false,false
#default,_result,,
,result,,_value
,,,` + tc.inputValue + `"

csv.from(csv: data) |> hex.toInt() `
			prog, err := lang.Compile(fluxString, runtime.Default, time.Now())
			if err != nil {
				t.Fatal(err)
			}
			ctx := flux.NewDefaultDependencies().Inject(context.Background())
			query, err := prog.Start(ctx, &memory.Allocator{})

			if err != nil {
				t.Fatal(err)
			}
			res := <-query.Results()
			_ = res
			values := 0
			err = res.Tables().Do(func(table flux.Table) error {
				return table.Do(func(reader flux.ColReader) error {
					if reader == nil {
						return nil
					}
					for i, meta := range reader.Cols() {
						if meta.Label == "_value" {
							values++
							if v := reader.Ints(i).Value(0); v != tc.outputValue {
								t.Fatalf("expecting _value=%v but got _value=%v", tc.outputValue, v)
							}
						}
					}
					return nil
				})
			})
			if values != 1 {
				t.Fatalf("expected one _value but get %v", values)
			}
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}
