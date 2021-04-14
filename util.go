package reflectx

import (
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"unicode"
)

// Alloc allocate a reflect.Value which has the reflect.Type specified by t.
// The return value can always be set either t is a point or not.
func Alloc(t reflect.Type) (v reflect.Value) {
	if t.Kind() == reflect.Ptr {
		return reflect.New(t.Elem())
	}
	return reflect.New(t).Elem()
}

// IndirectAlloc returns the value that v points to.
// It is used before v is going to set value.
// If v is a nil pointer, IndirectAlloc returns a allocated Value.
func AllocIndirect(v reflect.Value) reflect.Value {
	// if this is a pointer and it's nil, allocate a new value and set it
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			// v.Elem().Type will panic because v.Elem() is zero.
			v.Set(reflect.New(v.Type().Elem()))
		}
		return v.Elem()
	}
	//note that the slice or map type is still nil
	return v
}

// Deref is Indirect for reflect.Types
func Deref(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

func Indirect(v interface{}) (reflect.Type, reflect.Value) {
	return Deref(reflect.TypeOf(v)), reflect.Indirect(reflect.ValueOf(v))
}

type kinder interface {
	Kind() reflect.Kind
}

// mustBe checks a value against a kind, panicing with a reflect.ValueError
// if the kind isn't that which is required.
func MustBe(v kinder, expected reflect.Kind) {
	if k := v.Kind(); k != expected {
		panic(&reflect.ValueError{Method: methodName(), Kind: k})
	}
}

// methodName returns the caller of the function calling methodName
func methodName() string {
	pc, _, _, _ := runtime.Caller(2)
	f := runtime.FuncForPC(pc)
	if f == nil {
		return "unknown method"
	}
	return f.Name()
}

// Convert the value of field to string
func ValueToStr(field reflect.Value) (string, error) {
	if field.Kind() == reflect.Ptr && field.IsNil() {
		return "", nil
	}

	v := reflect.Indirect(field)
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10), nil
	case reflect.Bool:
		return strconv.FormatBool(v.Bool()), nil
	case reflect.Float32:
		return strconv.FormatFloat(v.Float(), 'g', -1, 32), nil
	case reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'g', -1, 64), nil
	case reflect.String:
		return v.String(), nil
	default:
		return "", errors.New("reflectx: unexpected type for ValueToStr: " + field.Type().String())
	}
}

// Set value of field by string, field must can be set.
func StrToValue(str string, field reflect.Value) (err error) {
	// empty string means nil point
	if field.Kind() == reflect.Ptr && str == "" {
		field.Set(reflect.Zero(field.Type()))
		return nil
	}

	v := AllocIndirect(field)
	switch v.Kind() {
	case reflect.Int:
		err = setIntField(str, 0, v)
	case reflect.Int8:
		err = setIntField(str, 8, v)
	case reflect.Int16:
		err = setIntField(str, 16, v)
	case reflect.Int32:
		err = setIntField(str, 32, v)
	case reflect.Int64:
		err = setIntField(str, 64, v)
	case reflect.Uint:
		err = setUintField(str, 0, v)
	case reflect.Uint8:
		err = setUintField(str, 8, v)
	case reflect.Uint16:
		err = setUintField(str, 16, v)
	case reflect.Uint32:
		err = setUintField(str, 32, v)
	case reflect.Uint64:
		err = setUintField(str, 64, v)
	case reflect.Bool:
		err = setBoolField(str, v)
	case reflect.Float32:
		err = setFloatField(str, 32, v)
	case reflect.Float64:
		err = setFloatField(str, 64, v)
	case reflect.String:
		v.SetString(str)
	default:
		return errors.New("reflectx: unexpected type for StrToValue: " + field.Type().String())
	}

	if err != nil {
		return errors.New("reflectx: " + err.Error())
	}
	return nil
}

// UnderscoreToCamelCase converts from underscore separated form to camel case form.
// Ex.: my_func => MyFunc
func UnderscoreToCamelCase(s string) string {
	return strings.Replace(strings.Title(strings.Replace(strings.ToLower(s), "_", " ", -1)), " ", "", -1)
}

// CamelCaseToUnderscore converts from camel case form to underscore separated form.
// Ex.: MyFunc => my_func
func CamelCaseToUnderscore(str string) string {
	var output []rune
	var segment []rune
	for _, r := range str {

		// not treat number as separate segment
		if !unicode.IsLower(r) && string(r) != "_" && !unicode.IsNumber(r) {
			output = addSegment(output, segment)
			segment = nil
		}
		segment = append(segment, unicode.ToLower(r))
	}
	output = addSegment(output, segment)
	return string(output)
}

func addSegment(inrune, segment []rune) []rune {
	if len(segment) == 0 {
		return inrune
	}
	if len(inrune) != 0 {
		inrune = append(inrune, '_')
	}
	inrune = append(inrune, segment...)
	return inrune
}

func setIntField(s string, bitSize int, field reflect.Value) error {
	intVal, err := parseInt(s, bitSize)
	if err == nil {
		field.SetInt(intVal)
	}
	return err
}

func setUintField(s string, bitSize int, field reflect.Value) error {
	uintVal, err := parseUint(s, bitSize)
	if err == nil {
		field.SetUint(uintVal)
	}
	return err
}

func setBoolField(s string, field reflect.Value) error {
	boolVal, err := strconv.ParseBool(s)
	if err == nil {
		field.SetBool(boolVal)
	}
	return err
}

func setFloatField(s string, bitSize int, field reflect.Value) error {
	floatVal, err := strconv.ParseFloat(s, bitSize)
	if err == nil {
		field.SetFloat(floatVal)
	}
	return err
}

// IsStruct returns whether v or point of v is a struct.
func IsStruct(v interface{}) bool {
	return reflect.Indirect(reflect.ValueOf(v)).Kind() == reflect.Struct
}

// set dst to field, auto convert types.
// interface{} --> base type(int, uint, float32....)
// float, int, uint... ->base type
// []interface{}, []int, []float... -> []int
// map[int]float... ---> map[float]float
func SetValue(field, v reflect.Value) error {
	v = reflect.Indirect(v)
	// get real type of  the interface{} points to.
	if v.Kind() == reflect.Interface {
		v = reflect.ValueOf(v.Interface())
	}
	if !v.IsValid() {
		//do nothing if v is nil or invalid
		return nil
	}

	field = AllocIndirect(field)
	fT := field.Type()
	if fT == v.Type() || field.Kind() == reflect.Interface {
		field.Set(v)
		return nil
	}

	switch fT.Kind() {
	case reflect.Struct:
		if fT != v.Type() {
			return fmt.Errorf("reflectx: type mismatch, expected %v, got %v", fT, v.Type())
		}
		field.Set(v)
	case reflect.Slice:
		if v.Kind() != reflect.Slice {
			return fmt.Errorf("reflectx: type mismatch, expected %v, got %v", fT, v.Type())
		}
		if v.IsNil() {
			return nil
		}

		numElems := v.Len()
		s := reflect.MakeSlice(field.Type(), numElems, numElems)
		for i := 0; i < numElems; i++ {
			if err := SetValue(s.Index(i), v.Index(i)); err != nil {
				return err
			}
		}
		field.Set(s)
	case reflect.Map:
		if v.Kind() != reflect.Map {
			return fmt.Errorf("reflectx: type mismatch, expect %v, got %v", fT, v.Type())
		}
		if v.IsNil() {
			return nil
		}

		keyT := Deref(fT.Key())
		elemT := Deref(fT.Elem())
		numElems := v.Len()

		//convert v to map[key]elem
		mapType := reflect.MapOf(keyT, elemT)
		mp := reflect.MakeMapWithSize(mapType, numElems)
		for _, vk := range v.MapKeys() {
			ve := v.MapIndex(vk)
			key := reflect.New(keyT).Elem()
			elem := reflect.New(elemT).Elem()
			// set key and elem
			if err := SetValue(key, vk); err != nil {
				return err
			}
			if err := SetValue(elem, ve); err != nil {
				return err
			}
			mp.SetMapIndex(key, elem)
		}
		field.Set(mp)
	default:
		str, err := ValueToStr(v)
		if err != nil {
			return err
		}
		return StrToValue(str, field)
	}

	return nil
}

const minUint53 = 0
const maxUint53 = 4503599627370495
const minInt53 = -2251799813685248
const maxInt53 = 2251799813685247

func floatToUint(f float64) (n uint64, err error) {
	n = uint64(f)
	if float64(n) == f && n >= minUint53 && n <= maxUint53 {
		return n, nil
	}
	return 0, fmt.Errorf("cann't convert %v to uint", f)
}

func floatToInt(f float64) (n int64, err error) {
	n = int64(f)
	if float64(n) == f && n >= minInt53 && n <= maxInt53 {
		return n, nil
	}
	return 0, fmt.Errorf("cann't convert %v to int", f)
}

// paresInt returns a int from string, including float values like "10.000".
func parseInt(s string, bitSize int) (int64, error) {
	i, err := strconv.ParseInt(s, 0, bitSize)
	if err != nil {
		// try again if s is a float value like "10.000"
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return floatToInt(f)
		}
	}
	return i, err
}

// paresInt returns a int from string, including float values like "10.000".
func parseUint(s string, bitSize int) (uint64, error) {
	//n, err = strconv.ParseUint(s, 0, bitSize)
	//connot convert "-1", for example:
	// vs.Set("-1", -1)
	// fmt.Print(vs.ValueOf("-1").Uint()
	// if vs is form, it returns an error.
	n, err := strconv.ParseInt(s, 0, bitSize)
	if err != nil {
		// try again if s is a float value like "10.000"
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return floatToUint(f)
		}
	}
	return uint64(n), err
}
