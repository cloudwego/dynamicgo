package generic

import (
	"github.com/cloudwego/dynamicgo/meta"
	"github.com/cloudwego/dynamicgo/proto"
	"github.com/cloudwego/dynamicgo/proto/binary"
)

func (self Node) iterFields() (fi structIterator) {
	fi.p.Buf = self.raw()
	return
}

func (self Node) iterElems() (fi listIterator) {
	buf := self.raw()
	// fi = listIteratorPool.Get().(*ListIterator)
	fi.p.Buf = buf
	if _, wtyp, _, err := fi.p.ConsumeTag(); err != nil {
		fi.Err = meta.NewError(meta.ErrRead, "", err)
		return
	} else {
		if wtyp != proto.BytesType {
			fi.Err = wrapError(meta.ErrUnsupportedType, "ListIterator: wire type is not bytes.", nil)
			return
		}
		// maybe we could calculate fi.size in the fulture.
		fi.size = -1
		fi.et = proto.Type(self.et)
		kind := fi.et.TypeToKind()
		fi.ewt = proto.Kind2Wire[kind]
	}
	return
}

func (self Node) iterPairs() (fi mapIterator) {
	buf := self.raw()
	// fi = pairIteratorPool.Get().(*PairIterator)
	fi.p.Buf = buf
	if _, wtyp, _, err := fi.p.ConsumeTagWithoutMove(); err != nil {
		fi.Err = meta.NewError(meta.ErrRead, "", err)
		return
	} else {
		if wtyp != proto.BytesType {
			fi.Err = wrapError(meta.ErrUnsupportedType, "MapIterator: wire type is not bytes.", nil)
		}
		// maybe we could calculate fi.size in the fulture.
		fi.size = -1
		fi.vt = proto.Type(self.et)
		fi.vwt = proto.Kind2Wire[fi.vt.TypeToKind()]
		fi.kt = proto.Type(self.kt)
		fi.kwt = proto.Kind2Wire[fi.kt.TypeToKind()]
	}
	return
}

type structIterator struct {
	Err error
	p   binary.BinaryProtocol
}

func (it structIterator) HasNext() bool {
	return it.Err == nil && it.p.Left() > 0
}

// start:end containg the tag
func (it *structIterator) Next(useNative bool) (id proto.FieldNumber, typ proto.WireType, start int, end int, tagPos int) {
	tagPos = it.p.Read
	fieldId, wireType, _, err := it.p.ConsumeTag()
	if err != nil {
		it.Err = meta.NewError(meta.ErrRead, "", err)
		return
	}

	start = it.p.Read
	typ = wireType
	err = it.p.Skip(wireType, useNative)
	if err != nil {
		it.Err = meta.NewError(meta.ErrRead, "", err)
		return
	}
	end = it.p.Read

	id = proto.FieldNumber(fieldId)
	return
}


type listIterator struct {
	Err  error
	size int
	k    int
	et   proto.Type
	ewt  proto.WireType
	p    binary.BinaryProtocol
}

func (it listIterator) HasNext() bool {
	return it.Err == nil && it.p.Left() > 0
}


func (it listIterator) Size() int {
	return it.size
}

func (it listIterator) Pos() int {
	return it.k
}

func (it listIterator) WireType() proto.WireType {
	return it.ewt
}


func (it *listIterator) Next(useNative bool) (start int, end int) {
	start = it.p.Read
	err := it.p.Skip(it.ewt, useNative)
	if err != nil {
		it.Err = meta.NewError(meta.ErrRead, "", err)
		return
	}
	end = it.p.Read
	it.k++
	return
}


type mapIterator struct {
	Err  error
	size int
	i    int
	kt   proto.Type
	kwt  proto.WireType
	vt   proto.Type
	vwt  proto.WireType
	p    binary.BinaryProtocol
}


func (it mapIterator) HasNext() bool {
	return it.Err == nil && it.p.Left() > 0
}

func (it mapIterator) Size() int {
	return it.size
}

func (it mapIterator) Pos() int {
	return it.i
}

func (it *mapIterator) NextStr(useNative bool) (keyStart int, keyString string, start int, end int) {
	if _, _, _, err := it.p.ConsumeTag(); err != nil {
		it.Err = meta.NewError(meta.ErrRead, "", err)
		return
	}

	if _, err := it.p.ReadLength(); err != nil {
		it.Err = meta.NewError(meta.ErrRead, "", err)
		return
	}

	keyStart = it.p.Read
	var err error
	if it.kt == proto.STRING {
		// need read tag?
		_,kwType,_,err := it.p.ConsumeTag()
		if err != nil {
			it.Err = meta.NewError(meta.ErrRead, "", err)
			return
		}

		if kwType != it.kwt {
			it.Err = wrapError(meta.ErrUnsupportedType, "MapIterator.nextStr: key type is not expected", nil)
			return
		}

		keyString, err = it.p.ReadString(false)
		if err != nil {
			it.Err = meta.NewError(meta.ErrRead, "", err)
			return
		}
	} else {
		it.Err = wrapError(meta.ErrUnsupportedType, "MapIterator.nextStr: key type is not string", nil)
		return
	}

	// read value tag
	_,ewType,_,err := it.p.ConsumeTag()
	if err != nil {
		it.Err = meta.NewError(meta.ErrRead, "", err)
		return
	}

	if ewType != it.vwt {
		it.Err = wrapError(meta.ErrUnsupportedType, "MapIterator.nextStr: value type is not expected", nil)
		return
	}

	start = it.p.Read
	
	if err = it.p.Skip(it.vwt,useNative); err != nil {
		it.Err = meta.NewError(meta.ErrRead, "", err)
		return
	}
	end = it.p.Read
	it.i++
	return
}


func (it *mapIterator) NextInt(useNative bool) (keyStart int, keyInt int, start int, end int) {
	if _, _, _, err := it.p.ConsumeTag(); err != nil {
		it.Err = meta.NewError(meta.ErrRead, "", err)
		return
	}

	if _, err := it.p.ReadLength(); err != nil {
		it.Err = meta.NewError(meta.ErrRead, "", err)
		return
	}
	
	keyStart = it.p.Read
	var err error
	if it.kt.IsInt() {
		// need read tag?
		_,kwType,_,err := it.p.ConsumeTag()
		if err != nil {
			it.Err = meta.NewError(meta.ErrRead, "", err)
			return
		}

		if kwType != it.kwt {
			it.Err = wrapError(meta.ErrUnsupportedType, "MapIterator.nextStr: key type is not expected", nil)
			return
		}

		keyInt, err = it.p.ReadInt(it.kt)
		if err != nil {
			it.Err = meta.NewError(meta.ErrRead, "", err)
			return
		}

	} else {
		it.Err = wrapError(meta.ErrUnsupportedType, "MapIterator.nextStr: key type is not string", nil)
		return
	}

	// read value tag
	_,ewType,_,err := it.p.ConsumeTag()
	if err != nil {
		it.Err = meta.NewError(meta.ErrRead, "", err)
		return
	}

	if ewType != it.vwt {
		it.Err = wrapError(meta.ErrUnsupportedType, "MapIterator.nextStr: value type is not expected", nil)
		return
	}

	start = it.p.Read
	err = it.p.Skip(it.vwt,
		 useNative)
	if err != nil {
		it.Err = meta.NewError(meta.ErrRead, "", err)
		return
	}
	end = it.p.Read
	it.i++
	return
}

// TODO: test forearch and foreachKV