package plan

import (
	"fmt"
	"runtime/debug"
	"strings"
)

type FormatOption func(*formatter)

// Formatted accepts a plan.Spec and options, and returns a Formatter
// that can be used with the standard fmt package, e.g.,
//
//	fmt.Println(Formatted(plan, WithDetails())
func Formatted(p *Spec, opts ...FormatOption) fmt.Formatter {
	f := formatter{
		p: p,
	}
	for _, o := range opts {
		o(&f)
	}
	return f
}

// WithDetails returns a FormatOption that can be used to provide extra details
// in a formatted plan.
func WithDetails() FormatOption {
	return func(f *formatter) {
		f.withDetails = true
	}
}

// Detailer provides an optional interface that ProcedureSpecs can implement.
// Implementors of this interface will have their details appear in the
// formatted output for a plan if the WithDetails() option is set.
type Detailer interface {
	PlanDetails() string
}

type formatter struct {
	withDetails bool
	p           *Spec
}

func formatAsDOT(id NodeID) string {
	// The DOT language does not allow "." or "/" in IDs
	// so quote the node IDs.
	return fmt.Sprintf("%q", id)
}

func (f formatter) Format(fs fmt.State, c rune) {
	// Panicking while producing debug output is frustrating, so catch any panics and
	// continue if that happens.
	defer func() {
		if e := recover(); e != nil {
			_, _ = fmt.Fprintf(fs, "panic while formatting plan: %v\n", e)
			_, _ = fmt.Fprintf(fs, "stack: %s\n", string(debug.Stack()))
		}
	}()

	_, _ = fmt.Fprintf(fs, "digraph {\n")
	var edges []string
	_ = f.p.BottomUpWalk(func(pn Node) error {
		_, _ = fmt.Fprintf(fs, "  %v\n", formatAsDOT(pn.ID()))
		if f.withDetails {
			details := ""
			if d, ok := pn.ProcedureSpec().(Detailer); ok {
				details += d.PlanDetails() + "\n"
			}

			if ppn, ok := pn.(*PhysicalPlanNode); ok {
				for _, attr := range ppn.outputAttrs() {
					if d, ok := attr.(Detailer); ok {
						details += d.PlanDetails() + "\n"
					}
				}
			}

			lines := strings.Split(strings.TrimSpace(details), "\n")

			for _, line := range lines {
				if len(line) > 0 {
					_, _ = fmt.Fprintf(fs, "  // %s\n", line)
				}
			}
		}
		for _, pred := range pn.Predecessors() {
			edges = append(edges, fmt.Sprintf("  %v -> %v", formatAsDOT(pred.ID()), formatAsDOT(pn.ID())))
		}
		return nil
	})

	_, _ = fmt.Fprintf(fs, "\n")
	for _, e := range edges {
		_, _ = fmt.Fprintf(fs, "%v\n", e)
	}
	_, _ = fmt.Fprintf(fs, "}\n")
}
