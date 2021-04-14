// Package reflectx implements extensions to the standard reflect lib suitable
// for implementing marshalling and unmarshalling packages.  The main Mapper type
// allows for Go-compatible named attribute access, including accessing embedded
// struct attributes and the ability to use  functions and struct tags to
// customize field names.
//
package reflectx

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

var (
	IgnoreThisField = "-"
	//the field is skipped if empty.
	OmitEmpty = "omitempty"
	// Field is not processed further by this package.
	OmitNested = "omitnested"
	// The FieldStruct's fields will be flattened into the parent level.
	Flatten = "flatten"
	// No tag, use the original name of field
	StdMapper = NewMapper("", nil)
)

// A FieldInfo is metadata for a struct field.
type FieldInfo struct {
	Index []int
	Path  string
	IsPtr bool // is a point or not
	Type  reflect.Type
	// It is not the same as reflect.Zero, If this is a pointer, allocate a new value.
	// Zero = reflect.Zero(Deref(Type))
	Zero     reflect.Value
	Name     string
	Parts    []string          //parts of tag splite by ",", exclusive of name.
	Options  map[string]string // options parsed from Parts, as "k=v".
	Embedded bool
	Children []*FieldInfo
	Parent   *FieldInfo
}

func (fi *FieldInfo) StringsToField(strs []string, v reflect.Value) error {
	fV := reflect.Indirect(FieldByIndexes(v, fi.Index))

	if fV.Kind() == reflect.Slice {
		fV.Set(reflect.MakeSlice(fV.Type(), len(strs), len(strs)))
		for i := range strs {
			if err := StrToValue(strs[i], fV.Index(i)); err != nil {
				return fmt.Errorf("reflectx: can not convert %v to %v: %v", strs, v.Type().String()+"."+fi.Path, err)
			}
		}
		return nil
	}

	if len(strs) > 1 {
		return fmt.Errorf("reflectx: can not convert %v to %v", strs, v.Type().String()+"."+fi.Path)
	}
	if err := StrToValue(strs[0], fV); err != nil {
		return err
	}

	return nil
}

// A StructMap is an index of field metadata for a struct.
type StructMap struct {
	Tree   *FieldInfo
	Fields []*FieldInfo          //all the fields of the tree.
	Paths  map[string]*FieldInfo // equal to Fields.
	Leaves map[string]*FieldInfo // all the leaves of the tree, not including struct type.
}

// GetByPath returns a *FieldInfo for a given string path.
func (f StructMap) GetByPath(path string) *FieldInfo {
	if path == "" {
		return f.Tree
	}
	return f.Paths[path]
}

// GetByTraversal returns a *FieldInfo for a given integer path.  It is
// analogous to reflect.FieldByIndex, but using the cached traversal
// rather than re-executing the reflect machinery each time.
func (f StructMap) GetByTraversal(index []int) *FieldInfo {
	if len(index) == 0 {
		return f.Tree
	}

	for _, fi := range f.Fields {
		if indexEqual(index, fi.Index) {
			return fi
		}
	}
	return nil
}

// Mapper is a general purpose mapper of names to struct fields.  A Mapper
// behaves like most marshallers in the standard library, obeying a field tag
// for name mapping but also providing a basic transform function.
type Mapper struct {
	tagName string
	tagFunc TagFunc
	cache   sync.Map //map[reflect.Type]StructMap
}

// the input fieldName is equal to reflect.Field.Name() of the struct.
type TagFunc func(fieldName, tag string) (name string, parts []string)

// StdTagFunc parses the target name and tag parts spliting by "," for the given field.
// e.g.
// Call with "foo,bar,size=64" returns ("foo", []string{"bar", "szie=64"})
// Call with ",bar" returns (fieldName, []string{"bar"})
func StdTagFunc(fieldName, tag string) (string, []string) {
	// if there's no tag to look for, return the field name
	if tag == "" {
		return fieldName, nil
	}

	// splits the parts of tag. The first part supposed to be the target name.
	parts := strings.Split(tag, ",")
	if parts[0] != "" {
		fieldName = parts[0]
	}

	return fieldName, parts[1:]
}

// same as StdTagFunc, but mapped fieldName to their lower case.
func FieldNameToLower(fieldName, tag string) (string, []string) {
	return StdTagFunc(strings.ToLower(fieldName), tag)
}

// Same as StdTagFunc, but mapped fieldName to underscore.
// Ex.: MyFunc => my_func
func FieldNameToUnderscore(fieldName, tag string) (string, []string) {
	return StdTagFunc(CamelCaseToUnderscore(fieldName), tag)
}

// NewMapper returns a new mapper using the tagName as its struct field tag.
// If tagName is the empty string, it is ignored.
// tagFunc is used to get target name and parts of tag from the field.
// if tagFunc is nil, StdTagFunc will be used.
func NewMapper(tagName string, tagFunc TagFunc) *Mapper {
	if tagFunc == nil {
		tagFunc = StdTagFunc
	}
	return &Mapper{
		tagName: tagName,
		tagFunc: tagFunc,
	}
}

// TypeMap returns a mapping of field strings to int slices representing
// the traversal down the struct to reach the field.
func (m *Mapper) TypeMap(t reflect.Type) StructMap {
	MustBe(Deref(t), reflect.Struct)

	if mapping, ok := m.cache.Load(t); ok {
		return mapping.(StructMap)
	}

	mapping := getMapping(t, m.tagName, m.tagFunc)
	m.cache.Store(t, mapping)
	return mapping
}

// NameMap returns StructMap by struct name.
func (m *Mapper) NameMap(name string) (sm StructMap, exist bool) {
	m.cache.Range(func(key, value interface{}) bool {
		if key.(reflect.Type).String() == name {
			sm = value.(StructMap)
			exist = true
			return false
		}
		return true
	})

	return
}

// FieldMap returns the mapper's mapping of field names to reflect values.
// Panics if v's Kind is not Struct, or v is not Indirectable to a struct kind.
func (m *Mapper) FieldMap(v reflect.Value) map[string]reflect.Value {
	v = reflect.Indirect(v)
	tm := m.TypeMap(v.Type())
	r := map[string]reflect.Value{}

	for tagName, fi := range tm.Paths {
		r[tagName] = FieldByIndexes(v, fi.Index)
	}
	return r
}

// FieldByName returns a field by its mapped name as a reflect.Value.
// Panics if v's Kind is not Struct or v is not Indirectable to a struct Kind.
// Returns zero Value if the name is not found.
func (m *Mapper) FieldByPath(v reflect.Value, path string) (rv reflect.Value) {
	tm := m.TypeMap(Deref(v.Type()))
	if fi := tm.GetByPath(path); fi != nil {
		return FieldByIndexes(v, fi.Index)
	}
	return rv
}

// FieldsByName returns a slice of values corresponding to the slice of names
// for the value.  Panics if v's Kind is not Struct or v is not Indirectable
// to a struct Kind.  Returns zero Value for each name not found.
func (m *Mapper) FieldsByPath(v reflect.Value, paths []string) []reflect.Value {
	v = reflect.Indirect(v)
	tm := m.TypeMap(v.Type())
	vals := make([]reflect.Value, 0, len(paths))
	for _, path := range paths {
		fi := tm.GetByPath(path)
		if fi == nil {
			vals = append(vals, *new(reflect.Value))
		} else {
			vals = append(vals, FieldByIndexes(v, fi.Index))
		}
	}
	return vals
}

// TraversalsByName returns a slice of int slices which represent the struct
// traversals for each mapped name.  Panics if t is not a struct or Indirectable
// to a struct.  Returns empty int slice for each name not found.
func (m *Mapper) TraversalsByPath(t reflect.Type, paths []string) [][]int {
	t = Deref(t)
	tm := m.TypeMap(t)

	r := make([][]int, 0, len(paths))
	for _, path := range paths {
		fi, ok := tm.Paths[path]
		if !ok {
			r = append(r, []int{})
		} else {
			r = append(r, fi.Index)
		}
	}
	return r
}

// FieldByIndexes returns a value for the field given by the struct traversal
// for the given value.
func FieldByIndexes(v reflect.Value, indexes []int) reflect.Value {
	for _, i := range indexes {
		v = reflect.Indirect(v).Field(i)
		// if this is a pointer and it's nil, allocate a new value and set it
		if v.Kind() == reflect.Ptr && v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
	}
	return v
}

// FieldByIndexesReadOnly returns a value for a particular struct traversal,
// but is not concerned with allocating nil pointers because the value is
// going to be used for reading and not setting.
// It returns a invalid value in case of v is a nil point.
func FieldByIndexesReadOnly(v reflect.Value, indexes []int) reflect.Value {
	for _, i := range indexes {
		if v = reflect.Indirect(v); v.IsValid() {
			v = v.Field(i)
		} else {
			break
		}
	}
	return v
}

// CopyTo copys fields of src to dst, these fields have the same name.
func CopyStruct(dst, src interface{}) {
	srcT, srcV := Indirect(src)
	// dst must be pointer.
	dstV := reflect.ValueOf(dst).Elem()
	dstT := dstV.Type()

	srcMap := StdMapper.TypeMap(srcT)
	dstMap := StdMapper.TypeMap(dstT)

	if len(srcMap.Leaves) < len(dstMap.Leaves) {
		for k, srcField := range srcMap.Leaves {
			if dstField, ok := dstMap.Leaves[k]; ok {
				srcFieldV := reflect.Indirect(FieldByIndexesReadOnly(srcV, srcField.Index))
				dstFieldV := reflect.Indirect(FieldByIndexes(dstV, dstField.Index))
				dstFieldV.Set(srcFieldV)
			}
		}
	} else {
		for k, dstField := range dstMap.Leaves {
			if srcField, ok := srcMap.Leaves[k]; ok {
				srcFieldV := reflect.Indirect(FieldByIndexesReadOnly(srcV, srcField.Index))
				dstFieldV := reflect.Indirect(FieldByIndexes(dstV, dstField.Index))
				dstFieldV.Set(srcFieldV)
			}
		}
	}

	return
}

type typeQueue struct {
	t      reflect.Type
	fi     *FieldInfo
	pIndex []int //parent index
}

// A copying append that creates a new slice each time.
func apnd(is []int, i int) []int {
	x := make([]int, len(is)+1)
	for p, n := range is {
		x[p] = n
	}
	x[len(x)-1] = i
	return x
}

// indexEqual returns whether a deep equal to b.
func indexEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}

	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// parseOptions parses options out of a tag string, skipping the name
func parseOptions(parts []string) map[string]string {
	if len(parts) == 0 {
		return nil
	}

	options := make(map[string]string, len(parts))
	for _, opt := range parts {
		// short circuit potentially expensive split op
		if strings.Contains(opt, "=") {
			kv := strings.SplitN(opt, "=", 2)
			options[kv[0]] = kv[1]
		} else {
			options[opt] = ""
		}
	}
	return options
}

// getMapping returns a mapping for the t type, using the tagName, mapFunc and
// tagMapFunc to determine the canonical names of fields.
func getMapping(t reflect.Type, tagName string, tagFunc func(string, string) (string, []string)) StructMap {
	root := &FieldInfo{
		IsPtr: t.Kind() == reflect.Ptr,
		Type:  t,
		Zero:  reflect.Zero(Deref(t)),
		Name:  t.String(),
	}
	m := []*FieldInfo{}
	queue := []typeQueue{}
	queue = append(queue, typeQueue{Deref(t), root, nil})

QueueLoop:
	for len(queue) != 0 {
		// pop the first item off of the queue
		tq := queue[0]
		queue = queue[1:]

		// ignore recursive field
		for p := tq.fi.Parent; p != nil; p = p.Parent {
			if tq.fi.Type == p.Type {
				continue QueueLoop
			}
		}

		nChildren := 0
		if tq.t.Kind() == reflect.Struct {
			nChildren = tq.t.NumField()
		}

		// iterate through all of its fields
		for fieldPos := 0; fieldPos < nChildren; fieldPos++ {
			f := tq.t.Field(fieldPos)

			// skip unexported fields
			if len(f.PkgPath) != 0 && !f.Anonymous {
				continue
			}

			// if this tag is not set using the normal convention in the tag,
			// then return the fieldname..  this check is done because according
			// to the reflect documentation:
			// If the tag does not have the conventional format,
			// the value returned by Get is unspecified.
			// which doesn't sound great.
			tag := f.Tag.Get(tagName)
			if !strings.Contains(string(f.Tag), tagName+":") {
				tag = ""
			}

			// parse the tag and the target name using the mapping options for this field
			name, parts := tagFunc(f.Name, tag)

			// if the name is "-" or empty, skip it
			if name == IgnoreThisField || name == "" {
				continue
			}

			fi := &FieldInfo{
				Index:    apnd(tq.pIndex, fieldPos),
				IsPtr:    f.Type.Kind() == reflect.Ptr,
				Type:     f.Type,
				Zero:     reflect.Zero(Deref(f.Type)),
				Name:     name,
				Parts:    parts,
				Options:  parseOptions(parts),
				Embedded: f.Anonymous,
				Parent:   tq.fi,
			}

			// if the path is empty, this path is just the name
			if tq.fi.Path == "" {
				fi.Path = fi.Name
			} else {
				fi.Path = tq.fi.Path + "." + fi.Name
			}

			owner := fi
			// go on mapping fields of child struct
			if Deref(fi.Type).Kind() == reflect.Struct {
				_, ok := fi.Options[Flatten]
				// case one:
				// the child struct has "flatten" tag.
				//	type Foo struct {
				//		B Bar `tagName:"flatten"`
				//	}
				//	type Bar struct {
				//		A int // this path is A, not B.A
				//	}
				//
				if _, ok := fi.Options[Flatten]; ok {
					owner = tq.fi
				}
				// case two:
				// if the child struct is embedded and name is equal to field name,
				// shortcut the name, for example:
				//	type Foo struct {
				//		Bar
				//	}
				//	type Bar struct {
				//		A int // this path is A, not Bar.A
				//	}
				// Parent of A is root.
				if fi.Embedded {
					// tagFunc may be FieldNameToLower or FieldNameToUnderscore.
					defaultName, _ := tagFunc(f.Name, "")
					if defaultName == name {
						owner = tq.fi
					}
				}
				// Field is not processed further by this package in case of OmitNested.
				if _, ok = fi.Options[OmitNested]; !ok {
					queue = append(queue, typeQueue{Deref(fi.Type), owner, fi.Index})
				}
			}

			//skip flatten struct field.
			if owner == fi {
				tq.fi.Children = append(tq.fi.Children, fi)
			}
			m = append(m, fi)
		}
	}

	flds := StructMap{
		Fields: m,
		Tree:   root,
		Paths:  map[string]*FieldInfo{},
		Leaves: map[string]*FieldInfo{},
	}
	for _, fi := range flds.Fields {
		if v, ok := flds.Paths[fi.Path]; ok {
			panic(fmt.Errorf("duplicated path: %v, indexs are %v and %v", t.String()+"."+fi.Name, v.Index, fi.Index))
		}
		flds.Paths[fi.Path] = fi
		if Deref(fi.Type).Kind() != reflect.Struct {
			flds.Leaves[fi.Path] = fi
		}
	}

	return flds
}
