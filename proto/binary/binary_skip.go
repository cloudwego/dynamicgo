package binary

import (
	"github.com/cloudwego/dynamicgo/proto"
	"github.com/cloudwego/dynamicgo/proto/protowire"
)

func (p *BinaryProtocol) SkipFixed32Type() (int, error) {
	_, err := p.next(4)
	return 4, err
}

func (p *BinaryProtocol) SkipFixed64Type() (int, error) {
	_, err := p.next(8)
	return 8, err
}

func (p *BinaryProtocol) SkipBytesType() (int, error) {
	v, n := protowire.ConsumeVarint((p.Buf)[p.Read:])
	if n < 0 {
		return n, errDecodeField
	}
	all := int(v) + n
	_, err := p.next(all)
	return all, err
}

// skip (L)V once by wireType, useNative is not implemented
func (p *BinaryProtocol) Skip(wireType proto.WireType, useNative bool) (err error) {
	switch wireType {
	case proto.VarintType:
		_, err = p.ReadVarint()
	case proto.Fixed32Type:
		_, err = p.SkipFixed32Type() // the same as p.ReadFixed32Type() but without slice bytes
	case proto.Fixed64Type:
		_, err = p.SkipFixed64Type() // the same as p.ReadFixed64Type() but without slice bytes
	case proto.BytesType:
		_, err = p.SkipBytesType() // the same as p.ReadBytesType() but without slice bytes
	}
	return
}

// fast skip all elements in LIST/MAP
func (p *BinaryProtocol) SkipAllElements(fieldNumber proto.FieldNumber, ispacked bool) (size int, err error) {
	size = 0
	if ispacked {
		if _, _, _, err := p.ConsumeTag(); err != nil {
			return -1, err
		}
		bytelen, err := p.ReadLength()
		if err != nil {
			return -1, err
		}
		start := p.Read
		for p.Read < start+int(bytelen) {
			if _, err := p.ReadVarint(); err != nil {
				return -1, err
			}
			size++
		}
	} else {
		for p.Read < len(p.Buf) {
			elementFieldNumber, ewt, n, err := p.ConsumeTagWithoutMove()
			if err != nil {
				return -1, err
			}
			if elementFieldNumber != fieldNumber {
				break
			}
			if _, moveTagErr := p.next(n); moveTagErr != nil {
				return -1, moveTagErr
			}
			if err = p.Skip(ewt, false); err != nil {
				return -1, err
			}
			size++
		}
	}

	return size, nil
}
