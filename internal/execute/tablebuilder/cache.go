package tablebuilder

import (
	"fmt"

	"github.com/influxdata/flux"
	"github.com/influxdata/flux/execute"
	"github.com/influxdata/flux/memory"
)

// Cache will contain a group of TableBuilder instances.
type Cache struct {
	// tables holds the instantiated tables.
	tables *execute.GroupLookup
	alloc  *memory.Allocator

	triggerSpec flux.TriggerSpec
}

// NewCache creates a new Cache for retrieving table builders.
func NewCache(alloc *memory.Allocator) *Cache {
	return &Cache{
		tables: execute.NewGroupLookup(),
		alloc:  alloc,
	}
}

type tableState struct {
	builder *Instance
	trigger execute.Trigger
}

// Get will retrieve a TableBuilder instance or, if it hasn't been
// created yet, it will create a new one. If a Table has already
// been instantiated for a specific key, it cannot be retrieved again.
func (c *Cache) Get(key flux.GroupKey) *Instance {
	b, ok := c.tables.Lookup(key)
	if !ok {
		builder, _ := New(c.alloc).WithGroupKey(key)
		t := execute.NewTriggerFromSpec(c.triggerSpec)
		b = tableState{
			builder: builder,
			trigger: t,
		}
		c.tables.Set(key, b)
	}
	return b.(tableState).builder
}

func (c *Cache) Table(key flux.GroupKey) (flux.Table, error) {
	b, ok := c.tables.Lookup(key)
	if !ok {
		return nil, fmt.Errorf("table not found with key %v", key)
	}
	return b.(tableState).builder.Table()
}

func (c *Cache) ForEach(fn func(key flux.GroupKey)) {
	c.tables.Range(func(key flux.GroupKey, _ interface{}) {
		fn(key)
	})
}

func (c *Cache) ForEachWithContext(fn func(key flux.GroupKey, trigger execute.Trigger, bc execute.TableContext)) {
	c.tables.Range(func(key flux.GroupKey, value interface{}) {
		b := value.(tableState)
		fn(key, b.trigger, execute.TableContext{
			Key:   key,
			Count: b.builder.NRows(),
		})
	})
}

func (*Cache) DiscardTable(flux.GroupKey) {
	//panic("implement me")
}

func (*Cache) ExpireTable(flux.GroupKey) {
	//panic("implement me")
}

func (c *Cache) SetTriggerSpec(t flux.TriggerSpec) {
	c.triggerSpec = t
}
