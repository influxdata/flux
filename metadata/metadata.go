package metadata

import (
	"sync"
)

// Metadata is made as a standalone package to avoid import cycle:
// influxd -> flux -> flux/interpreter -> flux/execute -> flux
type Metadata map[string][]interface{}

func (md Metadata) Add(key string, value interface{}) {
	md[key] = append(md[key], value)
}

func (md Metadata) AddAll(other Metadata) {
	for key, values := range other {
		md[key] = append(md[key], values...)
	}
}

// Range will iterate over the Metadata. It will invoke the function for each
// key/value pair. If there are multiple values for a single key, then this will
// be called with the same key once for each value.
func (md Metadata) Range(fn func(key string, value interface{}) bool) {
	for key, values := range md {
		for _, value := range values {
			if ok := fn(key, value); !ok {
				return
			}
		}
	}
}

func (md Metadata) Del(key string) {
	delete(md, key)
}

func (md Metadata) Get(key string) (interface{}, bool) {
	if values, ok := md[key]; ok && len(values) != 0 {
		return values[0], true
	}
	return nil, false
}

func (md Metadata) GetAll(key string) []interface{} {
	if values, ok := md[key]; ok {
		return values
	}
	return []interface{}{}
}

// SyncMetadata is a version of `Metadata` which allows concurrent modifications to it
type SyncMetadata struct {
	lock sync.RWMutex
	meta Metadata
}

func NewSyncMetadata() *SyncMetadata {
	return NewSyncMetadataWith(make(Metadata))
}

func NewSyncMetadataWith(meta Metadata) *SyncMetadata {
	return &SyncMetadata{
		meta: meta,
	}
}

func (md *SyncMetadata) Add(key string, value interface{}) {
	md.lock.Lock()
	defer md.lock.Unlock()

	md.meta.Add(key, value)
}

func (md *SyncMetadata) AddAll(other Metadata) {
	md.lock.Lock()
	defer md.lock.Unlock()

	md.meta.AddAll(other)
}

// Range will iterate over the SyncMetadata. It will invoke the function for each
// key/value pair. If there are multiple values for a single key, then this will
// be called with the same key once for each value.
func (md *SyncMetadata) Range(fn func(key string, value interface{}) bool) {
	md.lock.RLock()
	defer md.lock.RUnlock()

	md.meta.Range(fn)
}

func (md *SyncMetadata) Del(key string) {
	md.lock.Lock()
	defer md.lock.Unlock()

	md.meta.Del(key)
}

func (md *SyncMetadata) Get(key string) (interface{}, bool) {
	md.lock.RLock()
	defer md.lock.RUnlock()

	return md.meta.Get(key)
}

func (md *SyncMetadata) GetAll(key string) []interface{} {
	md.lock.RLock()
	defer md.lock.RUnlock()

	return md.meta.GetAll(key)
}

// ReadView provides read access to the underlying `Metadata` map.
// Since the map may be concurrently modified outside of the closure
// it should not be allowed to escape it.
func (md *SyncMetadata) ReadView(fn func(meta Metadata)) {
	md.lock.RLock()
	defer md.lock.RUnlock()

	fn(md.meta)
}

func (md *SyncMetadata) ReadWriteView(fn func(meta *Metadata)) {
	md.lock.Lock()
	defer md.lock.Unlock()

	fn(&md.meta)
}
