package generic

import (
	"fmt"

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
			// Error or unsupport unpacked List mode now (only support packed mode)
			fi.Err = wrapError(meta.ErrUnsupportedType, "ListIterator: wire type is not bytes, ReadError or unsupport unpacked List mode now (only support packed mode)", nil)
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
			// Error or unsupport unpacked Map mode now (only support packed mode)
			fi.Err = wrapError(meta.ErrUnsupportedType, "MapIterator: wire type is not bytes, ReadError or unsupport unpacked Map mode now (only support packed mode)", nil)
			return
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

// Foreach scan each element of a complex type (LIST/SET/MAP/STRUCT),
// and call handler sequentially with corresponding path and value
func (self Value) Foreach(handler func(path Path, val Value) bool, opts *Options) error {
	switch self.t {
	case proto.MESSAGE:
		it := self.iterFields()
		if it.Err != nil {
			return it.Err
		}
		for it.HasNext() {
			id, wtyp, start, end, tagPos := it.Next(opts.UseNativeSkip)
			if it.Err != nil {
				return it.Err
			}

			var f proto.FieldDescriptor
			if self.rootDesc != nil {
				f = (*self.rootDesc).Fields().ByNumber(id)
			} else {
				f = (*self.Desc).Message().Fields().ByNumber(id)
			}
			
			if f == nil {
				if opts.DisallowUnknow {
					return wrapError(meta.ErrUnknownField, fmt.Sprintf("unknown field %d", id), nil)
				}
				continue
			}
			
			if proto.Kind2Wire[f.Kind()] != wtyp {
				return wrapError(meta.ErrDismatchType, fmt.Sprintf("field %d type %v mismatch read type %v", id, proto.Kind2Wire[f.Kind()], wtyp), nil)
			}
			var v Value
			if f.IsMap() {
				it.p.Read = tagPos
				if _, err := it.p.ReadMap(&f, false,false,false); err != nil {
					return wrapError(meta.ErrRead, "", err)
				}
				start = tagPos
				end = it.p.Read
			}

			v = self.slice(start, end, &f)
			var p Path
			if opts.IterateStructByName && f != nil {
				p = NewPathFieldName(f.TextName())
			} else {
				p = NewPathFieldId(id)
			}
			if !handler(p, v) {
				return nil
			}
		}
	case proto.LIST:
		it := self.iterElems()
		if it.Err != nil {
			return it.Err
		}

		for i := 0; it.HasNext(); i++ {
			start, end := it.Next(opts.UseNativeSkip)
			if it.Err != nil {
				return it.Err
			}
			v := self.slice(start, end, self.Desc)
			p := NewPathIndex(i)
			if !handler(p, v) {
				return nil
			}
		}
	case proto.MAP:
		it := self.iterPairs()
		if it.Err != nil {
			return it.Err
		}
		keyDesc := (*self.Desc).MapKey()
		valueDesc := (*self.Desc).MapValue()
		kt := proto.FromProtoKindToType(keyDesc.Kind(),false,false)
		vt := proto.FromProtoKindToType(valueDesc.Kind(),valueDesc.IsList(),valueDesc.IsMap())
		if kt != it.kt {
			return wrapError(meta.ErrDismatchType, "dismatched descriptor key type and binary type", nil)
		}
		if vt != it.vt {
			return wrapError(meta.ErrDismatchType, "dismatched descriptor elem type and binary type", nil)
		}

		if kt == proto.STRING {
			for i := 0; it.HasNext(); i++ {
				_, keyString, start, end := it.NextStr(opts.UseNativeSkip)
				if it.Err != nil {
					return it.Err
				}
				v := self.slice(start, end, &valueDesc)
				p := NewPathStrKey(keyString)
				if !handler(p, v) {
					return nil
				}
			}
		} else if kt.IsInt() {
			for i := 0; it.HasNext(); i++ {
				_, keyInt, start, end := it.NextInt(opts.UseNativeSkip)
				if it.Err != nil {
					return it.Err
				}
				v := self.slice(start, end, &valueDesc)
				p := NewPathIntKey(keyInt)
				if !handler(p, v) {
					return nil
				}
			}
		}
	}
	return nil
}


// ForeachKV scan each element of a MAP type, and call handler sequentially with corresponding key and value
func (self Value) ForeachKV(handler func(key Value, val Value) bool, opts *Options) error {
	if err := self.should("MAP", proto.MAP); err != "" {
		return wrapError(meta.ErrUnsupportedType, err, nil)
	}
	switch self.t {
	case proto.MAP:
		p := binary.BinaryProtocol{Buf: self.raw()}

		keyDesc := (*self.Desc).MapKey()
		valueDesc := (*self.Desc).MapValue()
		kt := proto.FromProtoKindToType(keyDesc.Kind(),false,false)
		vt := proto.FromProtoKindToType(valueDesc.Kind(),valueDesc.IsList(),valueDesc.IsMap())

		if kt != self.kt {
			return wrapError(meta.ErrDismatchType, "dismatched descriptor key type and binary type", nil)
		}

		if vt != self.et {
			return wrapError(meta.ErrDismatchType, "dismatched descriptor elem type and binary type", nil)
		}
		fieldNumebr := (*self.Desc).Number()

		for p.Read < len(p.Buf) {
			pairId, _, pairLen, err := p.ConsumeTagWithoutMove()
			if err != nil {
				return errNode(meta.ErrRead, "", nil)
			}

			if pairId != fieldNumebr {
				break
			}
			p.Read += pairLen
			_, pairErr := p.ReadLength()
			if pairErr != nil {
				return errNode(meta.ErrRead, "", nil)
			}

			// read key tag
			if _, _, _, err := p.ConsumeTag(); err != nil {
				return errNode(meta.ErrRead, "", nil)
			}

			ks := p.Read
			kwt := proto.Kind2Wire[keyDesc.Kind()]

			if err := p.Skip(kwt, opts.UseNativeSkip); err != nil {
				return errNode(meta.ErrRead, "", nil)
			}
			key := self.slice(ks, p.Read, &keyDesc)
			
			// read value tag
			if _, _, _, err := p.ConsumeTag(); err != nil {
				return errNode(meta.ErrRead, "", nil)
			}

			vs := p.Read
			vwt := proto.Kind2Wire[valueDesc.Kind()]
			if err := p.Skip(vwt, opts.UseNativeSkip); err != nil {
				return errNode(meta.ErrRead, "", nil)
			}
			val := self.slice(vs, p.Read, &valueDesc)
			if !handler(key, val) {
				return nil
			}
		}
		return nil
	default:
		return errNode(meta.ErrUnsupportedType, "", nil)
	}
}