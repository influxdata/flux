package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/prometheus/prometheus/promql"
)

func TestEscapeLabelName(t *testing.T) {
	escapedNames := map[string]string{
		"test":     "test",
		"_test":    "~_test",
		"__name__": "_measurement",
		"__foo__":  "~__foo__",
	}

	for ln, want := range escapedNames {
		got := escapeLabelName(ln)

		if got != want {
			t.Fatalf("want %q, got %q", want, got)
		}
	}
}

func TestUnescapeLabelName(t *testing.T) {
	escapedNames := map[string]string{
		"test":         "test",
		"~_test":       "_test",
		"_measurement": "__name__",
		"~__foo__":     "__foo__",
	}

	for ln, want := range escapedNames {
		got := unescapeLabelName(ln)

		if got != want {
			t.Fatalf("want %q, got %q", want, got)
		}
	}
}

func TestEscapeLabelNames(t *testing.T) {
	tests := []struct {
		labelNames []string
		want       []string
	}{
		{
			labelNames: nil,
			want:       []string{},
		},
		{
			labelNames: []string{},
			want:       []string{},
		},
		{
			labelNames: []string{"test", "_test", "__name__", "__foo__"},
			want:       []string{"test", "~_test", "_measurement", "~__foo__"},
		},
	}

	for _, test := range tests {
		got := escapeLabelNames(test.labelNames)

		if diff := cmp.Diff(test.want, got); diff != "" {
			t.Fatalf("unexpected escaped label names:\n%s", diff)
		}
	}
}

func TestEscapeExpression(t *testing.T) {
	tests := []struct {
		expr string
		want string
	}{
		{
			expr: `foo{bar!="baz",_value="value"}`,
			want: `{_measurement="foo",bar!="baz",~_value="value"}`,
		},
		{
			expr: `{__name__=~".+"}`,
			want: `{_measurement=~".+"}`,
		},
		{
			expr: `foo{bar!="baz",_value="value"}[5m]`,
			want: `{_measurement="foo",bar!="baz",~_value="value"}[5m]`,
		},
		{
			expr: `{__name__=~".+"}[5m]`,
			want: `{_measurement=~".+"}[5m]`,
		},
		{
			expr: `sum by(test, _value, __name__, __foo__) (foo)`,
			want: `sum by(test, ~_value, _measurement, ~__foo__) ({_measurement="foo"})`,
		},
		{
			expr: `sum without(test, _value, __name__, __foo__) (foo)`,
			want: `sum without(test, ~_value, _measurement, ~__foo__) ({_measurement="foo"})`,
		},
		{
			expr: `foo / on(test, _value, __name__, __foo__) group_left(_time) bar`,
			want: `{_measurement="foo"} / on(test, ~_value, _measurement, ~__foo__) group_left(~_time) {_measurement="bar"}`,
		},
		{
			expr: `foo / ignoring(test, _value, __name__, __foo__) group_right(_time) bar`,
			want: `{_measurement="foo"} / ignoring(test, ~_value, _measurement, ~__foo__) group_right(~_time) {_measurement="bar"}`,
		},
	}

	for _, test := range tests {
		node, err := promql.ParseExpr(test.expr)
		if err != nil {
			t.Fatal(err)
		}
		promql.Walk(labelNameEscaper{}, node, nil)
		got := node.String()

		if diff := cmp.Diff(test.want, got); diff != "" {
			t.Fatalf("unexpected escaped expression:\n%s", diff)
		}
	}
}
