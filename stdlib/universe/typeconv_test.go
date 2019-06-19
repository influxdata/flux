package universe

import (
	"testing"
	"github.com/influxdata/flux/values"
)

func TestTypeconv_Bool(t *testing.T)  {
	testCases := []struct {
		name   string
		v      interface{}
		want   bool
	} {
		{ 
			name : "bool(v:1)",
			v : int64(1),
			want : true,
		},
		{
			name : "bool(v:2)",
			v : "true",
			want : true,
		},
		{
			name : "bool(v:3)",
			v : uint64(1),
			want : true,
		},
		{
			name : "bool(v:4)", 
			v : float64(1),
			want : true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t*testing.T) {
			myMap := map[string]values.Value {
				"v" : values.New(tc.v),
			}
			args := values.NewObjectWithValues(myMap)
			c := &boolConv{}
			got, err := c.Call(args)
			if err != nil {
				t.Fatal(err)
			}
			want := values.NewBool(tc.want)
			if !got.Equal(want) {
				t.Fatal("Test failed")
			}
		})
	}
}
