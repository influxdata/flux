package main

import (
	"context"
	"encoding/json"

	"github.com/InfluxCommunity/flux"
	"github.com/InfluxCommunity/flux/ast"
	"github.com/InfluxCommunity/flux/cmd/flux/cmd"
	"github.com/InfluxCommunity/flux/codes"
	"github.com/InfluxCommunity/flux/dependencies/testing"
	"github.com/InfluxCommunity/flux/dependency"
	"github.com/InfluxCommunity/flux/execute/executetest"
	"github.com/InfluxCommunity/flux/internal/errors"
	"github.com/InfluxCommunity/flux/lang"
	"github.com/InfluxCommunity/flux/memory"
	"github.com/InfluxCommunity/flux/runtime"
)

func NewTestExecutor(ctx context.Context) (cmd.TestExecutor, error) {
	return testExecutor{}, nil
}

type testExecutor struct{}

func (testExecutor) Run(pkg *ast.Package, fn cmd.TestResultFunc) error {
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

	results := flux.NewResultIteratorFromQuery(query)
	defer results.Release()

	return fn(ctx, results)
}

func (testExecutor) Close() error { return nil }
