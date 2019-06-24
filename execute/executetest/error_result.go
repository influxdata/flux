package executetest

import (
	"fmt"
	"github.com/influxdata/flux"
)

// ErrorResultIterator will inject errors into various
// parts of the results.  Multiple results will be returned,
// with one result for each element in Tbls, which is a multi-dimensional array.
// - Errors in TblsErrs[i] will be returned by table iterator after processing the tables in Tbls[i]
// - The error in ResErr will be returned by the result iterator after all the results
//   have been returned.
type ErrorResultIterator struct {
	Tbls    [][]*Table
	TblErrs []error
	ResErr  error

	i int
}

func (r *ErrorResultIterator) More() bool {
	return len(r.Tbls) > 0
}

func (r *ErrorResultIterator) Next() flux.Result {
	if len(r.Tbls) == 0 {
		panic("no results")
	}

	tbls := r.Tbls[0]
	r.Tbls = r.Tbls[1:]
	var e error
	if len(r.TblErrs) > 0 {
		e = r.TblErrs[0]
		r.TblErrs = r.TblErrs[1:]
	}

	res := &errorResult{
		name: fmt.Sprintf("_result%v", r.i),
		tbls: tbls,
		err:  e,
	}
	r.i++
	return res
}

func (r *ErrorResultIterator) Release() {
}

func (r *ErrorResultIterator) Err() error {
	if len(r.Tbls) > 0 {
		return nil
	}
	return r.ResErr
}

func (r *ErrorResultIterator) Statistics() flux.Statistics {
	return flux.Statistics{}
}

type errorResult struct {
	name string
	tbls []*Table
	err  error
}

func (er *errorResult) Name() string {
	return er.name
}

func (er *errorResult) Tables() flux.TableIterator {
	return &errorTableIterator{
		tbls: er.tbls,
		err:  er.err,
	}
}

type errorTableIterator struct {
	tbls []*Table
	err  error
}

func (eti *errorTableIterator) Do(f func(flux.Table) error) error {
	for _, t := range eti.tbls {
		if err := f(t); err != nil {
			return err
		}
	}
	return eti.err
}
