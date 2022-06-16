package plan

// Attributes provide a way to model different aspects of the data
// flowing though the physical operations in a plan graph.
//
// For example, if a node requires its input tables to be sorted, it
// will have the CollationAttr among its required attributes.
// Likewise, if a node produces sorted tables, its output attributes
// will have CollationAttr.
//
// Operations can require or provide attributes by implementing interfaces
// in the corresponding PhysicalProcedureSpec:
// - OutputAttributer is to be implemented if the procedure will provide an attribiute. E.g.,
//   the SortProcedureSpec will provide CollationAttr for the columns on which the data is
//   to be sorted.
// - PassThroughAttributeris to be implemented if a procedure will not perturb a given attribute.
//   E.g., if data with CollationAttr flows into a filter, it will still be sorted. Therefore,
//   FilterProcedureSpec can implement PassThroughAttributer for collation.
// - RequiredAttributer is to be implemented by procedures that require particular attreibutes.
//   There is one set of required physical attributes for each input, since they may be different.
//   E.g., SortMergeJoinProcedureSpec requires that the left and right inputs be sorted on the
//   columns being joined on (and they may in fact have different names on either side).
//
// It's the obligation of planner rules to ensure that required attributes are satisified by
// a procedure's inputs. If a node has required attribute that are not satisfied, it will be
// caught by ValidatePhysicalPlan(), and an internal error will be returned.

// PhysicalAttr represents an attribute (collation, parallel execution)
// of a plan node.
type PhysicalAttr interface {
	Key() string
	SuccessorsMustRequire() bool
	SatisfiedBy(attr PhysicalAttr) bool
}

// PhysicalAttributes encapsulates any physical attributes of the result produced
// by a physical plan node, such as collation, etc.
type PhysicalAttributes map[string]PhysicalAttr

// OutputAttributer is an interface to be implemented by PhysicalProcedureSpec implementations
// that produce output that has particular attributes.
type OutputAttributer interface {
	OutputAttributes() PhysicalAttributes
}

// PassThroughAttributer is an interface to be implemented by PhysicalProcedureSpec implementations
// that allow attributes to propagate from input to output.
type PassThroughAttributer interface {
	PassThroughAttribute(attrKey string) bool
}

// RequiredAttributer is an interface to be implemented by PhysicalProcedureSpec implementations
// that require physical attributes to be provided by inputs. The return value here is a slice,
// since each input may be required to have a different set of attributes.
type RequiredAttributer interface {
	RequiredAttributes() []PhysicalAttributes
}

// NodeSatisfiesRequiredAttribute returns true if the given node can provide the given attribute.
// An attribute can be provided if the node provides it directly, or if the node passes through
// an attreibute from a predecessor.
func NodeSatisfiesRequiredAttribute(node *PhysicalPlanNode, requiredAttr PhysicalAttr) bool {
	gotAttr := getAttribute(node, requiredAttr.Key())
	return requiredAttr.SatisfiedBy(gotAttr)
}

func getAttribute(node *PhysicalPlanNode, attrKey string) PhysicalAttr {
	if attr, ok := node.OutputAttrs()[attrKey]; ok {
		return attr
	}

	if passer, ok := node.Spec.(PassThroughAttributer); ok {
		if passer.PassThroughAttribute(attrKey) && len(node.Predecessors()) == 1 {
			// TODO(cwolff): consider what it means for nodes with multiple predecessors
			//   (e.g. join or union) to pass on attributes.
			return getAttribute(node.Predecessors()[0].(*PhysicalPlanNode), attrKey)
		}
	}

	return nil
}
