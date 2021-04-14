package values

import (
	"encoding/json"

	"github.com/zltgo/reflectx"
)

var _ Values = &JsonMap{}

type JsonMap map[string]interface{}

func (m JsonMap) Encode() ([]byte, error) {
	return json.Marshal(m)
}

func (m *JsonMap) Decode(b []byte) error {
	return json.Unmarshal(b, m)
}

// Get value from values, it returns nil in case of non-existed.
func (m JsonMap) Get(key string) interface{} {
	if len(m) == 0 {
		return nil
	}
	v, _ := m[key]
	return v
}

// ValueOf gets the reflectx.Value associated with the given key.
// If there are no values associated with the key, ValueOf should return
// a nil Value.
func (m JsonMap) ValueOf(key string) reflectx.Value {
	return reflectx.ValueOf(m.Get(key))
}

// Set sets the key to value. It replaces any existing values.
func (m JsonMap) Set(key string, value interface{}) {
	m[key] = value
}

// Get the value of key, use it as input parameter to call fn.
// If the return value of fn is not nil, save it with key.
// Update returns the value of key in the end.
func (m JsonMap) Update(key string, fn func(interface{}) interface{}) interface{} {
	old := m.Get(key)
	rv := fn(old)
	if rv != nil {
		m.Set(key, rv)
		return rv
	}
	return old
}

// Del deletes the values associated with key.
func (m JsonMap) Remove(key string) {
	// it's ok to delete nil map.
	delete(m, key)
}

func (m *JsonMap) RemoveAll() {
	*m = make(map[string]interface{})
}

func (m JsonMap) Len() int {
	return len(m)
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
func (m JsonMap) Range(f func(key string, value interface{}) bool) {
	// it's ok to use range when the map is nil.
	for k, v := range m {
		if f(k, v) == false {
			return
		}
	}
}
