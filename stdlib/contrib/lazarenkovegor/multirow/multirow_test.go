package multirow_test

import (
	"context"
	"fmt"
	"github.com/influxdata/flux"
	_ "github.com/influxdata/flux/csv"
	_ "github.com/influxdata/flux/fluxinit/static"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/runtime"
	"testing"
	"time"
)

func TestMap(t *testing.T) {

	const script = `
import "csv"
import "contrib/lazarenkovegor/multirow"

inData = "
#datatype,string,long,string,long,dateTime:RFC3339,string
#group,false,false,false,false,false,false
#default,_result,0,,,2000-01-01T00:00:00Z,m0
,result,table,_field,_value,_time,_measurement
,,,test1,1,2020-08-02T17:22:00Z,
,,,test1,2,2020-08-02T17:22:00Z,
,,,test2,3,2020-08-02T17:22:01Z,
,,,test2,4,2020-08-02T17:22:01Z,
,,,test2,5,2020-08-02T17:22:01Z,
,,,test2,6,2020-08-02T17:22:02Z,
,,,test2,7,2020-08-02T17:22:03Z,
,,,test2,8,2020-08-02T17:22:03Z,
,,,test2,9,2020-08-02T17:22:04Z,
"
d = csv.from(csv:inData)

d   |> drop(columns: ["_start", "_stop"])
    |> multirow.map(left: 1s, right: 1, fn: (window, row) => 
		window |> count() |> map(fn: (r)=>({r with _time: row._time, _field: row._field}))
    )
`

	prog, err := lang.Compile(script, runtime.Default, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	ctx := flux.NewDefaultDependencies().Inject(context.Background())
	query, err := prog.Start(ctx, &memory.Allocator{})

	if err != nil {
		t.Fatal(err)
	}
	res := <-query.Results()
	if query.Err() != nil {
		t.Fatal(err)
	}

	err = res.Tables().Do(func(table flux.Table) error {
		return table.Do(func(reader flux.ColReader) error {
			rc := reader.Len()
			for row := 0; row < rc; row++ {
				cc := len(reader.Cols())
				for col := 0; col < cc; col++ {
					var v interface{}

					switch reader.Cols()[col].Type {
					case flux.TBool:
						column := reader.Bools(col)
						if column.IsNull(row) {
							break
						}
						v = column.Value(row)
					case flux.TInt:
						column := reader.Ints(col)
						if column.IsNull(row) {
							break
						}
						v = column.Value(row)
					case flux.TUInt:
						column := reader.UInts(col)
						if column.IsNull(row) {
							break
						}
						v = column.Value(row)
					case flux.TFloat:
						column := reader.Floats(col)
						if column.IsNull(row) {
							break
						}
						v = column.Value(row)
					case flux.TString:
						column := reader.Strings(col)
						if column.IsNull(row) {
							break
						}
						v = column.Value(row)
					case flux.TTime:
						column := reader.Times(col)
						if column.IsNull(row) {
							break
						}
						v = column.Value(row)
					default:
						panic(fmt.Errorf("unsupported column type %v", reader.Cols()[col].Type))
					}
					fmt.Print(reader.Cols()[col].Label, ":", reader.Cols()[col].Type, "=", v, "\t")
				}
				fmt.Print("\n")
			}
			return nil
		})
	})

	if err != nil {
		t.Fatal(err)
	}

}
