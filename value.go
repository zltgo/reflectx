package reflectx

import (
	"errors"
	"fmt"
	"strconv"
)

var (
	ErrNil = errors.New("nil value")
)

// Value represents a value that is returned from Get().
type Value struct {
	v interface{}
}

//ValueOf returns the Value of v.
func ValueOf(v interface{}) Value {
	return Value{v}
}

// Interface returns v's current value as an interface{}.
// It is equivalent to:
//	var i interface{} = (v's underlying value)
func (v Value) Interface() interface{} {
	return v.v
}

// IsNill returns true if v is nil.
func (v Value) IsNil() bool {
	return v.v == nil
}

// Default sets defaultVal to v if v is nil.
func (v Value) Default(defaultVal interface{}) Value {
	if v.v == nil {
		return Value{defaultVal}
	}
	return v
}

// Plus adds step to v and returns the result.
// The return value has the same type as v.
// It panics if step is not numberical.
// If v is nil, step will be returned.
func (v Value) Plus(step interface{}) Value {
	i := Value{step}.MustInt64()
	var result interface{}

	switch val := v.v.(type) {
	case nil:
		result = step
	case int:
		result = int(int64(val) + i)
	case int8:
		result = int8(int64(val) + i)
	case int16:
		result = int16(int64(val) + i)
	case int32:
		result = int32(int64(val) + i)
	case int64:
		result = val + i
	case uint:
		result = uint(int64(val) + i)
	case uint8:
		result = uint8(int64(val) + i)
	case uint16:
		result = uint16(int64(val) + i)
	case uint32:
		result = uint32(int64(val) + i)
	case uint64:
		result = uint64(int64(val) + i)
	case float32:
		if orig, err := floatToInt(float64(val)); err == nil {
			result = float32(orig + i)
		} else {
			result = step
		}
	case float64: //if m is decode by json, evey number has float64 type
		if orig, err := floatToInt(val); err == nil {
			result = float64(orig + i)
		} else {
			result = step
		}
	case string:
		if orig, err := parseInt(val, 64); err == nil {
			result = fmt.Sprint(orig + i)
		} else {
			result = step
		}
	default:
		result = step
	}

	return Value{result}
}

// String returns string from v.
// It returns "" if v is nil.
func (v Value) String() string {
	if v.v == nil {
		return ""
	}
	return fmt.Sprint(v.v)
}

// ToInt64 returns an int64 number from v.
// It returns ErrNil if v is nil.
// It returns an error if the type of value is not numeric.
// Note that, if v is decode by json, evey number has float64 type.
// If v is decode by bson, number only has int, int64 type.
func (v Value) ToInt64() (int64, error) {
	switch val := v.v.(type) {
	case nil:
		return 0, ErrNil
	case int:
		return int64(val), nil
	case int8:
		return int64(val), nil
	case int16:
		return int64(val), nil
	case int32:
		return int64(val), nil
	case int64:
		return val, nil
	case uint:
		return int64(val), nil
	case uint8:
		return int64(val), nil
	case uint16:
		return int64(val), nil
	case uint32:
		return int64(val), nil
	case uint64:
		return int64(val), nil
	case float32:
		return floatToInt(float64(val))
	case float64: //if m is decode by json, evey number has float64 type
		return floatToInt(val)
	case string:
		return parseInt(val, 64)
	default:
		return 0, fmt.Errorf("can't get int from %v", val)
	}
}

//Int64 returns zero in case of any error.
func (v Value) Int64() int64 {
	i, _ := v.ToInt64()
	return i
}

//MustInt64 returns the int64 format number, it panics in case of any error.
func (v Value) MustInt64() int64 {
	i, err := v.ToInt64()
	if err != nil {
		panic(err)
	}
	return i
}

// ToInt returns an int number from v.
// It returns ErrNil if v is nil.
// It returns an error if the type of value is not numeric.
// Note that, if v is decode by json, evey number has float64 type.
// If v is decode by bson, number only has int, int64 type.
func (v Value) ToInt() (int, error) {
	i, err := v.ToInt64()
	return int(i), err
}

//Int returns zero in case of any error.
func (v Value) Int() int {
	i, _ := v.ToInt64()
	return int(i)
}

// MustInt returns the int format number, it panics in case of any error.
func (v Value) MustInt() int {
	i, err := v.ToInt64()
	if err != nil {
		panic(err)
	}
	return int(i)
}

// ToUint64 returns ErrNil if v is nil.
// It returns an error if the  type of value is not numeric.
// Note that, if v is decode by json, evey number has float64 type.
// If v is decode by bson, number only has int, int64 type.
func (v Value) ToUint64() (uint64, error) {
	switch val := v.v.(type) {
	case nil:
		return 0, ErrNil
	case int:
		return uint64(val), nil
	case int8:
		return uint64(val), nil
	case int16:
		return uint64(val), nil
	case int32:
		return uint64(val), nil
	case int64:
		return uint64(val), nil
	case uint:
		return uint64(val), nil
	case uint8:
		return uint64(val), nil
	case uint16:
		return uint64(val), nil
	case uint32:
		return uint64(val), nil
	case uint64:
		return val, nil
	case float32:
		return floatToUint(float64(val))
	case float64: //if m is decode by json, evey number has float64 type
		return floatToUint(val)
	case string:
		return parseUint(val, 64)
	default:
		return 0, fmt.Errorf("can't get uint from %v", val)
	}
}

//Uint64 returns zero in case of any error.
func (v Value) Uint64() uint64 {
	n, _ := v.ToUint64()
	return n
}

// MustUint64 returns the uint64 format number, it panics in case of any error.
func (v Value) MustUint64() uint64 {
	n, err := v.ToUint64()
	if err != nil {
		panic(err)
	}
	return n
}

// ToUint returns an int number from v.
// It returns ErrNil if v is nil.
// It returns an error if the type of value is not numeric.
// Note that, if v is decode by json, evey number has float64 type.
// If v is decode by bson, number only has int, int64 type.
func (v Value) ToUint() (uint, error) {
	n, err := v.ToUint64()
	return uint(n), err
}

//Uint returns zero in case of any error.
func (v Value) Uint() uint {
	n, _ := v.ToUint64()
	return uint(n)
}

// MustUint returns the uint format number, it panics in case of any error.
func (v Value) MustUint() uint {
	n, err := v.ToUint64()
	if err != nil {
		panic(err)
	}
	return uint(n)
}

// Float returns returns ErrNil if v is nil.
// It returns an error if the  type of value is not numeric.
// Note that, if map is decode by json, evey number has float64 type.
// If map is decode by bson, number only has int, int64 type.
func (v Value) ToFloat() (float64, error) {
	switch val := v.v.(type) {
	case nil:
		return 0, ErrNil
	case int:
		return float64(val), nil
	case int8:
		return float64(val), nil
	case int16:
		return float64(val), nil
	case int32:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case uint:
		return float64(val), nil
	case uint8:
		return float64(val), nil
	case uint16:
		return float64(val), nil
	case uint32:
		return float64(val), nil
	case uint64:
		return float64(val), nil
	case float32:
		return float64(val), nil
	case float64:
		return val, nil
	case string:
		return strconv.ParseFloat(val, 64)
	default:
		return 0, fmt.Errorf("can't get float from %v", val)
	}
}

// Bool returns false in case of non-existed.
func (v Value) ToBool() (bool, error) {
	s := v.String()
	if s == "" {
		return false, ErrNil
	}
	return strconv.ParseBool(s)
}

// Float returns a float format number, it returns 0.0 in case of any error.
func (v Value) Float() float64 {
	f, _ := v.ToFloat()
	return f
}

// MustFloat returns the float format number, it panics in case of any error.
func (v Value) MustFloat() float64 {
	f, err := v.ToFloat()
	if err != nil {
		panic(err)
	}
	return f
}

// Bool returns a bool format value, it returns 0.0 in case of any error.
func (v Value) Bool() bool {
	b, _ := v.ToBool()
	return b
}

// MustBool returns the bool format value, it panics in case of any error.
func (v Value) MustBool() bool {
	b, err := v.ToBool()
	if err != nil {
		panic(err)
	}
	return b
}
