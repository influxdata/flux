// Code generated by the FlatBuffers compiler. DO NOT EDIT.

package fbast

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type UnaryExpression struct {
	_tab flatbuffers.Table
}

func GetRootAsUnaryExpression(buf []byte, offset flatbuffers.UOffsetT) *UnaryExpression {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &UnaryExpression{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *UnaryExpression) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *UnaryExpression) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *UnaryExpression) BaseNode(obj *BaseNode) *BaseNode {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		x := rcv._tab.Indirect(o + rcv._tab.Pos)
		if obj == nil {
			obj = new(BaseNode)
		}
		obj.Init(rcv._tab.Bytes, x)
		return obj
	}
	return nil
}

func (rcv *UnaryExpression) Operator() OperatorKind {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		return rcv._tab.GetInt8(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *UnaryExpression) MutateOperator(n OperatorKind) bool {
	return rcv._tab.MutateInt8Slot(6, n)
}

func (rcv *UnaryExpression) ArgumentType() byte {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		return rcv._tab.GetByte(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *UnaryExpression) MutateArgumentType(n byte) bool {
	return rcv._tab.MutateByteSlot(8, n)
}

func (rcv *UnaryExpression) Argument(obj *flatbuffers.Table) bool {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(10))
	if o != 0 {
		rcv._tab.Union(obj, o)
		return true
	}
	return false
}

func UnaryExpressionStart(builder *flatbuffers.Builder) {
	builder.StartObject(4)
}
func UnaryExpressionAddBaseNode(builder *flatbuffers.Builder, baseNode flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(0, flatbuffers.UOffsetT(baseNode), 0)
}
func UnaryExpressionAddOperator(builder *flatbuffers.Builder, operator int8) {
	builder.PrependInt8Slot(1, operator, 0)
}
func UnaryExpressionAddArgumentType(builder *flatbuffers.Builder, argumentType byte) {
	builder.PrependByteSlot(2, argumentType, 0)
}
func UnaryExpressionAddArgument(builder *flatbuffers.Builder, argument flatbuffers.UOffsetT) {
	builder.PrependUOffsetTSlot(3, flatbuffers.UOffsetT(argument), 0)
}
func UnaryExpressionEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}
