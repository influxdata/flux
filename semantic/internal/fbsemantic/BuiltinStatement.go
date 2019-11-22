// Code generated by the FlatBuffers compiler. DO NOT EDIT.

package fbsemantic

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type BuiltinStatement struct {
	_tab flatbuffers.Table
}

func GetRootAsBuiltinStatement(buf []byte, offset flatbuffers.UOffsetT) *BuiltinStatement {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &BuiltinStatement{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *BuiltinStatement) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *BuiltinStatement) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *BuiltinStatement) Loc(obj *SourceLocation) *SourceLocation {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		x := rcv._tab.Indirect(o + rcv._tab.Pos)
		if obj == nil {
			obj = new(SourceLocation)
		}
		obj.Init(rcv._tab.Bytes, x)
		return obj
	}
	return nil
}

func (rcv *BuiltinStatement) Id(obj *IdentifierExpression) *IdentifierExpression {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		x := rcv._tab.Indirect(o + rcv._tab.Pos)
		if obj == nil {
			obj = new(IdentifierExpression)
		}
		obj.Init(rcv._tab.Bytes, x)
		return obj
	}
	return nil
}

func BuiltinStatementStart(builder *flatbuffers.Builder) {
	builder.StartObject(2)
}
func BuiltinStatementAddLoc(builder *flatbuffers.Builder, loc flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(0, flatbuffers.UOffsetT(loc), 0)
}
func BuiltinStatementAddId(builder *flatbuffers.Builder, id flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(1, flatbuffers.UOffsetT(id), 0)
}
func BuiltinStatementEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}
