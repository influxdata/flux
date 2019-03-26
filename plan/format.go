package plan

import "fmt"

type FormatOption func(*formatter)

// TODO(cwolff): enhance the this output to make it more useful
func Formatted(p *Spec, opts ...FormatOption) fmt.Formatter {
	f := formatter{
		p: p,
	}
	for _, o := range opts {
		o(&f)
	}
	return f
}

type formatter struct {
	p *Spec
}

func (f formatter) Format(fs fmt.State, c rune) {
	fmt.Fprintf(fs, "\ndigraph {\n")
	var edges []string
	f.p.BottomUpWalk(func(pn Node) error {
		fmt.Fprintf(fs, "  %v\n", pn.ID())
		for _, pred := range pn.Predecessors() {
			edges = append(edges, fmt.Sprintf("  %v -> %v", pred.ID(), pn.ID()))
		}
		return nil
	})

	fmt.Fprintf(fs, "\n")
	for _, e := range edges {
		fmt.Fprintf(fs, "%v\n", e)
	}
	fmt.Fprintf(fs, "}\n")
}
