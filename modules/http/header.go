package http

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/d5/tengo/v2"
)

type Header struct {
	Value http.Header
	tengo.ObjectImpl
}

// TypeName returns the name of the type.
func (o *Header) TypeName() string {
	return "header"
}

func (o *Header) String() string {
	var pairs []string
	for k, v := range o.Value {
		pairs = append(pairs, fmt.Sprintf("%s: %v", k, v))
	}
	return fmt.Sprintf("header{%s}", strings.Join(pairs, ", "))
}

// Copy returns a copy of the type.
func (o *Header) Copy() tengo.Object {
	c := o.Value.Clone()
	return &Header{Value: c}
}

// IsFalsy returns true if the value of the type is falsy.
func (o *Header) IsFalsy() bool {
	return len(o.Value) == 0 // empty header is falsy, like empty map
}

// Equals returns true if the value of the type is equal to the value of
// another object.
func (o *Header) Equals(x tengo.Object) bool {
	xo, ok := x.(*Header)
	if !ok {
		return false
	}
	if len(o.Value) != len(xo.Value) {
		return false
	}
	for k, v := range o.Value {
		xv, ok := xo.Value[k]
		if !ok {
			return false
		}
		if len(v) != len(xv) {
			return false
		}
		for i, vv := range v {
			xvv := xv[i]
			if vv != xvv {
				return false
			}
		}
	}
	return true
}

// IndexGet returns the value for the given key.
func (o *Header) IndexGet(index tengo.Object) (tengo.Object, error) {
	key, ok := tengo.ToString(index)
	if !ok {
		return nil, tengo.ErrInvalidIndexType
	}
	res := o.Value.Get(key)
	return &tengo.String{Value: res}, nil
}

// IndexSet sets the value for the given key. If value is undefined, index
// is deleted.
func (o *Header) IndexSet(index, value tengo.Object) error {
	key, ok := tengo.ToString(index)
	if !ok {
		return tengo.ErrInvalidIndexType
	}

	if value == tengo.UndefinedValue {
		o.Value.Del(key)
		return nil
	}

	val, ok := value.(*tengo.String)
	if !ok {
		return tengo.ErrInvalidArgumentType{
			Name:     "value",
			Expected: "string",
			Found:    value.TypeName(),
		}
	}
	o.Value.Set(key, val.Value)
	return nil
}

// Iterate creates a header iterator.
func (o *Header) Iterate() tengo.Iterator {
	var keys []string
	for k := range o.Value {
		keys = append(keys, k)
	}
	return &HeaderIterator{
		v: o.Value,
		k: keys,
		l: len(keys),
	}
}

// CanIterate returns whether the Object can be Iterated.
func (o *Header) CanIterate() bool {
	return true
}

// HeaderIterator represents an iterator for the header.
type HeaderIterator struct {
	tengo.ObjectImpl
	v http.Header
	k []string
	i int
	l int
}

// TypeName returns the name of the type.
func (i *HeaderIterator) TypeName() string {
	return "header-iterator"
}

func (i *HeaderIterator) String() string {
	return "<header-iterator>"
}

// IsFalsy returns true if the value of the type is falsy.
func (i *HeaderIterator) IsFalsy() bool {
	return true
}

// Equals returns true if the value of the type is equal to the value of
// another object.
func (i *HeaderIterator) Equals(tengo.Object) bool {
	return false
}

// Copy returns a copy of the type.
func (i *HeaderIterator) Copy() tengo.Object {
	return &HeaderIterator{v: i.v, k: i.k, i: i.i, l: i.l}
}

// Next returns true if there are more elements to iterate.
func (i *HeaderIterator) Next() bool {
	i.i++
	return i.i <= i.l
}

// Key returns the key or index value of the current element.
func (i *HeaderIterator) Key() tengo.Object {
	k := i.k[i.i-1]
	return &tengo.String{Value: k}
}

// Value returns the value of the current element.
func (i *HeaderIterator) Value() tengo.Object {
	k := i.k[i.i-1]
	return &tengo.String{Value: i.v.Get(k)}
}
