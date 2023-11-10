package binary

import (
	"github.com/cloudwego/dynamicgo/proto"
)

// skip (L)V once by wireType, useNative is not implemented
func (p *BinaryProtocol) Skip(wireType proto.WireType, useNative bool) (err error) {
	switch wireType {
	case proto.VarintType:
		_, err = p.ReadVarint()
	case proto.Fixed32Type:
		_, err = p.ReadFixed32()
	case proto.Fixed64Type:
		_, err = p.ReadFixed64()
	case proto.BytesType:
		_, err = p.ReadBytes()
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
