package proto

import (
	"errors"
	"fmt"
	"io"
)

// Errors in encoding and decoding Protobuf wiretypes.
const (
	_ = -iota
	ErrCodeTruncated
	ErrCodeFieldNumber
	ErrCodeOverflow
	ErrCodeReserved
	ErrCodeEndGroup
	ErrCodeRecursionDepth
)

var (
	errFieldNumber = errors.New("invalid field number")
	errOverflow    = errors.New("variable length integer overflow")
	errReserved    = errors.New("cannot parse reserved wire type")
	errEndGroup    = errors.New("mismatching end group marker")
	errParse       = errors.New("parse error")
)

// prefixError add prefix string to specfic error message
type prefixError struct{ s string }

type wrapError struct {
	s   string
	err error
}

// ParseError converts an error code into an error value.
// This returns nil if n is a non-negative number.
func ParseError(n int) error {
	if n >= 0 {
		return nil
	}
	switch n {
	case ErrCodeTruncated:
		return io.ErrUnexpectedEOF
	case ErrCodeFieldNumber:
		return errFieldNumber
	case ErrCodeOverflow:
		return errOverflow
	case ErrCodeReserved:
		return errReserved
	case ErrCodeEndGroup:
		return errEndGroup
	default:
		return errParse
	}
}

func (e *prefixError) Error() string {
	return "proto: " + e.s
}

func New(f string, x ...interface{}) error {
	return &prefixError{s: format(f, x...)}
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
