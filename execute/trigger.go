package execute

import (
	"fmt"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/plan"
)

type Trigger interface {
	Triggered(TriggerContext) bool
	Finished() bool
	Reset()
}

type TriggerContext struct {
	Table                 TableContext
	Watermark             Time
	CurrentProcessingTime Time
}

type TableContext struct {
	Key   flux.GroupKey
	Count int
}

func NewTriggerFromSpec(spec plan.TriggerSpec) Trigger {
	switch s := spec.(type) {
	case plan.NarrowTransformationTriggerSpec:
		return &narrowTransformationTrigger{}
	case plan.AfterWatermarkTriggerSpec:
		return &afterWatermarkTrigger{
			allowedLateness: Duration(s.AllowedLateness),
		}
	case plan.RepeatedTriggerSpec:
		return &repeatedlyForever{
			t: NewTriggerFromSpec(s.Trigger),
		}
	case plan.AfterProcessingTimeTriggerSpec:
		return &afterProcessingTimeTrigger{
			duration: Duration(s.Duration),
		}
	case plan.AfterAtLeastCountTriggerSpec:
		return &afterAtLeastCount{
			atLeast: s.Count,
		}
	case plan.OrFinallyTriggerSpec:
		return &orFinally{
			main:    NewTriggerFromSpec(s.Main),
			finally: NewTriggerFromSpec(s.Finally),
		}
	default:
		//TODO(nathanielc): Add proper error handling here.
		// Maybe separate validation of a spec and creation of a spec so we know we cannot error during creation?
		panic(fmt.Sprintf("unsupported trigger spec provided %T", spec))
	}
}

// Informally a narrow transformation is one where each output table originates
// from a single input table. Once an input table is processed, the resulting
// output table may be sent downstream immediately. The trigger associated with
// these kinds of transformations fires immediately as well as finishes
// immediately. This behavior ensures there is at most one table in the
// transformation's data cache at any given time.
type narrowTransformationTrigger struct{}

func (t *narrowTransformationTrigger) Triggered(c TriggerContext) bool {
	return true
}

func (t *narrowTransformationTrigger) Finished() bool {
	return true
}

func (t *narrowTransformationTrigger) Reset() {}

// afterWatermarkTrigger triggers once the watermark is greater than the bounds of the block.
type afterWatermarkTrigger struct {
	allowedLateness Duration
	finished        bool
}

func (t *afterWatermarkTrigger) Triggered(c TriggerContext) bool {
	timeIdx := ColIdx(DefaultStopColLabel, c.Table.Key.Cols())
	if timeIdx < 0 {
		return false
	}
	stop := c.Table.Key.ValueTime(timeIdx)
	if c.Watermark >= stop+Time(t.allowedLateness) {
		t.finished = true
	}
	return c.Watermark >= stop
}
func (t *afterWatermarkTrigger) Finished() bool {
	return t.finished
}
func (t *afterWatermarkTrigger) Reset() {
	t.finished = false
}

type repeatedlyForever struct {
	t Trigger
}

func (t *repeatedlyForever) Triggered(c TriggerContext) bool {
	return t.t.Triggered(c)
}
func (t *repeatedlyForever) Finished() bool {
	if t.t.Finished() {
		t.Reset()
	}
	return false
}
func (t *repeatedlyForever) Reset() {
	t.t.Reset()
}

type afterProcessingTimeTrigger struct {
	duration       Duration
	triggerTimeSet bool
	triggerTime    Time
	current        Time
}

func (t *afterProcessingTimeTrigger) Triggered(c TriggerContext) bool {
	if !t.triggerTimeSet {
		t.triggerTimeSet = true
		t.triggerTime = c.CurrentProcessingTime + Time(t.duration)
	}
	t.current = c.CurrentProcessingTime
	return t.current >= t.triggerTime
}
func (t *afterProcessingTimeTrigger) Finished() bool {
	return t.triggerTimeSet && t.current >= t.triggerTime
}
func (t *afterProcessingTimeTrigger) Reset() {
	t.triggerTimeSet = false
}

type afterAtLeastCount struct {
	n, atLeast int
}

func (t *afterAtLeastCount) Triggered(c TriggerContext) bool {
	t.n = c.Table.Count
	return t.n >= t.atLeast
}
func (t *afterAtLeastCount) Finished() bool {
	return t.n >= t.atLeast
}
func (t *afterAtLeastCount) Reset() {
	t.n = 0
}

type orFinally struct {
	main     Trigger
	finally  Trigger
	finished bool
}

func (t *orFinally) Triggered(c TriggerContext) bool {
	if t.finally.Triggered(c) {
		t.finished = true
		return true
	}
	return t.main.Triggered(c)
}

func (t *orFinally) Finished() bool {
	return t.finished
}
func (t *orFinally) Reset() {
	t.finished = false
}
