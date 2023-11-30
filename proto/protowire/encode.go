package protowire

import (
	"math"
	"math/bits"
)

type BinaryEncoder struct{}

// VarintEncode encodes a uint64 value as a varint-encoded byte slice.
// AppendVarint appends v to b as a varint-encoded uint64.
func AppendVarint(b []byte, v uint64) []byte {
	switch {
	case v < 1<<7:
		b = append(b, byte(v))
	case v < 1<<14:
		b = append(b,
			byte((v>>0)&0x7f|0x80),
			byte(v>>7))
	case v < 1<<21:
		b = append(b,
			byte((v>>0)&0x7f|0x80),
			byte((v>>7)&0x7f|0x80),
			byte(v>>14))
	case v < 1<<28:
		b = append(b,
			byte((v>>0)&0x7f|0x80),
			byte((v>>7)&0x7f|0x80),
			byte((v>>14)&0x7f|0x80),
			byte(v>>21))
	case v < 1<<35:
		b = append(b,
			byte((v>>0)&0x7f|0x80),
			byte((v>>7)&0x7f|0x80),
			byte((v>>14)&0x7f|0x80),
			byte((v>>21)&0x7f|0x80),
			byte(v>>28))
	case v < 1<<42:
		b = append(b,
			byte((v>>0)&0x7f|0x80),
			byte((v>>7)&0x7f|0x80),
			byte((v>>14)&0x7f|0x80),
			byte((v>>21)&0x7f|0x80),
			byte((v>>28)&0x7f|0x80),
			byte(v>>35))
	case v < 1<<49:
		b = append(b,
			byte((v>>0)&0x7f|0x80),
			byte((v>>7)&0x7f|0x80),
			byte((v>>14)&0x7f|0x80),
			byte((v>>21)&0x7f|0x80),
			byte((v>>28)&0x7f|0x80),
			byte((v>>35)&0x7f|0x80),
			byte(v>>42))
	case v < 1<<56:
		b = append(b,
			byte((v>>0)&0x7f|0x80),
			byte((v>>7)&0x7f|0x80),
			byte((v>>14)&0x7f|0x80),
			byte((v>>21)&0x7f|0x80),
			byte((v>>28)&0x7f|0x80),
			byte((v>>35)&0x7f|0x80),
			byte((v>>42)&0x7f|0x80),
			byte(v>>49))
	case v < 1<<63:
		b = append(b,
			byte((v>>0)&0x7f|0x80),
			byte((v>>7)&0x7f|0x80),
			byte((v>>14)&0x7f|0x80),
			byte((v>>21)&0x7f|0x80),
			byte((v>>28)&0x7f|0x80),
			byte((v>>35)&0x7f|0x80),
			byte((v>>42)&0x7f|0x80),
			byte((v>>49)&0x7f|0x80),
			byte(v>>56))
	default:
		b = append(b,
			byte((v>>0)&0x7f|0x80),
			byte((v>>7)&0x7f|0x80),
			byte((v>>14)&0x7f|0x80),
			byte((v>>21)&0x7f|0x80),
			byte((v>>28)&0x7f|0x80),
			byte((v>>35)&0x7f|0x80),
			byte((v>>42)&0x7f|0x80),
			byte((v>>49)&0x7f|0x80),
			byte((v>>56)&0x7f|0x80),
			1)
	}
	return b
}

// AppendFixed32 appends v to b as a little-endian uint32.
func AppendFixed32(b []byte, v uint32) []byte {
	return append(b,
		byte(v>>0),
		byte(v>>8),
		byte(v>>16),
		byte(v>>24))
}

// AppendFixed64 appends v to b as a little-endian uint64.
func AppendFixed64(b []byte, v uint64) []byte {
	return append(b,
		byte(v>>0),
		byte(v>>8),
		byte(v>>16),
		byte(v>>24),
		byte(v>>32),
		byte(v>>40),
		byte(v>>48),
		byte(v>>56))
}

// EncodeZigZag encodes an int64 as a zig-zag-encoded uint64.
//
//	Input:  {…, -3, -2, -1,  0, +1, +2, +3, …}
//	Output: {…,  5,  3,  1,  0,  2,  4,  6, …}
func EncodeZigZag(v int64) uint64 {
	return uint64(v<<1) ^ uint64(v>>63)
}

// SizeVarint returns the encoded size of a varint.
// The size is guaranteed to be within 1 and 10, inclusive.
func SizeVarint(v uint64) int {
	// This computes 1 + (bits.Len64(v)-1)/7.
	// 9/64 is a good enough approximation of 1/7
	return int(9*uint32(bits.Len64(v))+64) / 64
}

// SizeFixed32 returns the encoded size of a fixed32; which is always 4.
func SizeFixed32() int {
	return 4
}

// SizeFixed64 returns the encoded size of a fixed64; which is always 8.
func SizeFixed64() int {
	return 8
}

// encode each proto kind into bytes
func (BinaryEncoder) EncodeBool(b []byte, v bool) []byte {
	if v {
		return append(b, 1)
	}
	return append(b, 0)
}

func (BinaryEncoder) EncodeByte(b []byte, v byte) []byte {
	return append(b, v)
}

func (BinaryEncoder) EncodeEnum(b []byte, v int32) []byte {
	return AppendVarint(b, uint64(v))
}

func (BinaryEncoder) EncodeInt32(b []byte, v int32) []byte {
	return AppendVarint(b, uint64(v))
}

func (BinaryEncoder) EncodeSint32(b []byte, v int32) []byte {
	return AppendVarint(b, EncodeZigZag(int64(v)))
}

func (BinaryEncoder) EncodeUint32(b []byte, v uint32) []byte {
	return AppendVarint(b, uint64(v))
}

func (BinaryEncoder) EncodeInt64(b []byte, v int64) []byte {
	return AppendVarint(b, uint64(v))
}

func (BinaryEncoder) EncodeSint64(b []byte, v int64) []byte {
	return AppendVarint(b, EncodeZigZag(v))
}

func (BinaryEncoder) EncodeUint64(b []byte, v uint64) []byte {
	return AppendVarint(b, v)
}

func (BinaryEncoder) EncodeSfixed32(b []byte, v int32) []byte {
	return AppendFixed32(b, uint32(v))
}

func (BinaryEncoder) EncodeFixed32(b []byte, v uint32) []byte {
	return AppendFixed32(b, v)
}

func (BinaryEncoder) EncodeFloat32(b []byte, v float32) []byte {
	return AppendFixed32(b, math.Float32bits(v))
}

func (BinaryEncoder) EncodeSfixed64(b []byte, v int64) []byte {
	return AppendFixed64(b, uint64(v))
}

func (BinaryEncoder) EncodeFixed64(b []byte, v uint64) []byte {
	return AppendFixed64(b, v)
}

func (BinaryEncoder) EncodeDouble(b []byte, v float64) []byte {
	return AppendFixed64(b, math.Float64bits(v))
}

func (BinaryEncoder) EncodeString(b []byte, v string) []byte {
	return append(AppendVarint(b, uint64(len(v))), v...)
}

func (BinaryEncoder) EncodeBytes(b []byte, v []byte) []byte {
	return append(AppendVarint(b, uint64(len(v))), v...)
}
