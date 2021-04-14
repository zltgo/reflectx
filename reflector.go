package reflectx

import (
	"encoding/json"
	"errors"
	"reflect"

	"gopkg.in/mgo.v2/bson"
)

var (
	StructNameKey = "_struct_name"
)

// json.Unmarshal or xml.Unmarshal only can decode bytes to map[]interface{},
// the type of interface{} contained in map  still is map[]intefface{}.
// Reflector can decode bytes to registed structs.
type Reflector interface {
	Register(v interface{})
	Encode(v interface{}) ([]byte, error)
	Decode(b []byte) (interface{}, error)
}

type reflector struct {
	mapper    *Mapper
	marshal   func(v interface{}) ([]byte, error)
	unmarshal func(b []byte, v interface{}) error
}

// opts can set format and tagName
// default format is json
// default tagName is "reflector"
// default tagFunc is StdTagfunc
func NewReflector(format, tagName string, tagFunc TagFunc) Reflector {
	if format == "" {
		format = "indentedjson"
	}
	if tagName == "" {
		tagName = "reflector"
	}
	r := reflector{
		mapper: NewMapper(tagName, tagFunc),
	}

	switch format {
	case "bson":
		r.marshal = bson.Marshal
		r.unmarshal = bson.Unmarshal
	//xml: unsupported type: map[string]interface {}
	//case "xml":
	//r.marshal = xml.Marshal
	//r.unmarshal = xml.Unmarshal
	case "json":
		r.marshal = json.Marshal
		r.unmarshal = json.Unmarshal
	case "indentedjson":
		r.marshal = func(v interface{}) ([]byte, error) {
			return json.MarshalIndent(v, "", "    ")
		}
		r.unmarshal = json.Unmarshal
	default:
		panic("unknown format name: " + format)
	}
	return r
}

// Before Decode, Register or Encode must be called first.
// v surpossed to have a struct type.
func (r reflector) Register(v interface{}) {
	r.mapper.TypeMap(Deref(reflect.TypeOf(v)))
}

// obj can be a struct or map[string]interface{}
func (r reflector) Encode(obj interface{}) ([]byte, error) {
	typ, val := Indirect(obj)
	if typ.Kind() == reflect.Struct {
		fi := r.mapper.TypeMap(typ).Tree
		rv := r.encode(fi, val)
		return r.marshal(rv)
	}

	if mp, ok := obj.(map[string]interface{}); ok {
		rv := make(map[string]interface{}, len(mp))
		for k, v := range mp {
			_, vv := Indirect(v)
			rv[k] = r.encode(nil, vv)
		}
		return r.marshal(rv)
	}

	return nil, errors.New("reflectx: expected struct or map[string]interface{} type, got " + typ.String())
}

// Decode decodes bytes to a struct.
// Before Decode, Register or Encode must be called first.
func (r reflector) Decode(b []byte) (interface{}, error) {
	var mp map[string]interface{}
	if err := r.unmarshal(b, &mp); err != nil {
		return nil, err
	}

	return r.decode(mp)
}

// decode decodes a interface{} to reflect.Value.
// The name of the struct store in the map with StructNameKey.
func (r reflector) decode(val interface{}) (rv interface{}, err error) {
	switch vv := val.(type) {
	case map[string]interface{}:
		if len(vv) == 0 {
			return vv, nil
		}
		structName := getStructName(vv)
		if structName == "" {
			mp := make(map[string]interface{}, len(vv))
			for k, v := range vv {
				if mp[k], err = r.decode(v); err != nil {
					return nil, err
				}
			}
			return mp, nil
		}
		sm, ok := r.mapper.NameMap(structName)
		if !ok {
			return nil, errors.New("reflectx: unknown struct name: " + structName)
		}
		return r.mapToStruct(vv, &sm, "")
	case []interface{}:
		if len(vv) == 0 {
			return vv, nil
		}
		s := make([]interface{}, len(vv))
		for i := 0; i < len(vv); i++ {
			if s[i], err = r.decode(vv[i]); err != nil {
				return nil, err
			}
		}
		return s, nil
	default:
		return vv, nil
	}
}

// mapToStruct converts a map[string]interface{} to a struct.
// The name of the struct store in the map with StructNameKey.
func (r reflector) mapToStruct(mp map[string]interface{}, sm *StructMap, prefix string) (interface{}, error) {
	structV := Alloc(sm.Tree.Type)
	structT := Deref(structV.Type())
	for k, v := range mp {
		if k == StructNameKey {
			continue
		}
		if v == nil {
			// nil field
			continue
		}

		if prefix != "" {
			k = prefix + "." + k
		}

		fi := sm.GetByPath(k)
		if fi == nil {
			return nil, errors.New("reflectx: " + k + " is not a path in struct " + structT.String())
		}

		fieldV := FieldByIndexes(structV, fi.Index)
		field, err := r.decode(v)
		if err != nil {
			return nil, err
		}
		if err := SetValue(fieldV, reflect.ValueOf(field)); err != nil {
			return nil, err
		}
	}
	return structV.Interface(), nil
}

//encode encodes a struct to map[string]interface{}, mark the name of the struct with StructNameKey.
func (r reflector) encode(fi *FieldInfo, val reflect.Value) interface{} {
	if val.Kind() == reflect.Interface {
		//get the real type of the interface
		val = reflect.ValueOf(val.Interface())
	}

	fT, fV := getField(fi, val)
	//skip invalid value or nil ptr.
	if !fV.IsValid() {
		return nil
	}

	switch fT.Kind() {
	case reflect.Struct:
		if fi == nil {
			fi = r.mapper.TypeMap(fT).Tree
		}
		mp := make(map[string]interface{}, len(fi.Children)+1)
		mp[StructNameKey] = fT.String()

		for _, child := range fi.Children {
			if elem := r.encode(child, val); elem != nil {
				mp[child.Name] = elem
			}
		}
		return mp
	case reflect.Slice:
		numElems := fV.Len()
		elemT := Deref(fT.Elem())
		if numElems == 0 {
			return nil
		}
		var elemFi *FieldInfo
		if elemT.Kind() == reflect.Struct {
			elemFi = r.mapper.TypeMap(elemT).Tree
		}
		var slice []interface{}
		for i := 0; i < numElems; i++ {
			// append nil elem or not?
			elem := r.encode(elemFi, fV.Index(i))
			slice = append(slice, elem)
		}
		return slice
	case reflect.Map:
		keyT := fT.Key()
		elemT := Deref(fT.Elem())
		numElems := fV.Len()
		if numElems == 0 {
			return nil
		}

		//convert to map[key]interface{}
		mapType := reflect.MapOf(keyT, reflect.TypeOf((*interface{})(nil)).Elem())
		mp := reflect.MakeMapWithSize(mapType, numElems)
		var elemFi *FieldInfo
		if elemT.Kind() == reflect.Struct {
			elemFi = r.mapper.TypeMap(elemT).Tree
		}

		for _, k := range fV.MapKeys() {
			// append nil elem or not?
			elem := r.encode(elemFi, fV.MapIndex(k))
			mp.SetMapIndex(k, reflect.ValueOf(elem))
		}
		return mp.Interface()
	default:
		return fV.Interface()
	}
}

func getField(fi *FieldInfo, val reflect.Value) (fT reflect.Type, fV reflect.Value) {
	if fi != nil {
		fV = FieldByIndexesReadOnly(val, fi.Index)
	} else {
		fV = val
	}
	fV = reflect.Indirect(fV)
	//skip invalid value or nil ptr.
	if fV.IsValid() {
		fT = fV.Type()
		if fT.Kind() == reflect.Interface {
			//get the real type of the interface, fV may be invalid because fV.Interface() may be nil.
			fV = reflect.Indirect(reflect.ValueOf(fV.Interface()))
			fT = fV.Type()
		}
	}
	return
}

func getStructName(mp map[string]interface{}) (name string) {
	structName := mp[StructNameKey]
	if structName != nil {
		name, _ = structName.(string)
	}
	return name
}
