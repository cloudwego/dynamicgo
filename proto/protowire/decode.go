package protowire

import (
	"math"
)

const (
	_ = -iota
	errCodeTruncated
	errCodeFieldNumber
	errCodeOverflow
	errCodeReserved
	errCodeEndGroup
	errCodeRecursionDepth
)

type BinaryDecoder struct{}

// ConsumeVarint parses b as a varint-encoded uint64, reporting its length.
// This returns a negative length upon an error (see ParseError).
func ConsumeVarint(b []byte) (v uint64, n int) {
	var y uint64
	if len(b) <= 0 {
		return 0, errCodeTruncated
	}
	v = uint64(b[0])
	if v < 0x80 {
		return v, 1
	}
	v -= 0x80

	if len(b) <= 1 {
		return 0, errCodeTruncated
	}
	y = uint64(b[1])
	v += y << 7
	if y < 0x80 {
		return v, 2
	}
	v -= 0x80 << 7

	if len(b) <= 2 {
		return 0, errCodeTruncated
	}
	y = uint64(b[2])
	v += y << 14
	if y < 0x80 {
		return v, 3
	}
	v -= 0x80 << 14

	if len(b) <= 3 {
		return 0, errCodeTruncated
	}
	y = uint64(b[3])
	v += y << 21
	if y < 0x80 {
		return v, 4
	}
	v -= 0x80 << 21

	if len(b) <= 4 {
		return 0, errCodeTruncated
	}
	y = uint64(b[4])
	v += y << 28
	if y < 0x80 {
		return v, 5
	}
	v -= 0x80 << 28

	if len(b) <= 5 {
		return 0, errCodeTruncated
	}
	y = uint64(b[5])
	v += y << 35
	if y < 0x80 {
		return v, 6
	}
	v -= 0x80 << 35

	if len(b) <= 6 {
		return 0, errCodeTruncated
	}
	y = uint64(b[6])
	v += y << 42
	if y < 0x80 {
		return v, 7
	}
	v -= 0x80 << 42

	if len(b) <= 7 {
		return 0, errCodeTruncated
	}
	y = uint64(b[7])
	v += y << 49
	if y < 0x80 {
		return v, 8
	}
	v -= 0x80 << 49

	if len(b) <= 8 {
		return 0, errCodeTruncated
	}
	y = uint64(b[8])
	v += y << 56
	if y < 0x80 {
		return v, 9
	}
	v -= 0x80 << 56

	if len(b) <= 9 {
		return 0, errCodeTruncated
	}
	y = uint64(b[9])
	v += y << 63
	if y < 2 {
		return v, 10
	}
	return 0, errCodeOverflow
}

// ConsumeFixed32 parses b as a little-endian uint32, reporting its length.
// This returns a negative length upon an error (see ParseError).
func ConsumeFixed32(b []byte) (v uint32, n int) {
	if len(b) < 4 {
		return 0, errCodeTruncated
	}
	v = uint32(b[0])<<0 | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
	return v, 4
}

// ConsumeFixed64 parses b as a little-endian uint64, reporting its length.
// This returns a negative length upon an error (see ParseError).
func ConsumeFixed64(b []byte) (v uint64, n int) {
	if len(b) < 8 {
		return 0, errCodeTruncated
	}
	v = uint64(b[0])<<0 | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 | uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56
	return v, 8
}

// ConsumeBytes parses b as a length-prefixed bytes value, reporting its length.
// This returns a negative length upon an error (see ParseError).
func ConsumeBytes(b []byte) (v []byte, n int, all int) {
	m, n := ConsumeVarint(b)
	if n < 0 {
		return nil, n, n
	}
	if m > uint64(len(b[n:])) {
		return nil, errCodeTruncated, errCodeTruncated
	}
	return b[n:][:m], n, n + int(m) // V's bytes, L's bytelen, (L+V)'s bytelen
}

// DecodeZigZag decodes a zig-zag-encoded uint64 as an int64.
//
//	Input:  {…,  5,  3,  1,  0,  2,  4,  6, …}
//	Output: {…, -3, -2, -1,  0, +1, +2, +3, …}
func DecodeZigZag(x uint64) int64 {
	return int64(x>>1) ^ int64(x)<<63>>63
}

func (BinaryDecoder) DecodeBool(b []byte) (bool, int) {
	v, n := ConsumeVarint(b)
	return int8(v) == 1, n
}

func (BinaryDecoder) DecodeByte(b []byte) byte {
	return byte(b[0])
}

func (BinaryDecoder) DecodeInt32(b []byte) (int32, int) {
	v, n := ConsumeVarint(b)
	return int32(v), n
}

func (BinaryDecoder) DecodeSint32(b []byte) (int32, int) {
	v, n := ConsumeVarint(b)
	return int32(DecodeZigZag(v & math.MaxUint32)), n
}

func (BinaryDecoder) DecodeUint32(b []byte) (uint32, int) {
	v, n := ConsumeVarint(b)
	return uint32(v), n
}

func (BinaryDecoder) DecodeInt64(b []byte) (int64, int) {
	v, n := ConsumeVarint(b)
	return int64(v), n
}

func (BinaryDecoder) DecodeSint64(b []byte) (int64, int) {
	v, n := ConsumeVarint(b)
	return DecodeZigZag(v), n
}

func (BinaryDecoder) DecodeUint64(b []byte) (uint64, int) {
	v, n := ConsumeVarint(b)
	return v, n
}

func (BinaryDecoder) DecodeSfixed32(b []byte) (int32, int) {
	v, n := ConsumeFixed32(b)
	return int32(v), n
}

func (BinaryDecoder) DecodeFixed32(b []byte) (uint32, int) {
	v, n := ConsumeFixed32(b)
	return v, n
}

func (BinaryDecoder) DecodeFloat32(b []byte) (float32, int) {
	v, n := ConsumeFixed32(b)
	return math.Float32frombits(v), n
}

func (BinaryDecoder) DecodeSfixed64(b []byte) (int64, int) {
	v, n := ConsumeFixed64(b)
	return int64(v), n
}

func (BinaryDecoder) DecodeFixed64(b []byte) (uint64, int) {
	v, n := ConsumeFixed64(b)
	return v, n
}

func (BinaryDecoder) DecodeDouble(b []byte) (float64, int) {
	v, n := ConsumeFixed64(b)
	return math.Float64frombits(v), n
}

func (BinaryDecoder) DecodeString(b []byte) (string, int, int) {
	v, n, all := ConsumeBytes(b)
	return string(v), n, all
}

func (BinaryDecoder) DecodeBytes(b []byte) ([]byte, int, int) {
	v, n, all := ConsumeBytes(b)
	return v, n, all
}
