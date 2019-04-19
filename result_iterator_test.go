package flux_test

import (
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/mock"
	"github.com/pkg/errors"
)

// ---- Helpers.

func moreIsIdempotentHelper(t *testing.T, ri flux.ResultIterator, steps int) {
	t.Helper()
	defer ri.Release()

	moreValue := ri.More()
	for i := 0; i < steps; i++ {
		if ri.More() != moreValue {
			t.Errorf("More() return value has changed at step %d", i)
		}
	}
}

func nextWhenNoDataHelper(t *testing.T, ri flux.ResultIterator) {
	t.Helper()
	defer ri.Release()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("call to Next() when there is no more data did not panic")
		}
	}()

	for ri.More() {
		_ = ri.Next()
	}
	// this should panic
	_ = ri.Next()
}

func callNextWithoutMoreHelper(t *testing.T, ri flux.ResultIterator, dataLen int) {
	t.Helper()
	defer ri.Release()
	if dataLen < 1 {
		t.Errorf("please pass non-empty data in")
	}

	for i := 0; i < dataLen; i++ {
		_ = ri.Next()
	}
	// data should be finished
	if ri.More() {
		t.Errorf("call to More() returned true when data was supposed to be consumed")
	}
}

// ---- ResultIterator providers.

type resultIteratorProvider struct {
	riName string
	riFn   func(data [][]*executetest.Table) flux.ResultIterator
}

func newQueryResultIterator(data [][]*executetest.Table) flux.ResultIterator {
	q := &mock.Query{}
	q.ProduceResults(func(results chan<- flux.Result, canceled <-chan struct{}) {
		for _, d := range data {
			select {
			case <-canceled:
				return
			default:
				results <- executetest.NewResult(d)
			}
		}
	})
	return flux.NewResultIteratorFromQuery(q)
}

func newMapResultIterator(data [][]*executetest.Table) flux.ResultIterator {
	rm := make(map[string]flux.Result, len(data))
	for i, tables := range data {
		rm["result"+strconv.Itoa(i)] = executetest.NewResult(tables)
	}
	return flux.NewMapResultIterator(rm)
}

func newSliceResultIterator(data [][]*executetest.Table) flux.ResultIterator {
	rs := make([]flux.Result, len(data))
	for i, tables := range data {
		rs[i] = executetest.NewResult(tables)
	}
	return flux.NewSliceResultIterator(rs)
}

var (
	providers = []resultIteratorProvider{
		{riName: "queryResultIterator", riFn: newQueryResultIterator},
		{riName: "mapResultIterator", riFn: newMapResultIterator},
		{riName: "sliceResultIterator", riFn: newSliceResultIterator},
	}
	sampleData = [][]*executetest.Table{
		{
			{
				ColMeta: []flux.ColMeta{
					{Label: "value", Type: flux.TInt},
					{Label: "tag", Type: flux.TString},
				},
				KeyCols: []string{"tag"},
				Data: [][]interface{}{
					{int64(10), "a"},
				},
			},
			{
				ColMeta: []flux.ColMeta{
					{Label: "value", Type: flux.TInt},
					{Label: "tag", Type: flux.TString},
				},
				KeyCols: []string{"tag"},
				Data: [][]interface{}{
					{int64(20), "b"},
				},
			},
			{
				ColMeta: []flux.ColMeta{
					{Label: "value", Type: flux.TInt},
					{Label: "tag", Type: flux.TString},
				},
				KeyCols: []string{"tag"},
				Data: [][]interface{}{
					{int64(30), "c"},
				},
			},
		},
	}
)

// ---- Property tests.

func TestResultIterator_MoreIsIdempotent(t *testing.T) {
	for _, p := range providers {
		p := p
		t.Run(p.riName+" - more is idempotent", func(t *testing.T) {
			t.Parallel()
			ri := p.riFn(sampleData)
			moreIsIdempotentHelper(t, ri, len(sampleData)*10)
		})
	}
}

func TestResultIterator_PanicWhenCallingNextAndThereIsNoDataLeft(t *testing.T) {
	for _, p := range providers {
		p := p
		t.Run(p.riName+" - next when no data left", func(t *testing.T) {
			t.Parallel()
			ri := p.riFn(sampleData)
			nextWhenNoDataHelper(t, ri)
		})
	}
}

func TestResultIterator_CanCallNextWithoutCallingMoreFirst(t *testing.T) {
	for _, p := range providers {
		p := p
		t.Run(p.riName+" - next without calling more first", func(t *testing.T) {
			t.Parallel()
			ri := p.riFn(sampleData)
			callNextWithoutMoreHelper(t, ri, len(sampleData))
		})
	}
}

// ---- QueryResultIterator-specific tests.

func TestQueryResultIterator_Results(t *testing.T) {
	type row struct {
		Value int64
		Tag   string
	}

	ri := newQueryResultIterator(sampleData)
	defer ri.Release()

	// Create a slice with elements for every row in tables.
	got := make([]row, 0)
	for ri.More() {
		if err := ri.Next().Tables().Do(func(table flux.Table) error {
			return table.Do(func(cr flux.ColReader) error {
				for i := 0; i < cr.Len(); i++ {
					r := row{
						Value: cr.Ints(0).Value(i),
						Tag:   cr.Strings(1).ValueString(i),
					}
					got = append(got, r)
				}
				return nil
			})
		}); err != nil {
			t.Fatal(err)
		}
	}

	if ri.Err() != nil {
		t.Fatal(errors.Wrap(ri.Err(), "unexpected error in result iterator"))
	}

	want := []row{
		{Value: 10, Tag: "a"},
		{Value: 20, Tag: "b"},
		{Value: 30, Tag: "c"},
	}

	if !cmp.Equal(want, got) {
		t.Fatalf("got unexpected results -want/got:\n%s\n", cmp.Diff(want, got))
	}
}

func TestQueryResultIterator_Cancel(t *testing.T) {
	const sleepInterval = 1 * time.Millisecond

	q := &mock.Query{}
	ri := flux.NewResultIteratorFromQuery(q)
	defer ri.Release()

	var cancelCalled bool
	q.ProduceResults(func(results chan<- flux.Result, canceled <-chan struct{}) {
		for {
			select {
			case <-canceled:
				cancelCalled = true
				return
			default:
				time.Sleep(sleepInterval)
				results <- executetest.NewResult([]*executetest.Table{})
			}
		}
	})

	go func() {
		time.Sleep(sleepInterval * 10)
		q.Cancel()
	}()

	for ri.More() {
		_ = ri.Next()
	}

	if !cancelCalled {
		t.Fatalf("cancel wasn't called")
	}
}

func TestQueryResultIterator_Error(t *testing.T) {
	expectedErr := errors.New("hello, I am an error")
	q := &mock.Query{}
	ri := flux.NewResultIteratorFromQuery(q)
	defer ri.Release()

	q.ProduceResults(func(results chan<- flux.Result, canceled <-chan struct{}) {
		// wait until the query hasn't been canceled
		<-canceled
	})

	go func() {
		time.Sleep(100 * time.Millisecond)
		q.SetErr(expectedErr)
	}()

	for ri.More() {
		_ = ri.Next()
	}

	if ri.Err() != expectedErr {
		t.Fatalf("didnt' get the expected error: -want/got:\n%s\n", cmp.Diff(expectedErr, ri.Err()))
	}
}

func TestQueryResultIterator_Statistics(t *testing.T) {
	const sleepInterval = 1 * time.Millisecond

	vanillaStats := flux.Statistics{}
	expectedStats := flux.Statistics{ExecuteDuration: 1 * time.Second}
	q := &mock.Query{}
	ri := flux.NewResultIteratorFromQuery(q)

	q.SetStatistics(expectedStats)
	q.ProduceResults(func(results chan<- flux.Result, canceled <-chan struct{}) {
		for {
			select {
			case <-canceled:
				return
			default:
				time.Sleep(sleepInterval)
				results <- executetest.NewResult([]*executetest.Table{})
			}
		}
	})

	go func() {
		time.Sleep(sleepInterval * 100)
		q.Cancel()
	}()

	for ri.More() {
		if diff := cmp.Diff(vanillaStats, q.Statistics()); diff != "" {
			t.Fatalf("unexpected stats: -want/got:\n%s\n", diff)
		}
		_ = ri.Next()
	}
	ri.Release()

	// stats changed only after the query has been Done.
	if diff := cmp.Diff(expectedStats, q.Statistics()); diff != "" {
		t.Fatalf("unexpected stats: -want/got:\n%s\n", diff)
	}
}
