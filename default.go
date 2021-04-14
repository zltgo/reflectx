package reflectx

import (
	"reflect"
	"strings"
	"sync"
)

var (
	defaultMapper = NewDefaultMapper("", DefaultTagFunc)
)

// DefaultTagFunc parses  tag parts spliting by "," for the given field.
// Unlike StdTagFunc, the first part is not the field name.
// It simply returns the input fieldName.
func DefaultTagFunc(fieldName, tag string) (string, []string) {
	if tag == IgnoreThisField {
		return IgnoreThisField, nil
	}
	// we can't ignore empty tag, because this field may be a child struct.
	if tag == "" {
		return fieldName, nil
	}
	return fieldName, strings.Split(tag, ",")
}

// the index and default value of the field
type defaultField struct {
	index []int
	// it is not the same as reflect.Zero, see Alloc
	// if this is a pointer and it's nil, allocate a new value and set it
	zero       reflect.Value
	defaultVal reflect.Value
}

// the default values of the fields in the struct
type defaultStruct struct {
	fields     []*defaultField
	zero       reflect.Value
	defaultVal reflect.Value
}

type DefaultMapper struct {
	mapper *Mapper
	cache  sync.Map //map[reflect.Type]defaultStruct
}

func NewDefaultMapper(tagName string, tagFunc TagFunc) *DefaultMapper {
	if tagName == "" {
		tagName = "default"
	}
	if tagFunc == nil {
		tagFunc = DefaultTagFunc
	}
	return &DefaultMapper{
		mapper: NewMapper(tagName, tagFunc),
	}
}

// Set fileds of struct to default value provided by tag if they are not zero.
//	type Test struct {
//		Field string `default:"some string"`
//		Slice []int `default:"1,2,3,4,5"`
//		Map map[string]int`default:"x=1,y=2,z=3"`
//	}
// Note that the zero value of slice is nil, not []int{}.
func SetDefault(ptr interface{}) {
	defaultMapper.SetDefault(ptr)
}

// Allocate a struct with default value, faster than SetDefault.
func AllocDefault(typ reflect.Type) reflect.Value {
	return defaultMapper.AllocDefault(typ)
}

func (dm *DefaultMapper) SetDefault(ptr interface{}) {
	val := reflect.ValueOf(ptr).Elem()
	typ := val.Type()

	ds := dm.getDefaultStruct(typ)
	if reflect.DeepEqual(ds.zero.Interface(), val.Interface()) {
		val.Set(ds.defaultVal)
		return
	}

	for _, df := range ds.fields {
		v := val.Field(df.index[0])
		switch v.Kind() {
		case reflect.Ptr:
			if v.IsNil() {
				v.Set(df.defaultVal)
			} else if v.Elem().Kind() == reflect.Struct {
				// skip the non-nil point of normal kind
				dm.SetDefault(v.Interface())
			}
		case reflect.Struct:
			dm.SetDefault(v.Addr().Interface())
		default:
			// skip the non-zero field, set it to default value
			if reflect.DeepEqual(df.zero.Interface(), v.Interface()) {
				v.Set(df.defaultVal)
			}
		}
	}
	return
}

// Allocate a struct with default value, faster than SetDefault.
func (dm *DefaultMapper) AllocDefault(typ reflect.Type) reflect.Value {
	v := Alloc(typ)
	ds := dm.getDefaultStruct(Deref(typ))
	reflect.Indirect(v).Set(ds.defaultVal)
	return v
}

// Get default info of the struct specified by typ.
func (dm *DefaultMapper) getDefaultStruct(typ reflect.Type) defaultStruct {
	MustBe(typ, reflect.Struct)

	mapping, ok := dm.cache.Load(typ)
	if ok {
		return mapping.(defaultStruct)
	}

	// if the struct info not found in the cache, create and cache it.
	structMap := dm.mapper.TypeMap(typ)
	ds := defaultStruct{
		fields:     make([]*defaultField, 0),
		defaultVal: reflect.New(typ).Elem(),
		zero:       reflect.Zero(typ),
	}

	for _, fi := range structMap.Tree.Children {
		fT := Deref(fi.Type)
		if fT.Kind() != reflect.Struct && len(fi.Parts) == 0 {
			// skip the nromal field which doesn't have default tag.
			continue
		}

		df := &defaultField{
			index:      fi.Index,
			zero:       fi.Zero,
			defaultVal: Alloc(fi.Type),
		}
		dv := reflect.Indirect(df.defaultVal)
		// set default value to df.defaultVal
		switch fT.Kind() {
		case reflect.Struct:
			child := dm.getDefaultStruct(fT)
			if len(child.fields) == 0 {
				// In this case, the child struct doesn't have any default tags.
				// So let's skip it.
				continue
			}
			dv.Set(child.defaultVal)
		case reflect.Map:
			dv.Set(reflect.MakeMapWithSize(fT, len(fi.Options)))
			for k, v := range fi.Options {
				// parse key of map
				kv := Alloc(fT.Key())
				if err := StrToValue(k, kv); err != nil {
					panic(err)
				}
				// parse value of map
				vv := Alloc(fT.Elem())
				if err := StrToValue(v, vv); err != nil {
					panic(err)
				}
				dv.SetMapIndex(kv, vv)
			}
		case reflect.Slice:
			numElems := len(fi.Parts)
			dv.Set(reflect.MakeSlice(fT, numElems, numElems))
			// only the key of Options is useful.
			for i := range fi.Parts {
				if err := StrToValue(fi.Parts[i], dv.Index(i)); err != nil {
					panic(err)
				}
			}
		default:
			if err := StrToValue(fi.Parts[0], dv); err != nil {
				panic(err)
			}
		}

		//save default field to default struct
		FieldByIndexes(ds.defaultVal, df.index).Set(df.defaultVal)
		ds.fields = append(ds.fields, df)
	}

	// cache and return
	dm.cache.Store(typ, ds)
	return ds
}
