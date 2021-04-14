package reflectx

import (
	"encoding/json"
	"reflect"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func equal(a, b interface{}, t *testing.T) bool {
	t.Helper()
	type value struct {
		V interface{}
	}
	as, err := json.Marshal(value{a})
	if err != nil {
		t.Log(err)
		return false
	}
	bs, err := json.Marshal(value{a})
	if err != nil {
		t.Log(err)
		return false
	}
	return string(as) == string(bs)
}

func TestSetValue(t *testing.T) {
	Convey("should set base type correctly", t, func() {
		a := reflect.New(reflect.TypeOf(int64(0)))
		err := SetValue(a, reflect.ValueOf(1))
		So(err, ShouldBeNil)
		So(equal(a.Interface(), 1, t), ShouldBeTrue)

		err = SetValue(a, reflect.ValueOf(uint(2)))
		So(err, ShouldBeNil)
		So(equal(a.Interface(), uint(2), t), ShouldBeTrue)

		err = SetValue(a, reflect.ValueOf(3.0))
		So(err, ShouldBeNil)
		So(equal(a.Interface(), 3.0, t), ShouldBeTrue)

		err = SetValue(a, reflect.ValueOf("4"))
		So(err, ShouldBeNil)
		So(equal(a.Interface(), "4", t), ShouldBeTrue)
	})

	Convey("should set struct type correctly", t, func() {
		a := reflect.New(reflect.TypeOf(All{})).Elem()
		err := SetValue(a, reflect.ValueOf(allValue))
		So(err, ShouldBeNil)
		So(equal(a.Interface(), allValue, t), ShouldBeTrue)
	})

	Convey("should set nil point correctly", t, func() {
		a := reflect.New(reflect.TypeOf(&All{})).Elem()
		err := SetValue(a, reflect.ValueOf((*interface{})(nil)))
		So(err, ShouldBeNil)
		So(a.Interface(), ShouldBeNil)

		err = SetValue(a, reflect.Value{})
		So(err, ShouldBeNil)
		So(a.Interface(), ShouldBeNil)
	})

	Convey("should set map correctly", t, func() {
		a := reflect.New(reflect.TypeOf(map[int]string{})).Elem()
		mp := map[string]interface{}{
			"1": 1,
			"2": 2.0,
			"3": true,
			"4": false,
			"5": new(int),
			"6": new(float64),
		}
		err := SetValue(a, reflect.ValueOf(mp))
		So(err, ShouldBeNil)
		So(equal(a.Interface(), mp, t), ShouldBeTrue)
	})

	Convey("should set slice correctly", t, func() {
		a := reflect.New(reflect.TypeOf([]interface{}{})).Elem()
		s := []interface{}{1, 2.0, true, false, new(int), new(float64)}

		err := SetValue(a, reflect.ValueOf(s))
		So(err, ShouldBeNil)
		So(equal(a.Interface(), s, t), ShouldBeTrue)
	})

	Convey("should return type mismatch error", t, func() {
		a := reflect.New(reflect.TypeOf([]interface{}{})).Elem()
		err := SetValue(a, reflect.ValueOf(allValue))
		So(err.Error(), ShouldEqual, "reflectx: type mismatch, expected []interface {}, got reflectx.All")
	})

	Convey("should return type mismatch error", t, func() {
		a := reflect.New(reflect.TypeOf(int64(0)))
		err := SetValue(a, reflect.ValueOf(func() {}))
		So(err.Error(), ShouldEqual, "reflectx: unexpected type for ValueToStr: func()")
	})
}
