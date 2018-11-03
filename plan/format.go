package plan

import "fmt"

type FormatOption func(*formatter)

// TODO(cwolff): enhance the this output to make it more useful
func Formatted(p *PlanSpec, opts ...FormatOption) fmt.Formatter {
	f := formatter{
		p: p,
	}
	for _, o := range opts {
		o(&f)
	}
	return f
}

type formatter struct {
	p *PlanSpec
}

func (f formatter) Format(fs fmt.State, c rune) {
	fmt.Fprintf(fs, "\ndigraph {\n")
	f.p.BottomUpWalk(func(pn PlanNode) error {
		fmt.Fprintf(fs, "  %v\n", pn.ID())
		for _, pred := range pn.Predecessors() {
			fmt.Fprintf(fs, "  %v -> %v\n", pred.ID(), pn.ID())
		}
		return nil
	})
	fmt.Fprintf(fs, "}\n")
}
