package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/cmd/flux/cmd"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/dependencies/testing"
	"github.com/influxdata/flux/dependency"
	"github.com/influxdata/flux/execute/executetest"
	"github.com/influxdata/flux/execute/table"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/lang"
	"github.com/influxdata/flux/memory"
	"github.com/influxdata/flux/runtime"
)

func NewTestExecutor(ctx context.Context) (cmd.TestExecutor, error) {
	return testExecutor{}, nil
}

type testExecutor struct{}

func (testExecutor) Run(pkg *ast.Package) error {
	jsonAST, err := json.Marshal(pkg)
	if err != nil {
		return err
	}
	c := lang.ASTCompiler{AST: jsonAST}

	ctx, span := dependency.Inject(context.Background(),
		executetest.NewTestExecuteDependencies(),
		testing.FrameworkConfig{},
	)
	defer span.Finish()
	program, err := c.Compile(ctx, runtime.Default)
	if err != nil {
		return errors.Wrap(err, codes.Invalid, "failed to compile")
	}

	alloc := &memory.ResourceAllocator{}
	query, err := program.Start(ctx, alloc)
	if err != nil {
		return errors.Wrap(err, codes.Inherit, "error while executing program")
	}
	defer query.Done()

	var output strings.Builder
	results := flux.NewResultIteratorFromQuery(query)

	foundErrorResult := false
	for results.More() {
		result := results.Next()
		if result.Name() == "error" {
			foundErrorResult = true

			err := result.Tables().Do(func(tbl flux.Table) error {
				// The data returned here is the result of `testing.diff`, so any result means that
				// a comparison of two tables showed inequality. Capture that inequality as part of the error.
				// XXX: rockstar (08 Dec 2020) - This could use some ergonomic work, as the diff output
				// is not exactly "human readable."
				_, _ = fmt.Fprint(&output, table.Stringify(tbl))
				return nil
			})
			if err != nil {
				return err
			}
		} else {
			err := result.Tables().Do(func(tbl flux.Table) error {
				return nil
			})
			if err != nil {
				return err
			}
		}
	}
	results.Release()

	err = results.Err()
	if err == nil {
		if output.Len() > 0 {
			err = errors.Newf(codes.FailedPrecondition, "%s", output.String())
		} else if !foundErrorResult {
			err = errors.Newf(codes.FailedPrecondition, "`yield(name: \"error\")` was never called. Did you forget to add an assertion to the test?")
		}
	}

	return err
}

func (testExecutor) Close() error { return nil }
