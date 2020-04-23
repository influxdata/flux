package main

//
// 1. cd to package that contains the program to analyze
// 2. call with paths to command mains
//
// For example, to analyze influx and influxd in the influxdb package:
//   cd ~/devel/influxdb
//   ../flux/calls ./cmd/influx ./cmd/influxd > calls.txt
//
// The output is a list of flux functions with non-flux call points indented
// below. Each function and call point is followed by the file location. The
// number preceding each function name is an identifier assigned by the
// analysis.
//

import (
	"flag"
	"fmt"
	"go/token"
	"os"
	"sort"
	"strings"

	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/callgraph/rta"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

func inFlux(path string) bool {
	return path == "github.com/influxdata/flux" ||
		strings.HasPrefix(path, "github.com/influxdata/flux/")
}

func shorten(targ string) string {
	return strings.Replace(targ, "github.com/influxdata/", "_/", 1)
}

func nameForCmp(node *callgraph.Node) string {
	return node.Func.Package().Pkg.Path() + node.Func.Name()
}

func walkGraph(fset *token.FileSet, cg *callgraph.Graph) {
	fluxNodes := []*callgraph.Node{}

	// Iterate all functions.
	for _, n := range cg.Nodes {
		// Skip init functions.
		if n.Func.Name() == "init" || strings.HasPrefix(n.Func.Name(), "init#") {
			continue
		}

		// Skip anonymous functions.
		if n.Func.Parent() != nil {
			continue
		}

		// Skip methods (declared or wrapper)
		if n.Func.Signature.Recv() != nil {
			continue
		}

		// Consider functions in the flux package.
		path := n.Func.Package().Pkg.Path()
		if inFlux(path) {
			fluxNodes = append(fluxNodes, n)
		}
	}

	// Sorted order.
	sort.SliceStable(fluxNodes, func(i, j int) bool {
		return nameForCmp(fluxNodes[i]) < nameForCmp(fluxNodes[j])
	})

	for _, n := range fluxNodes {
		// Collect the calls from outside flux
		nonFluxEdges := []*callgraph.Edge{}
		for _, e := range n.In {
			if !inFlux(e.Caller.Func.Package().Pkg.Path()) {
				nonFluxEdges = append(nonFluxEdges, e)
			}
		}

		// Sorted order.
		sort.SliceStable(nonFluxEdges, func(i, j int) bool {
			return nameForCmp(nonFluxEdges[i].Caller) <
				nameForCmp(nonFluxEdges[j].Caller)
		})

		// Display them.
		if len(nonFluxEdges) > 0 {
			// The called function.
			fmt.Printf("%s  %s\n", shorten(n.String()),
				fset.Position(n.Func.Pos()))

			// The call points.
			for _, e := range nonFluxEdges {
				position := fset.Position(e.Site.Common().Pos())
				fmt.Printf("  %s  %s\n", shorten(e.Caller.String()),
					position.String())
			}
		}
	}
}

func showFluxUsage(args []string) error {
	cfg := &packages.Config{
		Mode: packages.NeedName |
			packages.NeedFiles |
			packages.NeedCompiledGoFiles |
			packages.NeedImports |
			packages.NeedDeps |
			packages.NeedExportsFile |
			packages.NeedTypes |
			packages.NeedSyntax |
			packages.NeedTypesInfo |
			packages.NeedTypesSizes,
		Tests: false,
		Dir:   "",
	}

	initial, err := packages.Load(cfg, args...)
	if err != nil {
		return err
	}
	if packages.PrintErrors(initial) > 0 {
		return fmt.Errorf("packages contain errors")
	}

	// Create and build SSA-form program representation.
	prog, pkgs := ssautil.AllPackages(initial, 0)
	prog.Build()

	// Verify we were supplied main packages
	for _, p := range pkgs {
		if p.Pkg.Name() != "main" || p.Func("main") == nil {
			return fmt.Errorf("supplied non-main package")
		}
	}

	// Use all mains and their corresponding inits as roots.
	var roots []*ssa.Function
	for _, main := range pkgs {
		roots = append(roots, main.Func("init"), main.Func("main"))
	}

	// Run the analysis and pull the callgraph.
	var cg *callgraph.Graph
	rtares := rta.Analyze(roots, true)
	cg = rtares.CallGraph
	cg.DeleteSyntheticNodes()

	walkGraph(prog.Fset, cg)

	return nil
}

func main() {
	flag.Parse()

	if err := showFluxUsage(flag.Args()); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}
