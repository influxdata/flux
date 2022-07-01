package executetest

import (
	"fmt"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
)

type Result struct {
	Nm   string
	Tbls []*Table
	Err  error
}

// ConvertResult produces a result object from any flux.Result type.
func ConvertResult(result flux.Result) *Result {
	var tbls []*Table
	err := result.Tables().Do(func(tbl flux.Table) error {
		t, err := ConvertTable(tbl)
		if err != nil {
			return err
		}
		tbls = append(tbls, t)
		return nil
	})
	return &Result{
		Nm:   result.Name(),
		Tbls: tbls,
		Err:  err,
	}
}

func NewResult(tables []*Table) *Result {
	return &Result{Tbls: tables}
}

func (r *Result) Name() string {
	return r.Nm
}

func (r *Result) Tables() flux.TableIterator {
	return &TableIterator{
		r.Tbls,
		r.Err,
	}
}

func (r *Result) Normalize() {
	NormalizeTables(r.Tbls)
}

type TableIterator struct {
	Tables []*Table
	Err    error
}

func (ti *TableIterator) Do(f func(flux.Table) error) error {
	if ti.Err != nil {
		return ti.Err
	}
	for _, t := range ti.Tables {
		if err := f(t); err != nil {
			return err
		}
	}
	return nil
}

// EqualResults compares two lists of Flux Results for equality
func EqualResults(want, got []flux.Result) error {
	wantTables := convertResults(want)
	gotTables := convertResults(got)
	if diff := cmp.Diff(wantTables, gotTables, floatOptions); diff != "" {
		return fmt.Errorf("unexpected iterator results; -want/+got\n%s", diff)
	}
	return nil
}

func convertResults(rs []flux.Result) []*Result {
	tables := make([]*Result, len(rs))
	for i, r := range rs {
		tables[i] = ConvertResult(r)
	}
	return tables
}

// EqualResultIterators compares two ResultIterators for equality
func EqualResultIterators(want, got flux.ResultIterator) error {
	wantResults, wantErr := readAllIterator(want)
	gotResults, gotErr := readAllIterator(got)

	if diff := cmp.Diff(wantResults, gotResults, floatOptions); diff != "" {
		return fmt.Errorf("unexpected iterator results; -want/+got\n%s", diff)
	}
	if wantErr == nil && gotErr == nil {
		return nil
	}
	if wantErr == nil || gotErr == nil || wantErr.Error() != gotErr.Error() {
		return fmt.Errorf("unexpected errors got %v; want: %v", gotErr, wantErr)
	}
	return nil
}

func readAllIterator(iter flux.ResultIterator) ([][]*Table, error) {
	results := [][]*Table{}
	for iter.More() {
		tables := []*Table{}
		err := iter.Next().Tables().Do(func(tbl flux.Table) error {
			t, err := ConvertTable(tbl)
			if err != nil {
				return fmt.Errorf("cannot convert table: %v", err)
			}
			tables = append(tables, t)
			return nil
		})
		if err != nil {
			return nil, err
		}
		NormalizeTables(tables)
		results = append(results, tables)
	}
	return results, iter.Err()
}

// EqualResult compares to results for equality
func EqualResult(w, g flux.Result) error {
	want := ConvertResult(w)
	got := ConvertResult(g)
	if diff := cmp.Diff(want, got, floatOptions); diff != "" {
		return fmt.Errorf("unexpected tables -want/+got\n%s", diff)
	}
	return nil
}
