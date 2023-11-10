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
	fi.p.Buf = buf
	if _, wtyp, _, err := fi.p.ConsumeTagWithoutMove(); err != nil {
		fi.Err = wrapError(meta.ErrRead, "ListIterator.iterElems: consume list tag error.", err)
		return
	} else {
		if wtyp != proto.BytesType {
			fi.Err = wrapError(meta.ErrUnsupportedType, "ListIterator.iterElems: wire type is not bytes.", nil)
			return
		}
		
		size, err := self.Len()
		if err != nil {
			fi.Err = wrapError(meta.ErrRead, "ListIterator.iterElems: get list size error.", err)
			return
		}
		fi.size = size
		fi.et = proto.Type(self.et)
		kind := fi.et.TypeToKind()
		fi.ewt = proto.Kind2Wire[kind]
		fi.isPacked = self.et.NeedVarint()
	}
	return
}

func (self Node) iterPairs() (fi mapIterator) {
	fi.p.Buf = self.raw()
	// fi = pairIteratorPool.Get().(*PairIterator)
	if _, wtyp, _, err := fi.p.ConsumeTagWithoutMove(); err != nil {
		fi.Err = wrapError(meta.ErrRead, "MapIterator.iterPairs: consume map tag error.", err)
		return
	} else {
		if wtyp != proto.BytesType {
			fi.Err = wrapError(meta.ErrUnsupportedType, "MapIterator.iterPairs: wire type is not bytes.", nil)
		}

		size, err := self.Len()
		if err != nil {
			fi.Err = wrapError(meta.ErrRead, "MapIterator.iterPairs: get map size error.", err)
			return
		}
		fi.size = size
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
		it.Err = wrapError(meta.ErrRead, "StructIterator.Next: consume field tag error.", err)
		return
	}

	start = it.p.Read
	typ = wireType
	if err = it.p.Skip(wireType, useNative); err != nil {
		it.Err = wrapError(meta.ErrRead, "StructIterator.Next: skip field data error.", err)
		return
	}
	end = it.p.Read

	id = proto.FieldNumber(fieldId)
	return
}


type listIterator struct {
	Err  error
	p    binary.BinaryProtocol
	size int // elements count, when lazyload mode, the size is also 0, so we use k to count
	k    int // count of elements that has been read
	isPacked bool // list mode
	et   proto.Type // element type
	ewt  proto.WireType // element wire type
}

func (it listIterator) HasNext() bool {
	return it.Err == nil && it.p.Left() > 0
}

func (it listIterator) IsPacked() bool {
	return it.isPacked
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

// compatible for packed/unpacked list, get element value: (L)V or V without tag
func (it *listIterator) Next(useNative bool) (start int, end int) {
	if !it.isPacked {
		if _, _, _, err := it.p.ConsumeTag(); err != nil {
			it.Err = wrapError(meta.ErrRead, "ListIterator: consume list tag error.", err)
			return
		}
	}

	start = it.p.Read
	if err := it.p.Skip(it.ewt, useNative); err != nil {
		it.Err = wrapError(meta.ErrRead, "ListIterator: skip list element error.", err)
		return
	}
	end = it.p.Read
	it.k++
	return
}


type mapIterator struct {
	Err  error
	p    binary.BinaryProtocol
	size int // elements count
	i    int // count of elements that has been read
	kt   proto.Type // key type
	kwt  proto.WireType // key wire type
	vt   proto.Type // value type
	vwt  proto.WireType // value wire type
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
	// pair tag
	if _, _, _, err := it.p.ConsumeTag(); err != nil {
		it.Err = wrapError(meta.ErrRead, "MapIterator: consume pair tag error.", err)
		return
	}

	if _, err := it.p.ReadLength(); err != nil {
		it.Err = wrapError(meta.ErrRead, "MapIterator: consume pair length error.", err)
		return
	}
	
	// read key tag
	keyStart = it.p.Read
	var err error
	if it.kt == proto.STRING {
		_,kwType,_,err := it.p.ConsumeTag()
		if err != nil {
			it.Err = wrapError(meta.ErrRead, "MapIterator: consume key tag error.", err)
			return
		}

		if kwType != it.kwt {
			it.Err = wrapError(meta.ErrUnsupportedType, "MapIterator.nextStr: key type is not expected", nil)
			return
		}

		keyString, err = it.p.ReadString(false)
		if err != nil {
			it.Err = wrapError(meta.ErrRead, "", err)
			return
		}
	} else {
		it.Err = wrapError(meta.ErrUnsupportedType, "MapIterator.nextStr: key type is not string.", nil)
		return
	}

	// read value tag
	_,ewType,_,err := it.p.ConsumeTag()
	if err != nil {
		it.Err = wrapError(meta.ErrRead, "", err)
		return
	}

	if ewType != it.vwt {
		it.Err = wrapError(meta.ErrUnsupportedType, "MapIterator.nextStr: value type is not expected", nil)
		return
	}

	start = it.p.Read
	
	if err = it.p.Skip(it.vwt, useNative); err != nil {
		it.Err = wrapError(meta.ErrRead, "", err)
		return
	}
	end = it.p.Read
	it.i++
	return
}


func (it *mapIterator) NextInt(useNative bool) (keyStart int, keyInt int, start int, end int) {
	if _, _, _, err := it.p.ConsumeTag(); err != nil {
		it.Err = wrapError(meta.ErrRead, "", err)
		return
	}

	if _, err := it.p.ReadLength(); err != nil {
		it.Err = wrapError(meta.ErrRead, "", err)
		return
	}
	
	keyStart = it.p.Read
	var err error
	if it.kt.IsInt() {
		// read key tag
		_,kwType,_,err := it.p.ConsumeTag()
		if err != nil {
			it.Err = wrapError(meta.ErrRead, "", err)
			return
		}

		if kwType != it.kwt {
			it.Err = wrapError(meta.ErrUnsupportedType, "MapIterator.nextStr: key type is not expected", nil)
			return
		}

		keyInt, err = it.p.ReadInt(it.kt)
		if err != nil {
			it.Err = wrapError(meta.ErrRead, "", err)
			return
		}

	} else {
		it.Err = wrapError(meta.ErrUnsupportedType, "MapIterator.nextStr: key type is not string", nil)
		return
	}

	// read value tag
	_,ewType,_,err := it.p.ConsumeTag()
	if err != nil {
		it.Err = wrapError(meta.ErrRead, "", err)
		return
	}

	if ewType != it.vwt {
		it.Err = wrapError(meta.ErrUnsupportedType, "MapIterator.nextStr: value type is not expected", nil)
		return
	}

	start = it.p.Read
	
	if err = it.p.Skip(it.vwt, useNative); err != nil {
		it.Err = wrapError(meta.ErrRead, "", err)
		return
	}
	end = it.p.Read
	it.i++
	return
}