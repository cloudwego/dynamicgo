package test

import (
	"testing"

	"github.com/cloudwego/dynamicgo/proto/binary"
	"github.com/cloudwego/dynamicgo/proto/generic"
	"github.com/cloudwego/prutal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	b := make([]byte, 0, 4096)
	sobj := getPbSimpleValue()

	b, _ = prutal.MarshalAppend(b[:0], sobj)
	println("small protobuf data size: ", len(b))

	psobj := getPbPartialSimpleValue()
	b, _ = prutal.MarshalAppend(b[:0], psobj)
	println("small protobuf data size: ", len(b))

	nobj := getPbNestingValue()
	b, _ = prutal.MarshalAppend(b[:0], nobj)
	println("medium protobuf data size: ", len(b))

	pnobj := getPbPartialNestingValue()
	b, _ = prutal.MarshalAppend(b[:0], pnobj)
	println("medium protobuf data size: ", len(b))
}

/*
 * performance test in DynamicGo
 * 1. ProtoSkip
 * 2. ProtoGetOne
 * 3. ProtoGetMany
 * 4. ProtoSetOne
 * 5. ProtoSetMany
 * 6. ProtoMarshalMany
 * 7. ProtoMarshalTo, compared with ProtoBufGo, Prutal
 */
func BenchmarkProtoSkip_DynamicGo(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		desc := getPbSimpleDesc()
		obj := getPbSimpleValue()

		data, err := prutal.Marshal(obj)
		assert.Nil(b, err)

		p := binary.NewBinaryProtol(data)
		for p.Left() > 0 {
			fieldNumber, wireType, _, err := p.ConsumeTag()
			if err != nil {
				b.Fatal(err)
			}

			if desc.Message().ByNumber(fieldNumber) == nil {
				b.Fatal("field not found")
			}

			if err := p.Skip(wireType, false); err != nil {
				b.Fatal(err)
			}
		}

		require.Equal(b, len(data), p.Read)

		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		b.Run("go", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				p := binary.NewBinaryProtol(data)
				for p.Left() > 0 {
					_, wireType, _, _ := p.ConsumeTag()
					_ = p.Skip(wireType, false)
				}
			}
		})
	})

	b.Run("medium", func(b *testing.B) {
		desc := getPbNestingDesc()
		obj := getPbNestingValue()
		data, err := prutal.Marshal(obj)
		assert.Nil(b, err)

		p := binary.NewBinaryProtol(data)
		for p.Left() > 0 {
			fieldNumber, wireType, _, err := p.ConsumeTag()
			if err != nil {
				b.Fatal(err)
			}

			if desc.Message().ByNumber(fieldNumber) == nil {
				b.Fatal("field not found")
			}

			if err := p.Skip(wireType, false); err != nil {
				b.Fatal(err)
			}
		}

		require.Equal(b, len(data), p.Read)

		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		b.Run("go", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				p := binary.NewBinaryProtol(data)
				for p.Left() > 0 {
					_, wireType, _, _ := p.ConsumeTag()
					_ = p.Skip(wireType, false)
				}
			}
		})
	})
}

func BenchmarkProtoGetOne_DynamicGo(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		desc := getPbSimpleDesc()
		obj := getPbSimpleValue()
		data, err := prutal.Marshal(obj)
		assert.Nil(b, err)

		v := generic.NewRootValue(desc, data)
		vv := v.GetByPath(generic.NewPathFieldId(6))
		require.Nil(b, vv.Check())
		bs, err := vv.Binary()
		require.Nil(b, err)
		require.Equal(b, obj.BinaryField, bs)
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		b.Run("go", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = v.GetByPath(generic.NewPathFieldId(6))
			}
		})
	})

	b.Run("medium", func(b *testing.B) {
		desc := getPbNestingDesc()
		obj := getPbNestingValue()
		data, err := prutal.Marshal(obj)
		assert.Nil(b, err)

		v := generic.NewRootValue(desc, data)
		vv := v.GetByPath(generic.NewPathFieldId(15), generic.NewPathStrKey("15"), generic.NewPathFieldId(6))
		require.Nil(b, vv.Check())
		bs, err := vv.Binary()
		require.Nil(b, err)
		require.Equal(b, obj.MapStringSimple["15"].BinaryField, bs)
		b.SetBytes(int64(len(data)))
		b.ResetTimer()
		b.Run("go", func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = v.GetByPath(generic.NewPathFieldId(15), generic.NewPathStrKey("15"), generic.NewPathFieldId(6))
			}
		})
	})
}
