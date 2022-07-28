package join

import (
	"context"
	"fmt"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/interpreter"
	"github.com/influxdata/flux/plan"
	"github.com/influxdata/flux/semantic"
)

const EquiJoinKind = "equijoin"

func init() {
	plan.RegisterPhysicalRules(EquiJoinPredicateRule{})
	execute.RegisterTransformation(EquiJoinKind, createJoinTransformation)
}

type ColumnPair struct {
	Left, Right string
}

type EquiJoinProcedureSpec struct {
	On     []ColumnPair
	As     interpreter.ResolvedFunction
	Left   *flux.TableObject
	Right  *flux.TableObject
	Method string
}

func (p *EquiJoinProcedureSpec) Kind() plan.ProcedureKind {
	return plan.ProcedureKind(EquiJoinKind)
}

func (p *EquiJoinProcedureSpec) Copy() plan.ProcedureSpec {
	return &EquiJoinProcedureSpec{
		On:     p.On,
		As:     p.As,
		Left:   p.Left,
		Right:  p.Right,
		Method: p.Method,
	}
}

func (p *EquiJoinProcedureSpec) Cost(inStats []plan.Statistics) (cost plan.Cost, outStats plan.Statistics) {
	return plan.Cost{}, plan.Statistics{}
}

func newEquiJoinProcedureSpec(spec *JoinProcedureSpec, cols []ColumnPair) *EquiJoinProcedureSpec {
	return &EquiJoinProcedureSpec{
		On:     cols,
		As:     spec.As,
		Left:   spec.Left,
		Right:  spec.Right,
		Method: spec.Method,
	}
}

type EquiJoinPredicateRule struct{}

func (EquiJoinPredicateRule) Name() string {
	return "equiJoinPredicate"
}

func (EquiJoinPredicateRule) Pattern() plan.Pattern {
	return plan.MultiSuccessor(Join2Kind, plan.AnyMultiSuccessor(), plan.AnyMultiSuccessor())
}

func (EquiJoinPredicateRule) Rewrite(ctx context.Context, n plan.Node) (plan.Node, bool, error) {
	s := n.ProcedureSpec()
	spec, ok := s.(*JoinProcedureSpec)
	if !ok {
		return nil, false, errors.New(codes.Internal, "invalid spec type on join node")
	}

	fnBody := spec.On.Fn.Block.Body

	if len(fnBody) != 1 {
		return nil, false, wrapErr(
			codes.Invalid,
			"function body should be a single logical expression that compares columns from each table",
		)
	}
	rs, ok := fnBody[0].(*semantic.ReturnStatement)
	if !ok {
		return nil, false, wrapErr(
			codes.Invalid,
			"function body should be a single logical expression that compares columns from each table",
		)
	}

	cols := []ColumnPair{}
	expr := rs.Argument
	var walkErr error
	semantic.Walk(semantic.CreateVisitor(func(n semantic.Node) {
		switch e := n.(type) {
		case *semantic.LogicalExpression:
			if e.Operator != ast.AndOperator {
				walkErr = wrapErr(
					codes.Invalid,
					fmt.Sprintf("unsupported operator in join predicate: %s", e.Operator.String()),
				)

				return
			}
		case *semantic.BinaryExpression:
			if e.Operator != ast.EqualOperator {
				walkErr = wrapErr(
					codes.Invalid,
					fmt.Sprintf("unsupported operator in join predicate: %s", e.Operator.String()),
				)
				return
			}

			lhs, ok := e.Left.(*semantic.MemberExpression)
			if !ok {
				walkErr = wrapErr(codes.Invalid, "left side of comparison is not a member expression")
				return
			}
			rhs, ok := e.Right.(*semantic.MemberExpression)
			if !ok {
				walkErr = wrapErr(codes.Invalid, "right side of comparison is not a member expression")
				return
			}
			lob, err := getObjectName(lhs)
			if err != nil {
				walkErr = err
				return
			}
			rob, err := getObjectName(rhs)
			if err != nil {
				walkErr = err
				return
			}

			// Each side of the binary expression should reference either the `l` or `r` object,
			// but they should not reference the same object.
			if !((lob == "l") != (rob == "l") && (lob == "r") != (rob == "r")) {
				walkErr = wrapErr(
					codes.Invalid,
					"binary expression operands must reference `l` or `r` only, and may not reference the same object",
				)
				return
			}

			lcol := lhs.Property.LocalName
			rcol := rhs.Property.LocalName
			pair := ColumnPair{}

			if lob == "l" {
				pair.Left = lcol
				pair.Right = rcol
			} else {
				pair.Left = rcol
				pair.Right = lcol
			}
			cols = append(cols, pair)
		case *semantic.MemberExpression:
		case *semantic.IdentifierExpression:
		default:
			walkErr = wrapErr(
				codes.Invalid,
				fmt.Sprintf("illegal expression type in join predicate: %s", e.NodeType()),
			)
		}
	}), expr)
	if walkErr != nil {
		return nil, false, walkErr
	}
	n.ReplaceSpec(newEquiJoinProcedureSpec(spec, cols))
	return n, true, nil
}

func wrapErr(code codes.Code, msg string) error {
	return errors.Newf(code, fmt.Sprintf("error in join function - some expressions are not yet supported in the `on` parameter: %s", msg))
}

func getObjectName(me *semantic.MemberExpression) (string, error) {
	id, ok := me.Object.(*semantic.IdentifierExpression)
	if !ok {
		return "", errors.New(codes.Internal, "invalid member expression")
	}
	name := id.Name.LocalName
	return name, nil
}
