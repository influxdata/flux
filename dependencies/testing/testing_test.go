package testing

import (
	"context"
	"testing"
)

func MustExpectPlannerRule(ctx context.Context, name string, n int) {
	if err := ExpectPlannerRule(ctx, name, n); err != nil {
		panic(err)
	}
}

func TestExpectPlannerRule(t *testing.T) {
	for _, tt := range []struct {
		name    string
		fn      func(ctx context.Context)
		wantErr string
	}{
		{
			name: "single rule",
			fn: func(ctx context.Context) {
				MustExpectPlannerRule(ctx, "A", 1)
				MarkInvokedPlannerRule(ctx, "A")
			},
		},
		{
			name: "multiple rules, single expect",
			fn: func(ctx context.Context) {
				MustExpectPlannerRule(ctx, "A", 1)
				MarkInvokedPlannerRule(ctx, "A")
				MarkInvokedPlannerRule(ctx, "B")
			},
		},
		{
			name: "single rule, multiple times",
			fn: func(ctx context.Context) {
				MustExpectPlannerRule(ctx, "A", 2)
				MarkInvokedPlannerRule(ctx, "A")
				MarkInvokedPlannerRule(ctx, "A")
			},
		},
		{
			name: "multiple rules, multiple expectations",
			fn: func(ctx context.Context) {
				MustExpectPlannerRule(ctx, "A", 2)
				MustExpectPlannerRule(ctx, "B", 1)
				MarkInvokedPlannerRule(ctx, "A")
				MarkInvokedPlannerRule(ctx, "A")
				MarkInvokedPlannerRule(ctx, "B")
			},
		},
		{
			name: "expect no planner rule",
			fn: func(ctx context.Context) {
				MustExpectPlannerRule(ctx, "A", 0)
			},
		},
		{
			name: "single rule, wrong number of times",
			fn: func(ctx context.Context) {
				MustExpectPlannerRule(ctx, "A", 3)
				MarkInvokedPlannerRule(ctx, "A")
				MarkInvokedPlannerRule(ctx, "A")
			},
			wantErr: "planner rule invoked an unexpected number of times: 3 (want) != 2 (got)",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx := Inject(context.Background())
			tt.fn(ctx)

			got := Check(ctx)
			if got != nil {
				if gotErr, wantErr := got.Error(), tt.wantErr; gotErr != wantErr {
					t.Errorf("unexpected error -want/+got:\n\t- %s\n\t+ %s", wantErr, gotErr)
				}
			} else if tt.wantErr != "" {
				t.Error("expected error")
			}
		})
	}
}

func TestNoTestingFramework_ExpectPlannerRule(t *testing.T) {
	// An error should happen if we call expect planner
	// rule without a testing framework.
	ctx := context.Background()
	if err := ExpectPlannerRule(ctx, "A", 1); err == nil {
		t.Error("expected error")
	}

	// Marking a planner rule as invoked does not panic or anything.
	MarkInvokedPlannerRule(ctx, "A")

	// It also doesn't affect expect planner rule.
	if err := ExpectPlannerRule(ctx, "A", 1); err == nil {
		t.Error("expected error")
	}

	// Check does not return an error.
	if err := Check(ctx); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}
