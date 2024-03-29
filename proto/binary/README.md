<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# binary

```go
import "github.com/cloudwego/dynamicgo/proto/binary"
```

## Index

- [binary](#binary)
  - [Index](#index)
  - [Variables](#variables)
  - [func AppendSpeculativeLength](#func-appendspeculativelength)
  - [func FinishSpeculativeLength](#func-finishspeculativelength)
  - [func FreeBinaryProtocol](#func-freebinaryprotocol)
  - [type BinaryProtocol](#type-binaryprotocol)
    - [func NewBinaryProtocolBuffer](#func-newbinaryprotocolbuffer)
    - [func NewBinaryProtol](#func-newbinaryprotol)
    - [func (\*BinaryProtocol) AppendTag](#func-binaryprotocol-appendtag)
    - [func (\*BinaryProtocol) AppendTagByKind](#func-binaryprotocol-appendtagbykind)
    - [func (\*BinaryProtocol) ConsumeTag](#func-binaryprotocol-consumetag)
    - [func (\*BinaryProtocol) ConsumeTagWithoutMove](#func-binaryprotocol-consumetagwithoutmove)
    - [func (\*BinaryProtocol) Left](#func-binaryprotocol-left)
    - [func (\*BinaryProtocol) RawBuf](#func-binaryprotocol-rawbuf)
    - [func (\*BinaryProtocol) ReadAnyWithDesc](#func-binaryprotocol-readanywithdesc)
    - [func (\*BinaryProtocol) ReadBaseTypeWithDesc](#func-binaryprotocol-readbasetypewithdesc)
    - [func (\*BinaryProtocol) ReadBool](#func-binaryprotocol-readbool)
    - [func (\*BinaryProtocol) ReadByte](#func-binaryprotocol-readbyte)
    - [func (\*BinaryProtocol) ReadBytes](#func-binaryprotocol-readbytes)
    - [func (\*BinaryProtocol) ReadDouble](#func-binaryprotocol-readdouble)
    - [func (\*BinaryProtocol) ReadEnum](#func-binaryprotocol-readenum)
    - [func (\*BinaryProtocol) ReadFixed32](#func-binaryprotocol-readfixed32)
    - [func (\*BinaryProtocol) ReadFixed64](#func-binaryprotocol-readfixed64)
    - [func (\*BinaryProtocol) ReadFloat](#func-binaryprotocol-readfloat)
    - [func (\*BinaryProtocol) ReadInt](#func-binaryprotocol-readint)
    - [func (\*BinaryProtocol) ReadInt32](#func-binaryprotocol-readint32)
    - [func (\*BinaryProtocol) ReadInt64](#func-binaryprotocol-readint64)
    - [func (\*BinaryProtocol) ReadLength](#func-binaryprotocol-readlength)
    - [func (\*BinaryProtocol) ReadList](#func-binaryprotocol-readlist)
    - [func (\*BinaryProtocol) ReadMap](#func-binaryprotocol-readmap)
    - [func (\*BinaryProtocol) ReadPair](#func-binaryprotocol-readpair)
    - [func (\*BinaryProtocol) ReadSfixed32](#func-binaryprotocol-readsfixed32)
    - [func (\*BinaryProtocol) ReadSfixed64](#func-binaryprotocol-readsfixed64)
    - [func (\*BinaryProtocol) ReadSint32](#func-binaryprotocol-readsint32)
    - [func (\*BinaryProtocol) ReadSint64](#func-binaryprotocol-readsint64)
    - [func (\*BinaryProtocol) ReadString](#func-binaryprotocol-readstring)
    - [func (\*BinaryProtocol) ReadUint32](#func-binaryprotocol-readuint32)
    - [func (\*BinaryProtocol) ReadUint64](#func-binaryprotocol-readuint64)
    - [func (\*BinaryProtocol) ReadVarint](#func-binaryprotocol-readvarint)
    - [func (\*BinaryProtocol) Recycle](#func-binaryprotocol-recycle)
    - [func (\*BinaryProtocol) Reset](#func-binaryprotocol-reset)
    - [func (\*BinaryProtocol) Skip](#func-binaryprotocol-skip)
    - [func (\*BinaryProtocol) SkipAllElements](#func-binaryprotocol-skipallelements)
    - [func (\*BinaryProtocol) SkipBytesType](#func-binaryprotocol-skipbytestype)
    - [func (\*BinaryProtocol) SkipFixed32Type](#func-binaryprotocol-skipfixed32type)
    - [func (\*BinaryProtocol) SkipFixed64Type](#func-binaryprotocol-skipfixed64type)
    - [func (\*BinaryProtocol) WriteAnyWithDesc](#func-binaryprotocol-writeanywithdesc)
    - [func (\*BinaryProtocol) WriteBaseTypeWithDesc](#func-binaryprotocol-writebasetypewithdesc)
    - [func (\*BinaryProtocol) WriteBool](#func-binaryprotocol-writebool)
    - [func (\*BinaryProtocol) WriteBytes](#func-binaryprotocol-writebytes)
    - [func (\*BinaryProtocol) WriteDouble](#func-binaryprotocol-writedouble)
    - [func (\*BinaryProtocol) WriteEnum](#func-binaryprotocol-writeenum)
    - [func (\*BinaryProtocol) WriteFixed32](#func-binaryprotocol-writefixed32)
    - [func (\*BinaryProtocol) WriteFixed64](#func-binaryprotocol-writefixed64)
    - [func (\*BinaryProtocol) WriteFloat](#func-binaryprotocol-writefloat)
    - [func (\*BinaryProtocol) WriteInt32](#func-binaryprotocol-writeint32)
    - [func (\*BinaryProtocol) WriteInt64](#func-binaryprotocol-writeint64)
    - [func (\*BinaryProtocol) WriteList](#func-binaryprotocol-writelist)
    - [func (\*BinaryProtocol) WriteMap](#func-binaryprotocol-writemap)
    - [func (\*BinaryProtocol) WriteMessageFields](#func-binaryprotocol-writemessagefields)
    - [func (\*BinaryProtocol) WriteSfixed32](#func-binaryprotocol-writesfixed32)
    - [func (\*BinaryProtocol) WriteSfixed64](#func-binaryprotocol-writesfixed64)
    - [func (\*BinaryProtocol) WriteSint32](#func-binaryprotocol-writesint32)
    - [func (\*BinaryProtocol) WriteSint64](#func-binaryprotocol-writesint64)
    - [func (\*BinaryProtocol) WriteString](#func-binaryprotocol-writestring)
    - [func (\*BinaryProtocol) WriteUint32](#func-binaryprotocol-writeuint32)
    - [func (\*BinaryProtocol) WriteUint64](#func-binaryprotocol-writeuint64)


## Variables

<a name="ErrConvert"></a>

```go
var (
    ErrConvert = meta.NewError(meta.ErrConvert, "convert type error", nil)
)
```

<a name="AppendSpeculativeLength"></a>
## func [AppendSpeculativeLength](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L188>)

```go
func AppendSpeculativeLength(b []byte) ([]byte, int)
```

When encoding length-prefixed fields, we speculatively set aside some number of bytes for the length, encode the data, and then encode the length (shifting the data if necessary to make room).

<a name="FinishSpeculativeLength"></a>
## func [FinishSpeculativeLength](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L194>)

```go
func FinishSpeculativeLength(b []byte, pos int) []byte
```



<a name="FreeBinaryProtocol"></a>
## func [FreeBinaryProtocol](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L127>)

```go
func FreeBinaryProtocol(bp *BinaryProtocol)
```



<a name="BinaryProtocol"></a>
## type [BinaryProtocol](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L45-L48>)

Serizalize data to byte array and reuse the memory

```go
type BinaryProtocol struct {
    Buf  []byte
    Read int
}
```

<a name="NewBinaryProtocolBuffer"></a>
### func [NewBinaryProtocolBuffer](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L122>)

```go
func NewBinaryProtocolBuffer() *BinaryProtocol
```



<a name="NewBinaryProtol"></a>
### func [NewBinaryProtol](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L116>)

```go
func NewBinaryProtol(buf []byte) *BinaryProtocol
```

BinaryProtocol Method

<a name="BinaryProtocol.AppendTag"></a>
### func (*BinaryProtocol) [AppendTag](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L138>)

```go
func (p *BinaryProtocol) AppendTag(num proto.FieldNumber, typ proto.WireType) error
```

Append Tag

<a name="BinaryProtocol.AppendTagByKind"></a>
### func (*BinaryProtocol) [AppendTagByKind](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L148>)

```go
func (p *BinaryProtocol) AppendTagByKind(number proto.FieldNumber, kind proto.ProtoKind) error
```

Append Tag With FieldDescriptor by kind, you must use kind to write tag, because the typedesc when list has no tag

<a name="BinaryProtocol.ConsumeTag"></a>
### func (*BinaryProtocol) [ConsumeTag](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L153>)

```go
func (p *BinaryProtocol) ConsumeTag() (proto.FieldNumber, proto.WireType, int, error)
```

ConsumeTag parses b as a varint-encoded tag, reporting its length.

<a name="BinaryProtocol.ConsumeTagWithoutMove"></a>
### func (*BinaryProtocol) [ConsumeTagWithoutMove](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L170>)

```go
func (p *BinaryProtocol) ConsumeTagWithoutMove() (proto.FieldNumber, proto.WireType, int, error)
```

ConsumeChildTag parses b as a varint-encoded tag, don't move p.Read

<a name="BinaryProtocol.Left"></a>
### func (*BinaryProtocol) [Left](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L111>)

```go
func (p *BinaryProtocol) Left() int
```

Left returns the left bytes to read

<a name="BinaryProtocol.RawBuf"></a>
### func (*BinaryProtocol) [RawBuf](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L106>)

```go
func (p *BinaryProtocol) RawBuf() []byte
```

RawBuf returns the raw buffer of the protocol

<a name="BinaryProtocol.ReadAnyWithDesc"></a>
### func (*BinaryProtocol) [ReadAnyWithDesc](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L1131>)

```go
func (p *BinaryProtocol) ReadAnyWithDesc(desc *proto.TypeDescriptor, hasMessageLen bool, copyString bool, disallowUnknown bool, useFieldName bool) (interface{}, error)
```

ReadAnyWithDesc read any type by desc and val, the first Tag is parsed outside when use ReadBaseTypeWithDesc

- LIST/SET will be converted to []interface{}
- MAP will be converted to map[string]interface{} or map[int]interface{} or map[interface{}]interface{}
- MESSAGE will be converted to map[proto.FieldNumber]interface{} or map[string]interface{}

<a name="BinaryProtocol.ReadBaseTypeWithDesc"></a>
### func (*BinaryProtocol) [ReadBaseTypeWithDesc](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L1143>)

```go
func (p *BinaryProtocol) ReadBaseTypeWithDesc(desc *proto.TypeDescriptor, hasMessageLen bool, copyString bool, disallowUnknown bool, useFieldName bool) (interface{}, error)
```

ReadBaseType with desc, not thread safe

<a name="BinaryProtocol.ReadBool"></a>
### func (*BinaryProtocol) [ReadBool](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L775>)

```go
func (p *BinaryProtocol) ReadBool() (bool, error)
```

ReadBool

<a name="BinaryProtocol.ReadByte"></a>
### func (*BinaryProtocol) [ReadByte](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L766>)

```go
func (p *BinaryProtocol) ReadByte() (value byte, err error)
```

ReadByte

<a name="BinaryProtocol.ReadBytes"></a>
### func (*BinaryProtocol) [ReadBytes](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L947>)

```go
func (p *BinaryProtocol) ReadBytes() ([]byte, error)
```

ReadBytes return bytesData and the sum length of L、V in TLV

<a name="BinaryProtocol.ReadDouble"></a>
### func (*BinaryProtocol) [ReadDouble](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L937>)

```go
func (p *BinaryProtocol) ReadDouble() (float64, error)
```

ReadDouble

<a name="BinaryProtocol.ReadEnum"></a>
### func (*BinaryProtocol) [ReadEnum](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L984>)

```go
func (p *BinaryProtocol) ReadEnum() (proto.EnumNumber, error)
```

ReadEnum

<a name="BinaryProtocol.ReadFixed32"></a>
### func (*BinaryProtocol) [ReadFixed32](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L887>)

```go
func (p *BinaryProtocol) ReadFixed32() (int32, error)
```

ReadFixed32

<a name="BinaryProtocol.ReadFixed64"></a>
### func (*BinaryProtocol) [ReadFixed64](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L917>)

```go
func (p *BinaryProtocol) ReadFixed64() (int64, error)
```

ReadFixed64

<a name="BinaryProtocol.ReadFloat"></a>
### func (*BinaryProtocol) [ReadFloat](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L907>)

```go
func (p *BinaryProtocol) ReadFloat() (float32, error)
```

ReadFloat

<a name="BinaryProtocol.ReadInt"></a>
### func (*BinaryProtocol) [ReadInt](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L785>)

```go
func (p *BinaryProtocol) ReadInt(t proto.Type) (value int, err error)
```

ReadInt containing INT32, SINT32, SFIX32, INT64, SINT64, SFIX64, UINT32, UINT64

<a name="BinaryProtocol.ReadInt32"></a>
### func (*BinaryProtocol) [ReadInt32](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L817>)

```go
func (p *BinaryProtocol) ReadInt32() (int32, error)
```

ReadI32

<a name="BinaryProtocol.ReadInt64"></a>
### func (*BinaryProtocol) [ReadInt64](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L847>)

```go
func (p *BinaryProtocol) ReadInt64() (int64, error)
```

ReadI64

<a name="BinaryProtocol.ReadLength"></a>
### func (*BinaryProtocol) [ReadLength](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L957>)

```go
func (p *BinaryProtocol) ReadLength() (int, error)
```

ReadLength return dataLength, and move pointer in the begin of data

<a name="BinaryProtocol.ReadList"></a>
### func (*BinaryProtocol) [ReadList](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L994>)

```go
func (p *BinaryProtocol) ReadList(desc *proto.TypeDescriptor, copyString bool, disallowUnknown bool, useFieldName bool) ([]interface{}, error)
```

ReadList

<a name="BinaryProtocol.ReadMap"></a>
### func (*BinaryProtocol) [ReadMap](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L1073>)

```go
func (p *BinaryProtocol) ReadMap(desc *proto.TypeDescriptor, copyString bool, disallowUnknown bool, useFieldName bool) (map[interface{}]interface{}, error)
```

ReadMap

<a name="BinaryProtocol.ReadPair"></a>
### func (*BinaryProtocol) [ReadPair](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L1051>)

```go
func (p *BinaryProtocol) ReadPair(keyDesc *proto.TypeDescriptor, valueDesc *proto.TypeDescriptor, copyString bool, disallowUnknown bool, useFieldName bool) (interface{}, interface{}, error)
```



<a name="BinaryProtocol.ReadSfixed32"></a>
### func (*BinaryProtocol) [ReadSfixed32](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L897>)

```go
func (p *BinaryProtocol) ReadSfixed32() (int32, error)
```

ReadSFixed32

<a name="BinaryProtocol.ReadSfixed64"></a>
### func (*BinaryProtocol) [ReadSfixed64](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L927>)

```go
func (p *BinaryProtocol) ReadSfixed64() (int64, error)
```

ReadSFixed64

<a name="BinaryProtocol.ReadSint32"></a>
### func (*BinaryProtocol) [ReadSint32](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L827>)

```go
func (p *BinaryProtocol) ReadSint32() (int32, error)
```

ReadSint32

<a name="BinaryProtocol.ReadSint64"></a>
### func (*BinaryProtocol) [ReadSint64](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L857>)

```go
func (p *BinaryProtocol) ReadSint64() (int64, error)
```

ReadSint64

<a name="BinaryProtocol.ReadString"></a>
### func (*BinaryProtocol) [ReadString](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L967>)

```go
func (p *BinaryProtocol) ReadString(copy bool) (value string, err error)
```

ReadString

<a name="BinaryProtocol.ReadUint32"></a>
### func (*BinaryProtocol) [ReadUint32](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L837>)

```go
func (p *BinaryProtocol) ReadUint32() (uint32, error)
```

ReadUint32

<a name="BinaryProtocol.ReadUint64"></a>
### func (*BinaryProtocol) [ReadUint64](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L867>)

```go
func (p *BinaryProtocol) ReadUint64() (uint64, error)
```

ReadUint64

<a name="BinaryProtocol.ReadVarint"></a>
### func (*BinaryProtocol) [ReadVarint](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L877>)

```go
func (p *BinaryProtocol) ReadVarint() (uint64, error)
```

ReadVarint

<a name="BinaryProtocol.Recycle"></a>
### func (*BinaryProtocol) [Recycle](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L132>)

```go
func (p *BinaryProtocol) Recycle()
```



<a name="BinaryProtocol.Reset"></a>
### func (*BinaryProtocol) [Reset](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L100>)

```go
func (p *BinaryProtocol) Reset()
```

Reset resets the buffer and read position

<a name="BinaryProtocol.Skip"></a>
### func (*BinaryProtocol) [Skip](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary_skip.go#L29>)

```go
func (p *BinaryProtocol) Skip(wireType proto.WireType, useNative bool) (err error)
```

skip (L)V once by wireType, useNative is not implemented

<a name="BinaryProtocol.SkipAllElements"></a>
### func (*BinaryProtocol) [SkipAllElements](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary_skip.go#L44>)

```go
func (p *BinaryProtocol) SkipAllElements(fieldNumber proto.FieldNumber, ispacked bool) (size int, err error)
```

fast skip all elements in LIST/MAP

<a name="BinaryProtocol.SkipBytesType"></a>
### func (*BinaryProtocol) [SkipBytesType](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary_skip.go#L18>)

```go
func (p *BinaryProtocol) SkipBytesType() (int, error)
```



<a name="BinaryProtocol.SkipFixed32Type"></a>
### func (*BinaryProtocol) [SkipFixed32Type](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary_skip.go#L8>)

```go
func (p *BinaryProtocol) SkipFixed32Type() (int, error)
```



<a name="BinaryProtocol.SkipFixed64Type"></a>
### func (*BinaryProtocol) [SkipFixed64Type](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary_skip.go#L13>)

```go
func (p *BinaryProtocol) SkipFixed64Type() (int, error)
```



<a name="BinaryProtocol.WriteAnyWithDesc"></a>
### func (*BinaryProtocol) [WriteAnyWithDesc](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L750>)

```go
func (p *BinaryProtocol) WriteAnyWithDesc(desc *proto.TypeDescriptor, val interface{}, NeedMessageLen bool, cast bool, disallowUnknown bool, useFieldName bool) error
```

WriteAnyWithDesc explain desc and val and write them into buffer

- LIST will be converted from []interface{}
- MAP will be converted from map[string]interface{} or map[int]interface{} or map[interface{}]interface{}
- MESSAGE will be converted from map[FieldNumber]interface{} or map[string]interface{}

<a name="BinaryProtocol.WriteBaseTypeWithDesc"></a>
### func (*BinaryProtocol) [WriteBaseTypeWithDesc](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L499>)

```go
func (p *BinaryProtocol) WriteBaseTypeWithDesc(desc *proto.TypeDescriptor, val interface{}, NeedMessageLen bool, cast bool, disallowUnknown bool, useFieldName bool) error
```

WriteBaseType Fields with FieldDescriptor format: (L)V

<a name="BinaryProtocol.WriteBool"></a>
### func (*BinaryProtocol) [WriteBool](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L216>)

```go
func (p *BinaryProtocol) WriteBool(value bool) error
```

WriteBool

<a name="BinaryProtocol.WriteBytes"></a>
### func (*BinaryProtocol) [WriteBytes](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L330>)

```go
func (p *BinaryProtocol) WriteBytes(value []byte) error
```

WriteBytes

<a name="BinaryProtocol.WriteDouble"></a>
### func (*BinaryProtocol) [WriteDouble](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L311>)

```go
func (p *BinaryProtocol) WriteDouble(value float64) error
```

WriteDouble

<a name="BinaryProtocol.WriteEnum"></a>
### func (*BinaryProtocol) [WriteEnum](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L336>)

```go
func (p *BinaryProtocol) WriteEnum(value proto.EnumNumber) error
```

WriteEnum

<a name="BinaryProtocol.WriteFixed32"></a>
### func (*BinaryProtocol) [WriteFixed32](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L243>)

```go
func (p *BinaryProtocol) WriteFixed32(value int32) error
```

Writefixed32

<a name="BinaryProtocol.WriteFixed64"></a>
### func (*BinaryProtocol) [WriteFixed64](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L281>)

```go
func (p *BinaryProtocol) WriteFixed64(value uint64) error
```

Writefixed64

<a name="BinaryProtocol.WriteFloat"></a>
### func (*BinaryProtocol) [WriteFloat](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L301>)

```go
func (p *BinaryProtocol) WriteFloat(value float32) error
```

WriteFloat

<a name="BinaryProtocol.WriteInt32"></a>
### func (*BinaryProtocol) [WriteInt32](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L225>)

```go
func (p *BinaryProtocol) WriteInt32(value int32) error
```

WriteInt32

<a name="BinaryProtocol.WriteInt64"></a>
### func (*BinaryProtocol) [WriteInt64](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L263>)

```go
func (p *BinaryProtocol) WriteInt64(value int64) error
```

WriteInt64

<a name="BinaryProtocol.WriteList"></a>
### func (*BinaryProtocol) [WriteList](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L347>)

```go
func (p *BinaryProtocol) WriteList(desc *proto.TypeDescriptor, val interface{}, cast bool, disallowUnknown bool, useFieldName bool) error
```

* WriteList

- packed format：[tag][length][value value value value....]
- unpacked format：[tag][(length)][value][tag][(length)][value][tag][(length)][value]....
- accpet val type: []interface{}

<a name="BinaryProtocol.WriteMap"></a>
### func (*BinaryProtocol) [WriteMap](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L388>)

```go
func (p *BinaryProtocol) WriteMap(desc *proto.TypeDescriptor, val interface{}, cast bool, disallowUnknown bool, useFieldName bool) error
```

* WriteMap

- Map bytes format: [Pairtag][Pairlength][keyTag(L)V][valueTag(L)V] [Pairtag][Pairlength][T(L)V][T(L)V]...
- Pairtag = MapFieldnumber << 3 | wiretype, wiertype = proto.BytesType
- accpet val type: map[string]interface{} or map[int]interface{} or map[interface{}]interface{}

<a name="BinaryProtocol.WriteMessageFields"></a>
### func (*BinaryProtocol) [WriteMessageFields](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L452>)

```go
func (p *BinaryProtocol) WriteMessageFields(desc *proto.MessageDescriptor, val interface{}, cast bool, disallowUnknown bool, useFieldName bool) error
```

* Write Message

- accpet val type: map[string]interface{} or map[proto.FieldNumber]interface{}
- message fields format: [fieldTag(L)V][fieldTag(L)V]...

<a name="BinaryProtocol.WriteSfixed32"></a>
### func (*BinaryProtocol) [WriteSfixed32](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L253>)

```go
func (p *BinaryProtocol) WriteSfixed32(value int32) error
```

WriteSfixed32

<a name="BinaryProtocol.WriteSfixed64"></a>
### func (*BinaryProtocol) [WriteSfixed64](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L291>)

```go
func (p *BinaryProtocol) WriteSfixed64(value int64) error
```

WriteSfixed64

<a name="BinaryProtocol.WriteSint32"></a>
### func (*BinaryProtocol) [WriteSint32](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L231>)

```go
func (p *BinaryProtocol) WriteSint32(value int32) error
```

WriteSint32

<a name="BinaryProtocol.WriteSint64"></a>
### func (*BinaryProtocol) [WriteSint64](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L269>)

```go
func (p *BinaryProtocol) WriteSint64(value int64) error
```

WriteSint64

<a name="BinaryProtocol.WriteString"></a>
### func (*BinaryProtocol) [WriteString](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L321>)

```go
func (p *BinaryProtocol) WriteString(value string) error
```

WriteString

<a name="BinaryProtocol.WriteUint32"></a>
### func (*BinaryProtocol) [WriteUint32](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L237>)

```go
func (p *BinaryProtocol) WriteUint32(value uint32) error
```

WriteUint32

<a name="BinaryProtocol.WriteUint64"></a>
### func (*BinaryProtocol) [WriteUint64](<https://github.com/khan-yin/dynamicgo/blob/main/proto/binary/binary.go#L275>)

```go
func (p *BinaryProtocol) WriteUint64(value uint64) error
```

WriteUint64

Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)
