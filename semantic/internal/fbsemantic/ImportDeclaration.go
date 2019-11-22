// Code generated by the FlatBuffers compiler. DO NOT EDIT.

package fbsemantic

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type ImportDeclaration struct {
	_tab flatbuffers.Table
}

func GetRootAsImportDeclaration(buf []byte, offset flatbuffers.UOffsetT) *ImportDeclaration {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &ImportDeclaration{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *ImportDeclaration) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *ImportDeclaration) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *ImportDeclaration) Loc(obj *SourceLocation) *SourceLocation {
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

func (rcv *ImportDeclaration) Alias(obj *IdentifierExpression) *IdentifierExpression {
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

func (rcv *ImportDeclaration) Path(obj *StringLiteral) *StringLiteral {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		x := rcv._tab.Indirect(o + rcv._tab.Pos)
		if obj == nil {
			obj = new(StringLiteral)
		}
		obj.Init(rcv._tab.Bytes, x)
		return obj
	}
	return nil
}

func ImportDeclarationStart(builder *flatbuffers.Builder) {
	builder.StartObject(3)
}
func ImportDeclarationAddLoc(builder *flatbuffers.Builder, loc flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(0, flatbuffers.UOffsetT(loc), 0)
}
func ImportDeclarationAddAlias(builder *flatbuffers.Builder, alias flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(1, flatbuffers.UOffsetT(alias), 0)
}
func ImportDeclarationAddPath(builder *flatbuffers.Builder, path flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(2, flatbuffers.UOffsetT(path), 0)
}
func ImportDeclarationEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}
