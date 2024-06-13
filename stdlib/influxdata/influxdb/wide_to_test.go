package influxdb_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/andreyvit/diff"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/dependencies/dependenciestest"
	"github.com/influxdata/flux/dependency"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/execute/table/static"
	_ "github.com/influxdata/flux/fluxinit/static"
	"github.com/influxdata/flux/internal/operation"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/querytest"
	"github.com/influxdata/flux/stdlib/influxdata/influxdb"
	"github.com/influxdata/flux/stdlib/universe"
)

func TestWideTo_Query(t *testing.T) {
	tests := []querytest.NewQueryTestCase{
		{
			Name: "from range pivot wideTo",
			Raw: `import "influxdata/influxdb"
import "influxdata/influxdb/v1"
from(bucket:"mydb")
  |> range(start: -1h)
  |> v1.fieldsAsCols()
  |> wideTo(bucket:"series1", org:"fred", host:"localhost", token:"auth-token")`,
			Want: &operation.Spec{
				Operations: []*operation.Node{
					{
						ID: "from0",
						Spec: &influxdb.FromOpSpec{
							Bucket: influxdb.NameOrID{Name: "mydb"},
						},
					},
					{
						ID: "range1",
						Spec: &universe.RangeOpSpec{
							Start:       flux.Time{IsRelative: true, Relative: -time.Hour},
							Stop:        flux.Time{IsRelative: true},
							TimeColumn:  "_time",
							StartColumn: "_start",
							StopColumn:  "_stop",
						},
					},
					{
						ID: "pivot2",
						Spec: &universe.PivotOpSpec{
							RowKey:      []string{"_time"},
							ColumnKey:   []string{"_field"},
							ValueColumn: "_value"},
					},
					{
						ID: "wide-to3",
						Spec: &influxdb.WideToOpSpec{
							Bucket: influxdb.NameOrID{Name: "series1"},
							Org:    influxdb.NameOrID{Name: "fred"},
							Host:   "localhost",
							Token:  "auth-token",
						},
					},
				},
				Edges: []operation.Edge{
					{Parent: "from0", Child: "range1"},
					{Parent: "range1", Child: "pivot2"},
					{Parent: "pivot2", Child: "wide-to3"},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			querytest.NewQueryTestHelper(t, tc)
		})
	}
}

func TestWideToTransformation(t *testing.T) {
	var written bytes.Buffer
	deps := dependenciestest.Default()
	deps.Deps.Deps.HTTPClient = &http.Client{
		Transport: dependenciestest.RoundTripFunc(func(req *http.Request) *http.Response {
			if _, err := io.Copy(&written, req.Body); err != nil {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Status:     http.StatusText(http.StatusInternalServerError),
					Body:       io.NopCloser(strings.NewReader(err.Error())),
					Header:     make(http.Header),
				}
			}

			return &http.Response{
				StatusCode: http.StatusNoContent,
				Status:     http.StatusText(http.StatusNoContent),
				Body:       io.NopCloser(bytes.NewReader(nil)),
				Header:     make(http.Header),
			}
		}),
	}

	cache := execute.NewTableBuilderCache(&memory.ResourceAllocator{})
	d := execute.NewDataset(executetest.RandomDatasetID(), execute.DiscardingMode, cache)
	d.SetTriggerSpec(plan.DefaultTriggerSpec)

	ctx, span := dependency.Inject(context.Background(), deps)
	defer span.Finish()
	spec := &influxdb.WideToProcedureSpec{
		Config: influxdb.Config{
			Bucket: influxdb.NameOrID{Name: "mybucket"},
			Host:   "http://localhost:8086",
		},
	}
	tr, err := influxdb.NewWideToTransformation(ctx, d, cache, spec)
	if err != nil {
		t.Fatal(err)
	}

	parentID := executetest.RandomDatasetID()
	input := static.TableGroup{
		static.Times("_time", 0, 10, 20, 30),
		static.Floats("f0", 1.0, 2.0, 3.0, nil),
		static.Ints("f1", 1, nil, 3, nil),
		static.StringKey("_measurement", "m0"),
		static.TableList{
			static.StringKeys("t0", "a", nil, "b"),
		},
	}
	if err := input.Do(func(tbl flux.Table) error {
		return tr.Process(parentID, tbl)
	}); err != nil {
		t.Fatal(err)
	}
	tr.Finish(parentID, nil)

	want := `m0,t0=a f0=1,f1=1i 0
m0,t0=a f0=2 10000000000
m0,t0=a f0=3,f1=3i 20000000000
m0 f0=1,f1=1i 0
m0 f0=2 10000000000
m0 f0=3,f1=3i 20000000000
m0,t0=b f0=1,f1=1i 0
m0,t0=b f0=2 10000000000
m0,t0=b f0=3,f1=3i 20000000000
`
	if got := written.String(); got != want {
		t.Errorf("unexpected line protocol -want/+got:\n%s", diff.LineDiff(want, got))
	}
}

func TestWideToTransformation_Errors(t *testing.T) {
	for _, tc := range []struct {
		name  string
		input flux.TableIterator
		want  string
	}{
		{
			name: "FieldKeyPresent",
			input: static.Table{
				static.Times("_time", 0, 10, 20),
				static.StringKey("_measurement", "m0"),
				static.StringKey("_field", "f0"),
				static.Floats("_value", 1.0, 2.0, 3.0),
			},
			want: `found column "_field" in the group key; wideTo() expects pivoted data`,
		},
		{
			name: "MeasurementWrongType",
			input: static.Table{
				static.Times("_time", 0, 10, 20),
				static.IntKey("_measurement", 0),
				static.Floats("f0", 1.0, 2.0, 3.0),
			},
			want: `group key column "_measurement" has type int; type string is required`,
		},
		{
			name: "NonStringGroupKey",
			input: static.Table{
				static.Times("_time", 0, 10, 20),
				static.StringKey("_measurement", "m0"),
				static.IntKey("t0", 0),
				static.Floats("f0", 1.0, 2.0, 3.0),
			},
			want: `group key column "t0" has type int; type string is required`,
		},
		{
			name: "MeasurementColumnMissing",
			input: static.Table{
				static.Times("_time", 0, 10, 20),
				static.Floats("f0", 1.0, 2.0, 3.0),
			},
			want: `required column "_measurement" not in group key`,
		},
		{
			name: "MeasurementNotInGroupKey",
			input: static.Table{
				static.Times("_time", 0, 10, 20),
				static.Strings("_measurement", "m0", "m0", "m0"),
				static.Floats("f0", 1.0, 2.0, 3.0),
			},
			want: `required column "_measurement" not in group key`,
		},
		{
			name: "TimeColumnMissing",
			input: static.Table{
				static.StringKey("_measurement", "m0"),
				static.Floats("f0", 1.0, 2.0, 3.0),
			},
			want: `input table is missing required column "_time"`,
		},
		{
			name: "TimeColumnWrongType",
			input: static.Table{
				static.Ints("_time", 0, 10, 20),
				static.StringKey("_measurement", "m0"),
				static.Floats("f0", 1.0, 2.0, 3.0),
			},
			want: `column "_time" has type int; type time is required`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			deps := dependenciestest.Default()
			deps.Deps.Deps.HTTPClient = &http.Client{
				Transport: dependenciestest.RoundTripFunc(func(req *http.Request) *http.Response {
					return &http.Response{
						StatusCode: http.StatusNoContent,
						Status:     http.StatusText(http.StatusNoContent),
						Body:       io.NopCloser(bytes.NewReader(nil)),
						Header:     make(http.Header),
					}
				}),
			}

			cache := execute.NewTableBuilderCache(&memory.ResourceAllocator{})
			d := execute.NewDataset(executetest.RandomDatasetID(), execute.DiscardingMode, cache)
			d.SetTriggerSpec(plan.DefaultTriggerSpec)

			ctx, span := dependency.Inject(context.Background(), deps)
			defer span.Finish()
			spec := &influxdb.WideToProcedureSpec{
				Config: influxdb.Config{
					Bucket: influxdb.NameOrID{Name: "mybucket"},
					Host:   "http://localhost:8086",
				},
			}
			tr, err := influxdb.NewWideToTransformation(ctx, d, cache, spec)
			if err != nil {
				t.Fatal(err)
			}

			parentID := executetest.RandomDatasetID()
			got := tc.input.Do(func(tbl flux.Table) error {
				return tr.Process(parentID, tbl)
			})
			if got == nil {
				t.Fatal("expected error")
			} else if got.Error() != tc.want {
				t.Fatalf("unexpected error -want/+got:\n\t- %s\n\t+ %s", tc.want, got.Error())
			}
		})
	}
}

func TestWideToTransformation_CloseOnError(t *testing.T) {
	var closed bool
	deps := dependenciestest.Default()
	provider := influxdb.Dependency{
		Provider: MockProvider{
			WriterForFn: func(ctx context.Context, conf influxdb.Config) (influxdb.Writer, error) {
				return &MockWriter{
					WriteFn: func(metric ...influxdb.Metric) error {
						return errors.New("expected")
					},
					CloseFn: func() error {
						closed = true
						return nil
					},
				}, nil
			},
		},
	}

	cache := execute.NewTableBuilderCache(&memory.ResourceAllocator{})
	d := execute.NewDataset(executetest.RandomDatasetID(), execute.DiscardingMode, cache)
	d.SetTriggerSpec(plan.DefaultTriggerSpec)

	ctx, span := dependency.Inject(
		context.Background(),
		deps,
		provider,
	)
	defer span.Finish()
	spec := &influxdb.WideToProcedureSpec{
		Config: influxdb.Config{
			Bucket: influxdb.NameOrID{Name: "mybucket"},
			Host:   "http://localhost:8086",
		},
	}
	tr, err := influxdb.NewWideToTransformation(ctx, d, cache, spec)
	if err != nil {
		t.Fatal(err)
	}

	parentID := executetest.RandomDatasetID()
	input := static.TableGroup{
		static.Times("_time", 0, 10, 20, 30),
		static.Floats("f0", 1.0, 2.0, 3.0, nil),
		static.Ints("f1", 1, nil, 3, nil),
		static.StringKey("_measurement", "m0"),
		static.TableList{
			static.StringKeys("t0", "a", "b"),
		},
	}

	err = input.Do(func(tbl flux.Table) error {
		return tr.Process(parentID, tbl)
	})
	if err == nil {
		t.Error("expected error")
	}
	tr.Finish(parentID, err)

	if !closed {
		t.Error("writer was not closed")
	}
}

type MockProvider struct {
	influxdb.UnimplementedProvider
	WriterForFn func(ctx context.Context, conf influxdb.Config) (influxdb.Writer, error)
}

func (p MockProvider) WriterFor(ctx context.Context, conf influxdb.Config) (influxdb.Writer, error) {
	return p.WriterForFn(ctx, conf)
}

type MockWriter struct {
	WriteFn func(metric ...influxdb.Metric) error
	CloseFn func() error
}

func (m *MockWriter) Write(metric ...influxdb.Metric) error {
	return m.WriteFn(metric...)
}

func (m *MockWriter) Close() error {
	return m.CloseFn()
}
