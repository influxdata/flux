package ast

import (
	"fmt"

	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/influxdata/flux/ast/internal/fbast"
)

type stmtGetterFn func(obj *fbast.WrappedStatement, j int) bool

func statementArrayFromBuf(len int, g stmtGetterFn, arrName string) ([]Statement, []Error) {
	s := make([]Statement, len)
	err := make([]Error, 0)
	for i := 0; i < len; i++ {
		fbs := new(fbast.WrappedStatement)
		t := new(flatbuffers.Table)
		if !g(fbs, i) || !fbs.Statement(t) || fbs.StatementType() == fbast.StatementNONE {
			err = append(err, Error{fmt.Sprintf("Encountered error in deserializing %s[%d]", arrName, i)})
		} else {
			s[i] = statementFromBuf(t, fbs.StatementType())
		}
	}
	return s, err
}

func statementFromBuf(t *flatbuffers.Table, stype fbast.Statement) Statement {
	switch stype {
	case fbast.StatementBadStatement:
		b := new(fbast.BadStatement)
		b.Init(t.Bytes, t.Pos)
		return BadStatement{}.FromBuf(b)
	case fbast.StatementVariableAssignment:
		a := new(fbast.VariableAssignment)
		a.Init(t.Bytes, t.Pos)
		return VariableAssignment{}.FromBuf(a)
	case fbast.StatementMemberAssignment:
		a := new(fbast.MemberAssignment)
		a.Init(t.Bytes, t.Pos)
		return MemberAssignment{}.FromBuf(a)
	case fbast.StatementExpressionStatement:
		e := new(fbast.ExpressionStatement)
		e.Init(t.Bytes, t.Pos)
		return ExpressionStatement{}.FromBuf(e)
	case fbast.StatementReturnStatement:
		r := new(fbast.ReturnStatement)
		r.Init(t.Bytes, t.Pos)
		return ReturnStatement{}.FromBuf(r)
	case fbast.StatementOptionStatement:
		o := new(fbast.OptionStatement)
		o.Init(t.Bytes, t.Pos)
		return OptionStatement{}.FromBuf(o)
	case fbast.StatementBuiltinStatement:
		b := new(fbast.BuiltinStatement)
		b.Init(t.Bytes, t.Pos)
		return BuiltinStatement{}.FromBuf(b)
	case fbast.StatementTestStatement:
		s := new(fbast.TestStatement)
		s.Init(t.Bytes, t.Pos)
		return TestStatement{}.FromBuf(s)
	default:
		// Ultimately we want to use bad statement/expression to store errors?
		return nil
	}
}

type exprGetterFn func(obj *fbast.WrappedExpression, j int) bool

func exprArrayFromBuf(len int, g exprGetterFn, arrName string) ([]Expression, []Error) {
	s := make([]Expression, len)
	err := make([]Error, 0)
	for i := 0; i < len; i++ {
		e := new(fbast.WrappedExpression)
		t := new(flatbuffers.Table)
		if !g(e, i) || !e.Expr(t) || e.ExprType() == fbast.ExpressionNONE {
			err = append(err, Error{fmt.Sprintf("Encountered error in deserializing %s[%d]", arrName, i)})
		} else {
			s[i] = exprFromBufTable(t, e.ExprType())
		}
	}
	return s, err
}

type unionTableWriterFn func(t *flatbuffers.Table) bool

func exprFromBuf(label string, baseNode BaseNode, f unionTableWriterFn, etype fbast.Expression) Expression {
	t := new(flatbuffers.Table)
	if !f(t) || etype == fbast.ExpressionNONE {
		baseNode.Errors = append(baseNode.Errors,
			Error{fmt.Sprintf("Encountered error in deserializing %s", label)})
		return nil
	}
	return exprFromBufTable(t, etype)
}

func exprFromBufTable(t *flatbuffers.Table, etype fbast.Expression) Expression {
	switch etype {
	case fbast.ExpressionStringExpression:
		s := new(fbast.StringExpression)
		s.Init(t.Bytes, t.Pos)
		return StringExpression{}.FromBuf(s)
	case fbast.ExpressionParenExpression:
		p := new(fbast.ParenExpression)
		p.Init(t.Bytes, t.Pos)
		return ParenExpression{}.FromBuf(p)
	case fbast.ExpressionArrayExpression:
		a := new(fbast.ArrayExpression)
		a.Init(t.Bytes, t.Pos)
		return ArrayExpression{}.FromBuf(a)
	case fbast.ExpressionFunctionExpression:
		f := new(fbast.FunctionExpression)
		f.Init(t.Bytes, t.Pos)
		return FunctionExpression{}.FromBuf(f)
	case fbast.ExpressionBinaryExpression:
		b := new(fbast.BinaryExpression)
		b.Init(t.Bytes, t.Pos)
		return BinaryExpression{}.FromBuf(b)
	case fbast.ExpressionBooleanLiteral:
		b := new(fbast.BooleanLiteral)
		b.Init(t.Bytes, t.Pos)
		return BooleanLiteral{}.FromBuf(b)
	case fbast.ExpressionCallExpression:
		c := new(fbast.CallExpression)
		c.Init(t.Bytes, t.Pos)
		return CallExpression{}.FromBuf(c)
	case fbast.ExpressionConditionalExpression:
		c := new(fbast.ConditionalExpression)
		c.Init(t.Bytes, t.Pos)
		return ConditionalExpression{}.FromBuf(c)
	case fbast.ExpressionDateTimeLiteral:
		d := new(fbast.DateTimeLiteral)
		d.Init(t.Bytes, t.Pos)
		return DateTimeLiteral{}.FromBuf(d)
	case fbast.ExpressionDurationLiteral:
		d := new(fbast.DurationLiteral)
		d.Init(t.Bytes, t.Pos)
		return DurationLiteral{}.FromBuf(d)
	case fbast.ExpressionFloatLiteral:
		f := new(fbast.FloatLiteral)
		f.Init(t.Bytes, t.Pos)
		return FloatLiteral{}.FromBuf(f)
	case fbast.ExpressionIdentifier:
		i := new(fbast.Identifier)
		i.Init(t.Bytes, t.Pos)
		return Identifier{}.FromBuf(i)
	case fbast.ExpressionIntegerLiteral:
		i := new(fbast.IntegerLiteral)
		i.Init(t.Bytes, t.Pos)
		return IntegerLiteral{}.FromBuf(i)
	case fbast.ExpressionLogicalExpression:
		l := new(fbast.LogicalExpression)
		l.Init(t.Bytes, t.Pos)
		return LogicalExpression{}.FromBuf(l)
	case fbast.ExpressionMemberExpression:
		m := new(fbast.MemberExpression)
		m.Init(t.Bytes, t.Pos)
		return MemberExpression{}.FromBuf(m)
	case fbast.ExpressionIndexExpression:
		m := new(fbast.IndexExpression)
		m.Init(t.Bytes, t.Pos)
		return IndexExpression{}.FromBuf(m)
	case fbast.ExpressionObjectExpression:
		m := new(fbast.ObjectExpression)
		m.Init(t.Bytes, t.Pos)
		return ObjectExpression{}.FromBuf(m)
	case fbast.ExpressionPipeExpression:
		p := new(fbast.PipeExpression)
		p.Init(t.Bytes, t.Pos)
		return PipeExpression{}.FromBuf(p)
	case fbast.ExpressionPipeLiteral:
		p := new(fbast.PipeLiteral)
		p.Init(t.Bytes, t.Pos)
		return PipeLiteral{}.FromBuf(p)
	case fbast.ExpressionRegexpLiteral:
		r := new(fbast.RegexpLiteral)
		r.Init(t.Bytes, t.Pos)
		return RegexpLiteral{}.FromBuf(r)
	case fbast.ExpressionStringLiteral:
		r := new(fbast.StringLiteral)
		r.Init(t.Bytes, t.Pos)
		return StringLiteral{}.FromBuf(r)
	case fbast.ExpressionUnaryExpression:
		u := new(fbast.UnaryExpression)
		u.Init(t.Bytes, t.Pos)
		return UnaryExpression{}.FromBuf(u)
	case fbast.ExpressionUnsignedIntegerLiteral:
		u := new(fbast.UnsignedIntegerLiteral)
		u.Init(t.Bytes, t.Pos)
		return UnsignedIntegerLiteral{}.FromBuf(u)
	case fbast.ExpressionBadExpression:
		fallthrough
	default:
		return nil
	}
}

func assignmentFromBuf(label string, baseNode BaseNode, f unionTableWriterFn, atype fbast.Assignment) Assignment {
	t := new(flatbuffers.Table)
	if !f(t) || atype == fbast.AssignmentNONE {
		baseNode.Errors = append(baseNode.Errors,
			Error{fmt.Sprintf("Encountered error in deserializing %s", label)})
		return nil
	}
	switch atype {
	case fbast.AssignmentMemberAssignment:
		fba := new(fbast.MemberAssignment)
		fba.Init(t.Bytes, t.Pos)
		return MemberAssignment{}.FromBuf(fba)
	case fbast.AssignmentVariableAssignment:
		fba := new(fbast.VariableAssignment)
		fba.Init(t.Bytes, t.Pos)
		return VariableAssignment{}.FromBuf(fba)
	default:
		return nil
	}
}

func propertyKeyFromBuf(label string, baseNode BaseNode, f unionTableWriterFn, atype fbast.PropertyKey) PropertyKey {
	t := new(flatbuffers.Table)
	if !f(t) || atype == fbast.PropertyKeyNONE {
		baseNode.Errors = append(baseNode.Errors,
			Error{fmt.Sprintf("Encountered error in deserializing %s", label)})
		return nil
	}
	switch atype {
	case fbast.PropertyKeyIdentifier:
		fbk := new(fbast.Identifier)
		fbk.Init(t.Bytes, t.Pos)
		return Identifier{}.FromBuf(fbk)
	case fbast.PropertyKeyStringLiteral:
		fbs := new(fbast.StringLiteral)
		fbs.Init(t.Bytes, t.Pos)
		return StringLiteral{}.FromBuf(fbs)
	default:
		return nil
	}
}
