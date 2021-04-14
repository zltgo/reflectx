package reflectx

import (
	"reflect"
)

var (
	formMapper = NewFormMapper("", nil)
)

type FormMapper struct {
	mapper *Mapper
}

//default tagName is "form"
func NewFormMapper(tagName string, tagFunc TagFunc) FormMapper {
	if tagName == "" {
		tagName = "form"
	}
	return FormMapper{
		NewMapper(tagName, tagFunc),
	}
}

//convert map to struct
func FormToStruct(form map[string][]string, ptr interface{}) error {
	return formMapper.FormToStruct(form, ptr)
}

//convert struct to form
func StructToForm(obj interface{}, form map[string][]string) {
	formMapper.StructToForm(obj, form)
}

func (fm FormMapper) FormToStruct(form map[string][]string, ptr interface{}) error {
	typ := reflect.TypeOf(ptr).Elem()
	val := reflect.ValueOf(ptr).Elem()
	structMap := fm.mapper.TypeMap(typ)

	if len(form) < len(structMap.Leaves) {
		for k, strs := range form {
			fi, ok := structMap.Leaves[k]
			//ignore if input string is empty, the field will not rewrite
			if !ok || len(strs) == 0 || (len(strs) == 1 && strs[0] == "") {
				continue
			}
			if err := fi.StringsToField(strs, val); err != nil {
				return err
			}
		}
	} else {
		for k, fi := range structMap.Leaves {
			strs, ok := form[k]
			//ignore if input string is empty, the field will not rewrite
			if !ok || len(strs) == 0 || (len(strs) == 1 && strs[0] == "") {
				continue
			}
			if err := fi.StringsToField(strs, val); err != nil {
				return err
			}
		}
	}

	return nil
}

// Convert struct to form.
func (fm FormMapper) StructToForm(obj interface{}, form map[string][]string) {
	typ, val := Indirect(obj)
	structMap := fm.mapper.TypeMap(typ)

	for k, fi := range structMap.Leaves {
		fV := FieldByIndexesReadOnly(val, fi.Index)
		fV = reflect.Indirect(fV)
		// skip invalid value or nil ptr
		if !fV.IsValid() {
			continue
		}

		switch fV.Kind() {
		case reflect.Slice:
			//omitempty??
			numElems := fV.Len()
			if numElems > 0 {
				slice := make([]string, numElems)
				var err error
				for i := 0; i < numElems; i++ {
					if slice[i], err = ValueToStr(fV.Index(i)); err != nil {
						panic(err)
					}
				}
				form[k] = slice
			}
		default:
			//omitempty
			_, ok := fi.Options[OmitEmpty]
			if !ok || !reflect.DeepEqual(fV.Interface(), fi.Zero.Interface()) {
				str, err := ValueToStr(fV)
				if err != nil {
					panic(err)
				}
				form[k] = []string{str}
			}
		}
	}
	return
}
