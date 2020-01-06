package table_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/internal/execute/table"
	"github.com/influxdata/flux/values"
)

type Builder struct {
	Data *executetest.Table
}

func (b *Builder) Table() (flux.Table, error) {
	data := b.Data
	b.Data = nil
	return data, nil
}

func (b *Builder) Release() {
	b.Data = nil
}

func TestTableBuilderCache(t *testing.T) {
	cache := table.BuilderCache{
		New: func(key flux.GroupKey) table.Builder {
			return &Builder{
				Data: &executetest.Table{
					GroupKey: key,
				},
			}
		},
	}

	key1 := execute.NewGroupKey(
		[]flux.ColMeta{
			{Label: "_measurement", Type: flux.TString},
			{Label: "_field", Type: flux.TString},
		},
		[]values.Value{
			values.NewString("m0"),
			values.NewString("f0"),
		},
	)

	var b *Builder
	if created := cache.Get(key1, &b); !created {
		t.Fatal("table builder was supposed to be created, but reported that it was not")
	} else if want, got := key1, b.Data.GroupKey; !cmp.Equal(want, got) {
		t.Fatalf("unexpected group key -want/+got:\n%s", cmp.Diff(want, got))
	}
}
