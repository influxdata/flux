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

// EqualResults compares two lists of Flux Results for equlity
func EqualResults(want, got []flux.Result) error {
	if len(want) != len(got) {
		return fmt.Errorf("unexpected number of results - want %d results, got %d results", len(want), len(got))
	}
	for i := range want {
		err := EqualResult(want[i], got[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// EqualResultIterators compares two ResultIterators for equlity
func EqualResultIterators(want, got flux.ResultIterator) error {
	for {
		if w, g := want.More(), got.More(); w != g {
			return fmt.Errorf("unexpected number of results: want more %t, got more %t", w, g)
		} else if w {
			err := EqualResult(want.Next(), got.Next())
			if err != nil {
				return err
			}
		} else {
			if w, g := want.Err(), got.Err(); !(w == nil && g == nil || w != nil && g != nil && w.Error() == g.Error()) {
				return fmt.Errorf("unexpected errors want: %s got: %s", w, g)
			}
			return nil
		}
	}
}

// EqualResult compares to results for equality
func EqualResult(w, g flux.Result) error {
	if w.Name() != g.Name() {
		return fmt.Errorf("unexpected result name - want %s, got %s", w.Name(), g.Name())
	}
	var wt, gt []*Table
	if err := w.Tables().Do(func(tbl flux.Table) error {
		t, err := ConvertTable(tbl)
		if err != nil {
			return err
		}
		wt = append(wt, t)
		return nil
	}); err != nil {
		return err
	}
	if err := g.Tables().Do(func(tbl flux.Table) error {
		t, err := ConvertTable(tbl)
		if err != nil {
			return err
		}
		gt = append(gt, t)
		return nil
	}); err != nil {
		return err
	}
	NormalizeTables(wt)
	NormalizeTables(gt)
	if len(wt) != len(gt) {
		return fmt.Errorf("unexpected size for result %s - want %d tables, got %d tables", w.Name(), len(wt), len(gt))
	}
	if !cmp.Equal(wt, gt, floatOptions) {
		return fmt.Errorf("unexpected tables -want/+got\n%s", cmp.Diff(wt, gt))
	}
	return nil
}
