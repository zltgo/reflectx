package values

import (
	"net/url"
	"reflect"

	"github.com/zltgo/reflectx"
)

var _ Values = &Form{}

type Form map[string][]string

// Get value from values, it returns nil in case of non-existed.
func (m Form) Get(key string) interface{} {
	if len(m) == 0 {
		return nil
	}
	v, ok := m[key]
	if !ok || len(v) == 0 {
		return nil
	}

	//return string instead of []string{"just one element"}
	if len(v) == 1 {
		return v[0]
	}
	return v
}

// ValueOf gets the reflectx.Value associated with the given key.
// If there are no values associated with the key, ValueOf should return
// a nil Value.
func (m Form) ValueOf(key string) reflectx.Value {
	return reflectx.ValueOf(m.Get(key))
}

// Set sets the key to value. It replaces any existing values.
func (m Form) Set(key string, value interface{}) {
	switch v := value.(type) {
	case string:
		m[key] = []string{v}
	case []string:
		m[key] = v
	case []interface{}:
		m.SetValues(key, v...)
	default:
		s, err := reflectx.ValueToStr(reflect.ValueOf(v))
		if err != nil {
			panic(err)
		}
		m[key] = []string{s}
	}
}

// Get the value of key, use it as input parameter to call fn.
// If the return value of fn is not nil, save it with key.
// Update returns the value of key in the end.
func (m Form) Update(key string, fn func(interface{}) interface{}) interface{} {
	old := m.Get(key)
	rv := fn(old)
	if rv != nil {
		m.Set(key, rv)
		return rv
	}
	return old
}

// Add adds the value to key. It appends to any existing
// values associated with key.
func (m Form) AddString(key, value string) {
	m[key] = append(m[key], value)
}

// Del deletes the values associated with key.
func (m Form) Remove(key string) {
	// it's ok to delete nil map.
	delete(m, key)
}

func (m *Form) RemoveAll() {
	// it's ok to delete nil map.
	*m = make(map[string][]string)
}

func (m Form) Len() int {
	return len(m)
}

// Encode encodes the values into URL-encoded form
// ("bar=baz&foo=quux") sorted by key.
func (m Form) Encode() ([]byte, error) {
	return []byte(url.Values(m).Encode()), nil
}

// Decode parses the URL-encoded query string and returns
// a map listing the values specified for each key.
func (m *Form) Decode(b []byte) error {
	values, err := url.ParseQuery(string(b))
	if err != nil {
		return err
	}

	*m = Form(values)
	return nil
}

func (m Form) SetValues(key string, values ...interface{}) {
	if len(values) == 0 {
		return
	}
	m[key] = make([]string, len(values))
	var err error
	for i, v := range values {
		if m[key][i], err = reflectx.ValueToStr(reflect.ValueOf(v)); err != nil {
			panic(err)
		}
	}
}

func (m Form) AddValues(key string, values ...interface{}) {
	for _, v := range values {
		tmp, err := reflectx.ValueToStr(reflect.ValueOf(v))
		if err != nil {
			panic(err)
		}
		m.AddString(key, tmp)
	}
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
func (m Form) Range(f func(key string, value interface{}) bool) {
	// it's ok to use range when the map is nil.
	for k, v := range m {
		if f(k, v) == false {
			return
		}
	}
}
