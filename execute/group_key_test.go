package execute_test

import (
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/semantic"
	"github.com/influxdata/flux/values"
)

func TestGroupKey_Equal(t *testing.T) {
	for _, tt := range []struct {
		name        string
		left, right flux.GroupKey
		want        bool
	}{
		{
			name: "Identical",
			left: execute.NewGroupKey(
				[]flux.ColMeta{
					{Label: "a", Type: flux.TString},
					{Label: "b", Type: flux.TString},
				},
				[]values.Value{
					values.NewString("b"),
					values.NewString("c"),
				},
			),
			right: execute.NewGroupKey(
				[]flux.ColMeta{
					{Label: "a", Type: flux.TString},
					{Label: "b", Type: flux.TString},
				},
				[]values.Value{
					values.NewString("b"),
					values.NewString("c"),
				},
			),
			want: true,
		},
		{
			name: "Transposed",
			left: execute.NewGroupKey(
				[]flux.ColMeta{
					{Label: "a", Type: flux.TString},
					{Label: "b", Type: flux.TString},
				},
				[]values.Value{
					values.NewString("b"),
					values.NewString("c"),
				},
			),
			right: execute.NewGroupKey(
				[]flux.ColMeta{
					{Label: "b", Type: flux.TString},
					{Label: "a", Type: flux.TString},
				},
				[]values.Value{
					values.NewString("c"),
					values.NewString("b"),
				},
			),
			want: true,
		},
		{
			name: "Unequal",
			left: execute.NewGroupKey(
				[]flux.ColMeta{
					{Label: "a", Type: flux.TString},
				},
				[]values.Value{
					values.NewString("b"),
				},
			),
			right: execute.NewGroupKey(
				[]flux.ColMeta{
					{Label: "a", Type: flux.TString},
				},
				[]values.Value{
					values.NewString("c"),
				},
			),
			want: false,
		},
		{
			name: "DifferentKeys",
			left: execute.NewGroupKey(
				[]flux.ColMeta{
					{Label: "a", Type: flux.TString},
				},
				[]values.Value{
					values.NewString("b"),
				},
			),
			right: execute.NewGroupKey(
				[]flux.ColMeta{
					{Label: "b", Type: flux.TString},
				},
				[]values.Value{
					values.NewString("b"),
				},
			),
			want: false,
		},
		{
			name: "DifferentLengths",
			left: execute.NewGroupKey(
				[]flux.ColMeta{
					{Label: "a", Type: flux.TString},
					{Label: "b", Type: flux.TString},
				},
				[]values.Value{
					values.NewString("b"),
				},
			),
			right: execute.NewGroupKey(
				[]flux.ColMeta{
					{Label: "b", Type: flux.TString},
				},
				[]values.Value{
					values.NewString("b"),
				},
			),
			want: false,
		},
		{
			name: "NullValue_Equal",
			left: execute.NewGroupKey(
				[]flux.ColMeta{
					{Label: "a", Type: flux.TString},
				},
				[]values.Value{
					values.NewNull(semantic.String),
				},
			),
			right: execute.NewGroupKey(
				[]flux.ColMeta{
					{Label: "a", Type: flux.TString},
				},
				[]values.Value{
					values.NewNull(semantic.String),
				},
			),
			want: true,
		},
		{
			name: "NullValue_NotEqual",
			left: execute.NewGroupKey(
				[]flux.ColMeta{
					{Label: "a", Type: flux.TString},
				},
				[]values.Value{
					values.NewNull(semantic.String),
				},
			),
			right: execute.NewGroupKey(
				[]flux.ColMeta{
					{Label: "a", Type: flux.TString},
				},
				[]values.Value{
					values.NewString("b"),
				},
			),
			want: false,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if want, got := tt.want, tt.left.Equal(tt.right); want != got {
				t.Errorf("unexpected result: want=%v got=%v", want, got)
			}
		})
	}
}

func TestGroupKey_Less(t *testing.T) {
	for _, tt := range []struct {
		name        string
		left, right flux.GroupKey
		want        [2]bool
	}{
		{
			name: "Identical",
			left: execute.NewGroupKey(
				[]flux.ColMeta{
					{Label: "a", Type: flux.TString},
					{Label: "b", Type: flux.TString},
				},
				[]values.Value{
					values.NewString("b"),
					values.NewString("c"),
				},
			),
			right: execute.NewGroupKey(
				[]flux.ColMeta{
					{Label: "a", Type: flux.TString},
					{Label: "b", Type: flux.TString},
				},
				[]values.Value{
					values.NewString("b"),
					values.NewString("c"),
				},
			),
			want: [2]bool{false, false},
		},
		{
			name: "Identical_Transposed",
			left: execute.NewGroupKey(
				[]flux.ColMeta{
					{Label: "a", Type: flux.TString},
					{Label: "b", Type: flux.TString},
				},
				[]values.Value{
					values.NewString("b"),
					values.NewString("c"),
				},
			),
			right: execute.NewGroupKey(
				[]flux.ColMeta{
					{Label: "b", Type: flux.TString},
					{Label: "a", Type: flux.TString},
				},
				[]values.Value{
					values.NewString("c"),
					values.NewString("b"),
				},
			),
			want: [2]bool{false, false},
		},
		{
			name: "LessThan",
			left: execute.NewGroupKey(
				[]flux.ColMeta{
					{Label: "a", Type: flux.TString},
				},
				[]values.Value{
					values.NewString("b"),
				},
			),
			right: execute.NewGroupKey(
				[]flux.ColMeta{
					{Label: "a", Type: flux.TString},
				},
				[]values.Value{
					values.NewString("c"),
				},
			),
			want: [2]bool{true, false},
		},
		{
			name: "LessThan_Transposed",
			left: execute.NewGroupKey(
				[]flux.ColMeta{
					{Label: "a", Type: flux.TString},
					{Label: "b", Type: flux.TString},
				},
				[]values.Value{
					values.NewString("b"),
					values.NewString("c"),
				},
			),
			right: execute.NewGroupKey(
				[]flux.ColMeta{
					{Label: "b", Type: flux.TString},
					{Label: "a", Type: flux.TString},
				},
				[]values.Value{
					values.NewString("b"),
					values.NewString("c"),
				},
			),
			want: [2]bool{true, false},
		},
		{
			name: "DifferentKeys",
			left: execute.NewGroupKey(
				[]flux.ColMeta{
					{Label: "a", Type: flux.TString},
				},
				[]values.Value{
					values.NewString("b"),
				},
			),
			right: execute.NewGroupKey(
				[]flux.ColMeta{
					{Label: "b", Type: flux.TString},
				},
				[]values.Value{
					values.NewString("b"),
				},
			),
			want: [2]bool{false, true},
		},
		{
			name: "NullValue_Equal",
			left: execute.NewGroupKey(
				[]flux.ColMeta{
					{Label: "a", Type: flux.TString},
				},
				[]values.Value{
					values.NewNull(semantic.String),
				},
			),
			right: execute.NewGroupKey(
				[]flux.ColMeta{
					{Label: "a", Type: flux.TString},
				},
				[]values.Value{
					values.NewNull(semantic.String),
				},
			),
			want: [2]bool{false, false},
		},
		{
			name: "NullValue_LessThan",
			left: execute.NewGroupKey(
				[]flux.ColMeta{
					{Label: "a", Type: flux.TString},
				},
				[]values.Value{
					values.NewNull(semantic.String),
				},
			),
			right: execute.NewGroupKey(
				[]flux.ColMeta{
					{Label: "a", Type: flux.TString},
				},
				[]values.Value{
					values.NewString("b"),
				},
			),
			want: [2]bool{true, false},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if want, got := tt.want[0], tt.left.Less(tt.right); want != got {
				t.Errorf("unexpected result for left < right: want=%v got=%v", want, got)
			}
			if want, got := tt.want[1], tt.right.Less(tt.left); want != got {
				t.Errorf("unexpected result for right < left: want=%v got=%v", want, got)
			}
		})
	}
}
