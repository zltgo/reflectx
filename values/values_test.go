package values

import (
	"math"
	"sync"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestForm(t *testing.T) {
	gf := Form{
		"1": []string{"1"},
		"2": []string{"2"},
		"3": []string{"3", "2333"},
		"4": []string{"4", "true", "3.14", "pi"},
	}
	Convey("test form set", t, func() {
		fm := Form{}
		fm.Set("1", 1)
		fm.Set("2", "2")
		fm.Set("3", []string{"3", "2333"})
		fm.Set("4", []interface{}{4, true, 3.14, "pi"})
		So(fm, ShouldResemble, gf)

		So(func() { fm.Set("5", time.Now()) }, ShouldPanic)
	})

	Convey("test form add string", t, func() {
		fm := Form{}
		fm.Set("3", 3)
		fm.AddString("3", "2333")
		So(fm["3"], ShouldResemble, []string{"3", "2333"})
	})

	Convey("test form add values", t, func() {
		fm := Form{}
		fm.Set("4", 4)
		fm.AddValues("4", true, 3.14, "pi")
		So(fm["4"], ShouldResemble, []string{"4", "true", "3.14", "pi"})
	})

	Convey("test form get", t, func() {
		v := gf.Get("0")
		So(v, ShouldBeNil)

		v = gf.Get("1")
		So(v, ShouldEqual, "1")

		v = gf.Get("4")
		So(v, ShouldResemble, []string{"4", "true", "3.14", "pi"})
	})

	Convey("test form encode and decode", t, func() {
		b, err := gf.Encode()
		So(err, ShouldBeNil)
		So(string(b), ShouldEqual, "1=1&2=2&3=3&3=2333&4=4&4=true&4=3.14&4=pi")

		fm := Form{}
		err = fm.Decode(b)
		So(err, ShouldBeNil)
		So(fm, ShouldResemble, gf)
	})

	Convey("test form range", t, func() {
		fm := Form{}
		gf.Range(func(key string, v interface{}) bool {
			fm.Set(key, v)
			return true
		})
		So(fm, ShouldResemble, gf)
	})
}

func TestJsonMap(t *testing.T) {
	gf := JsonMap{
		"1": 1,
		"2": "2",
		"3": []string{"3", "2333"},
		"4": []interface{}{4, true, 3.14, "pi"},
	}
	Convey("test JsonMap set", t, func() {
		fm := JsonMap{}
		fm.Set("1", 1)
		fm.Set("2", "2")
		fm.Set("3", []string{"3", "2333"})
		fm.Set("4", []interface{}{4, true, 3.14, "pi"})
		So(fm, ShouldResemble, gf)
	})

	Convey("test JsonMap get", t, func() {
		v := gf.Get("0")
		So(v, ShouldBeNil)

		v = gf.Get("1")
		So(v, ShouldEqual, 1)

		v = gf.Get("4")
		So(v, ShouldResemble, []interface{}{4, true, 3.14, "pi"})
	})

	Convey("test JsonMap encode and decode", t, func() {
		b, err := gf.Encode()
		So(err, ShouldBeNil)

		fm := JsonMap{}
		err = fm.Decode(b)
		So(err, ShouldBeNil)

		So(fm["1"], ShouldEqual, 1)
		So(fm["2"], ShouldEqual, "2")
		So(fm["3"], ShouldResemble, []interface{}{"3", "2333"})
		//So(fm["4"], ShouldResemble, []interface{}{4, true, 3.14, "pi"})
	})

	Convey("test JsonMap range", t, func() {
		fm := JsonMap{}
		gf.Range(func(key string, v interface{}) bool {
			fm.Set(key, v)
			return true
		})
		So(fm, ShouldResemble, gf)
	})
}

func TestSafeMap(t *testing.T) {
	sm := SafeMap{}
	vs := &sm

	Convey("should be thread-safe", t, func() {
		c := 100
		n := 100000
		wg := sync.WaitGroup{}

		for i := 0; i < c; i++ {
			wg.Add(1)

			go func(thread int) {
				defer wg.Done()
				for j := 0; j < n; j++ {
					AddInt(vs, "1", 1)
				}
			}(i)
		}
		wg.Wait()
		So(Get(vs, "1").MustInt(), ShouldEqual, c*n)
	})
}

func TestFormValues(t *testing.T) {
	fm := Form{}
	vs := &fm
	testValues(t, vs)
}

func TestJsonMapValues(t *testing.T) {
	jm := JsonMap{}
	vs := &jm
	testValues(t, vs)
}

func TestSafeMapValues(t *testing.T) {
	sm := SafeMap{}
	vs := &sm
	testValues(t, vs)
}

func testValues(t *testing.T, vs Values) {
	t.Helper()
	vs.Set("0", 0.618)
	vs.Set("1", 1)
	vs.Set("2", "2")
	vs.Set("3", []string{"3", "2333"})
	vs.Set("4", []interface{}{4, true, 3.14, "pi"})
	vs.Set("5", "not a number")

	Convey("test values DefaultGet", t, func() {
		So(Get(vs, "-1").Default(-1).Interface(), ShouldEqual, -1)
		So(Get(vs, "2").Default(3).Interface(), ShouldEqual, "2")
		So(vs.ValueOf("3").Interface(), ShouldResemble, []string{"3", "2333"})
		So(vs.ValueOf("6").Interface(), ShouldEqual, nil)
	})

	Convey("test values get string", t, func() {
		So(Get(vs, "0").String(), ShouldEqual, "0.618")
		So(Get(vs, "1").String(), ShouldEqual, "1")
		So(Get(vs, "2").String(), ShouldEqual, "2")
		So(Get(vs, "3").String(), ShouldEqual, "[3 2333]")
		So(Get(vs, "4").String(), ShouldEqual, "[4 true 3.14 pi]")
		So(Get(vs, "5").String(), ShouldEqual, "not a number")
	})

	Convey("test values get int", t, func() {
		So(Get(vs, "0").Int(), ShouldEqual, 0)
		So(Get(vs, "1").Int(), ShouldEqual, 1)
		So(Get(vs, "2").Int(), ShouldEqual, 2)
		So(Get(vs, "3").Int(), ShouldEqual, 0)
		So(func() { Get(vs, "3").MustInt() }, ShouldPanic)
		So(func() { Get(vs, "5").MustInt() }, ShouldPanic)

		_, err := Get(vs, "0").ToInt()
		So(err.Error(), ShouldEqual, "cann't convert 0.618 to int")

		_, err = Get(vs, "3").ToInt()
		So(err.Error(), ShouldEqual, "can't get int from [3 2333]")
	})

	Convey("test values get uint", t, func() {
		So(Get(vs, "0").Uint(), ShouldEqual, 0)
		So(Get(vs, "1").Uint(), ShouldEqual, 1)
		So(Get(vs, "2").Uint(), ShouldEqual, 2)
		So(Get(vs, "3").Uint(), ShouldEqual, 0)
		So(func() { Get(vs, "3").MustUint() }, ShouldPanic)
		So(func() { Get(vs, "5").MustUint() }, ShouldPanic)

		_, err := Get(vs, "0").ToUint()
		So(err.Error(), ShouldEqual, "cann't convert 0.618 to uint")

		_, err = Get(vs, "3").ToUint()
		So(err.Error(), ShouldEqual, "can't get uint from [3 2333]")

		vs.Set("-1", -1)
		So(Get(vs, "-1").Uint64(), ShouldEqual, uint(math.MaxUint64))
	})

	Convey("test values get float", t, func() {
		So(vs.ValueOf("0").MustFloat(), ShouldEqual, 0.618)
		So(vs.ValueOf("1").MustFloat(), ShouldEqual, 1)
		So(vs.ValueOf("2").MustFloat(), ShouldEqual, 2)
		So(vs.ValueOf("3").Float(), ShouldEqual, 0)
		So(func() { vs.ValueOf("3").MustFloat() }, ShouldPanic)
		So(func() { vs.ValueOf("5").MustFloat() }, ShouldPanic)

		_, err := vs.ValueOf("3").ToFloat()
		So(err.Error(), ShouldEqual, "can't get float from [3 2333]")
	})

	Convey("test values get bool", t, func() {
		vs.Set("false", false)
		vs.Set("true", "true")
		So(vs.ValueOf("1").Bool(), ShouldBeTrue)
		So(vs.ValueOf("false").Bool(), ShouldBeFalse)
		So(vs.ValueOf("true").Bool(), ShouldBeTrue)
		So(vs.ValueOf("3").Bool(), ShouldBeFalse)

		So(func() { vs.ValueOf("2").MustBool() }, ShouldPanic)
		So(func() { vs.ValueOf("3").MustBool() }, ShouldPanic)
		So(func() { vs.ValueOf("0").MustBool() }, ShouldPanic)

		_, err := vs.ValueOf("3").ToBool()
		So(err.Error(), ShouldEqual, `strconv.ParseBool: parsing "[3 2333]": invalid syntax`)
	})

	Convey("test values add int", t, func() {
		vs.Set("cnt", 9527)
		So(AddInt(vs, "cnt", 1), ShouldEqual, 9528)
		So(AddInt(vs, "cnt", -2), ShouldEqual, 9526)
		So(AddInt(vs, "notfound", 123), ShouldEqual, 123)
		So(vs.ValueOf("notfound").Int(), ShouldEqual, 123)
		vs.Remove("notfound")

		vs.Set("0", 0.618)
		So(AddInt(vs, "0", 10), ShouldEqual, 10)
	})

	Convey("test values add uint", t, func() {
		vs.Set("cnt", 9527)
		So(AddUint(vs, "cnt", 1), ShouldEqual, 9528)
		So(AddUint(vs, "notfound", 123), ShouldEqual, 123)
		So(vs.ValueOf("notfound").Uint(), ShouldEqual, 123)
		vs.Remove("notfound")

		vs.Set("0", 0.618)
		So(AddInt(vs, "0", 10), ShouldEqual, 10)

		vs.Set("-1", -1)
		So(AddUint(vs, "-1", 9), ShouldEqual, 8)
	})

	Convey("test getsert", t, func() {
		v := Getsert(vs, "getsert", func() interface{} {
			return "a new one"
		})
		So(v, ShouldEqual, "a new one")

		v = Getsert(vs, "getsert", func() interface{} {
			return "can not replace"
		})
		So(v, ShouldEqual, "a new one")
	})
}
