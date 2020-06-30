package universe_test

import (
	"context"
	"errors"
	"testing"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/stdlib/universe"
	"github.com/influxdata/flux/values/valuestest"
)

func TestDropRenameKeep_Deprecated_Process(t *testing.T) {
	testCases := []struct {
		name    string
		spec    plan.ProcedureSpec
		data    []flux.Table
		want    []*executetest.Table
		wantErr error
	}{
		{
			name: "rename multiple cols",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.RenameOpSpec{
						Columns: map[string]string{
							"1a": "1b",
							"2a": "2b",
							"3a": "3b",
						},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "2a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "1b", Type: flux.TFloat},
					{Label: "2b", Type: flux.TFloat},
					{Label: "3b", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
		},

		{
			name: "drop multiple cols",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DropOpSpec{
						Columns: []string{"a", "b"},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TFloat},
					{Label: "b", Type: flux.TFloat},
					{Label: "c", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "c", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{3.0},
					{13.0},
					{23.0},
				},
			}},
		},
		{
			name: "drop key col merge tables",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DropOpSpec{
						Columns: []string{"b"},
					},
				},
			},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TString},
						{Label: "b", Type: flux.TString},
						{Label: "c", Type: flux.TFloat},
					},
					KeyCols: []string{"a", "b"},
					Data: [][]interface{}{
						{"one", "two", 3.0},
						{"one", "two", 13.0},
					},
				},
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TString},
						{Label: "b", Type: flux.TString},
						{Label: "c", Type: flux.TFloat},
					},
					KeyCols: []string{"a", "b"},
					Data: [][]interface{}{
						{"one", "three", 5.0},
						{"one", "three", 15.0},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TString},
					{Label: "c", Type: flux.TFloat},
				},
				KeyCols: []string{"a"},
				Data: [][]interface{}{
					{"one", 3.0},
					{"one", 13.0},
					{"one", 5.0},
					{"one", 15.0},
				},
			}},
		},
		{
			name: "drop key col merge error column count",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DropOpSpec{
						Columns: []string{"b"},
					},
				},
			},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TString},
						{Label: "b", Type: flux.TString},
						{Label: "c", Type: flux.TFloat},
					},
					KeyCols: []string{"a", "b"},
					Data: [][]interface{}{
						{"one", "two", 3.0},
						{"one", "two", 13.0},
					},
				},
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TString},
						{Label: "b", Type: flux.TString},
					},
					KeyCols: []string{"a", "b"},
					Data: [][]interface{}{
						{"one", "three"},
						{"one", "three"},
					},
				},
			},
			wantErr: errors.New("requested operation merges tables with different numbers of columns for group key {a=one}"),
		},
		{
			name: "drop key col merge error column type",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DropOpSpec{
						Columns: []string{"b"},
					},
				},
			},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TString},
						{Label: "b", Type: flux.TString},
						{Label: "c", Type: flux.TFloat},
					},
					KeyCols: []string{"a", "b"},
					Data: [][]interface{}{
						{"one", "two", 3.0},
						{"one", "two", 13.0},
					},
				},
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TString},
						{Label: "b", Type: flux.TString},
						{Label: "c", Type: flux.TString},
					},
					KeyCols: []string{"a", "b"},
					Data: [][]interface{}{
						{"one", "three", "val"},
						{"one", "three", "val"},
					},
				},
			},
			wantErr: errors.New("requested operation merges tables with different schemas for group key {a=one}"),
		},
		{
			name: "drop no exist",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DropOpSpec{
						Columns: []string{"boo"},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TString},
					{Label: "b", Type: flux.TString},
					{Label: "c", Type: flux.TFloat},
				},
				KeyCols: []string{"a", "b"},
				Data: [][]interface{}{
					{"one", "two", 3.0},
					{"one", "two", 13.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TString},
					{Label: "b", Type: flux.TString},
					{Label: "c", Type: flux.TFloat},
				},
				KeyCols: []string{"a", "b"},
				Data: [][]interface{}{
					{"one", "two", 3.0},
					{"one", "two", 13.0},
				},
			}},
		},
		{
			name: "keep multiple cols",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.KeepOpSpec{
						Columns: []string{"a"},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TFloat},
					{Label: "b", Type: flux.TFloat},
					{Label: "c", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0},
					{11.0},
					{21.0},
				},
			}},
		},
		{
			name: "keep one key col merge tables",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.KeepOpSpec{
						Columns: []string{"a", "c"},
					},
				},
			},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TString},
						{Label: "b", Type: flux.TString},
						{Label: "c", Type: flux.TFloat},
					},
					KeyCols: []string{"a", "b"},
					Data: [][]interface{}{
						{"one", "two", 3.0},
						{"one", "two", 13.0},
					},
				},
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TString},
						{Label: "b", Type: flux.TString},
						{Label: "c", Type: flux.TFloat},
					},
					KeyCols: []string{"a", "b"},
					Data: [][]interface{}{
						{"one", "three", 5.0},
						{"one", "three", 15.0},
					},
				},
			},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TString},
					{Label: "c", Type: flux.TFloat},
				},
				KeyCols: []string{"a"},
				Data: [][]interface{}{
					{"one", 3.0},
					{"one", 13.0},
					{"one", 5.0},
					{"one", 15.0},
				},
			}},
		},
		{
			name: "keep one key col merge error column count",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.KeepOpSpec{
						Columns: []string{"a", "c"},
					},
				},
			},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TString},
						{Label: "b", Type: flux.TString},
						{Label: "c", Type: flux.TFloat},
					},
					KeyCols: []string{"a", "b"},
					Data: [][]interface{}{
						{"one", "two", 3.0},
						{"one", "two", 13.0},
					},
				},
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TString},
						{Label: "b", Type: flux.TString},
					},
					KeyCols: []string{"a", "b"},
					Data: [][]interface{}{
						{"one", "three"},
						{"one", "three"},
					},
				},
			},
			wantErr: errors.New("requested operation merges tables with different numbers of columns for group key {a=one}"),
		},
		{
			name: "keep one key col merge error column type",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.KeepOpSpec{
						Columns: []string{"a", "c"},
					},
				},
			},
			data: []flux.Table{
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TString},
						{Label: "b", Type: flux.TString},
						{Label: "c", Type: flux.TFloat},
					},
					KeyCols: []string{"a", "b"},
					Data: [][]interface{}{
						{"one", "two", 3.0},
						{"one", "two", 13.0},
					},
				},
				&executetest.Table{
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TString},
						{Label: "b", Type: flux.TString},
						{Label: "c", Type: flux.TString},
					},
					KeyCols: []string{"a", "b"},
					Data: [][]interface{}{
						{"one", "three", "foo"},
						{"one", "three", "bar"},
					},
				},
			},
			wantErr: errors.New("requested operation merges tables with different schemas for group key {a=one}"),
		},
		{
			name: "duplicate single col",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DuplicateOpSpec{
						Column: "a",
						As:     "a_1",
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TFloat},
					{Label: "b", Type: flux.TFloat},
					{Label: "c", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TFloat},
					{Label: "b", Type: flux.TFloat},
					{Label: "c", Type: flux.TFloat},
					{Label: "a_1", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0, 1.0},
					{11.0, 12.0, 13.0, 11.0},
					{21.0, 22.0, 23.0, 21.0},
				},
			}},
		},
		{
			name: "rename map fn (column) => name",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.RenameOpSpec{
						Fn: interpreter.ResolvedFunction{
							Fn:    executetest.FunctionExpression(t, `(column) => "new_name"`),
							Scope: valuestest.Scope(),
						},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "2a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			wantErr: errors.New("table builder already has column with label new_name"),
		},
		{
			name: "drop predicate (column) => column ~= /reg/",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DropOpSpec{
						Predicate: interpreter.ResolvedFunction{
							Fn:    executetest.FunctionExpression(t, `(column) => column =~ /server*/`),
							Scope: valuestest.Scope(),
						},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "server1", Type: flux.TFloat},
					{Label: "local", Type: flux.TFloat},
					{Label: "server2", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "local", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{2.0},
					{12.0},
					{22.0},
				},
			}},
		},
		{
			name: "keep predicate (column) => column ~= /reg/",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.KeepOpSpec{
						Predicate: interpreter.ResolvedFunction{
							Fn:    executetest.FunctionExpression(t, `(column) => column =~ /server*/`),
							Scope: valuestest.Scope(),
						},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "server1", Type: flux.TFloat},
					{Label: "local", Type: flux.TFloat},
					{Label: "server2", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "server1", Type: flux.TFloat},
					{Label: "server2", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 3.0},
					{11.0, 13.0},
					{21.0, 23.0},
				},
			}},
		},
		{
			name: "drop and rename",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DropOpSpec{
						Columns: []string{"server1", "server2"},
					},
					&universe.RenameOpSpec{
						Columns: map[string]string{
							"local": "localhost",
						},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "server1", Type: flux.TFloat},
					{Label: "local", Type: flux.TFloat},
					{Label: "server2", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "localhost", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{2.0},
					{12.0},
					{22.0},
				},
			}},
		},
		{
			name: "drop no exist",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DropOpSpec{
						Columns: []string{"no_exist"},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "server1", Type: flux.TFloat},
					{Label: "local", Type: flux.TFloat},
					{Label: "server2", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "server1", Type: flux.TFloat},
					{Label: "local", Type: flux.TFloat},
					{Label: "server2", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
		},
		{
			name: "rename no exist",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.RenameOpSpec{
						Columns: map[string]string{
							"no_exist": "noexist",
						},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "server1", Type: flux.TFloat},
					{Label: "local", Type: flux.TFloat},
					{Label: "server2", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			want:    []*executetest.Table(nil),
			wantErr: errors.New(`rename error: column "no_exist" doesn't exist`),
		},
		{
			name: "keep no exist",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.KeepOpSpec{
						Columns: []string{"no_exist"},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "server1", Type: flux.TFloat},
					{Label: "local", Type: flux.TFloat},
					{Label: "server2", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			want: []*executetest.Table{{
				Data: [][]interface{}(nil),
			}},
		},
		{
			name: "keep no exist along with all other columns",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.KeepOpSpec{
						Columns: []string{"no_exist", "server1", "local", "server2"},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "server1", Type: flux.TFloat},
					{Label: "local", Type: flux.TFloat},
					{Label: "server2", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "server1", Type: flux.TFloat},
					{Label: "local", Type: flux.TFloat},
					{Label: "server2", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
		},
		{
			name: "duplicate no exist",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DuplicateOpSpec{
						Column: "no_exist",
						As:     "no_exist_2",
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "server1", Type: flux.TFloat},
					{Label: "local", Type: flux.TFloat},
					{Label: "server2", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 12.0, 13.0},
					{21.0, 22.0, 23.0},
				},
			}},
			want:    []*executetest.Table(nil),
			wantErr: errors.New(`duplicate error: column "no_exist" doesn't exist`),
		},
		{
			name: "rename group key",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.RenameOpSpec{
						Columns: map[string]string{
							"1a": "1b",
							"2a": "2b",
							"3a": "3b",
						},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"1a"},
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "2a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{1.0, 12.0, 13.0},
					{1.0, 22.0, 23.0},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string{"1b"},
				ColMeta: []flux.ColMeta{
					{Label: "1b", Type: flux.TFloat},
					{Label: "2b", Type: flux.TFloat},
					{Label: "3b", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{1.0, 12.0, 13.0},
					{1.0, 22.0, 23.0},
				},
			}},
		},
		{
			name: "drop group key",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DropOpSpec{
						Columns: []string{"2a"},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"2a"},
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "2a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{11.0, 2.0, 13.0},
					{21.0, 2.0, 23.0},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string(nil),
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 3.0},
					{11.0, 13.0},
					{21.0, 23.0},
				},
			}},
		},
		{
			name: "keep group key",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.KeepOpSpec{
						Columns: []string{"1a"},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"1a", "3a"},
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "2a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{1.0, 12.0, 3.0},
					{1.0, 22.0, 3.0},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string{"1a"},
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0},
					{1.0},
					{1.0},
				},
			}},
		},
		{
			name: "duplicate group key",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DuplicateOpSpec{
						Column: "1a",
						As:     "1a_1",
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"1a", "3a"},
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "2a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0},
					{1.0, 12.0, 3.0},
					{1.0, 22.0, 3.0},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string{"1a", "3a"},
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "2a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
					{Label: "1a_1", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, 3.0, 1.0},
					{1.0, 12.0, 3.0, 1.0},
					{1.0, 22.0, 3.0, 1.0},
				},
			}},
		},
		{
			name: "keep with changing schema",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.KeepOpSpec{
						Columns: []string{"a"},
					},
				},
			},
			data: []flux.Table{
				&executetest.Table{
					KeyCols: []string{"a"},
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TInt},
						{Label: "b", Type: flux.TFloat},
						{Label: "c", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{int64(1), 10.0, 3.0},
						{int64(1), 12.0, 4.0},
						{int64(1), 22.0, 5.0},
					},
				},
				&executetest.Table{
					KeyCols: []string{"a"},
					ColMeta: []flux.ColMeta{
						{Label: "a", Type: flux.TInt},
						{Label: "b", Type: flux.TFloat},
					},
					Data: [][]interface{}{
						{int64(2), 11.0},
						{int64(2), 13.0},
						{int64(2), 23.0},
					},
				},
			},
			want: []*executetest.Table{
				{
					KeyCols: []string{"a"},
					ColMeta: []flux.ColMeta{{Label: "a", Type: flux.TInt}},
					Data: [][]interface{}{
						{int64(1)},
						{int64(1)},
						{int64(1)},
					},
				},
				{
					KeyCols: []string{"a"},
					ColMeta: []flux.ColMeta{{Label: "a", Type: flux.TInt}},
					Data: [][]interface{}{
						{int64(2)},
						{int64(2)},
						{int64(2)},
					},
				},
			},
		},
		{
			name: "rename with nulls",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.RenameOpSpec{
						Columns: map[string]string{
							"1a": "1b",
							"2a": "2b",
							"3a": "3b",
						},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "2a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{nil, 2.0, 3.0},
					{11.0, 12.0, nil},
					{21.0, nil, nil},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "1b", Type: flux.TFloat},
					{Label: "2b", Type: flux.TFloat},
					{Label: "3b", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{nil, 2.0, 3.0},
					{11.0, 12.0, nil},
					{21.0, nil, nil},
				},
			}},
		},

		{
			name: "drop with nulls",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DropOpSpec{
						Columns: []string{"a", "b"},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TFloat},
					{Label: "b", Type: flux.TFloat},
					{Label: "c", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{nil, 2.0, 3.0},
					{nil, nil, nil},
					{nil, 22.0, nil},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "c", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{3.0},
					{nil},
					{nil},
				},
			}},
		},
		{
			name: "keep with nulls",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.KeepOpSpec{
						Columns: []string{"a"},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TFloat},
					{Label: "b", Type: flux.TFloat},
					{Label: "c", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, nil},
					{nil, 12.0, 13.0},
					{21.0, nil, 23.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0},
					{nil},
					{21.0},
				},
			}},
		},
		{
			name: "duplicate with nulls",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DuplicateOpSpec{
						Column: "a",
						As:     "a_1",
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TFloat},
					{Label: "b", Type: flux.TFloat},
					{Label: "c", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{nil, nil, 3.0},
					{nil, 12.0, nil},
					{21.0, nil, 23.0},
				},
			}},
			want: []*executetest.Table{{
				ColMeta: []flux.ColMeta{
					{Label: "a", Type: flux.TFloat},
					{Label: "b", Type: flux.TFloat},
					{Label: "c", Type: flux.TFloat},
					{Label: "a_1", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{nil, nil, 3.0, nil},
					{nil, 12.0, nil, nil},
					{21.0, nil, 23.0, 21.0},
				},
			}},
		},
		{
			name: "rename group key with nulls",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.RenameOpSpec{
						Columns: map[string]string{
							"1a": "1b",
							"2a": "2b",
							"3a": "3b",
						},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"1a"},
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "2a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{nil, 2.0, 3.0},
					{nil, 12.0, nil},
					{nil, nil, 23.0},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string{"1b"},
				ColMeta: []flux.ColMeta{
					{Label: "1b", Type: flux.TFloat},
					{Label: "2b", Type: flux.TFloat},
					{Label: "3b", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{nil, 2.0, 3.0},
					{nil, 12.0, nil},
					{nil, nil, 23.0},
				},
			}},
		},
		{
			name: "drop group key with nulls",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DropOpSpec{
						Columns: []string{"2a"},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"2a"},
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "2a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, nil, 3.0},
					{nil, nil, 13.0},
					{21.0, nil, nil},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string(nil),
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 3.0},
					{nil, 13.0},
					{21.0, nil},
				},
			}},
		},
		{
			name: "keep group key with nulls",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.KeepOpSpec{
						Columns: []string{"1a"},
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"1a", "3a"},
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "2a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{nil, 2.0, nil},
					{nil, 12.0, nil},
					{nil, 22.0, nil},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string{"1a"},
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{nil},
					{nil},
					{nil},
				},
			}},
		},
		{
			name: "duplicate group key with nulls",
			spec: &universe.SchemaMutationProcedureSpec{
				Mutations: []universe.SchemaMutation{
					&universe.DuplicateOpSpec{
						Column: "3a",
						As:     "3a_1",
					},
				},
			},
			data: []flux.Table{&executetest.Table{
				KeyCols: []string{"1a", "3a"},
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "2a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, nil},
					{1.0, 12.0, nil},
					{1.0, 22.0, nil},
				},
			}},
			want: []*executetest.Table{{
				KeyCols: []string{"1a", "3a"},
				ColMeta: []flux.ColMeta{
					{Label: "1a", Type: flux.TFloat},
					{Label: "2a", Type: flux.TFloat},
					{Label: "3a", Type: flux.TFloat},
					{Label: "3a_1", Type: flux.TFloat},
				},
				Data: [][]interface{}{
					{1.0, 2.0, nil, nil},
					{1.0, 12.0, nil, nil},
					{1.0, 22.0, nil, nil},
				},
			}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			executetest.ProcessTestHelper(
				t,
				tc.data,
				tc.want,
				tc.wantErr,
				func(d execute.Dataset, cache execute.TableBuilderCache) execute.Transformation {
					spec := tc.spec.(*universe.SchemaMutationProcedureSpec)
					tr, err := universe.NewDeprecatedSchemaMutationTransformation(context.Background(), spec, d, cache)
					if err != nil {
						t.Fatal(err)
					}
					return tr
				},
			)
		})
	}
}
