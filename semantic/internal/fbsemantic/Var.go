// Code generated by the FlatBuffers compiler. DO NOT EDIT.

package fbsemantic

import (
	flatbuffers "github.com/google/flatbuffers/go"
)

type Var struct {
	_tab flatbuffers.Table
}

func GetRootAsVar(buf []byte, offset flatbuffers.UOffsetT) *Var {
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Var{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *Var) Init(buf []byte, i flatbuffers.UOffsetT) {
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Var) Table() flatbuffers.Table {
	return rcv._tab
}

func (rcv *Var) I() uint64 {
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		return rcv._tab.GetUint64(o + rcv._tab.Pos)
	}
	return 0
}

func (rcv *Var) MutateI(n uint64) bool {
	return rcv._tab.MutateUint64Slot(4, n)
}

func VarStart(builder *flatbuffers.Builder) {
	builder.StartObject(1)
}
func VarAddI(builder *flatbuffers.Builder, i uint64) {
	builder.PrependUint64Slot(0, i, 0)
}
func VarEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	return builder.EndObject()
}
