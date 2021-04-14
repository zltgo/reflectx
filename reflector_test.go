package reflectx

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type Foo struct {
	F int `reflector:"f"`
	O map[string]interface{}
}

type Bar struct {
	B  *int
	Ar *string
}

type FooBar struct {
	Foo `reflector:"F"`
	//PFoo   *Foo
	Bar    `reflector:",flatten"`
	PBar   *Bar
	SFoo   []Foo
	PSFoo  *[]Foo
	PSPFoo *[]*Foo
}

func TestEncode(t *testing.T) {
	Convey("should encode struct correctly", t, func() {
		i := 1
		s := "string"
		b := Bar{&i, &s}
		mp := map[string]interface{}{
			"int":    1,
			"struct": struct{}{},
			"bar":    b,
		}
		f := Foo{
			F: 1,
			O: mp,
		}

		r := NewReflector("", "", nil)
		rv, err := r.Encode(map[string]interface{}{
			"foo":           f,
			"bar":           b,
			"int":           1,
			"string":        "string",
			"[]interface{}": []interface{}{1, "string", &b},
		})
		So(err, ShouldBeNil)
		So(string(rv), ShouldEqual, `{
    "[]interface{}": [
        1,
        "string",
        {
            "Ar": "string",
            "B": 1,
            "_struct_name": "reflectx.Bar"
        }
    ],
    "bar": {
        "Ar": "string",
        "B": 1,
        "_struct_name": "reflectx.Bar"
    },
    "foo": {
        "O": {
            "bar": {
                "Ar": "string",
                "B": 1,
                "_struct_name": "reflectx.Bar"
            },
            "int": 1,
            "struct": {
                "_struct_name": "struct {}"
            }
        },
        "_struct_name": "reflectx.Foo",
        "f": 1
    },
    "int": 1,
    "string": "string"
}`)
	})

	Convey("should encode struct correctly", t, func() {
		i := 1
		s := "string"
		b := Bar{&i, &s}
		mp := map[string]interface{}{
			"int":    1,
			"struct": struct{}{},
			"bar":    b,
		}
		f := Foo{
			F: 1,
			O: mp,
		}

		sf := []Foo{f, f}
		psf := []*Foo{&f, nil, &f}
		fb := FooBar{
			Foo:    f,
			Bar:    b,
			PBar:   &b,
			SFoo:   sf,
			PSFoo:  &sf,
			PSPFoo: &psf,
		}

		r := NewReflector("", "", nil)
		rv, err := r.Encode(fb)
		So(err, ShouldBeNil)
		So(string(rv), ShouldEqual, `{
    "Ar": "string",
    "B": 1,
    "F": {
        "O": {
            "bar": {
                "Ar": "string",
                "B": 1,
                "_struct_name": "reflectx.Bar"
            },
            "int": 1,
            "struct": {
                "_struct_name": "struct {}"
            }
        },
        "_struct_name": "reflectx.Foo",
        "f": 1
    },
    "PBar": {
        "Ar": "string",
        "B": 1,
        "_struct_name": "reflectx.Bar"
    },
    "PSFoo": [
        {
            "O": {
                "bar": {
                    "Ar": "string",
                    "B": 1,
                    "_struct_name": "reflectx.Bar"
                },
                "int": 1,
                "struct": {
                    "_struct_name": "struct {}"
                }
            },
            "_struct_name": "reflectx.Foo",
            "f": 1
        },
        {
            "O": {
                "bar": {
                    "Ar": "string",
                    "B": 1,
                    "_struct_name": "reflectx.Bar"
                },
                "int": 1,
                "struct": {
                    "_struct_name": "struct {}"
                }
            },
            "_struct_name": "reflectx.Foo",
            "f": 1
        }
    ],
    "PSPFoo": [
        {
            "O": {
                "bar": {
                    "Ar": "string",
                    "B": 1,
                    "_struct_name": "reflectx.Bar"
                },
                "int": 1,
                "struct": {
                    "_struct_name": "struct {}"
                }
            },
            "_struct_name": "reflectx.Foo",
            "f": 1
        },
        null,
        {
            "O": {
                "bar": {
                    "Ar": "string",
                    "B": 1,
                    "_struct_name": "reflectx.Bar"
                },
                "int": 1,
                "struct": {
                    "_struct_name": "struct {}"
                }
            },
            "_struct_name": "reflectx.Foo",
            "f": 1
        }
    ],
    "SFoo": [
        {
            "O": {
                "bar": {
                    "Ar": "string",
                    "B": 1,
                    "_struct_name": "reflectx.Bar"
                },
                "int": 1,
                "struct": {
                    "_struct_name": "struct {}"
                }
            },
            "_struct_name": "reflectx.Foo",
            "f": 1
        },
        {
            "O": {
                "bar": {
                    "Ar": "string",
                    "B": 1,
                    "_struct_name": "reflectx.Bar"
                },
                "int": 1,
                "struct": {
                    "_struct_name": "struct {}"
                }
            },
            "_struct_name": "reflectx.Foo",
            "f": 1
        }
    ],
    "_struct_name": "reflectx.FooBar"
}`)
	})
}

func TestDecode(t *testing.T) {
	testDecode(t, "json")
	testDecode(t, "bson")
	testDecode(t, "indentedjson")
}

func testDecode(t *testing.T, format string) {
	r := NewReflector(format, "", nil)
	r.Register(Bar{})
	r.Register(Foo{})
	r.Register(FooBar{})
	r.Register(struct{}{})

	i := 1
	s := "string"
	bar := Bar{&i, &s}
	mp := map[string]interface{}{
		"float64": 1.0, // int will be decoded to float64 for json.
		"struct":  struct{}{},
		"bar":     bar,
	}
	f := Foo{
		F: 1,
		O: mp,
	}

	sf := []Foo{f, f}
	psf := []*Foo{&f, nil, &f}
	fb := FooBar{
		Foo:    f,
		Bar:    bar,
		PBar:   &bar,
		SFoo:   sf,
		PSFoo:  &sf,
		PSPFoo: &psf,
	}

	Convey("should decode simple struct correctly", t, func() {
		b, err := r.Encode(bar)
		So(err, ShouldBeNil)

		rv, err := r.Decode(b)
		So(err, ShouldBeNil)

		rb, ok := rv.(Bar)
		So(ok, ShouldBeTrue)
		So(rb, ShouldResemble, bar)
	})

	Convey("should decode map[string]inteface{} correctly", t, func() {
		b, err := r.Encode(mp)
		So(err, ShouldBeNil)

		rv, err := r.Decode(b)
		So(err, ShouldBeNil)
		So(rv, ShouldResemble, mp)
	})

	Convey("should decode map field correctly", t, func() {
		b, err := r.Encode(f)
		So(err, ShouldBeNil)

		rv, err := r.Decode(b)
		So(err, ShouldBeNil)
		So(rv, ShouldResemble, f)
	})

	Convey("should decode complicated struct correctly", t, func() {
		b, err := r.Encode(fb)
		So(err, ShouldBeNil)

		rv, err := r.Decode(b)
		So(err, ShouldBeNil)
		So(rv, ShouldResemble, fb)
	})

	Convey("should return strconv.ParseInt error", t, func() {
		tmp := map[string]interface{}{
			"_struct_name": "reflectx.Bar",
			"B":            "expected int, got string",
			"Ar":           "just a string",
		}

		b, err := r.Encode(tmp)
		So(err, ShouldBeNil)

		_, err = r.Decode(b)
		So(err.Error(), ShouldEqual, `reflectx: strconv.ParseInt: parsing "expected int, got string": invalid syntax`)
	})

	Convey("should return struct type mismatch", t, func() {
		tmp := map[string]interface{}{
			"_struct_name": "balabala",
			"B":            "expected int, got string",
			"Ar":           "just a string",
		}

		b, err := r.Encode(tmp)
		So(err, ShouldBeNil)

		_, err = r.Decode(b)
		So(err.Error(), ShouldEqual, `reflectx: unknown struct name: balabala`)
	})

	Convey("should return struct type mismatch", t, func() {
		fieldMap := map[string]interface{}{
			"_struct_name": "reflectx.Bar",
			"B":            1,
			"Ar":           "just a string",
		}
		structMap := map[string]interface{}{
			"_struct_name": "reflectx.FooBar",
			"F":            fieldMap,
		}

		b, err := r.Encode(structMap)
		So(err, ShouldBeNil)

		_, err = r.Decode(b)
		So(err.Error(), ShouldEqual, "reflectx: type mismatch, expected reflectx.Foo, got reflectx.Bar")
	})
}
