package protowire_test

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math"
	"reflect"
	"strconv"
	"testing"

	"github.com/cloudwego/dynamicgo/proto"
	"github.com/cloudwego/dynamicgo/proto/protowire"
)

type (
	testOps struct {
		// appendOps is a sequence of append operations, each appending to
		// the output of the previous append operation.
		appendOps []appendOp

		// wantRaw (if not nil) is the bytes that the appendOps should produce.
		wantRaw []byte

		// consumeOps are a sequence of consume operations, each consuming the
		// remaining output after the previous consume operation.
		// The first consume operation starts with the output of appendOps.
		consumeOps []consumeOp
	}

	appendOp = interface{}

	appendVarint struct {
		InVal uint64
	}
	appendFixed32 struct {
		InVal uint32
	}
	appendFixed64 struct {
		InVal uint64
	}
	encodeBool struct {
		InVal bool
	}
	encodeByte struct {
		InVal byte
	}
	encodeEnum struct {
		InVal int32
	}
	encodeInt32 struct {
		InVal int32
	}
	encodeSint32 struct {
		InVal int32
	}
	encodeUint32 struct {
		InVal uint32
	}
	encodeInt64 struct {
		InVal int64
	}
	encodeSint64 struct {
		InVal int64
	}
	encodeUint64 struct {
		InVal uint64
	}
	encodeFixed32 struct {
		InVal uint32
	}
	encodeSFixed32 struct {
		InVal int32
	}
	encodeFloat32 struct {
		InVal float32
	}
	encodeFixed64 struct {
		InVal uint64
	}
	encodeSFixed64 struct {
		InVal int64
	}
	encodeDouble struct {
		InVal float64
	}
	encodeString struct {
		InVal string
	}
	encodeBytes struct {
		intVal []byte
	}

	appendRaw []byte

	consumeOp     = interface{}
	consumeVarint struct {
		wantVal uint64
		wantCnt int
		wantErr error
	}
	consumeFixed32 struct {
		wantVal uint32
		wantCnt int
		wantErr error
	}
	consumeFixed64 struct {
		wantVal uint64
		wantCnt int
		wantErr error
	}
	consumeBytes struct {
		wantVal []byte
		wantCnt int
		wantErr error
	}
	decodeBool struct {
		wantVal bool
		wantCnt int
		wantErr error
	}
	decodeByte struct {
		wantVal byte
		wantCnt int
		wantErr error
	}
	decodeEnum struct {
		wantVal int
		wantCnt int
		wantErr error
	}
	decodeInt32 struct {
		wantVal int32
		wantCnt int
		wantErr error
	}
	decodeSint32 struct {
		wantVal int32
		wantCnt int
		wantErr error
	}
	decodeUint32 struct {
		wantVal uint32
		wantCnt int
		wantErr error
	}
	decodeInt64 struct {
		wantVal int64
		wantCnt int
		wantErr error
	}
	decodeSint64 struct {
		wantVal int64
		wantCnt int
		wantErr error
	}
	decodeUint64 struct {
		wantVal uint64
		wantCnt int
		wantErr error
	}
	decodeFixed32 struct {
		wantVal uint32
		wantCnt int
		wantErr error
	}
	decodeSfixed32 struct {
		wantVal int32
		wantCnt int
		wantErr error
	}
	decodeFloat32 struct {
		wantVal float32
		wantCnt int
		wantErr error
	}
	decodeFixed64 struct {
		wantVal uint64
		wantCnt int
		wantErr error
	}
	decodeSfixed64 struct {
		wantVal int64
		wantCnt int
		wantErr error
	}
	decodeDouble struct {
		wantVal float64
		wantCnt int
		wantErr error
	}
	decodeString struct {
		wantVal string
		wantCnt int
		wantErr error
	}
	decodeBytes struct {
		wantVal []byte
		wantCnt int
		wantErr error
	}
	ops []interface{}
)

// dhex decodes a hex-string and returns the bytes and panics if s is invalid.
func dhex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

var (
	errFieldNumber = errors.New("invalid field number")
	errOverflow    = errors.New("variable length integer overflow")
	errReserved    = errors.New("cannot parse reserved wire type")
	errEndGroup    = errors.New("mismatching end group marker")
	errParse       = errors.New("parse error")
)

// ParseError converts an error code into an error value.
// This returns nil if n is a non-negative number.
func ParseError(n int) error {
	if n >= 0 {
		return nil
	}
	switch n {
	case proto.ErrCodeTruncated:
		return io.ErrUnexpectedEOF
	case proto.ErrCodeFieldNumber:
		return errFieldNumber
	case proto.ErrCodeOverflow:
		return errOverflow
	case proto.ErrCodeReserved:
		return errReserved
	case proto.ErrCodeEndGroup:
		return errEndGroup
	default:
		return errParse
	}
}

var OpType2Name = map[interface{}]string{
	reflect.TypeOf(appendVarint{}):   "Varint_en",
	reflect.TypeOf(appendFixed32{}):  "Fixed32_en",
	reflect.TypeOf(appendFixed64{}):  "Fixed64_en",
	reflect.TypeOf(encodeBool{}):     "Bool_en",
	reflect.TypeOf(encodeByte{}):     "Byte_en",
	reflect.TypeOf(encodeEnum{}):     "Enum_en",
	reflect.TypeOf(encodeInt32{}):    "Int32_en",
	reflect.TypeOf(encodeSint32{}):   "Sint32_en",
	reflect.TypeOf(encodeUint32{}):   "Uint32_en",
	reflect.TypeOf(encodeInt64{}):    "Int64_en",
	reflect.TypeOf(encodeSint64{}):   "Sint64_en",
	reflect.TypeOf(encodeUint64{}):   "Uint64_en",
	reflect.TypeOf(encodeSFixed32{}): "Sfixed32_en",
	reflect.TypeOf(encodeFixed32{}):  "EncodeFixed32_en",
	reflect.TypeOf(encodeSFixed64{}): "Sfixed64_en",
	reflect.TypeOf(encodeFixed64{}):  "EncodeFixed64_en",
	reflect.TypeOf(encodeDouble{}):   "Double_en",
	reflect.TypeOf(encodeString{}):   "String_en",
	reflect.TypeOf(encodeBytes{}):    "Bytes_en",
}

func TestVarint(t *testing.T) {
	runTests(t, []testOps{
		{
			// Test varints at various boundaries where the length changes.
			appendOps:  ops{appendVarint{InVal: 0x0}},
			wantRaw:    dhex("00"),
			consumeOps: ops{consumeVarint{wantVal: 0, wantCnt: 1}},
		}, {
			appendOps:  ops{appendVarint{InVal: 0x1}},
			wantRaw:    dhex("01"),
			consumeOps: ops{consumeVarint{wantVal: 1, wantCnt: 1}},
		}, {
			appendOps:  ops{appendVarint{InVal: 0x7f}},
			wantRaw:    dhex("7f"),
			consumeOps: ops{consumeVarint{wantVal: 0x7f, wantCnt: 1}},
		}, {
			appendOps:  ops{appendVarint{InVal: 0x7f + 1}},
			wantRaw:    dhex("8001"),
			consumeOps: ops{consumeVarint{wantVal: 0x7f + 1, wantCnt: 2}},
		}, {
			appendOps:  ops{appendVarint{InVal: 0x3fff}},
			wantRaw:    dhex("ff7f"),
			consumeOps: ops{consumeVarint{wantVal: 0x3fff, wantCnt: 2}},
		}, {
			appendOps:  ops{appendVarint{InVal: 0x3fff + 1}},
			wantRaw:    dhex("808001"),
			consumeOps: ops{consumeVarint{wantVal: 0x3fff + 1, wantCnt: 3}},
		}, {
			appendOps:  ops{appendVarint{InVal: 0x1fffff}},
			wantRaw:    dhex("ffff7f"),
			consumeOps: ops{consumeVarint{wantVal: 0x1fffff, wantCnt: 3}},
		}, {
			appendOps:  ops{appendVarint{InVal: 0x1fffff + 1}},
			wantRaw:    dhex("80808001"),
			consumeOps: ops{consumeVarint{wantVal: 0x1fffff + 1, wantCnt: 4}},
		}, {
			appendOps:  ops{appendVarint{InVal: 0xfffffff}},
			wantRaw:    dhex("ffffff7f"),
			consumeOps: ops{consumeVarint{wantVal: 0xfffffff, wantCnt: 4}},
		}, {
			appendOps:  ops{appendVarint{InVal: 0xfffffff + 1}},
			wantRaw:    dhex("8080808001"),
			consumeOps: ops{consumeVarint{wantVal: 0xfffffff + 1, wantCnt: 5}},
		}, {
			appendOps:  ops{appendVarint{InVal: 0x7ffffffff}},
			wantRaw:    dhex("ffffffff7f"),
			consumeOps: ops{consumeVarint{wantVal: 0x7ffffffff, wantCnt: 5}},
		}, {
			appendOps:  ops{appendVarint{InVal: 0x7ffffffff + 1}},
			wantRaw:    dhex("808080808001"),
			consumeOps: ops{consumeVarint{wantVal: 0x7ffffffff + 1, wantCnt: 6}},
		}, {
			appendOps:  ops{appendVarint{InVal: 0x3ffffffffff}},
			wantRaw:    dhex("ffffffffff7f"),
			consumeOps: ops{consumeVarint{wantVal: 0x3ffffffffff, wantCnt: 6}},
		}, {
			appendOps:  ops{appendVarint{InVal: 0x3ffffffffff + 1}},
			wantRaw:    dhex("80808080808001"),
			consumeOps: ops{consumeVarint{wantVal: 0x3ffffffffff + 1, wantCnt: 7}},
		}, {
			appendOps:  ops{appendVarint{InVal: 0x1ffffffffffff}},
			wantRaw:    dhex("ffffffffffff7f"),
			consumeOps: ops{consumeVarint{wantVal: 0x1ffffffffffff, wantCnt: 7}},
		}, {
			appendOps:  ops{appendVarint{InVal: 0x1ffffffffffff + 1}},
			wantRaw:    dhex("8080808080808001"),
			consumeOps: ops{consumeVarint{wantVal: 0x1ffffffffffff + 1, wantCnt: 8}},
		}, {
			appendOps:  ops{appendVarint{InVal: 0xffffffffffffff}},
			wantRaw:    dhex("ffffffffffffff7f"),
			consumeOps: ops{consumeVarint{wantVal: 0xffffffffffffff, wantCnt: 8}},
		}, {
			appendOps:  ops{appendVarint{InVal: 0xffffffffffffff + 1}},
			wantRaw:    dhex("808080808080808001"),
			consumeOps: ops{consumeVarint{wantVal: 0xffffffffffffff + 1, wantCnt: 9}},
		}, {
			appendOps:  ops{appendVarint{InVal: 0x7fffffffffffffff}},
			wantRaw:    dhex("ffffffffffffffff7f"),
			consumeOps: ops{consumeVarint{wantVal: 0x7fffffffffffffff, wantCnt: 9}},
		}, {
			appendOps:  ops{appendVarint{InVal: 0x7fffffffffffffff + 1}},
			wantRaw:    dhex("80808080808080808001"),
			consumeOps: ops{consumeVarint{wantVal: 0x7fffffffffffffff + 1, wantCnt: 10}},
		}, {
			appendOps:  ops{appendVarint{InVal: math.MaxUint64}},
			wantRaw:    dhex("ffffffffffffffffff01"),
			consumeOps: ops{consumeVarint{wantVal: math.MaxUint64, wantCnt: 10}},
		},
	})
}

func TestFixed32(t *testing.T) {
	runTests(t, []testOps{{
		appendOps:  ops{appendFixed32{0}},
		wantRaw:    dhex("00000000"),
		consumeOps: ops{consumeFixed32{wantVal: 0, wantCnt: 4}},
	}, {
		appendOps:  ops{appendFixed32{math.MaxUint32}},
		wantRaw:    dhex("ffffffff"),
		consumeOps: ops{consumeFixed32{wantVal: math.MaxUint32, wantCnt: 4}},
	}, {
		appendOps:  ops{appendFixed32{0xf0e1d2c3}},
		wantRaw:    dhex("c3d2e1f0"),
		consumeOps: ops{consumeFixed32{wantVal: 0xf0e1d2c3, wantCnt: 4}},
	}})
}

func TestFixed64(t *testing.T) {
	runTests(t, []testOps{{
		appendOps:  ops{appendFixed64{0}},
		wantRaw:    dhex("0000000000000000"),
		consumeOps: ops{consumeFixed64{wantVal: 0, wantCnt: 8}},
	}, {
		appendOps:  ops{appendFixed64{math.MaxUint64}},
		wantRaw:    dhex("ffffffffffffffff"),
		consumeOps: ops{consumeFixed64{wantVal: math.MaxUint64, wantCnt: 8}},
	}, {
		appendOps:  ops{appendFixed64{0xf0e1d2c3b4a59687}},
		wantRaw:    dhex("8796a5b4c3d2e1f0"),
		consumeOps: ops{consumeFixed64{wantVal: 0xf0e1d2c3b4a59687, wantCnt: 8}},
	}})
}

func TestBool(t *testing.T) {
	runTests(t, []testOps{{
		appendOps:  ops{encodeBool{true}},
		wantRaw:    dhex("01"),
		consumeOps: ops{decodeBool{wantVal: true, wantCnt: 1}},
	}, {
		appendOps:  ops{encodeBool{false}},
		wantRaw:    dhex("00"),
		consumeOps: ops{decodeBool{wantVal: false, wantCnt: 1}},
	}})
}

func TestByte(t *testing.T) {
	runTests(t, []testOps{{
		appendOps:  ops{encodeByte{0x12}},
		wantRaw:    []byte{0x12},
		consumeOps: ops{decodeByte{wantVal: 0x12, wantCnt: 1}},
	}, {
		appendOps:  ops{encodeByte{0xff}},
		wantRaw:    []byte{0xff},
		consumeOps: ops{decodeByte{wantVal: 0xff, wantCnt: 1}},
	}})
}

func TestInt32(t *testing.T) {
	runTests(t, []testOps{
		{
			appendOps:  ops{encodeInt32{26984}},
			wantRaw:    dhex("e8d201"),
			consumeOps: ops{decodeInt32{wantVal: 26984, wantCnt: 3}},
		},
		{
			appendOps:  ops{encodeInt32{123}},
			wantRaw:    []byte{0x7b},
			consumeOps: ops{decodeInt32{wantVal: 123, wantCnt: 1}},
		},
		{
			appendOps:  ops{encodeInt32{2147483647}},
			wantRaw:    dhex("ffffffff07"),
			consumeOps: ops{decodeInt32{wantVal: 2147483647, wantCnt: 5}},
		},
	})
}

func TestSint32(t *testing.T) {
	runTests(t, []testOps{
		{
			appendOps:  ops{encodeSint32{26984}},
			wantRaw:    dhex("d0a503"),
			consumeOps: ops{decodeSint32{wantVal: 26984, wantCnt: 3}},
		},
		{
			appendOps:  ops{encodeSint32{-1}},
			wantRaw:    []byte{0x1},
			consumeOps: ops{decodeSint32{wantVal: -1, wantCnt: 1}},
		},
		{
			appendOps:  ops{encodeSint32{-200}},
			wantRaw:    dhex("8f03"),
			consumeOps: ops{decodeSint32{wantVal: -200, wantCnt: 2}},
		},
	})
}

func TestUint32(t *testing.T) {
	runTests(t, []testOps{
		{
			appendOps:  ops{encodeUint32{26984}},
			wantRaw:    dhex("e8d201"),
			consumeOps: ops{decodeUint32{wantVal: 26984, wantCnt: 3}},
		},
		{
			appendOps:  ops{encodeUint32{10000}},
			wantRaw:    dhex("904e"),
			consumeOps: ops{decodeUint32{wantVal: 10000, wantCnt: 2}},
		},
		{
			appendOps:  ops{encodeUint32{200}},
			wantRaw:    dhex("c801"),
			consumeOps: ops{decodeUint32{wantVal: 200, wantCnt: 2}},
		},
	})
}

func TestInt64(t *testing.T) {
	runTests(t, []testOps{
		{
			appendOps:  ops{encodeInt64{26984}},
			wantRaw:    dhex("e8d201"),
			consumeOps: ops{decodeInt64{wantVal: 26984, wantCnt: 3}},
		},
		{
			appendOps:  ops{encodeInt64{123}},
			wantRaw:    []byte{0x7b},
			consumeOps: ops{decodeInt64{wantVal: 123, wantCnt: 1}},
		},
		{
			appendOps:  ops{encodeInt64{2147483647}},
			wantRaw:    dhex("ffffffff07"),
			consumeOps: ops{decodeInt64{wantVal: 2147483647, wantCnt: 5}},
		},
	})
}

func TestSint64(t *testing.T) {
	runTests(t, []testOps{
		{
			appendOps:  ops{encodeSint64{26984}},
			wantRaw:    dhex("d0a503"),
			consumeOps: ops{decodeSint64{wantVal: 26984, wantCnt: 3}},
		},
		{
			appendOps:  ops{encodeSint64{-1}},
			wantRaw:    []byte{0x1},
			consumeOps: ops{decodeSint64{wantVal: -1, wantCnt: 1}},
		},
		{
			appendOps:  ops{encodeSint64{-200}},
			wantRaw:    dhex("8f03"),
			consumeOps: ops{decodeSint64{wantVal: -200, wantCnt: 2}},
		},
	})
}

func TestUint64(t *testing.T) {
	runTests(t, []testOps{
		{
			appendOps:  ops{encodeUint64{26984}},
			wantRaw:    dhex("e8d201"),
			consumeOps: ops{decodeUint64{wantVal: 26984, wantCnt: 3}},
		},
		{
			appendOps:  ops{encodeUint64{10000}},
			wantRaw:    dhex("904e"),
			consumeOps: ops{decodeUint64{wantVal: 10000, wantCnt: 2}},
		},
		{
			appendOps:  ops{encodeUint64{200}},
			wantRaw:    dhex("c801"),
			consumeOps: ops{decodeUint64{wantVal: 200, wantCnt: 2}},
		},
	})
}

func TestSfixed32(t *testing.T) {
	runTests(t, []testOps{{
		appendOps:  ops{encodeSFixed32{0}},
		wantRaw:    dhex("00000000"),
		consumeOps: ops{decodeSfixed32{wantVal: 0, wantCnt: 4}},
	}, {
		appendOps:  ops{encodeSFixed32{math.MaxInt32}},
		wantRaw:    dhex("ffffff7f"),
		consumeOps: ops{decodeSfixed32{wantVal: math.MaxInt32, wantCnt: 4}},
	}, {
		appendOps:  ops{encodeSFixed32{1}},
		wantRaw:    dhex("01000000"),
		consumeOps: ops{decodeSfixed32{wantVal: 1, wantCnt: 4}},
	}})
}

func TestFloat32(t *testing.T) {
	runTests(t, []testOps{
		{
			appendOps:  ops{encodeFloat32{0.32}},
			wantRaw:    dhex("0ad7a33e"),
			consumeOps: ops{decodeFloat32{wantVal: 0.32, wantCnt: 4}},
		},
		{
			appendOps:  ops{encodeFloat32{1298.222}},
			wantRaw:    dhex("1b47a244"),
			consumeOps: ops{decodeFloat32{wantVal: 1298.222, wantCnt: 4}},
		},
		{
			appendOps:  ops{encodeFloat32{9898989.11}},
			wantRaw:    dhex("ed0b174b"),
			consumeOps: ops{decodeFloat32{wantVal: 9898989.11, wantCnt: 4}},
		},
	})
}

func TestSfixed64(t *testing.T) {
	runTests(t, []testOps{
		{
			appendOps:  ops{encodeSFixed64{0}},
			wantRaw:    dhex("0000000000000000"),
			consumeOps: ops{decodeSfixed64{wantVal: 0, wantCnt: 8}},
		},
		{
			appendOps:  ops{appendFixed64{math.MaxUint64}},
			wantRaw:    dhex("ffffffffffffffff"),
			consumeOps: ops{consumeFixed64{wantVal: math.MaxUint64, wantCnt: 8}},
		}, {
			appendOps:  ops{appendFixed64{0xf0e1d2c3b4a59687}},
			wantRaw:    dhex("8796a5b4c3d2e1f0"),
			consumeOps: ops{consumeFixed64{wantVal: 0xf0e1d2c3b4a59687, wantCnt: 8}},
		},
	})
}

func TestDouble(t *testing.T) {
	runTests(t, []testOps{
		{
			appendOps:  ops{encodeDouble{0.32}},
			wantRaw:    dhex("7b14ae47e17ad43f"),
			consumeOps: ops{decodeDouble{wantVal: 0.32, wantCnt: 8}},
		},
		{
			appendOps:  ops{encodeDouble{1298.222}},
			wantRaw:    dhex("d9cef753e3489440"),
			consumeOps: ops{decodeDouble{wantVal: 1298.222, wantCnt: 8}},
		},
		{
			appendOps:  ops{encodeDouble{9898989.11}},
			wantRaw:    dhex("b81e85a37de16241"),
			consumeOps: ops{decodeDouble{wantVal: 9898989.11, wantCnt: 8}},
		},
	})
}

func TestString(t *testing.T) {
	runTests(t, []testOps{
		{
			appendOps:  ops{encodeString{"aabbcc"}},
			wantRaw:    dhex("06616162626363"),
			consumeOps: ops{decodeString{wantVal: "aabbcc", wantCnt: 7}},
		},
		{
			appendOps:  ops{encodeString{"dau9bkn;jasdf"}},
			wantRaw:    dhex("0d64617539626b6e3b6a61736466"),
			consumeOps: ops{decodeString{wantVal: "dau9bkn;jasdf", wantCnt: 14}},
		},
	})
}

func TestBytes(t *testing.T) {
	runTests(t, []testOps{
		{
			appendOps:  ops{encodeBytes{[]byte{0x61, 0x61, 0x62, 0x62, 0x63, 0x63}}},
			wantRaw:    dhex("06616162626363"),
			consumeOps: ops{decodeBytes{wantVal: []byte{0x61, 0x61, 0x62, 0x62, 0x63, 0x63}, wantCnt: 7}},
		},
		{
			appendOps:  ops{encodeBytes{[]byte{0x64, 0x61, 0x75, 0x39, 0x62, 0x6b, 0x6e, 0x3b, 0x6a, 0x61, 0x73, 0x64, 0x66}}},
			wantRaw:    dhex("0d64617539626b6e3b6a61736466"),
			consumeOps: ops{decodeBytes{wantVal: []byte{0x64, 0x61, 0x75, 0x39, 0x62, 0x6b, 0x6e, 0x3b, 0x6a, 0x61, 0x73, 0x64, 0x66}, wantCnt: 14}},
		},
	})
}

func runTests(t *testing.T, tests []testOps) {
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			var b []byte
			for _, op := range tt.appendOps {
				b0 := b
				switch op := op.(type) {
				case appendVarint:
					b = protowire.AppendVarint(b, op.InVal)
				case appendFixed32:
					b = protowire.AppendFixed32(b, op.InVal)
				case appendFixed64:
					b = protowire.AppendFixed64(b, op.InVal)
				case encodeBool:
					b = protowire.BinaryEncoder{}.EncodeBool(b, op.InVal)
				case encodeByte:
					b = protowire.BinaryEncoder{}.EncodeByte(b, op.InVal)
				case encodeEnum:
					b = protowire.BinaryEncoder{}.EncodeEnum(b, op.InVal)
				case encodeInt32:
					b = protowire.BinaryEncoder{}.EncodeInt32(b, op.InVal)
				case encodeSint32:
					b = protowire.BinaryEncoder{}.EncodeSint32(b, op.InVal)
				case encodeUint32:
					b = protowire.BinaryEncoder{}.EncodeUint32(b, op.InVal)
				case encodeInt64:
					b = protowire.BinaryEncoder{}.EncodeInt64(b, op.InVal)
				case encodeSint64:
					b = protowire.BinaryEncoder{}.EncodeSint64(b, op.InVal)
				case encodeUint64:
					b = protowire.BinaryEncoder{}.EncodeUint64(b, op.InVal)
				case encodeSFixed32:
					b = protowire.BinaryEncoder{}.EncodeSfixed32(b, op.InVal)
				case encodeFixed32:
					b = protowire.BinaryEncoder{}.EncodeFixed32(b, op.InVal)
				case encodeSFixed64:
					b = protowire.BinaryEncoder{}.EncodeSfixed64(b, op.InVal)
				case encodeFixed64:
					b = protowire.BinaryEncoder{}.EncodeFixed64(b, op.InVal)
				case encodeFloat32:
					b = protowire.BinaryEncoder{}.EncodeFloat32(b, op.InVal)
				case encodeDouble:
					b = protowire.BinaryEncoder{}.EncodeDouble(b, op.InVal)
				case encodeString:
					b = protowire.BinaryEncoder{}.EncodeString(b, op.InVal)
				case encodeBytes:
					b = protowire.BinaryEncoder{}.EncodeBytes(b, op.intVal)
				case appendRaw:
					b = append(b, op...)
				}

				check := func(label string, want int) {
					t.Helper()
					if got := len(b) - len(b0); got != want {
						t.Errorf("len(Append%v) and Size%v mismatch: got %v, want %v", label, label, got, want)
					}
				}

				switch op := op.(type) {
				case appendVarint, encodeEnum, encodeInt32, encodeUint32, encodeInt64, encodeUint64:
					data_str := fmt.Sprint(reflect.ValueOf(op).Field(0).Interface())
					data_conv, _ := strconv.ParseUint(data_str, 10, 64)
					check(OpType2Name[op], protowire.SizeVarint(data_conv))
				case appendFixed32, encodeFixed32, encodeSFixed32, encodeFloat32:
					check(OpType2Name[op], protowire.SizeFixed32())
				case appendFixed64, encodeFixed64, encodeSFixed64:
					check(OpType2Name[op], protowire.SizeFixed64())
				case encodeBool, encodeByte:
					check(OpType2Name[op], 1)
				case encodeSint32, encodeSint64:
					data_str := fmt.Sprint(reflect.ValueOf(op).Field(0).Interface())
					data_conv, _ := strconv.ParseInt(data_str, 10, 64)
					check(OpType2Name[op], protowire.SizeVarint(protowire.EncodeZigZag(data_conv)))
				}
			}

			if tt.wantRaw != nil && !bytes.Equal(b, tt.wantRaw) {
				t.Errorf("raw output mismatch:\ngot  %x\nwant %x", b, tt.wantRaw)
			}

			for _, op := range tt.consumeOps {
				check := func(label string, gotCnt, wantCnt int, wantErr error) {
					t.Helper()
					gotErr := ParseError(gotCnt)
					if gotCnt < 0 {
						gotCnt = 0
					}
					if gotCnt != wantCnt {
						t.Errorf("Consume%v(): consumed %d bytes, want %d bytes consumed", label, gotCnt, wantCnt)
					}
					if gotErr != wantErr {
						t.Errorf("Consume%v(): got %v error, want %v error", label, gotErr, wantErr)
					}
					b = b[gotCnt:]
				}
				switch op := op.(type) {
				case consumeVarint:
					gotVal, n := protowire.ConsumeVarint(b)
					if gotVal != op.wantVal {
						t.Errorf("ConsumeVarint() = %d, want %d", gotVal, op.wantVal)
					}
					check("Varint_de", n, op.wantCnt, op.wantErr)
				case consumeFixed32:
					gotVal, n := protowire.ConsumeFixed32(b)
					if gotVal != op.wantVal {
						t.Errorf("ConsumeFixed32() = %d, want %d", gotVal, op.wantVal)
					}
					check("Fixed32_de", n, op.wantCnt, op.wantErr)
				case consumeFixed64:
					gotVal, n := protowire.ConsumeFixed64(b)
					if gotVal != op.wantVal {
						t.Errorf("ConsumeFixed64() = %d, want %d", gotVal, op.wantVal)
					}
					check("Fixed64_de", n, op.wantCnt, op.wantErr)
				case consumeBytes:
					gotVal, _, n := protowire.ConsumeBytes(b)
					if !bytes.Equal(gotVal, op.wantVal) {
						t.Errorf("ConsumeBytes() = %x, want %x", gotVal, op.wantVal)
					}
					check("Bytes_de", n, op.wantCnt, op.wantErr)
				case decodeBool:
					gotVal, n := protowire.BinaryDecoder{}.DecodeBool(b)
					if gotVal != op.wantVal {
						t.Errorf("DecodeBool() = %t, want %t", gotVal, op.wantVal)
					}
					check("Bool_de", n, op.wantCnt, op.wantErr)
				case decodeByte:
					gotVal := protowire.BinaryDecoder{}.DecodeByte(b)
					if gotVal != op.wantVal {
						t.Errorf("DecodeByte() = %d, want %d", gotVal, op.wantVal)
					}
					check("Byte_de", 1, op.wantCnt, op.wantErr)
				case decodeInt32:
					gotVal, n := protowire.BinaryDecoder{}.DecodeInt32(b)
					if gotVal != op.wantVal {
						t.Errorf("DecodeInt32() = %d, want %d", gotVal, op.wantVal)
					}
					check("Int32_de", n, op.wantCnt, op.wantErr)
				case decodeSint32:
					gotVal, n := protowire.BinaryDecoder{}.DecodeSint32(b)
					if gotVal != op.wantVal {
						t.Errorf("DecodeSint32() = %d, want %d", gotVal, op.wantVal)
					}
					check("Sint32_de", n, op.wantCnt, op.wantErr)
				case decodeUint32:
					gotVal, n := protowire.BinaryDecoder{}.DecodeUint32(b)
					if gotVal != op.wantVal {
						t.Errorf("DecodeUint32() = %d, want %d", gotVal, op.wantVal)
					}
					check("Uint32_de", n, op.wantCnt, op.wantErr)
				case decodeInt64:
					gotVal, n := protowire.BinaryDecoder{}.DecodeInt64(b)
					if gotVal != op.wantVal {
						t.Errorf("DecodeInt64() = %d, want %d", gotVal, op.wantVal)
					}
					check("Int64_de", n, op.wantCnt, op.wantErr)
				case decodeSint64:
					gotVal, n := protowire.BinaryDecoder{}.DecodeSint64(b)
					if gotVal != op.wantVal {
						t.Errorf("DecodeSint64() = %d, want %d", gotVal, op.wantVal)
					}
					check("Sint64_de", n, op.wantCnt, op.wantErr)
				case decodeUint64:
					gotVal, n := protowire.BinaryDecoder{}.DecodeUint64(b)
					if gotVal != op.wantVal {
						t.Errorf("DecodeUint64() = %d, want %d", gotVal, op.wantVal)
					}
					check("Uint64_de", n, op.wantCnt, op.wantErr)
				case decodeFixed32:
					gotVal, n := protowire.BinaryDecoder{}.DecodeFixed32(b)
					if gotVal != op.wantVal {
						t.Errorf("DecodeFixed32() = %d, want %d", gotVal, op.wantVal)
					}
					check("DecodeFixed32_de", n, op.wantCnt, op.wantErr)
				case decodeSfixed32:
					gotVal, n := protowire.BinaryDecoder{}.DecodeSfixed32(b)
					if gotVal != op.wantVal {
						t.Errorf("DecodeSfixed32() = %d, want %d", gotVal, op.wantVal)
					}
					check("DecodeSfixed32_de", n, op.wantCnt, op.wantErr)
				case decodeFloat32:
					gotVal, n := protowire.BinaryDecoder{}.DecodeFloat32(b)
					if gotVal != op.wantVal {
						t.Errorf("DecodeFloat32() = %f, want %f", gotVal, op.wantVal)
					}
					check("Float32_de", n, op.wantCnt, op.wantErr)
				case decodeFixed64:
					gotVal, n := protowire.BinaryDecoder{}.DecodeFixed64(b)
					if gotVal != op.wantVal {
						t.Errorf("DecodeFixed64() = %d, want %d", gotVal, op.wantVal)
					}
					check("DecodeFixed64_de", n, op.wantCnt, op.wantErr)
				case decodeSfixed64:
					gotVal, n := protowire.BinaryDecoder{}.DecodeSfixed64(b)
					if gotVal != op.wantVal {
						t.Errorf("DecodeSfixed64() = %d, want %d", gotVal, op.wantVal)
					}
					check("DecodeSfixed64_de", n, op.wantCnt, op.wantErr)
				case decodeDouble:
					gotVal, n := protowire.BinaryDecoder{}.DecodeDouble(b)
					if gotVal != op.wantVal {
						t.Errorf("DecodeDouble() = %f, want %f", gotVal, op.wantVal)
					}
					check("Double_de", n, op.wantCnt, op.wantErr)
				case decodeString:
					gotVal, _, n := protowire.BinaryDecoder{}.DecodeString(b)
					if gotVal != op.wantVal {
						t.Errorf("DecodeString() = %s, want %s", gotVal, op.wantVal)
					}
					check("String_de", n, op.wantCnt, op.wantErr)
				case decodeBytes:
					gotVal, _, n := protowire.BinaryDecoder{}.DecodeBytes(b)
					if !bytes.Equal(gotVal, op.wantVal) {
						t.Errorf("DecodeBytes() = %x, want %x", gotVal, op.wantVal)
					}
					check("DecodeBytes_de", n, op.wantCnt, op.wantErr)
				}
			}
		})
	}
}


func TestZigZag(t *testing.T) {
	tests := []struct {
		dec int64
		enc uint64
	}{
		{math.MinInt64 + 0, math.MaxUint64 - 0},
		{math.MinInt64 + 1, math.MaxUint64 - 2},
		{math.MinInt64 + 2, math.MaxUint64 - 4},
		{-3, 5},
		{-2, 3},
		{-1, 1},
		{0, 0},
		{+1, 2},
		{+2, 4},
		{+3, 6},
		{math.MaxInt64 - 2, math.MaxUint64 - 5},
		{math.MaxInt64 - 1, math.MaxUint64 - 3},
		{math.MaxInt64 - 0, math.MaxUint64 - 1},
	}

	for _, tt := range tests {
		if enc := protowire.EncodeZigZag(tt.dec); enc != tt.enc {
			t.Errorf("EncodeZigZag(%d) = %d, want %d", tt.dec, enc, tt.enc)
		}
		if dec := protowire.DecodeZigZag(tt.enc); dec != tt.dec {
			t.Errorf("DecodeZigZag(%d) = %d, want %d", tt.enc, dec, tt.dec)
		}
	}
}