/**
 * Copyright 2023 CloudWeGo Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package meta

import (
	"fmt"
	"strings"
)

// CategoryBitOnErrorCode is used to shift category on error code
const CategoryBitOnErrorCode = 24

// ErrCode is the error code of dynamicgo.
// Usually the left 8 bits are used to represent the category of error,
// and the right 24 bits are used to represent the behavior of error.
type ErrCode uint32

// Error Behaviors
const (
	// ErrUnsupportedType represents unsupported type error
	ErrUnsupportedType ErrCode = 1 + iota
	// ErrStackOverflow represents exceed stack limit error
	ErrStackOverflow
	// ErrRead represents read error
	ErrRead
	// ErrWrite represents write error
	ErrWrite
	// ErrDismatchType represents dismatch type error
	ErrDismatchType
	// ErrConvert represents convert error
	ErrConvert
	// ErrNotFound represents not found error
	ErrNotFound
	// ErrMissRequiredField represents missing required field error
	ErrMissRequiredField
	// ErrUnknownField represents unknown field error
	ErrUnknownField
	// ErrInvalidParam represents invalid parameter error
	ErrInvalidParam
	// ErrNotImplemented represents not implemented error
	ErrNotImplemented
)

var errsMap = map[ErrCode]string{
	ErrUnsupportedType:   "unsupported type",
	ErrStackOverflow:     "exceed depth limit",
	ErrRead:              "read failed",
	ErrWrite:             "write failed",
	ErrDismatchType:      "dismatched type",
	ErrConvert:           "convert failed",
	ErrNotFound:          "not found",
	ErrMissRequiredField: "missing required field",
	ErrUnknownField:      "unknown field",
	ErrInvalidParam:      "invalid parameter",
}

// NewErrorCode created a new error code with category and behavior
func NewErrorCode(behavior ErrCode, category Category) ErrCode {
	return behavior | (ErrCode(category) << CategoryBitOnErrorCode)
}

// Category returns the category of error code
func (ec ErrCode) Category() Category {
	return Category(ec >> CategoryBitOnErrorCode)
}

// String returns the string representation of error code
func (ec ErrCode) String() string {
	if m, ok := errsMap[ec.Behavior()]; ok {
		return m
	} else {
		return fmt.Sprintf("error code %d", ec)
	}
}

// Error implement error interface
func (ec ErrCode) Error() string {
	return ec.String()
}

// Behavior returns the behavior of error code
func (ec ErrCode) Behavior() ErrCode {
	return ErrCode(ec & 0x00ffffff)
}

// Error is the error concrete type of dynamicgo
type Error struct {
	Code ErrCode
	Msg  string
	Err  error
}

// NewError creates a new error with error code, message and preceding error
//
//go:noinline
func NewError(code ErrCode, msg string, err error) error {
	return Error{
		Code: code,
		Msg:  msg,
		Err:  err,
	}
}

// Message return current message if has,
// otherwise return preceding error message
func (err Error) Message() string {
	output := []string{err.Msg}
	if err.Err != nil {
		output = append(output, err.Err.Error())
	}
	return strings.Join(output, "\n")
}

// Error return error message,
// combining category, behavior and message
func (err Error) Error() string {
	return fmt.Sprintf("[%s] %s: %s", err.Code.Category(), err.Code.Behavior(), err.Message())
}

// Unwrap implements errors.Unwrap interface
func (err Error) Unwrap() error {
	return err.Err
}
