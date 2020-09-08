package tasks_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb/tasks"
	"github.com/influxdata/flux/values"
)

func TestLastSuccess(t *testing.T) {
	for _, tt := range []struct {
		name string
		args map[string]values.Value
		want values.Value
	}{
		{
			name: "orTime",
			args: map[string]values.Value{
				"orTime":          values.NewTime(10),
				"lastSuccessTime": values.Null,
			},
			want: values.NewTime(10),
		},
		{
			name: "lastSuccessTime",
			args: map[string]values.Value{
				"orTime":          values.NewTime(10),
				"lastSuccessTime": values.NewTime(5),
			},
			want: values.NewTime(5),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			args := values.NewObjectWithValues(tt.args)
			got, err := tasks.LastSuccess(context.Background(), args)
			if err != nil {
				t.Fatal(err)
			} else if !cmp.Equal(tt.want, got) {
				t.Fatalf("unexpected value -want/+got:\n%s", cmp.Diff(tt.want, got))
			}
		})
	}
}
