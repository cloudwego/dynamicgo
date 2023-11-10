package errors

import (
	"errors"
	"fmt"
)

// not used this package
var (
	prefix = "proto: "
	Error  = errors.New("protobuf error")
)

type prefixError struct{ s string }

func (e *prefixError) Error() string {
	return prefix + e.s
}

func (e *prefixError) Unwrap() error {
	return Error
}

func New(f string, x ...interface{}) error {
	return &prefixError{s: format(f, x...)}
}

type wrapError struct {
	s   string
	err error
}

func format(f string, x ...interface{}) string {
	// avoid "proto: " prefix when chaining
	for i := 0; i < len(x); i++ {
		switch e := x[i].(type) {
		case *prefixError:
			x[i] = e.s
		case *wrapError:
			x[i] = format("%v: %v", e.s, e.err)
		}
	}
	return fmt.Sprintf(f, x...)
}

func InvalidUTF8(name string) error {
	return New("field %v contains invalid UTF-8", name)
}
