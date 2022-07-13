package plan

import (
	"fmt"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
)

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
	String() string
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

// CheckRequiredAttributes will check that if the given node requires any
// attributes from its predecessors, then they are provided, either directly or
// because a predecessor passes on the attribute from one of its own predecessors.
// If all requirements are met, nil is returned, otherwise an error
// wil an appopriate diagnostic is produced.
func CheckRequiredAttributes(node *PhysicalPlanNode) error {

	// If there are any required attributes for this node, there should be one set of
	// required attributes for each input.
	reqAttrsSlice := node.requiredAttrs() // one set of required attributes for each predecessor
	if lra, lpred := len(reqAttrsSlice), len(node.Predecessors()); lra != lpred {
		return &flux.Error{
			Code: codes.Internal,
			Msg:  fmt.Sprintf("node has %d predecessors but has %d sets of required attributes", lpred, lra),
		}
	}

	for i, reqAttrMap := range reqAttrsSlice {
		for _, reqAttr := range reqAttrMap {
			ppred := node.Predecessors()[i].(*PhysicalPlanNode)
			haveAttr, n := getOutputAttributeWithNode(node.Predecessors()[i].(*PhysicalPlanNode), reqAttr.Key())
			if haveAttr == nil {
				return errors.Newf(codes.Internal,
					"attribute %q, required by %q, is missing from predecessor %q",
					reqAttr.Key(), node.ID(), n.ID(),
				)
			}

			if !reqAttr.SatisfiedBy(haveAttr) {
				return errors.Newf(codes.Internal,
					"node %q requires attribute %v, which is not satisfied by predecessor %q, "+
						"which has attribute %v",
					node.ID(), reqAttr, ppred.ID(), haveAttr,
				)
			}
		}
	}

	return nil
}

// GetOutputAttribute will return the attribute with the given key
// provided by the given plan node, traversing backwards through predecessors
// as needed for attributes that may pass through. E.g.,
//   sort |> filter
// The "filter" node will still provide the collation attribute, even though it's
// the "sort" node that actually does the collating.
func GetOutputAttribute(node *PhysicalPlanNode, attrKey string) PhysicalAttr {
	attr, _ := getOutputAttributeWithNode(node, attrKey)
	return attr
}

func getOutputAttributeWithNode(node *PhysicalPlanNode, attrKey string) (PhysicalAttr, *PhysicalPlanNode) {
	if attr, ok := node.outputAttrs()[attrKey]; ok {
		return attr, nil
	}

	if node.passesThroughAttr(attrKey) && len(node.Predecessors()) == 1 {
		// TODO(cwolff): consider what it means for nodes with multiple predecessors
		//   (e.g. join or union) to pass on attributes.
		return getOutputAttributeWithNode(node.Predecessors()[0].(*PhysicalPlanNode), attrKey)
	}

	return nil, node
}

// CheckSuccessorsMustRequire will return an error if the node has an output attribute
// that must be required by a successor, and nil otherwise.. E.g., the parallel-run
// attribute is like this in that it must be required by a merge node.
// This function will walk forward through successors to find the requiring node.
func CheckSuccessorsMustRequire(node *PhysicalPlanNode) error {
	for _, attr := range node.outputAttrs() {
		if !attr.SuccessorsMustRequire() {
			continue
		}

		if len(node.Successors()) == 0 {
			return &flux.Error{
				Code: codes.Internal,
				Msg: fmt.Sprintf("node %q provides attribute %v that must be required but has no "+
					"successors to require it", node.ID(), attr.Key()),
			}
		}

		for _, succ := range node.Successors() {
			reqd, n := requiredBySuccessor(attr, node, succ.(*PhysicalPlanNode))
			if reqd {
				continue
			}

			if n != nil {
				return &flux.Error{
					Code: codes.Internal,
					Msg: fmt.Sprintf("plan node %q has attribute %q that must be required by successors, "+
						"but it is not required or propagated by successor %q",
						node.ID(), attr.Key(), n.ID(),
					),
				}
			}

			return &flux.Error{
				Code: codes.Internal,
				Msg: fmt.Sprintf("plan node %q has attribute %q that must be required by successors, "+
					"but no successors require it",
					node.ID(), attr.Key(),
				),
			}
		}
	}

	return nil
}

// requiredBySuccessor returns true if the given attribute is required by succ or
// succ passes through the attribute and one of its successors requires the attribute.
// If the attribute is not required, this function returns false and the node that neither passes
// along nor requires the attribute.
func requiredBySuccessor(requiredAttr PhysicalAttr, node, succ *PhysicalPlanNode) (bool, *PhysicalPlanNode) {
	i := indexOfNode(node, succ.Predecessors())
	if _, ok := succ.requiredAttrs()[i][requiredAttr.Key()]; ok {
		return true, succ
	}
	if succ.passesThroughAttr(requiredAttr.Key()) {
		if len(succ.Successors()) == 0 {
			return false, nil
		}
		// If this node does not require the attribute itself but passes it along,
		// see if any successors require it.
		for _, ssucc := range succ.Successors() {
			if reqd, n := requiredBySuccessor(requiredAttr, succ, ssucc.(*PhysicalPlanNode)); !reqd {
				return false, n
			}
		}
		return true, succ
	}
	return false, succ
}
