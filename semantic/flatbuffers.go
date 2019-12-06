package semantic

import (
	"fmt"
	"regexp"
	"time"

	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/influxdata/flux/ast"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/semantic/internal/fbsemantic"
)

func DeserializeFromFlatBuffer(buf []byte) (*Package, error) {
	fbPkg := fbsemantic.GetRootAsPackage(buf, 0)
	p := &Package{}
	if err := p.FromBuf(fbPkg); err != nil {
		return nil, err
	}
	return p, nil
}

func (l *loc) FromBuf(fb *fbsemantic.SourceLocation) error {
	l.File = string(fb.File())
	posFromBuf(&l.Start, fb.Start(nil))
	posFromBuf(&l.End, fb.End(nil))
	l.Source = string(fb.Source())
	return nil
}

func posFromBuf(p *ast.Position, fb *fbsemantic.Position) *ast.Position {
	p.Line = int(fb.Line())
	p.Column = int(fb.Column())
	return p
}

func fromWrappedStatement(fb *fbsemantic.WrappedStatement) (Statement, error) {
	tbl := new(flatbuffers.Table)
	if !fb.Statement(tbl) {
		return nil, fmt.Errorf("missing table in wrapped statement")
	}
	switch st := fb.StatementType(); st {
	case fbsemantic.StatementOptionStatement:
		fbStmt := new(fbsemantic.OptionStatement)
		fbStmt.Init(tbl.Bytes, tbl.Pos)
		s := &OptionStatement{}
		if err := s.FromBuf(fbStmt); err != nil {
			return nil, err
		}
		return s, nil
	case fbsemantic.StatementBuiltinStatement:
		fbStmt := new(fbsemantic.BuiltinStatement)
		fbStmt.Init(tbl.Bytes, tbl.Pos)
		s := &BuiltinStatement{}
		if err := s.FromBuf(fbStmt); err != nil {
			return nil, err
		}
		return s, nil
	case fbsemantic.StatementTestStatement:
		fbStmt := new(fbsemantic.TestStatement)
		fbStmt.Init(tbl.Bytes, tbl.Pos)
		s := &TestStatement{}
		if err := s.FromBuf(fbStmt); err != nil {
			return nil, err
		}
		return s, nil
	case fbsemantic.StatementExpressionStatement:
		fbStmt := new(fbsemantic.ExpressionStatement)
		fbStmt.Init(tbl.Bytes, tbl.Pos)
		s := &ExpressionStatement{}
		if err := s.FromBuf(fbStmt); err != nil {
			return nil, err
		}
		return s, nil
	case fbsemantic.StatementReturnStatement:
		fbStmt := new(fbsemantic.ReturnStatement)
		fbStmt.Init(tbl.Bytes, tbl.Pos)
		s := &ReturnStatement{}
		if err := s.FromBuf(fbStmt); err != nil {
			return nil, err
		}
		return s, nil
	default:
		if name, ok := fbsemantic.EnumNamesStatement[st]; ok {
			return nil, errors.Newf(codes.Internal, "unhandled statement type %v", name)
		} else {
			return nil, errors.Newf(codes.Internal, "unknown statement type (%v)", st)
		}
	}
}

type getTableFn func(*flatbuffers.Table) bool

func fromWrappedExpression(fb *fbsemantic.WrappedExpression) (Expression, error) {
	return fromExpressionTable(fb.Expression, fb.ExpressionType())
}

func fromExpressionTable(getTable getTableFn, exprType fbsemantic.Expression) (Expression, error) {
	tbl := new(flatbuffers.Table)
	if !getTable(tbl) {
		if name, ok := fbsemantic.EnumNamesExpression[exprType]; ok {
			return nil, errors.Newf(codes.Internal, "missing expr type %v", name)
		} else {
			return nil, errors.Newf(codes.Internal, "missing unknown expr type %v", exprType)
		}
	}
	switch exprType {
	case fbsemantic.ExpressionStringExpression:
		fbExpr := new(fbsemantic.StringExpression)
		fbExpr.Init(tbl.Bytes, tbl.Pos)
		e := &StringExpression{}
		if err := e.FromBuf(fbExpr); err != nil {
			return nil, err
		}
		return e, nil
	case fbsemantic.ExpressionArrayExpression:
		fbExpr := new(fbsemantic.ArrayExpression)
		fbExpr.Init(tbl.Bytes, tbl.Pos)
		e := &ArrayExpression{}
		if err := e.FromBuf(fbExpr); err != nil {
			return nil, err
		}
		return e, nil
	case fbsemantic.ExpressionFunctionExpression:
		fbExpr := new(fbsemantic.FunctionExpression)
		fbExpr.Init(tbl.Bytes, tbl.Pos)
		e := &FunctionExpression{}
		if err := e.FromBuf(fbExpr); err != nil {
			return nil, err
		}
		return e, nil
	case fbsemantic.ExpressionBinaryExpression:
		fbExpr := new(fbsemantic.BinaryExpression)
		fbExpr.Init(tbl.Bytes, tbl.Pos)
		e := &BinaryExpression{}
		if err := e.FromBuf(fbExpr); err != nil {
			return nil, err
		}
		return e, nil
	case fbsemantic.ExpressionCallExpression:
		fbExpr := new(fbsemantic.CallExpression)
		fbExpr.Init(tbl.Bytes, tbl.Pos)
		e := &CallExpression{}
		if err := e.FromBuf(fbExpr); err != nil {
			return nil, err
		}
		return e, nil
	case fbsemantic.ExpressionConditionalExpression:
		fbExpr := new(fbsemantic.ConditionalExpression)
		fbExpr.Init(tbl.Bytes, tbl.Pos)
		e := &ConditionalExpression{}
		if err := e.FromBuf(fbExpr); err != nil {
			return nil, err
		}
		return e, nil
	case fbsemantic.ExpressionIdentifierExpression:
		fbExpr := new(fbsemantic.IdentifierExpression)
		fbExpr.Init(tbl.Bytes, tbl.Pos)
		e := &IdentifierExpression{}
		if err := e.FromBuf(fbExpr); err != nil {
			return nil, err
		}
		return e, nil
	case fbsemantic.ExpressionLogicalExpression:
		fbExpr := new(fbsemantic.LogicalExpression)
		fbExpr.Init(tbl.Bytes, tbl.Pos)
		e := &LogicalExpression{}
		if err := e.FromBuf(fbExpr); err != nil {
			return nil, err
		}
		return e, nil
	case fbsemantic.ExpressionMemberExpression:
		fbExpr := new(fbsemantic.MemberExpression)
		fbExpr.Init(tbl.Bytes, tbl.Pos)
		e := &MemberExpression{}
		if err := e.FromBuf(fbExpr); err != nil {
			return nil, err
		}
		return e, nil
	case fbsemantic.ExpressionIndexExpression:
		fbExpr := new(fbsemantic.IndexExpression)
		fbExpr.Init(tbl.Bytes, tbl.Pos)
		e := &IndexExpression{}
		if err := e.FromBuf(fbExpr); err != nil {
			return nil, err
		}
		return e, nil
	case fbsemantic.ExpressionObjectExpression:
		fbExpr := new(fbsemantic.ObjectExpression)
		fbExpr.Init(tbl.Bytes, tbl.Pos)
		e := &ObjectExpression{}
		if err := e.FromBuf(fbExpr); err != nil {
			return nil, err
		}
		return e, nil
	case fbsemantic.ExpressionUnaryExpression:
		fbExpr := new(fbsemantic.UnaryExpression)
		fbExpr.Init(tbl.Bytes, tbl.Pos)
		e := &UnaryExpression{}
		if err := e.FromBuf(fbExpr); err != nil {
			return nil, err
		}
		return e, nil
	case fbsemantic.ExpressionBooleanLiteral:
		fbExpr := new(fbsemantic.BooleanLiteral)
		fbExpr.Init(tbl.Bytes, tbl.Pos)
		e := &BooleanLiteral{}
		if err := e.FromBuf(fbExpr); err != nil {
			return nil, err
		}
		return e, nil
	case fbsemantic.ExpressionDateTimeLiteral:
		fbExpr := new(fbsemantic.DateTimeLiteral)
		fbExpr.Init(tbl.Bytes, tbl.Pos)
		e := &DateTimeLiteral{}
		if err := e.FromBuf(fbExpr); err != nil {
			return nil, err
		}
		return e, nil
	case fbsemantic.ExpressionDurationLiteral:
		fbExpr := new(fbsemantic.DurationLiteral)
		fbExpr.Init(tbl.Bytes, tbl.Pos)
		e := &DurationLiteral{}
		if err := e.FromBuf(fbExpr); err != nil {
			return nil, err
		}
		return e, nil
	case fbsemantic.ExpressionFloatLiteral:
		fbExpr := new(fbsemantic.FloatLiteral)
		fbExpr.Init(tbl.Bytes, tbl.Pos)
		e := &FloatLiteral{}
		if err := e.FromBuf(fbExpr); err != nil {
			return nil, err
		}
		return e, nil
	case fbsemantic.ExpressionIntegerLiteral:
		fbExpr := new(fbsemantic.IntegerLiteral)
		fbExpr.Init(tbl.Bytes, tbl.Pos)
		e := &IntegerLiteral{}
		if err := e.FromBuf(fbExpr); err != nil {
			return nil, err
		}
		return e, nil
	case fbsemantic.ExpressionStringLiteral:
		fbExpr := new(fbsemantic.StringLiteral)
		fbExpr.Init(tbl.Bytes, tbl.Pos)
		e := &StringLiteral{}
		if err := e.FromBuf(fbExpr); err != nil {
			return nil, err
		}
		return e, nil
	case fbsemantic.ExpressionRegexpLiteral:
		fbExpr := new(fbsemantic.RegexpLiteral)
		fbExpr.Init(tbl.Bytes, tbl.Pos)
		e := &RegexpLiteral{}
		if err := e.FromBuf(fbExpr); err != nil {
			return nil, err
		}
		return e, nil
	case fbsemantic.ExpressionUnsignedIntegerLiteral:
		fbExpr := new(fbsemantic.UnsignedIntegerLiteral)
		fbExpr.Init(tbl.Bytes, tbl.Pos)
		e := &UnsignedIntegerLiteral{}
		if err := e.FromBuf(fbExpr); err != nil {
			return nil, err
		}
		return e, nil
	default:
		if name, ok := fbsemantic.EnumNamesExpression[exprType]; ok {
			return nil, errors.Newf(codes.Internal, "unhandled expr type %v", name)
		} else {
			return nil, errors.Newf(codes.Internal, "unknown expr type %v", exprType)
		}
	}
}

func fromAssignmentTable(getTable getTableFn, assignType fbsemantic.Assignment) (Assignment, error) {
	tbl := new(flatbuffers.Table)
	if !getTable(tbl) {
		if name, ok := fbsemantic.EnumNamesAssignment[assignType]; ok {
			return nil, errors.Newf(codes.Internal, "missing assignment with type %v", name)
		} else {
			return nil, errors.Newf(codes.Internal, "missing assignment with unknown type (%v)", assignType)
		}
	}
	switch assignType {
	case fbsemantic.AssignmentMemberAssignment:
		fbAssign := new(fbsemantic.MemberAssignment)
		fbAssign.Init(tbl.Bytes, tbl.Pos)
		a := &MemberAssignment{}
		if err := a.FromBuf(fbAssign); err != nil {
			return nil, err
		}
		return a, nil
	case fbsemantic.AssignmentNativeVariableAssignment:
		fbAssign := new(fbsemantic.NativeVariableAssignment)
		fbAssign.Init(tbl.Bytes, tbl.Pos)
		a := &NativeVariableAssignment{}
		if err := a.FromBuf(fbAssign); err != nil {
			return nil, err
		}
		return a, nil
	default:
		if name, ok := fbsemantic.EnumNamesAssignment[assignType]; ok {
			return nil, errors.Newf(codes.Internal, "unhandled assignment type %v", name)
		} else {
			return nil, errors.Newf(codes.Internal, "unknown assignment type (%v)", assignType)
		}
	}
}

func fromFBOperator(o fbsemantic.Operator) (ast.OperatorKind, error) {
	switch o {
	case fbsemantic.OperatorMultiplicationOperator:
		return ast.MultiplicationOperator, nil
	case fbsemantic.OperatorDivisionOperator:
		return ast.DivisionOperator, nil
	case fbsemantic.OperatorModuloOperator:
		return ast.ModuloOperator, nil
	case fbsemantic.OperatorPowerOperator:
		return ast.PowerOperator, nil
	case fbsemantic.OperatorAdditionOperator:
		return ast.AdditionOperator, nil
	case fbsemantic.OperatorSubtractionOperator:
		return ast.SubtractionOperator, nil
	case fbsemantic.OperatorLessThanEqualOperator:
		return ast.LessThanEqualOperator, nil
	case fbsemantic.OperatorLessThanOperator:
		return ast.LessThanOperator, nil
	case fbsemantic.OperatorGreaterThanEqualOperator:
		return ast.GreaterThanEqualOperator, nil
	case fbsemantic.OperatorGreaterThanOperator:
		return ast.GreaterThanOperator, nil
	case fbsemantic.OperatorStartsWithOperator:
		return ast.StartsWithOperator, nil
	case fbsemantic.OperatorInOperator:
		return ast.InOperator, nil
	case fbsemantic.OperatorNotOperator:
		return ast.NotOperator, nil
	case fbsemantic.OperatorExistsOperator:
		return ast.ExistsOperator, nil
	case fbsemantic.OperatorNotEmptyOperator:
		return ast.NotEmptyOperator, nil
	case fbsemantic.OperatorEmptyOperator:
		return ast.EmptyOperator, nil
	case fbsemantic.OperatorEqualOperator:
		return ast.EqualOperator, nil
	case fbsemantic.OperatorNotEqualOperator:
		return ast.NotEqualOperator, nil
	case fbsemantic.OperatorRegexpMatchOperator:
		return ast.RegexpMatchOperator, nil
	case fbsemantic.OperatorNotRegexpMatchOperator:
		return ast.NotRegexpMatchOperator, nil
	default:
		if name, ok := fbsemantic.EnumNamesOperator[o]; ok {
			return 0, errors.Newf(codes.Internal, "unsupported operator %v", name)
		} else {
			return 0, errors.Newf(codes.Internal, "unknown operator (%v)", o)
		}
	}
}

func fromFBLogicalOperator(o fbsemantic.Operator) (ast.LogicalOperatorKind, error) {
	switch o {
	case fbsemantic.LogicalOperatorAndOperator:
		return ast.AndOperator, nil
	case fbsemantic.LogicalOperatorOrOperator:
		return ast.OrOperator, nil
	default:
		if name, ok := fbsemantic.EnumNamesLogicalOperator[o]; ok {
			return 0, errors.Newf(codes.Internal, "unsupported logical operator %v", name)
		} else {
			return 0, errors.Newf(codes.Internal, "unknown logical operator (%v)", o)
		}
	}
}

func propertyKeyFromFBIdentifier(fbId *fbsemantic.Identifier) (PropertyKey, error) {
	id := &Identifier{}
	if err := id.FromBuf(fbId); err != nil {
		return nil, err
	}
	return id, nil
}

func fromFBTime(fbTime *fbsemantic.Time) time.Time {
	z := time.FixedZone("fbsem", int(fbTime.Offset()))
	t := time.Unix(fbTime.Secs(), int64(fbTime.Nsecs()))
	return t.In(z)
}

func fromFBDurationVector(fbDurLit *fbsemantic.DurationLiteral) ([]ast.Duration, error) {
	if fbDurLit.ValueLength() <= 0 {
		return nil, errors.New(codes.Internal, "missing duration vector")
	}

	durs := make([]ast.Duration, fbDurLit.ValueLength())
	for i := 0; i < fbDurLit.ValueLength(); i++ {
		fbDur := new(fbsemantic.Duration)
		if !fbDurLit.Value(fbDur, i) {
			return nil, errors.Newf(codes.Internal, "missing duration at position %v", i)
		}
		dur := ast.Duration{
			Magnitude: fbDur.Magnitude(),
			Unit:      fbsemantic.EnumNamesTimeUnit[fbDur.Unit()],
		}
		durs[i] = dur
	}
	return durs, nil
}

func fromFBStringExpressionPartVector(fbExpr *fbsemantic.StringExpression) ([]StringExpressionPart, error) {
	if fbExpr.PartsLength() <= 0 {
		return nil, errors.New(codes.Internal, "missing string expression part vector")
	}

	parts := make([]StringExpressionPart, fbExpr.PartsLength())
	for i := 0; i < fbExpr.PartsLength(); i++ {
		fbPart := new(fbsemantic.StringExpressionPart)
		if !fbExpr.Parts(fbPart, i) {
			return nil, errors.New(codes.Internal, "missing string expression part")
		}

		fbLoc := fbPart.Loc(nil)
		fbExprTy := fbPart.InterpolatedExpressionType()
		var part StringExpressionPart
		if text := fbPart.TextValue(); len(text) > 0 {
			if fbExprTy != fbsemantic.ExpressionNONE {
				return nil, errors.Newf(codes.Internal, "found both text part and interpolated expression")
			}
			tp := &TextPart{
				Value: string(text),
			}
			if fbLoc != nil {
				if err := tp.loc.FromBuf(fbLoc); err != nil {
					return nil, err
				}
			}
			part = tp
		} else if fbExprTy != fbsemantic.ExpressionNONE {
			expr, err := fromExpressionTable(fbPart.InterpolatedExpression, fbExprTy)
			if err != nil {
				return nil, err
			}
			ip := &InterpolatedPart{
				Expression: expr,
			}
			if fbLoc != nil {
				if err := ip.loc.FromBuf(fbLoc); err != nil {
					return nil, err
				}
			}
			part = ip
		} else {
			return nil, errors.New(codes.Internal, "expected to find either text or interpolated expression")
		}

		parts = append(parts, part)
	}
	return parts, nil
}

func fromFBRegexpLiteral(fbRegexp []byte) (*regexp.Regexp, error) {
	if len(fbRegexp) == 0 {
		return nil, errors.New(codes.Internal, "missing regular expression")
	}

	re, err := regexp.Compile(string(fbRegexp))
	if err != nil {
		return nil, errors.Wrap(err, codes.Internal)
	}
	return re, nil
}

func (e *FunctionExpression) FromBuf(fb *fbsemantic.FunctionExpression) error {
	if fbLoc := fb.Loc(nil); fbLoc != nil {
		if err := e.loc.FromBuf(fbLoc); err != nil {
			return errors.Wrap(err, codes.Inherit, "FunctionExpression.loc")
		}
	}

	bl := new(FunctionBlock)
	var defaults []*Property
	{
		bl.loc = e.loc
		ps := &FunctionParameters{
			loc: e.loc,
		}
		{
			nParams := fb.ParamsLength()
			ps.List = make([]*FunctionParameter, nParams)
			for i := 0; i < nParams; i++ {
				fbp := new(fbsemantic.FunctionParameter)
				if !fb.Params(fbp, i) {
					return errors.Newf(codes.Internal, "missing parameter at position %v", i)
				}
				p := new(FunctionParameter)
				if err := p.FromBuf(fbp); err != nil {
					return err
				}
				ps.List[i] = p

				if fbp.Default(&flatbuffers.Table{}) {
					e, err := fromExpressionTable(fbp.Default, fbp.DefaultType())
					if err != nil {
						return errors.Wrapf(err, codes.Inherit, "default for parameter at position %v", i)
					}
					defaults = append(defaults, &Property{
						loc:   p.loc,
						Key:   p.Key,
						Value: e,
					})
				}

				if fbp.IsPipe() {
					ps.Pipe = p.Key
				}
			}
		}
		bl.Parameters = ps

		fbBlock := fb.Body(nil)
		if fbBlock == nil {
			return errors.New(codes.Internal, "missing function body")
		}
		stmts := new(Block)
		if err := stmts.FromBuf(fbBlock); err != nil {
			return err
		}
		bl.Body = stmts
	}
	e.Block = bl

	e.Defaults = &ObjectExpression{
		loc:        e.loc,
		Properties: defaults,
	}

	return nil
}

func (p *FunctionParameter) FromBuf(fb *fbsemantic.FunctionParameter) error {
	if fbLoc := fb.Loc(nil); fbLoc != nil {
		if err := p.loc.FromBuf(fbLoc); err != nil {
			return errors.Wrap(err, codes.Inherit, "FunctionParameter.loc")
		}
	}

	fbKey := fb.Key(nil)
	if fbKey == nil {
		return errors.New(codes.Internal, "missing parameter")
	}
	p.Key = new(Identifier)
	if err := p.Key.FromBuf(fbKey); err != nil {
		return err
	}
	return nil
}
