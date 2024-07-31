// Code generated by thriftgo (0.3.15). DO NOT EDIT.

package deep

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
)

const (
	ConstString = "const string"
)

type FOO int64

const (
	FOO_B FOO = 0
	FOO_A FOO = 1
)

func (p FOO) String() string {
	switch p {
	case FOO_B:
		return "B"
	case FOO_A:
		return "A"
	}
	return "<UNSET>"
}

func FOOFromString(s string) (FOO, error) {
	switch s {
	case "B":
		return FOO_B, nil
	case "A":
		return FOO_A, nil
	}
	return FOO(0), fmt.Errorf("not a valid FOO string")
}

func FOOPtr(v FOO) *FOO { return &v }
func (p *FOO) Scan(value interface{}) (err error) {
	var result sql.NullInt64
	err = result.Scan(value)
	*p = FOO(result.Int64)
	return
}

func (p *FOO) Value() (driver.Value, error) {
	if p == nil {
		return nil, nil
	}
	return int64(*p), nil
}

type TestStruct struct {
	A int64  `thrift:"a,1" json:"a"`
	B string `thrift:"b,2" json:"b"`
}

func NewTestStruct() *TestStruct {
	return &TestStruct{}
}

func (p *TestStruct) InitDefault() {
}

func (p *TestStruct) GetA() (v int64) {
	return p.A
}

func (p *TestStruct) GetB() (v string) {
	return p.B
}

func (p *TestStruct) String() string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprintf("TestStruct(%+v)", *p)
}

var fieldIDToName_TestStruct = map[int16]string{
	1: "a",
	2: "b",
}
