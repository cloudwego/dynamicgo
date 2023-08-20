package binary

import (
	"github.com/cloudwego/dynamicgo/proto"
)


const MaxSkipDepth = 1023

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



