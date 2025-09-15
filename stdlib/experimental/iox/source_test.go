package iox

import (
	"testing"

	stdarrow "github.com/apache/arrow-go/v18/arrow"
	"github.com/google/go-cmp/cmp"
	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
)

func TestCreateSchema(t *testing.T) {
	tests := []struct {
		name    string
		fields  []stdarrow.Field
		wantErr bool
		errCode codes.Code
		errMsg  string
	}{
		{
			name: "no duplicates",
			fields: []stdarrow.Field{
				{Name: "id", Type: &stdarrow.Int64Type{}},
				{Name: "name", Type: &stdarrow.StringType{}},
				{Name: "value", Type: &stdarrow.Float64Type{}},
				{Name: "active", Type: &stdarrow.BooleanType{}},
				{Name: "time", Type: stdarrow.FixedWidthTypes.Timestamp_ns},
			},
			wantErr: false,
		},
		{
			name: "duplicate field names",
			fields: []stdarrow.Field{
				{Name: "id", Type: &stdarrow.Int64Type{}},
				{Name: "value", Type: &stdarrow.Float64Type{}},
				{Name: "id", Type: &stdarrow.Int64Type{}}, // duplicate
			},
			wantErr: true,
			errCode: codes.Invalid,
			errMsg:  "duplicate field name 'id' in schema",
		},
		{
			name: "multiple duplicates",
			fields: []stdarrow.Field{
				{Name: "id", Type: &stdarrow.Int64Type{}},
				{Name: "name", Type: &stdarrow.StringType{}},
				{Name: "id", Type: &stdarrow.Int64Type{}},    // first duplicate
				{Name: "name", Type: &stdarrow.StringType{}}, // second duplicate
			},
			wantErr: true,
			errCode: codes.Invalid,
			errMsg:  "duplicate field name 'id' in schema", // should catch first duplicate
		},
		{
			name: "case sensitive field names",
			fields: []stdarrow.Field{
				{Name: "ID", Type: &stdarrow.Int64Type{}},
				{Name: "id", Type: &stdarrow.Int64Type{}},
				{Name: "Id", Type: &stdarrow.Int64Type{}},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := stdarrow.NewSchema(tt.fields, nil)

			source := &sqlSource{}

			cols, err := source.createSchema(schema)

			if tt.wantErr {
				if err == nil {
					t.Fatal("createSchema() expected error but got none")
				}

				// Check error code
				if e, ok := err.(*errors.Error); ok {
					if e.Code != tt.errCode {
						t.Fatalf("createSchema() error code = %v, want %v", e.Code, tt.errCode)
					}
					if e.Msg != tt.errMsg {
						t.Fatalf("createSchema() error message = %q, want %q", e.Msg, tt.errMsg)
					}
				} else {
					t.Fatalf("createSchema() error type = %T, want *errors.Error", err)
				}
			} else {
				if err != nil {
					t.Fatalf("createSchema() unexpected error: %v", err)
				}

				// Verify columns were created correctly
				if len(cols) != len(tt.fields) {
					t.Fatalf("createSchema() returned %d columns, want %d", len(cols), len(tt.fields))
				}

				wantCols := make([]flux.ColMeta, len(tt.fields))
				for i, field := range tt.fields {
					wantCols[i].Label = field.Name
					switch field.Type.ID() {
					case stdarrow.INT64:
						wantCols[i].Type = flux.TInt
					case stdarrow.FLOAT64:
						wantCols[i].Type = flux.TFloat
					case stdarrow.STRING:
						wantCols[i].Type = flux.TString
					case stdarrow.BOOL:
						wantCols[i].Type = flux.TBool
					case stdarrow.TIMESTAMP:
						wantCols[i].Type = flux.TTime
					}
				}

				if !cmp.Equal(wantCols, cols) {
					t.Fatalf("createSchema() columns -want/+got\n%s", cmp.Diff(wantCols, cols))
				}
			}
		})
	}
}

func TestCreateSchema_UnsupportedType(t *testing.T) {
	// Test unsupported arrow type
	fields := []stdarrow.Field{
		{Name: "id", Type: &stdarrow.Int64Type{}},
		{Name: "data", Type: &stdarrow.Float16Type{}}, // unsupported type
	}

	schema := stdarrow.NewSchema(fields, nil)
	source := &sqlSource{}

	_, err := source.createSchema(schema)
	if err == nil {
		t.Fatal("createSchema() expected error for unsupported type but got none")
	}

	e, ok := err.(*errors.Error)
	if !ok {
		t.Fatalf("createSchema() error type = %T, want *errors.Error", err)
	}
	if e.Code != codes.Internal {
		t.Fatalf("createSchema() error code = %v, want %v", e.Code, codes.Internal)
	}
}
