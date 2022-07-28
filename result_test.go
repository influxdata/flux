package flux_test

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/andreyvit/diff"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/iocounter"
	"github.com/influxdata/flux/mock"
	"github.com/influxdata/flux/semantic"
)

// TestColumnType tests that the column type gets returned from a semantic type correctly.
func TestColumnType(t *testing.T) {
	for _, tt := range []struct {
		typ  semantic.MonoType
		want flux.ColType
	}{
		{typ: semantic.BasicString, want: flux.TString},
		{typ: semantic.BasicInt, want: flux.TInt},
		{typ: semantic.BasicUint, want: flux.TUInt},
		{typ: semantic.BasicFloat, want: flux.TFloat},
		{typ: semantic.BasicBool, want: flux.TBool},
		{typ: semantic.BasicTime, want: flux.TTime},
		{typ: semantic.BasicDuration, want: flux.TInvalid},
		{typ: semantic.BasicRegexp, want: flux.TInvalid},
		{typ: semantic.NewArrayType(semantic.BasicString), want: flux.TInvalid},
		{typ: semantic.NewObjectType([]semantic.PropertyType{{Key: []byte("a"), Value: semantic.BasicInt}}), want: flux.TInvalid},
		{typ: semantic.NewFunctionType(semantic.BasicInt, []semantic.ArgumentType{{Name: []byte("a"), Type: semantic.BasicInt}}), want: flux.TInvalid},
	} {
		t.Run(fmt.Sprint(tt.typ), func(t *testing.T) {
			if want, got := tt.want, flux.ColumnType(tt.typ); want != got {
				t.Fatalf("unexpected type -want/+got\n\t- %s\n\t+ %s", want, got)
			}
		})
	}
}

// ResultLineEncoder is a simple line encoder to encode the results.
type ResultLineEncoder struct {
	testing.TB
}

func (enc *ResultLineEncoder) Encode(w io.Writer, result flux.Result) (int64, error) {
	wc := &iocounter.Writer{Writer: w}
	err := result.Tables().Do(func(tbl flux.Table) error {
		return tbl.Do(func(cr flux.ColReader) error {
			for i, n := 0, cr.Len(); i < n; i++ {
				values := make([]string, len(cr.Cols()))
				for j, col := range cr.Cols() {
					v := execute.ValueForRow(cr, i, j)
					values[j] = fmt.Sprintf("%s=%v", col.Label, v)
				}
				_, _ = fmt.Fprintf(wc, "result(%s): %s\n", result.Name(), strings.Join(values, " "))
			}
			return nil
		})
	})
	return wc.Count(), err
}

func (enc *ResultLineEncoder) EncodeError(w io.Writer, err error) error {
	_, _ = fmt.Fprintf(w, "error: %s\n", err.Error())
	return nil
}

func TestDelimitedMultiResultEncoder_Encode(t *testing.T) {
	for _, tt := range []struct {
		name    string
		results func() flux.ResultIterator
		want    string
		wantErr string
	}{
		{
			name: "SingleResult",
			results: func() flux.ResultIterator {
				return flux.NewSliceResultIterator(
					[]flux.Result{
						&executetest.Result{
							Nm: "success",
							Tbls: []*executetest.Table{
								{
									ColMeta: []flux.ColMeta{
										{Label: "_time", Type: flux.TTime},
										{Label: "_value", Type: flux.TFloat},
									},
									Data: [][]interface{}{
										{execute.Time(0), 2.0},
									},
								},
							},
						},
					},
				)
			},
			want: `result(success): _time=1970-01-01T00:00:00.000000000Z _value=2

`,
		},
		{
			name: "MultipleResults",
			results: func() flux.ResultIterator {
				return flux.NewSliceResultIterator(
					[]flux.Result{
						&executetest.Result{
							Nm: "first",
							Tbls: []*executetest.Table{
								{
									ColMeta: []flux.ColMeta{
										{Label: "_time", Type: flux.TTime},
										{Label: "_value", Type: flux.TFloat},
									},
									Data: [][]interface{}{
										{execute.Time(0), 2.0},
									},
								},
							},
						},
						&executetest.Result{
							Nm: "second",
							Tbls: []*executetest.Table{
								{
									ColMeta: []flux.ColMeta{
										{Label: "_time", Type: flux.TTime},
										{Label: "_value", Type: flux.TFloat},
									},
									Data: [][]interface{}{
										{execute.Time(0), 3.0},
									},
								},
							},
						},
					},
				)
			},
			want: `result(first): _time=1970-01-01T00:00:00.000000000Z _value=2

result(second): _time=1970-01-01T00:00:00.000000000Z _value=3

`,
		},
		{
			name: "QueryError",
			results: func() flux.ResultIterator {
				results := make(chan flux.Result)
				close(results)
				q := &mock.Query{
					ResultsCh: results,
				}
				q.SetErr(errors.New("expected error"))
				return flux.NewResultIteratorFromQuery(q)
			},
			wantErr: "expected error",
		},
		{
			name: "ResultError",
			results: func() flux.ResultIterator {
				return flux.NewSliceResultIterator(
					[]flux.Result{
						&executetest.Result{
							Nm:  "test",
							Err: errors.New("expected error"),
						},
					},
				)
			},
			wantErr: "expected error",
		},
		{
			name: "ResultErrorOnSecondResult",
			results: func() flux.ResultIterator {
				return flux.NewSliceResultIterator(
					[]flux.Result{
						&executetest.Result{
							Nm: "success",
							Tbls: []*executetest.Table{
								{
									ColMeta: []flux.ColMeta{
										{Label: "_time", Type: flux.TTime},
										{Label: "_value", Type: flux.TFloat},
									},
									Data: [][]interface{}{
										{execute.Time(0), 2.0},
									},
								},
							},
						},
						&executetest.Result{
							Nm:  "error",
							Err: errors.New("expected error"),
						},
					},
				)
			},
			want: `result(success): _time=1970-01-01T00:00:00.000000000Z _value=2

error: expected error
`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			results := tt.results()
			enc := &flux.DelimitedMultiResultEncoder{
				Delimiter: []byte("\n"),
				Encoder:   &ResultLineEncoder{TB: t},
			}

			var got strings.Builder
			if _, err := enc.Encode(&got, results); err != nil {
				if tt.wantErr != "" {
					if got, want := err.Error(), tt.wantErr; got != want {
						t.Fatalf("unexpected error -want/+got:\n\t- %v\n\t+ %v", got, want)
					}
					return
				}
				t.Fatalf("unexpected error: %v", err)
			}

			if got, want := got.String(), tt.want; got != want {
				t.Fatalf("unexpected output -want/+got\n%s", diff.LineDiff(want, got))
			}
		})
	}
}
