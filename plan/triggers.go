package plan

import (
	"fmt"

	"github.com/influxdata/flux"
)

type TriggerSpec interface {
	Kind() TriggerKind
}

type TriggerKind int

const (
	NarrowTransformation TriggerKind = iota
	AfterWatermark
	Repeated
	AfterProcessingTime
	AfterAtLeastCount
	OrFinally
)

var DefaultTriggerSpec = AfterWatermarkTriggerSpec{}

type TriggerAwareProcedureSpec interface {
	TriggerSpec() TriggerSpec
}

func SetTriggerSpec(node PlanNode) error {
	ppn, ok := node.(*PhysicalPlanNode)
	if !ok {
		return fmt.Errorf("cannot set trigger spec on plan node of type %T", node)
	}
	spec := ppn.Spec
	if n, ok := spec.(TriggerAwareProcedureSpec); ok {
		ppn.TriggerSpec = n.TriggerSpec()
	} else {
		ppn.TriggerSpec = DefaultTriggerSpec
	}
	return nil
}

type NarrowTransformationTriggerSpec struct{}

func (NarrowTransformationTriggerSpec) Kind() TriggerKind {
	return NarrowTransformation
}

type AfterWatermarkTriggerSpec struct {
	AllowedLateness flux.Duration
}

func (AfterWatermarkTriggerSpec) Kind() TriggerKind {
	return AfterWatermark
}

type RepeatedTriggerSpec struct {
	Trigger TriggerSpec
}

func (RepeatedTriggerSpec) Kind() TriggerKind {
	return Repeated
}

type AfterProcessingTimeTriggerSpec struct {
	Duration flux.Duration
}

func (AfterProcessingTimeTriggerSpec) Kind() TriggerKind {
	return AfterProcessingTime
}

type AfterAtLeastCountTriggerSpec struct {
	Count int
}

func (AfterAtLeastCountTriggerSpec) Kind() TriggerKind {
	return AfterAtLeastCount
}

type OrFinallyTriggerSpec struct {
	Main    TriggerSpec
	Finally TriggerSpec
}

func (OrFinallyTriggerSpec) Kind() TriggerKind {
	return OrFinally
}
