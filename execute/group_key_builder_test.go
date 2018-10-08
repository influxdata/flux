package execute_test

import (
	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/values"
	"testing"

	"github.com/influxdata/flux/execute"
)

func TestGroupKeyBuilder_Empty(t *testing.T) {
	var gkb execute.GroupKeyBuilder
	gkb.AddKeyValue("_measurement", values.NewStringValue("cpu"))

	key, err := gkb.Build()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if got, want := len(key.Cols()), 1; got != want {
		t.Fatalf("unexpected number of columns -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	if got, want := key.Cols(), []flux.ColMeta{
		{Label: "_measurement", Type: flux.TString},
	}; !cmp.Equal(want, got) {
		t.Fatalf("unexpected columns -want/+got:\n%s", cmp.Diff(want, got))
	}

	if got, want := key.Values(), []values.Value{
		values.NewStringValue("cpu"),
	}; !cmp.Equal(want, got) {
		t.Fatalf("unexpected columns -want/+got:\n%s", cmp.Diff(want, got))
	}
}

func TestGroupKeyBuilder_Existing(t *testing.T) {
	gkb := execute.NewGroupKeyBuilder(
		execute.NewGroupKey(
			[]flux.ColMeta{
				{
					Label: "_measurement",
					Type:  flux.TString,
				},
			},
			[]values.Value{
				values.NewStringValue("cpu"),
			},
		),
	)
	gkb.AddKeyValue("_field", values.NewStringValue("usage_user"))

	key, err := gkb.Build()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if got, want := len(key.Cols()), 2; got != want {
		t.Fatalf("unexpected number of columns -want/+got:\n\t- %d\n\t+ %d", want, got)
	}

	if got, want := key.Cols(), []flux.ColMeta{
		{Label: "_measurement", Type: flux.TString},
		{Label: "_field", Type: flux.TString},
	}; !cmp.Equal(want, got) {
		t.Fatalf("unexpected columns -want/+got:\n%s", cmp.Diff(want, got))
	}

	if got, want := key.Values(), []values.Value{
		values.NewStringValue("cpu"),
		values.NewStringValue("usage_user"),
	}; !cmp.Equal(want, got) {
		t.Fatalf("unexpected columns -want/+got:\n%s", cmp.Diff(want, got))
	}
}
