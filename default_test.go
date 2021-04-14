package reflectx

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

type third struct {
	Third string `default:"Third"`
}

var thir = third{"Third"}

type FooStruct struct {
	Foo    string  `default:"foo"`
	PFoo   *string `default:"pfoo"`
	Third  third   `form:",inline"`
	Pthird *third
}

var pfoo = "pfoo"
var fs = FooStruct{"foo", &pfoo, thir, &thir}

type NoTagStruct struct {
	A int `form:",omitempty"`
	B int `form:",omitempty"`
}

var notag = NoTagStruct{}

type FooBarStruct struct {
	private   float32 `default:"1.23"`
	NoTag     string  `form:","`
	Ignore    bool    `default:"-" form:"-"`
	Int       int     `default:"9" form:",omitempty"`
	PInt      *int    `default:"10"`
	FooStruct `form:"FS"`
	PFS       *FooStruct `form:",inline"`
	Bar       []int16    `default:"1,2,3,4,5"`
	PBar      *[]int16   `default:"2,3,4,5,6"`
	BarP      []*int64   `default:"3,4,5,6,7"`
	PBarP     *[]*int64  `default:"4,5,6,7,8"`
	N         NoTagStruct
	PN        *NoTagStruct
}

var i = 10
var bar1 = []int16{1, 2, 3, 4, 5}
var bar2 = []int16{2, 3, 4, 5, 6}
var ii = []int64{3, 4, 5, 6, 7, 8}

var barp1 = []*int64{&ii[0], &ii[1], &ii[2], &ii[3], &ii[4]}
var barp2 = []*int64{&ii[1], &ii[2], &ii[3], &ii[4], &ii[5]}
var fbs = FooBarStruct{0, "", false, 9, &i, fs, &fs, bar1, &bar2, barp1, &barp2, notag, nil}

func toString(df *defaultField) string {
	return fmt.Sprintf("index: %v, zero: %v, default: %v", df.index, df.zero.Interface(), reflect.Indirect(df.defaultVal).Interface())
}

func toJson(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return err.Error()
	}
	return string(b)
}

func TestGetDeaultInfo(t *testing.T) {
	di := defaultMapper.getDefaultStruct(reflect.TypeOf(FooBarStruct{}))

	if !reflect.DeepEqual(di.zero.Interface(), FooBarStruct{}) {
		t.Error("not equal !")
		t.Error("expected: ", toJson(FooBarStruct{}))
		t.Error("         got:", toJson(di.zero.Interface()))
	}

	if !reflect.DeepEqual(di.defaultVal.Interface(), fbs) {
		t.Error("not equal !")
		t.Error("expected: ", toJson(fbs))
		t.Error("         got:", toJson(di.defaultVal.Interface()))
	}

	fls := di.fields
	for _, tmp := range fls {
		t.Log(tmp.index)
	}

	if len(fls) != 11 {
		t.Errorf("expected num of default fields is %d, got %d", 8, len(fls))
		return
	}

	expect := []*defaultField{
		&defaultField{[]int{3}, reflect.Zero(reflect.TypeOf(fbs.Int)), reflect.ValueOf(fbs.Int)},
		&defaultField{[]int{4}, reflect.Zero(reflect.TypeOf(fbs.Int)), reflect.ValueOf(fbs.PInt)},
		&defaultField{[]int{6}, reflect.Zero(reflect.TypeOf(fbs.PFS).Elem()), reflect.ValueOf(fbs.PFS)},
		&defaultField{[]int{7}, reflect.Zero(reflect.TypeOf(fbs.Bar)), reflect.ValueOf(fbs.Bar)},
		&defaultField{[]int{8}, reflect.Zero(reflect.TypeOf(fbs.Bar)), reflect.ValueOf(fbs.PBar)},
		&defaultField{[]int{9}, reflect.Zero(reflect.TypeOf(fbs.BarP)), reflect.ValueOf(fbs.BarP)},
		&defaultField{[]int{10}, reflect.Zero(reflect.TypeOf(fbs.BarP)), reflect.ValueOf(fbs.PBarP)},
		&defaultField{[]int{5, 0}, reflect.Zero(reflect.TypeOf(fbs.FooStruct.Foo)), reflect.ValueOf(fbs.FooStruct.Foo)},
		&defaultField{[]int{5, 1}, reflect.Zero(reflect.TypeOf(fbs.FooStruct.Foo)), reflect.ValueOf(fbs.FooStruct.PFoo)},
		&defaultField{[]int{5, 2}, reflect.Zero(reflect.TypeOf(fbs.FooStruct.Third)), reflect.ValueOf(fbs.FooStruct.Third)},
		&defaultField{[]int{5, 3}, reflect.Zero(reflect.TypeOf(fbs.FooStruct.Third)), reflect.ValueOf(fbs.FooStruct.Pthird)},
	}

	for i := range expect {
		if !reflect.DeepEqual(fls[i].index, expect[i].index) ||
			!reflect.DeepEqual(fls[i].zero.Interface(), expect[i].zero.Interface()) ||
			!reflect.DeepEqual(fls[i].defaultVal.Interface(), expect[i].defaultVal.Interface()) {
			t.Errorf("expected %v, got %v", toString(expect[i]), toString(fls[i]))
		}
	}
}

type All struct {
	Bool    bool    `default:"true"`
	Int     int     `default:"-1"`
	Int8    int8    `default:"-2"`
	Int16   int16   `default:"-3"`
	Int32   int32   `default:"-4"`
	Int64   int64   `default:"-5"`
	Uint    uint    `default:"6"`
	Uint8   uint8   `default:"7"`
	Uint16  uint16  `default:"8"`
	Uint32  uint32  `default:"9"`
	Uint64  uint64  `default:"10"`
	Float32 float32 `default:"11.32"`
	Float64 float64 `default:"11.64"`
	String  string  `default:"some string"`

	PBool    *bool    `default:"true"`
	PInt     *int     `default:"11"`
	PInt8    *int8    `default:"12"`
	PInt16   *int16   `default:"13"`
	PInt32   *int32   `default:"14"`
	PInt64   *int64   `default:"15"`
	PUint    *uint    `default:"16"`
	PUint8   *uint8   `default:"17"`
	PUint16  *uint16  `default:"18"`
	PUint32  *uint32  `default:"19"`
	PUint64  *uint64  `default:"20"`
	PFloat32 *float32 `default:"21.32"`
	PFloat64 *float64 `default:"-21.64"`
	PString  *string  `default:"some string ptr"`

	Struc  FooBarStruct
	PStruc *FooBarStruct
}

var (
	Bool    bool    = true
	Int     int     = 11
	Int8    int8    = 12
	Int16   int16   = 13
	Int32   int32   = 14
	Int64   int64   = 15
	Uint    uint    = 16
	Uint8   uint8   = 17
	Uint16  uint16  = 18
	Uint32  uint32  = 19
	Uint64  uint64  = 20
	Float32 float32 = 21.32
	Float64 float64 = -21.64
	String  string  = "some string ptr"
)

var allValue = All{
	Bool:    true,
	Int:     -1,
	Int8:    -2,
	Int16:   -3,
	Int32:   -4,
	Int64:   -5,
	Uint:    6,
	Uint8:   7,
	Uint16:  8,
	Uint32:  9,
	Uint64:  10,
	Float32: 11.32,
	Float64: 11.64,
	String:  "some string",

	PBool:    &Bool,
	PInt:     &Int,
	PInt8:    &Int8,
	PInt16:   &Int16,
	PInt32:   &Int32,
	PInt64:   &Int64,
	PUint:    &Uint,
	PUint8:   &Uint8,
	PUint16:  &Uint16,
	PUint32:  &Uint32,
	PUint64:  &Uint64,
	PFloat32: &Float32,
	PFloat64: &Float64,
	PString:  &String,

	Struc:  fbs,
	PStruc: &fbs,
}

func TestAllocDefault(t *testing.T) {
	b := false
	v := AllocDefault(reflect.TypeOf(All{
		PBool: &b,
	}))

	if !reflect.DeepEqual(v.Interface(), allValue) {
		t.Error("not equal !")
		t.Error("expected: ", toJson(allValue))
		t.Error("         got:", toJson(v.Interface()))
	}

	v = AllocDefault(reflect.TypeOf(&All{}))

	if !reflect.DeepEqual(v.Elem().Interface(), allValue) {
		t.Error("not equal !")
		t.Error("expected: ", toJson(allValue))
		t.Error("         got:", toJson(v.Elem().Interface()))
	}

	return
}

func TestSetDefault(t *testing.T) {
	tmp := All{}
	var i = -3
	tmp.Int = -3
	tmp.PInt = &i
	tmp.Struc.Bar = []int16{}
	tmp.Struc.BarP = []*int64{nil}

	SetDefault(&tmp)

	expect := allValue
	expect.Int = -3
	expect.PInt = &i
	expect.Struc.Bar = []int16{}
	expect.Struc.BarP = []*int64{nil}

	if !reflect.DeepEqual(tmp, expect) {
		t.Error("not equal !")
		t.Error("expected: ", toJson(expect))
		t.Error("         got:", toJson(tmp))
	}

	type Unnamed struct {
		A string `default:"a"`
		B struct {
			C int   `default:"1"`
			D []int `default:"2,3"`
		}
	}
	var unnamed = Unnamed{}
	SetDefault(&unnamed)
	t.Log(unnamed.A, unnamed.B.C, unnamed.B.D)

	form := make(map[string][]string)
	StructToForm(unnamed, form)
	t.Log(form)
	return
}

func Benchmark_AllocDefault(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AllocDefault(reflect.TypeOf(All{}))
	}
}

func Benchmark_SetDefault(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tmp := All{}
		var i = -3
		tmp.Int = -3
		tmp.PInt = &i
		tmp.Struc.Bar = []int16{}
		tmp.Struc.BarP = []*int64{nil}
		SetDefault(&tmp)
	}
}

func Benchmark_SetThird(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tmp := third{}
		SetDefault(&tmp)
	}
}

func Benchmark_SetFooStruct(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tmp := FooStruct{}
		SetDefault(&tmp)
	}
}

func Benchmark_SetFooBarStruct(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tmp := FooBarStruct{}
		SetDefault(&tmp)
	}
}
