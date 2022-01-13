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

data = "
#datatype,string,long,string,string,long
#group,false,false,false,false,false
#default,_result,0,,,
,result,table,strcol0,strcol1,intcol3
,,,test1,test10,1
,,,test1,test11,
,,,test2,test12,3
,,,test2,test13,4
"
 
d = csv.from(csv:data)
//d|>multirow.map(left:1, fn: (row,window)=> window 
//	|> mean(column:"intcol3")  
//	|> map(fn: (r) => ({ r with c0: row.strcol0, d2: exists row.intcol3 })) 
//)


//d|>multirow.rowNumber()

//d|>multirow.map(fn: (row)=> row) 

//d|>multirow.simpleAMA(n: 2, column: "intcol3") 


//d|> group(columns:["strcol0"])|> multirow.map(fn: (row, index)=> ({strcol0: row.strcol0 + string(v: index), a: "111"}))|> count(column:"a")


d|> multirow.map(fn: (previous, row) => {
x = previous.x_col*2 -1
return {row with 
	concat : (if exists previous.concat then previous.concat + "," else "")  + row.strcol1, 
	x_col : x,
	val :  x % 100
}}
, init : {x_col : 100}
, virtual : ["x_col"]
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
					fmt.Print(reader.Cols()[col].Label, "=", v, "\t")
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
