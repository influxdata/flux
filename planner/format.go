package planner

import "fmt"

type FormatOption func(*formatter)

// TODO: make this actually do something useful
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
	fmt.Fprintf(fs, "%#v", f.p)
}
