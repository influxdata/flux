package metadata

import "fmt"

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

func (md Metadata) Get(key string) (interface{}, error) {
	if values, ok := md[key]; ok && len(values) != 0 {
		return values[0], nil
	}
	return nil, fmt.Errorf("key %s does not exist in Metadata", key)
}

func (md Metadata) GetAll(key string) []interface{} {
	if values, ok := md[key]; ok {
		return values
	}
	return []interface{}{}
}
