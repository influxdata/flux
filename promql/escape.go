package promql

import (
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/prometheus/prometheus/promql"
)

func escapeLabelName(ln string) string {
	switch {
	case ln == "":
		// This can occur in parameters to functions (e.g. label_replace() empty "src" parameter).
		return ""
	case ln == "__name__":
		return "_field"
	case ln[0] == '_' || ln[0] == '~':
		return "~" + ln
	default:
		return ln
	}
}

func UnescapeLabelName(ln string) string {
	switch {
	case ln == "_field":
		return "__name__"
	case ln[0] == '~':
		return ln[1:]
	default:
		return ln
	}
}

func escapeLabelNames(in []string) []string {
	out := make([]string, len(in))
	for i, ln := range in {
		out[i] = escapeLabelName(ln)
	}
	return out
}

func escapeLabelMatchers(in []*labels.Matcher) []*labels.Matcher {
	out := make([]*labels.Matcher, len(in))
	var err error
	for i, m := range in {
		out[i], err = labels.NewMatcher(m.Type, escapeLabelName(m.Name), m.Value)
		if err != nil {
			panic("unable to create escaped label matcher")
		}
	}
	return out
}

type labelNameEscaper struct{}

func (s labelNameEscaper) Visit(node promql.Node, path []promql.Node) (promql.Visitor, error) {
	switch n := node.(type) {
	case *promql.AggregateExpr:
		n.Grouping = escapeLabelNames(n.Grouping)
	case *promql.BinaryExpr:
		if n.VectorMatching != nil {
			n.VectorMatching.MatchingLabels = escapeLabelNames(n.VectorMatching.MatchingLabels)
			n.VectorMatching.Include = escapeLabelNames(n.VectorMatching.Include)
		}
	case *promql.Call:
		// Nothing to do here - there are only two functions that take label names
		// as string parameters (label_replace() and label_join()), and those handle
		// escaping by themselves.
	case *promql.MatrixSelector:
		n.Name = ""
		n.LabelMatchers = escapeLabelMatchers(n.LabelMatchers)
	case *promql.VectorSelector:
		n.Name = ""
		n.LabelMatchers = escapeLabelMatchers(n.LabelMatchers)
	}
	return s, nil
}
