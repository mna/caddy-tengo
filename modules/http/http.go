package http

import (
	"net/http"

	"github.com/d5/tengo/v2"
)

var httpModule = map[string]tengo.Object{}

func NewResponseWriter(w http.ResponseWriter) *tengo.ImmutableMap {
	var hdr *Header

	return &tengo.ImmutableMap{
		Value: map[string]tengo.Object{
			"header": &tengo.UserFunction{
				Name: "header",
				Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
					if len(args) != 0 {
						return nil, tengo.ErrWrongNumArguments
					}
					if hdr == nil {
						hdr = &Header{Value: w.Header()}
					}
					return hdr, nil
				},
			},

			"write": &tengo.UserFunction{
				Name: "write",
				Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
					if len(args) != 1 {
						return nil, tengo.ErrWrongNumArguments
					}

					// allow writing a String, Bytes or Char
					var b []byte
					switch a0 := args[0].(type) {
					case *tengo.Bytes:
						b = a0.Value
					case *tengo.String:
						b = []byte(a0.Value)
					case *tengo.Char:
						b = []byte(string(a0.Value))
					default:
						return nil, tengo.ErrInvalidArgumentType{
							Name:     "first",
							Expected: "bytes, string or char",
							Found:    args[0].TypeName(),
						}
					}
					n, err := w.Write(b)
					return &tengo.Int{Value: int64(n)}, err
				},
			},

			"write_header": &tengo.UserFunction{
				Name: "write_header",
				Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
					if len(args) != 1 {
						return nil, tengo.ErrWrongNumArguments
					}
					a0, ok := args[0].(*tengo.Int)
					if !ok {
						return nil, tengo.ErrInvalidArgumentType{
							Name:     "first",
							Expected: "integer",
							Found:    args[0].TypeName(),
						}
					}
					w.WriteHeader(int(a0.Value))
					return tengo.UndefinedValue, nil
				},
			},
		},
	}
}
