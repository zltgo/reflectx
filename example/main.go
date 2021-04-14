package main

import (
	"fmt"

	"github.com/zltgo/reflectx"
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
	Bar `reflector:",flatten"`
	//PBar   *Bar
	//SFoo   []Foo
	//PSFoo  *[]Foo
	//PSPFoo *[]*Foo
}

func main() {
	i := 1
	s := "string"
	b := Bar{&i, &s}
	f := Foo{
		F: 1,
		O: map[string]interface{}{
			"int":    1,
			"struct": struct{}{},
			"bar":    b,
		},
	}

	fb := FooBar{
		Foo: f,
		Bar: b,
	}

	r := reflectx.NewReflector("", "", nil)
	rv, _ := r.Encode(fb)

	fmt.Println(string(rv))
}
