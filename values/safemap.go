package values

import (
	"encoding/json"
	"sync"

	"github.com/zltgo/reflectx"
	"gopkg.in/mgo.v2/bson"
)

var _ Values = &SafeMap{}

type SafeMap struct {
	marshal   func(v interface{}) (b []byte, err error)
	unmarshal func(b []byte, v interface{}) error

	data map[string]interface{}
	mu   sync.Mutex
}

func NewSafeMap(typ string, data map[string]interface{}) *SafeMap {
	sm := &SafeMap{data: data}

	switch typ {
	case "bson":
		sm.marshal = bson.Marshal
		sm.unmarshal = bson.Unmarshal
	//xml: unsupported type: map[string]interface {}
	//case "xml":
	//r.marshal = xml.Marshal
	//r.unmarshal = xml.Unmarshal
	case "json":
		sm.marshal = json.Marshal
		sm.unmarshal = json.Unmarshal
	default:
		panic("unknown type name: " + typ)
	}
	return sm
}

func (m *SafeMap) Encode() ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.marshal(m.data)
}

func (m *SafeMap) Decode(b []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.unmarshal(b, m.data)
}

//safe call
func (m *SafeMap) LockGuard(f func(data map[string]interface{})) {
	m.mu.Lock()
	defer m.mu.Unlock()
	//may panic
	f(m.data)
}

// Get value from values, it returns nil in case of non-existed.
func (m *SafeMap) Get(key string) interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.data) == 0 {
		return nil
	}
	v, _ := m.data[key]
	return v
}

// ValueOf gets the reflect.Value associated with the given key.
// If there are no values associated with the key, ValueOf should return
// a nil Value.
func (m *SafeMap) ValueOf(key string) reflectx.Value {
	return reflectx.ValueOf(m.Get(key))
}

// Set sets the key to value. It replaces any existing values.
func (m *SafeMap) Set(key string, value interface{}) {
	m.mu.Lock()
	//lazy init
	if m.data == nil {
		m.data = make(map[string]interface{})
	}
	m.data[key] = value
	m.mu.Unlock()
}

// Get the value of key, use it as input parameter to call fn.
// If the return value of fn is not nil, save it with key.
// Update returns the value of key in the end.
func (m *SafeMap) Update(key string, fn func(interface{}) interface{}) interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	// get old value
	var oldV interface{}
	if len(m.data) != 0 {
		oldV, _ = m.data[key]
	}

	// get new value
	newV := fn(oldV)
	if newV == nil {
		//nothing changed
		return oldV
	}

	// set new value
	if m.data == nil {
		m.data = make(map[string]interface{})
	}
	m.data[key] = newV
	return newV
}

// Del deletes the values associated with key.
func (m *SafeMap) Remove(key string) {
	// it's ok to delete nil map.
	m.mu.Lock()
	delete(m.data, key)
	m.mu.Unlock()
}

func (m *SafeMap) RemoveAll() {
	m.mu.Lock()
	m.data = nil
	m.mu.Unlock()
}

func (m *SafeMap) Len() int {
	m.mu.Lock()
	l := len(m.data)
	m.mu.Unlock()
	return l
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
func (m *SafeMap) Range(f func(key string, value interface{}) bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// it's ok to use range when the map is nil.
	for k, v := range m.data {
		if f(k, v) == false {
			return
		}
	}
}
