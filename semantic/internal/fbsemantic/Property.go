// Code generated by the FlatBuffers compiler. DO NOT EDIT.

package fbsemantic

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type Property struct {
	_tab flatbuffers.Table
}

func GetRootAsProperty(buf []byte, offset flatbuffers.UOffsetT) *Property {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Property{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *Property) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Property) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *Property) Loc(obj *SourceLocation) *SourceLocation {
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

func (rcv *Property) Key(obj *IdentifierExpression) *IdentifierExpression {
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

func (rcv *Property) ValueType() byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		return rcv._tab.GetByte(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *Property) MutateValueType(n byte) bool {
	return rcv._tab.MutateByteSlot(8, n)
}

func (rcv *Property) Value(obj *flatbuffers.Table) bool {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(10))
	if o != 0 {
		rcv._tab.Union(obj, o)
		return true
	}
	return false
}

func PropertyStart(builder *flatbuffers.Builder) {
	builder.StartObject(4)
}
func PropertyAddLoc(builder *flatbuffers.Builder, loc flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(0, flatbuffers.UOffsetT(loc), 0)
}
func PropertyAddKey(builder *flatbuffers.Builder, key flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(1, flatbuffers.UOffsetT(key), 0)
}
func PropertyAddValueType(builder *flatbuffers.Builder, valueType byte) {
	builder.PrependByteSlot(2, valueType, 0)
}
func PropertyAddValue(builder *flatbuffers.Builder, value flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(3, flatbuffers.UOffsetT(value), 0)
}
func PropertyEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}
