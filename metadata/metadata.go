package metadata

import (
	"fmt"
	"sync"
)

// Metadata is made as a standalone package to avoid import cycle:
// influxd -> flux -> flux/interpreter -> flux/execute -> flux
type Metadata struct {
	// Metadata is passed as a value so we must store the mutex as a pointer to ensure it does not get copied
	lock *sync.RWMutex
	meta map[string][]interface{}
}

func NewMetadata() Metadata {
	return Metadata{
		lock: &sync.RWMutex{},
		meta: make(map[string][]interface{}),
	}
}

func (md Metadata) Add(key string, value interface{}) {
	md.lock.Lock()
	defer md.lock.Unlock()

	md.meta[key] = append(md.meta[key], value)
}

func (md Metadata) AddAll(other Metadata) {
	md.lock.Lock()
	defer md.lock.Unlock()

	other.lock.RLock()
	defer other.lock.RUnlock()

	for key, values := range other.meta {
		md.meta[key] = append(md.meta[key], values...)
	}
}

// Range will iterate over the Metadata. It will invoke the function for each
// key/value pair. If there are multiple values for a single key, then this will
// be called with the same key once for each value.
func (md Metadata) Range(fn func(key string, value interface{}) bool) {
	md.RangeSlices(func(key string, values []interface{}) bool {
		for _, value := range values {
			if ok := fn(key, value); !ok {
				return false
			}
		}
		return true
	})
}

func (md Metadata) RangeSlices(fn func(key string, values []interface{}) bool) {
	md.lock.RLock()
	defer md.lock.RUnlock()

	for key, values := range md.meta {
		if ok := fn(key, values); !ok {
			return
		}
	}
}

func (md Metadata) Del(key string) {
	md.lock.Lock()
	defer md.lock.Unlock()

	delete(md.meta, key)
}

func (md Metadata) Get(key string) (interface{}, error) {
	md.lock.RLock()
	defer md.lock.RUnlock()

	if values, ok := md.meta[key]; ok && len(values) != 0 {
		return values[0], nil
	}
	return nil, fmt.Errorf("key %s does not exist in Metadata", key)
}

func (md Metadata) GetAll(key string) []interface{} {
	md.lock.RLock()
	defer md.lock.RUnlock()

	if values, ok := md.meta[key]; ok {
		return values
	}
	return []interface{}{}
}
