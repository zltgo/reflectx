package values

import (
	"github.com/zltgo/reflectx"
)

type Getter interface {
	// Get value from values, it returns nil in case of non-existed.
	Get(key string) interface{}
}

type Setter interface {
	// Set sets the key to value. It replaces any existing values.
	Set(key string, value interface{})
}

type Updater interface {
	// Get the value of key, use it as input parameter to call fn.
	// If the return value of fn is not nil, save it with key.6
	// Update returns the value of key in the end.
	Update(key string, fn func(interface{}) interface{}) interface{}
}

type Ranger interface {
	// Range calls f sequentially for each key and value present in the map.
	// If f returns false, range stops the iteration.
	Range(f func(key string, value interface{}) bool)
}

type Values interface {
	Getter
	Setter
	Updater
	Ranger

	// ValueOf gets the reflectx.Value associated with the given key.
	// If there are no values associated with the key, ValueOf should return
	// a nil Value.
	ValueOf(key string) reflectx.Value
	Encode() ([]byte, error)
	Decode([]byte) error

	// Removes the values associated with key.
	Remove(key string)
	RemoveAll()
	Len() int
}

// GetValue gets the value associated with the given key.
// If there are no values associated with the key, GetValue returns
// a nil value.
func Get(vs Getter, key string) reflectx.Value {
	return reflectx.ValueOf(vs.Get(key))
}

// Appends all the values from src, it replaces any existing values.
func Append(dst Setter, src Ranger) {
	src.Range(func(key string, val interface{}) bool {
		dst.Set(key, val)
		return true
	})
}

// If key does not exsit, create a new one by constructor.
func Getsert(vs Updater, key string, constructor func() interface{}) interface{} {
	return vs.Update(key, func(v interface{}) interface{} {
		// if the key dose not exist, create a new one
		if v == nil {
			return constructor()
		}
		// the key does exist, do nothing
		return nil
	})
}

// Plus adds step to v and returns the result.
// The return value has the same type with the value associated with the given key.
// It panics if step is not numberical.
// If v is nil, step will be returned.
func Plus(vs Updater, key string, step interface{}) (result reflectx.Value) {
	vs.Update(key, func(old interface{}) interface{} {
		result = reflectx.ValueOf(old).Plus(step)
		return result.Interface()
	})
	return result
}

// AddInt adds step to key,  returns the result as int.
// If the value of key is not exist or other error, value is set to step.
func AddInt(vs Updater, key string, step int) int {
	return Plus(vs, key, step).MustInt()
}

// AddInt64 adds step to key,  returns the result as int64.
// If the value of key is not exist or other error, value is set to step.
func AddInt64(vs Updater, key string, step int64) int64 {
	return Plus(vs, key, step).MustInt64()
}

// AddUint adds step to key,  returns the result as uint.
// If the value of key is not exist or other error, value is set to step.
func AddUint(vs Updater, key string, step uint) uint {
	return Plus(vs, key, step).MustUint()
}

// AddUint64 adds step to key,  returns the result as uint64.
// If the value of key is not exist or other error, value is set to step.
func AddUint64(vs Updater, key string, step uint64) uint64 {
	return Plus(vs, key, step).MustUint64()
}
