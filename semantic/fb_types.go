package semantic

import (
	"fmt"
	"strings"

	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/influxdata/flux/codes"
	"github.com/influxdata/flux/internal/errors"
	"github.com/influxdata/flux/semantic/internal/fbsemantic"
)

type fbTabler interface {
	Init(buf []byte, i flatbuffers.UOffsetT)
	Table() flatbuffers.Table
}

type MonoType struct {
	mt  fbsemantic.MonoType
	tbl fbTabler
}

func monoTypeFromVar(v *fbsemantic.Var) *MonoType {
	return &MonoType{
		mt:  fbsemantic.MonoTypeVar,
		tbl: v,
	}
}

func (mt *MonoType) ReturnType() (*MonoType, error) {
	f, ok := mt.tbl.(*fbsemantic.Fun)
	if !ok {
		return nil, errors.New(codes.Internal, "ReturnType() called on non-function MonoType")
	}
	tbl := new(flatbuffers.Table)
	if !f.Retn(tbl) {
		return nil, errors.New(codes.Internal, "missing return type")
	}
	return getMonoTypeFromTable(tbl, f.RetnType())
}

func (mt *MonoType) String() string {
	switch mt.mt {
	case fbsemantic.MonoTypeVar:
		v := mt.tbl.(*fbsemantic.Var)
		return fmt.Sprintf("t%d", v.I())
	case fbsemantic.MonoTypeBasic:
		b := mt.tbl.(*fbsemantic.Basic)
		return strings.ToLower(fbsemantic.EnumNamesType[b.T()])
	case fbsemantic.MonoTypeFun:
		var sb strings.Builder
		f := mt.tbl.(*fbsemantic.Fun)
		sb.WriteString("(")
		needComma := false
		for i := 0; i < f.ArgsLength(); i++ {
			arg := new(fbsemantic.Argument)
			if !f.Args(arg, i) {
				return "<missing arg>"
			}
			if needComma {
				sb.WriteString(", ")
			} else {
				needComma = true
			}
			if arg.Optional() {
				sb.WriteString("?")
			} else if arg.Pipe() {
				sb.WriteString("<-")
			}
			sb.WriteString(string(arg.Name()) + ": ")
			tbl := new(flatbuffers.Table)
			if !arg.T(tbl) {
				return "<fun arg missing type>"
			}
			argTy, err := getMonoTypeFromTable(tbl, arg.TType())
			if err != nil {
				return "<" + err.Error() + ">"
			}
			sb.WriteString(argTy.String())
		}
		sb.WriteString(") -> ")
		rt, err := mt.ReturnType()
		if err != nil {
			return "<" + err.Error() + ">"
		}
		sb.WriteString(rt.String())
		return sb.String()
	default:
		return "<" + fbsemantic.EnumNamesMonoType[mt.mt] + ">"
	}
}

func PolyTypeToString(pt *fbsemantic.PolyType) string {
	var sb strings.Builder
	sb.WriteString("forall [")
	needComma := false
	for i := 0; i < pt.VarsLength(); i++ {
		v := &fbsemantic.Var{}
		if !pt.Vars(v, i) {
			continue
		}
		if needComma {
			sb.WriteString(", ")
		} else {
			needComma = true
		}
		mt := monoTypeFromVar(v)
		sb.WriteString(mt.String())
	}

	sb.WriteString("] ")

	needWhere := true
	for i := 0; i < pt.ConsLength(); i++ {
		cons := &fbsemantic.Constraint{}
		if !pt.Cons(cons, i) {
			continue
		}
		tv := cons.Tvar(nil)
		if tv == nil {
			continue
		}
		k := cons.Kind()

		if needWhere {
			sb.WriteString("where ")
		} else {
			needWhere = false
		}
		mtv := monoTypeFromVar(tv)
		sb.WriteString(mtv.String())
		sb.WriteString(": ")
		sb.WriteString(fbsemantic.EnumNamesKind[k])

		if i < pt.ConsLength()-1 {
			sb.WriteString(", ")
		} else {
			sb.WriteString(" ")
		}
	}

	tbl := &flatbuffers.Table{}
	if !pt.Expr(tbl) {
		sb.WriteString("<missing type expr>")
	}
	mt, err := getMonoTypeFromTable(tbl, pt.ExprType())
	if err != nil {
		sb.WriteString(fmt.Sprintf("<%v>", err))
	}
	sb.WriteString(mt.String())

	return sb.String()
}
