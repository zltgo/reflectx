package reflectx

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2/bson"
)

var allForm = map[string][]string{
	"Bool":    []string{"true"},
	"Int":     []string{"-1"},
	"Int8":    []string{"-2"},
	"Int16":   []string{"-3"},
	"Int32":   []string{"-4"},
	"Int64":   []string{"-5"},
	"Uint":    []string{"6"},
	"Uint8":   []string{"7"},
	"Uint16":  []string{"8"},
	"Uint32":  []string{"9"},
	"Uint64":  []string{"10"},
	"Float32": []string{"11.32"},
	"Float64": []string{"11.64"},
	"String":  []string{"some string"},

	"PBool":    []string{"true"},
	"PInt":     []string{"11"},
	"PInt8":    []string{"12"},
	"PInt16":   []string{"13"},
	"PInt32":   []string{"14"},
	"PInt64":   []string{"15"},
	"PUint":    []string{"16"},
	"PUint8":   []string{"17"},
	"PUint16":  []string{"18"},
	"PUint32":  []string{"19"},
	"PUint64":  []string{"20"},
	"PFloat32": []string{"21.32"},
	"PFloat64": []string{"-21.64"},
	"PString":  []string{"some string ptr"},

	"Struc.Int":   []string{"9"},
	"Struc.PInt":  []string{"10"},
	"Struc.PBarP": []string{"4", "5", "6", "7", "8"},

	"Struc.BarP":  []string{"3", "4", "5", "6", "7"},
	"Struc.NoTag": []string{""},
	"Struc.PBar":  []string{"2", "3", "4", "5", "6"},
	"Struc.Bar":   []string{"1", "2", "3", "4", "5"},

	"Struc.PFS.Foo":  []string{"foo"},
	"Struc.PFS.PFoo": []string{"pfoo"},
	"Struc.FS.PFoo":  []string{"pfoo"},
	"Struc.FS.Foo":   []string{"foo"},

	"Struc.PFS.Third.Third":  []string{"Third"},
	"Struc.PFS.Pthird.Third": []string{"Third"},
	"Struc.FS.Third.Third":   []string{"Third"},
	"Struc.FS.Pthird.Third":  []string{"Third"},

	"PStruc.Int":   []string{"9"},
	"PStruc.PInt":  []string{"10"},
	"PStruc.PBarP": []string{"4", "5", "6", "7", "8"},

	"PStruc.BarP":  []string{"3", "4", "5", "6", "7"},
	"PStruc.NoTag": []string{""},
	"PStruc.PBar":  []string{"2", "3", "4", "5", "6"},
	"PStruc.Bar":   []string{"1", "2", "3", "4", "5"},

	"PStruc.PFS.Foo":  []string{"foo"},
	"PStruc.PFS.PFoo": []string{"pfoo"},
	"PStruc.FS.PFoo":  []string{"pfoo"},
	"PStruc.FS.Foo":   []string{"foo"},

	"PStruc.PFS.Third.Third":  []string{"Third"},
	"PStruc.PFS.Pthird.Third": []string{"Third"},
	"PStruc.FS.Third.Third":   []string{"Third"},
	"PStruc.FS.Pthird.Third":  []string{"Third"},
}

func checkEqualForm(t *testing.T, f1, f2 map[string][]string) {
	t.Helper()
	if len(f1) != len(f2) {
		t.Error("length of f1 is not equal to f2")
	}

	for k, v := range f1 {
		vv, ok := f2[k]
		if !ok {
			t.Errorf(k + " not found in f2")
			continue
		}
		if strings.Join(v, " ") != strings.Join(vv, " ") {
			t.Errorf("%s not equal: expect %v, got %v", k, vv, v)
		}
	}
}

func TestStructToForm(t *testing.T) {
	form := make(map[string][]string)

	StructToForm(allValue, form)
	checkEqualForm(t, form, allForm)
	t.Log(reflect.DeepEqual(form, allForm))
}

func TestMapForm(t *testing.T) {
	tmp := All{}
	if err := FormToStruct(allForm, &tmp); err != nil {
		t.Error(err)
		return
	}

	if !reflect.DeepEqual(tmp, allValue) {
		t.Error("not equal !")
		t.Error("expected: ", toJson(allValue))
		t.Error("         got:", toJson(tmp))
	}

	return
}

func TestCopyTo(t *testing.T) {
	type Foo struct {
		F string
		O int
	}
	type FooBar struct {
		Foo
		Bar int
	}

	f := Foo{"will", 2}
	fb := FooBar{Bar: 3}
	CopyStruct(&fb, f)

	expect := FooBar{
		Foo: f,
		Bar: 3,
	}
	if !reflect.DeepEqual(fb, expect) {
		t.Error("not equal !")
		t.Error("expected: ", toJson(expect))
		t.Error("         got:", toJson(fb))
	}

	return
}

func TestDuplicatedName(t *testing.T) {
	type DuplicatedName struct {
		A int
		B int `form:"A"`
	}

	tmp := DuplicatedName{1, 2}
	form := make(map[string][]string)

	defer func() {
		r := recover()
		if r.(error).Error() != "duplicated path: reflectx.DuplicatedName.A, indexs are [0] and [1]" {
			t.Error(r)
		}
	}()

	StructToForm(tmp, form)
	t.Error("got here, didn't expect to")
}

func TestDuplicatedNameInline(t *testing.T) {
	type InlineT struct {
		A int `form:",omitempty"`
	}

	type DuplicatedInline struct {
		A int
		InlineT
	}

	tmp := DuplicatedInline{}
	form := make(map[string][]string)

	defer func() {
		r := recover()
		if r.(error).Error() != "duplicated path: reflectx.DuplicatedInline.A, indexs are [0] and [1 0]" {
			t.Errorf("expected %v", r)
		}
	}()

	StructToForm(tmp, form)
	t.Error("got here, didn't expect to")
}

func TestOmitempty(t *testing.T) {
	type InlineT struct {
		A     int  `form:",omitempty"`
		Astar *int `form:",omitempty"`
		Bstar *int `form:",omitempty"`
	}
	it := InlineT{
		Bstar: new(int),
	}
	form := make(map[string][]string)

	StructToForm(it, form)
	t.Log(form)
	assert.Equal(t, len(form), 0)
}

func Benchmark_FormToStruct(b *testing.B) {

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tmp := All{}
		if err := FormToStruct(allForm, &tmp); err != nil {
			b.Error(err)
			return
		}
	}
}

func Benchmark_StructToForm(b *testing.B) {

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		form := make(map[string][]string)

		StructToForm(allValue, form)
	}
}

func Benchmark_MarshalJson(b *testing.B) {

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := json.Marshal(allValue); err != nil {
			b.Error(err)
			return
		}
	}
}

func Benchmark_UnmarshalJson(b *testing.B) {
	c, err := json.Marshal(allValue)
	if err != nil {
		b.Error(err)
		return
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tmp := All{}

		if err := json.Unmarshal(c, &tmp); err != nil {
			b.Error(err)
			return
		}
	}
}

type AliceStruct struct {
	Foo string `default:"foo"`
}

type AliceBobStruct struct {
	AliceStruct
	Bar []int `default:"1,2,3,4,5"`
}

type Form map[string][]string

// Encode encodes the values into URL-encoded form
// ("bar=baz&foo=quux") sorted by key.
func (m Form) Encode() string {
	return url.Values(m).Encode()
}

//It replaces any existing values.
func (m Form) Append(dest map[string][]string) {
	for k, v := range dest {
		m[k] = v
	}
	return
}

// Decode parses the URL-encoded query string and returns
// a map listing the values specified for each key.
// if m is not nil, values from query will append to m.
func (m *Form) Decode(query string) error {
	values, err := url.ParseQuery(query)
	if err != nil {
		return err
	}

	if len(*m) > 0 {
		m.Append(Form(values))
	} else {
		*m = Form(values)
	}
	return nil
}

func TestMapStruct(t *testing.T) {
	dest := make(Form)
	dest["Foo"] = []string{"foo"}
	dest["Bar"] = []string{"1", "2", "3", "4", "5"}

	var tmp AliceBobStruct
	tmp.Foo = "foo"
	tmp.Bar = []int{1, 2, 3, 4, 5}

	src := make(Form)
	StructToForm(tmp, src)
	checkEqualForm(t, src, dest)
}

func TestToStruct(t *testing.T) {
	f := make(Form)
	f["Foo"] = []string{"foo"}
	f["Bar"] = []string{"1", "2", "3", "4", "5"}

	var dest AliceBobStruct
	dest.Foo = "foo"
	dest.Bar = []int{1, 2, 3, 4, 5}

	var src AliceBobStruct
	err := FormToStruct(f, &src)
	if err != nil {
		t.Error(err)
		return
	}
	assert.Equal(t, src, dest)

	type ErrorStruct struct {
		Bar map[int]string
	}

	es := &ErrorStruct{}
	err = FormToStruct(f, es)
	if err == nil {
		t.Error("expect error here")
		return
	}
	assert.Equal(t, err.Error(), "reflectx: can not convert [1 2 3 4 5] to reflectx.ErrorStruct.Bar")
}

type AllStruct struct {
	Bool    bool
	Int     int
	Int8    int8
	Int16   int16
	Int32   int32
	Int64   int64
	Uint    uint
	Uint8   uint8
	Uint16  uint16
	Uint32  uint32
	Uint64  uint64
	Float32 float32
	Float64 float64

	PFloat64 *float64

	String string

	StringSlice []string
	ByteSlice   []byte

	Small Small
}

type Small struct {
	Tag string
}

var allVs = AllStruct{
	Bool:    true,
	Int:     2,
	Int8:    3,
	Int16:   4,
	Int32:   5,
	Int64:   6,
	Uint:    7,
	Uint8:   8,
	Uint16:  9,
	Uint32:  10,
	Uint64:  11,
	Float32: 14.1,
	Float64: 15.1,

	String:      "16",
	StringSlice: []string{"str24", "str25", "str26"},
	ByteSlice:   []byte{27, 28, 29},
	Small:       Small{Tag: "tag30"},
}

func Benchmark_Json(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		c, err := json.Marshal(allVs)
		if err != nil {
			b.Error(err)
			break
		}
		var tmp AllStruct
		err = json.Unmarshal(c, &tmp)
		if err != nil {
			b.Error(err)
			break
		}
	}
}

//Gob编码，为什么比JSON慢？
func EncodeGob(obj interface{}) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(obj)
	return buf.Bytes(), err
}

//不科学啊，84849 ns/op
func DecodeGob(encoded []byte, ptr interface{}) error {
	buf := bytes.NewBuffer(encoded)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(ptr)
	return err
}

func Benchmark_Gob(b *testing.B) {

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c, err := EncodeGob(allVs)
		if err != nil {
			b.Error(err)
			break
		}
		var tmp AllStruct
		err = DecodeGob(c, &tmp)
		if err != nil {
			b.Error(err)
			break
		}
	}
}

func Benchmark_Bson(b *testing.B) {

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c, err := bson.Marshal(allVs)
		if err != nil {
			b.Error(err)
			break
		}
		var tmp AllStruct
		err = bson.Unmarshal(c, &tmp)
		if err != nil {
			b.Error(err)
			break
		}
	}
}

func Benchmark_BJson(b *testing.B) {

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c, err := bson.MarshalJSON(allVs)
		if err != nil {
			b.Error(err)
			break
		}
		var tmp AllStruct
		err = bson.UnmarshalJSON(c, &tmp)
		if err != nil {
			b.Error(err)
			break
		}
	}
}

func Benchmark_EncodeForm(b *testing.B) {
	f := Form{}
	StructToForm(allVs, f)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		f.Encode()
	}
}

func Benchmark_DecodeForm(b *testing.B) {
	f := Form{}
	StructToForm(allVs, f)

	str := f.Encode()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tmp := Form{}
		err := tmp.Decode(str)
		if err != nil {
			b.Error(err)
			break
		}
	}
}

func Benchmark_FormTransfer(b *testing.B) {

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		f1 := make(Form)
		StructToForm(allVs, f1)

		var tmp AllStruct
		f2 := make(Form)
		if err := f2.Decode(f1.Encode()); err != nil {
			b.Error(err)
			break
		}
		if err := FormToStruct(f2, &tmp); err != nil {
			b.Error(err)
			break
		}
	}
}

func Benchmark_Form(b *testing.B) {

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		f := make(Form)
		StructToForm(allVs, f)
		var tmp AllStruct
		if err := FormToStruct(f, &tmp); err != nil {
			b.Error(err)
			break
		}
	}
}
