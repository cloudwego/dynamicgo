// Code generated by Fastpb v0.0.2. DO NOT EDIT.

package baseline

import (
	fmt "fmt"
	fastpb "github.com/cloudwego/fastpb"
)

var (
	_ = fmt.Errorf
	_ = fastpb.Skip
)

func (x *Simple) FastRead(buf []byte, _type int8, number int32) (offset int, err error) {
	switch number {
	case 1:
		offset, err = x.fastReadField1(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 2:
		offset, err = x.fastReadField2(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 3:
		offset, err = x.fastReadField3(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 4:
		offset, err = x.fastReadField4(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 5:
		offset, err = x.fastReadField5(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 6:
		offset, err = x.fastReadField6(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	default:
		offset, err = fastpb.Skip(buf, _type, number)
		if err != nil {
			goto SkipFieldError
		}
	}
	return offset, nil
SkipFieldError:
	return offset, fmt.Errorf("%T cannot parse invalid wire-format data, error: %s", x, err)
ReadFieldError:
	return offset, fmt.Errorf("%T read field %d '%s' error: %s", x, number, fieldIDToName_Simple[number], err)
}

func (x *Simple) fastReadField1(buf []byte, _type int8) (offset int, err error) {
	x.ByteField, offset, err = fastpb.ReadBytes(buf, _type)
	return offset, err
}

func (x *Simple) fastReadField2(buf []byte, _type int8) (offset int, err error) {
	x.I64Field, offset, err = fastpb.ReadInt64(buf, _type)
	return offset, err
}

func (x *Simple) fastReadField3(buf []byte, _type int8) (offset int, err error) {
	x.DoubleField, offset, err = fastpb.ReadDouble(buf, _type)
	return offset, err
}

func (x *Simple) fastReadField4(buf []byte, _type int8) (offset int, err error) {
	x.I32Field, offset, err = fastpb.ReadInt32(buf, _type)
	return offset, err
}

func (x *Simple) fastReadField5(buf []byte, _type int8) (offset int, err error) {
	x.StringField, offset, err = fastpb.ReadString(buf, _type)
	return offset, err
}

func (x *Simple) fastReadField6(buf []byte, _type int8) (offset int, err error) {
	x.BinaryField, offset, err = fastpb.ReadBytes(buf, _type)
	return offset, err
}

func (x *PartialSimple) FastRead(buf []byte, _type int8, number int32) (offset int, err error) {
	switch number {
	case 1:
		offset, err = x.fastReadField1(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 3:
		offset, err = x.fastReadField3(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 6:
		offset, err = x.fastReadField6(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	default:
		offset, err = fastpb.Skip(buf, _type, number)
		if err != nil {
			goto SkipFieldError
		}
	}
	return offset, nil
SkipFieldError:
	return offset, fmt.Errorf("%T cannot parse invalid wire-format data, error: %s", x, err)
ReadFieldError:
	return offset, fmt.Errorf("%T read field %d '%s' error: %s", x, number, fieldIDToName_PartialSimple[number], err)
}

func (x *PartialSimple) fastReadField1(buf []byte, _type int8) (offset int, err error) {
	x.ByteField, offset, err = fastpb.ReadBytes(buf, _type)
	return offset, err
}

func (x *PartialSimple) fastReadField3(buf []byte, _type int8) (offset int, err error) {
	x.DoubleField, offset, err = fastpb.ReadDouble(buf, _type)
	return offset, err
}

func (x *PartialSimple) fastReadField6(buf []byte, _type int8) (offset int, err error) {
	x.BinaryField, offset, err = fastpb.ReadBytes(buf, _type)
	return offset, err
}

func (x *Nesting) FastRead(buf []byte, _type int8, number int32) (offset int, err error) {
	switch number {
	case 1:
		offset, err = x.fastReadField1(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 2:
		offset, err = x.fastReadField2(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 3:
		offset, err = x.fastReadField3(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 4:
		offset, err = x.fastReadField4(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 5:
		offset, err = x.fastReadField5(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 6:
		offset, err = x.fastReadField6(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 7:
		offset, err = x.fastReadField7(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 8:
		offset, err = x.fastReadField8(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 9:
		offset, err = x.fastReadField9(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 10:
		offset, err = x.fastReadField10(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 11:
		offset, err = x.fastReadField11(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 12:
		offset, err = x.fastReadField12(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 13:
		offset, err = x.fastReadField13(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 14:
		offset, err = x.fastReadField14(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 15:
		offset, err = x.fastReadField15(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	default:
		offset, err = fastpb.Skip(buf, _type, number)
		if err != nil {
			goto SkipFieldError
		}
	}
	return offset, nil
SkipFieldError:
	return offset, fmt.Errorf("%T cannot parse invalid wire-format data, error: %s", x, err)
ReadFieldError:
	return offset, fmt.Errorf("%T read field %d '%s' error: %s", x, number, fieldIDToName_Nesting[number], err)
}

func (x *Nesting) fastReadField1(buf []byte, _type int8) (offset int, err error) {
	x.String_, offset, err = fastpb.ReadString(buf, _type)
	return offset, err
}

func (x *Nesting) fastReadField2(buf []byte, _type int8) (offset int, err error) {
	var v Simple
	offset, err = fastpb.ReadMessage(buf, _type, &v)
	if err != nil {
		return offset, err
	}
	x.ListSimple = append(x.ListSimple, &v)
	return offset, nil
}

func (x *Nesting) fastReadField3(buf []byte, _type int8) (offset int, err error) {
	x.Double, offset, err = fastpb.ReadDouble(buf, _type)
	return offset, err
}

func (x *Nesting) fastReadField4(buf []byte, _type int8) (offset int, err error) {
	x.I32, offset, err = fastpb.ReadInt32(buf, _type)
	return offset, err
}

func (x *Nesting) fastReadField5(buf []byte, _type int8) (offset int, err error) {
	offset, err = fastpb.ReadList(buf, _type,
		func(buf []byte, _type int8) (n int, err error) {
			var v int32
			v, offset, err = fastpb.ReadInt32(buf, _type)
			if err != nil {
				return offset, err
			}
			x.ListI32 = append(x.ListI32, v)
			return offset, err
		})
	return offset, err
}

func (x *Nesting) fastReadField6(buf []byte, _type int8) (offset int, err error) {
	x.I64, offset, err = fastpb.ReadInt64(buf, _type)
	return offset, err
}

func (x *Nesting) fastReadField7(buf []byte, _type int8) (offset int, err error) {
	if x.MapStringString == nil {
		x.MapStringString = make(map[string]string)
	}
	var key string
	var value string
	offset, err = fastpb.ReadMapEntry(buf, _type,
		func(buf []byte, _type int8) (offset int, err error) {
			key, offset, err = fastpb.ReadString(buf, _type)
			return offset, err
		},
		func(buf []byte, _type int8) (offset int, err error) {
			value, offset, err = fastpb.ReadString(buf, _type)
			return offset, err
		})
	if err != nil {
		return offset, err
	}
	x.MapStringString[key] = value
	return offset, nil
}

func (x *Nesting) fastReadField8(buf []byte, _type int8) (offset int, err error) {
	var v Simple
	offset, err = fastpb.ReadMessage(buf, _type, &v)
	if err != nil {
		return offset, err
	}
	x.SimpleStruct = &v
	return offset, nil
}

func (x *Nesting) fastReadField9(buf []byte, _type int8) (offset int, err error) {
	if x.MapI32I64 == nil {
		x.MapI32I64 = make(map[int32]int64)
	}
	var key int32
	var value int64
	offset, err = fastpb.ReadMapEntry(buf, _type,
		func(buf []byte, _type int8) (offset int, err error) {
			key, offset, err = fastpb.ReadInt32(buf, _type)
			return offset, err
		},
		func(buf []byte, _type int8) (offset int, err error) {
			value, offset, err = fastpb.ReadInt64(buf, _type)
			return offset, err
		})
	if err != nil {
		return offset, err
	}
	x.MapI32I64[key] = value
	return offset, nil
}

func (x *Nesting) fastReadField10(buf []byte, _type int8) (offset int, err error) {
	var v string
	v, offset, err = fastpb.ReadString(buf, _type)
	if err != nil {
		return offset, err
	}
	x.ListString = append(x.ListString, v)
	return offset, err
}

func (x *Nesting) fastReadField11(buf []byte, _type int8) (offset int, err error) {
	x.Binary, offset, err = fastpb.ReadBytes(buf, _type)
	return offset, err
}

func (x *Nesting) fastReadField12(buf []byte, _type int8) (offset int, err error) {
	if x.MapI64String == nil {
		x.MapI64String = make(map[int64]string)
	}
	var key int64
	var value string
	offset, err = fastpb.ReadMapEntry(buf, _type,
		func(buf []byte, _type int8) (offset int, err error) {
			key, offset, err = fastpb.ReadInt64(buf, _type)
			return offset, err
		},
		func(buf []byte, _type int8) (offset int, err error) {
			value, offset, err = fastpb.ReadString(buf, _type)
			return offset, err
		})
	if err != nil {
		return offset, err
	}
	x.MapI64String[key] = value
	return offset, nil
}

func (x *Nesting) fastReadField13(buf []byte, _type int8) (offset int, err error) {
	offset, err = fastpb.ReadList(buf, _type,
		func(buf []byte, _type int8) (n int, err error) {
			var v int64
			v, offset, err = fastpb.ReadInt64(buf, _type)
			if err != nil {
				return offset, err
			}
			x.ListI64 = append(x.ListI64, v)
			return offset, err
		})
	return offset, err
}

func (x *Nesting) fastReadField14(buf []byte, _type int8) (offset int, err error) {
	x.Byte, offset, err = fastpb.ReadBytes(buf, _type)
	return offset, err
}

func (x *Nesting) fastReadField15(buf []byte, _type int8) (offset int, err error) {
	if x.MapStringSimple == nil {
		x.MapStringSimple = make(map[string]*Simple)
	}
	var key string
	var value *Simple
	offset, err = fastpb.ReadMapEntry(buf, _type,
		func(buf []byte, _type int8) (offset int, err error) {
			key, offset, err = fastpb.ReadString(buf, _type)
			return offset, err
		},
		func(buf []byte, _type int8) (offset int, err error) {
			var v Simple
			offset, err = fastpb.ReadMessage(buf, _type, &v)
			if err != nil {
				return offset, err
			}
			value = &v
			return offset, nil
		})
	if err != nil {
		return offset, err
	}
	x.MapStringSimple[key] = value
	return offset, nil
}

func (x *PartialNesting) FastRead(buf []byte, _type int8, number int32) (offset int, err error) {
	switch number {
	case 2:
		offset, err = x.fastReadField2(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 8:
		offset, err = x.fastReadField8(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 15:
		offset, err = x.fastReadField15(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	default:
		offset, err = fastpb.Skip(buf, _type, number)
		if err != nil {
			goto SkipFieldError
		}
	}
	return offset, nil
SkipFieldError:
	return offset, fmt.Errorf("%T cannot parse invalid wire-format data, error: %s", x, err)
ReadFieldError:
	return offset, fmt.Errorf("%T read field %d '%s' error: %s", x, number, fieldIDToName_PartialNesting[number], err)
}

func (x *PartialNesting) fastReadField2(buf []byte, _type int8) (offset int, err error) {
	var v PartialSimple
	offset, err = fastpb.ReadMessage(buf, _type, &v)
	if err != nil {
		return offset, err
	}
	x.ListSimple = append(x.ListSimple, &v)
	return offset, nil
}

func (x *PartialNesting) fastReadField8(buf []byte, _type int8) (offset int, err error) {
	var v PartialSimple
	offset, err = fastpb.ReadMessage(buf, _type, &v)
	if err != nil {
		return offset, err
	}
	x.SimpleStruct = &v
	return offset, nil
}

func (x *PartialNesting) fastReadField15(buf []byte, _type int8) (offset int, err error) {
	if x.MapStringSimple == nil {
		x.MapStringSimple = make(map[string]*PartialSimple)
	}
	var key string
	var value *PartialSimple
	offset, err = fastpb.ReadMapEntry(buf, _type,
		func(buf []byte, _type int8) (offset int, err error) {
			key, offset, err = fastpb.ReadString(buf, _type)
			return offset, err
		},
		func(buf []byte, _type int8) (offset int, err error) {
			var v PartialSimple
			offset, err = fastpb.ReadMessage(buf, _type, &v)
			if err != nil {
				return offset, err
			}
			value = &v
			return offset, nil
		})
	if err != nil {
		return offset, err
	}
	x.MapStringSimple[key] = value
	return offset, nil
}

func (x *Nesting2) FastRead(buf []byte, _type int8, number int32) (offset int, err error) {
	switch number {
	case 1:
		offset, err = x.fastReadField1(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 2:
		offset, err = x.fastReadField2(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 3:
		offset, err = x.fastReadField3(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 4:
		offset, err = x.fastReadField4(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 5:
		offset, err = x.fastReadField5(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 6:
		offset, err = x.fastReadField6(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 7:
		offset, err = x.fastReadField7(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 8:
		offset, err = x.fastReadField8(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 9:
		offset, err = x.fastReadField9(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 10:
		offset, err = x.fastReadField10(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	case 11:
		offset, err = x.fastReadField11(buf, _type)
		if err != nil {
			goto ReadFieldError
		}
	default:
		offset, err = fastpb.Skip(buf, _type, number)
		if err != nil {
			goto SkipFieldError
		}
	}
	return offset, nil
SkipFieldError:
	return offset, fmt.Errorf("%T cannot parse invalid wire-format data, error: %s", x, err)
ReadFieldError:
	return offset, fmt.Errorf("%T read field %d '%s' error: %s", x, number, fieldIDToName_Nesting2[number], err)
}

func (x *Nesting2) fastReadField1(buf []byte, _type int8) (offset int, err error) {
	if x.MapSimpleNesting == nil {
		x.MapSimpleNesting = make(map[int64]*Nesting)
	}
	var key int64
	var value *Nesting
	offset, err = fastpb.ReadMapEntry(buf, _type,
		func(buf []byte, _type int8) (offset int, err error) {
			key, offset, err = fastpb.ReadInt64(buf, _type)
			return offset, err
		},
		func(buf []byte, _type int8) (offset int, err error) {
			var v Nesting
			offset, err = fastpb.ReadMessage(buf, _type, &v)
			if err != nil {
				return offset, err
			}
			value = &v
			return offset, nil
		})
	if err != nil {
		return offset, err
	}
	x.MapSimpleNesting[key] = value
	return offset, nil
}

func (x *Nesting2) fastReadField2(buf []byte, _type int8) (offset int, err error) {
	var v Simple
	offset, err = fastpb.ReadMessage(buf, _type, &v)
	if err != nil {
		return offset, err
	}
	x.SimpleStruct = &v
	return offset, nil
}

func (x *Nesting2) fastReadField3(buf []byte, _type int8) (offset int, err error) {
	x.Byte, offset, err = fastpb.ReadBytes(buf, _type)
	return offset, err
}

func (x *Nesting2) fastReadField4(buf []byte, _type int8) (offset int, err error) {
	x.Double, offset, err = fastpb.ReadDouble(buf, _type)
	return offset, err
}

func (x *Nesting2) fastReadField5(buf []byte, _type int8) (offset int, err error) {
	var v Nesting
	offset, err = fastpb.ReadMessage(buf, _type, &v)
	if err != nil {
		return offset, err
	}
	x.ListNesting = append(x.ListNesting, &v)
	return offset, nil
}

func (x *Nesting2) fastReadField6(buf []byte, _type int8) (offset int, err error) {
	x.I64, offset, err = fastpb.ReadInt64(buf, _type)
	return offset, err
}

func (x *Nesting2) fastReadField7(buf []byte, _type int8) (offset int, err error) {
	var v Nesting
	offset, err = fastpb.ReadMessage(buf, _type, &v)
	if err != nil {
		return offset, err
	}
	x.NestingStruct = &v
	return offset, nil
}

func (x *Nesting2) fastReadField8(buf []byte, _type int8) (offset int, err error) {
	x.Binary, offset, err = fastpb.ReadBytes(buf, _type)
	return offset, err
}

func (x *Nesting2) fastReadField9(buf []byte, _type int8) (offset int, err error) {
	x.String_, offset, err = fastpb.ReadString(buf, _type)
	return offset, err
}

func (x *Nesting2) fastReadField10(buf []byte, _type int8) (offset int, err error) {
	var v Nesting
	offset, err = fastpb.ReadMessage(buf, _type, &v)
	if err != nil {
		return offset, err
	}
	x.SetNesting = append(x.SetNesting, &v)
	return offset, nil
}

func (x *Nesting2) fastReadField11(buf []byte, _type int8) (offset int, err error) {
	x.I32, offset, err = fastpb.ReadInt32(buf, _type)
	return offset, err
}

func (x *Simple) FastWrite(buf []byte) (offset int) {
	if x == nil {
		return offset
	}
	offset += x.fastWriteField1(buf[offset:])
	offset += x.fastWriteField2(buf[offset:])
	offset += x.fastWriteField3(buf[offset:])
	offset += x.fastWriteField4(buf[offset:])
	offset += x.fastWriteField5(buf[offset:])
	offset += x.fastWriteField6(buf[offset:])
	return offset
}

func (x *Simple) fastWriteField1(buf []byte) (offset int) {
	if len(x.ByteField) == 0 {
		return offset
	}
	offset += fastpb.WriteBytes(buf[offset:], 1, x.GetByteField())
	return offset
}

func (x *Simple) fastWriteField2(buf []byte) (offset int) {
	if x.I64Field == 0 {
		return offset
	}
	offset += fastpb.WriteInt64(buf[offset:], 2, x.GetI64Field())
	return offset
}

func (x *Simple) fastWriteField3(buf []byte) (offset int) {
	if x.DoubleField == 0 {
		return offset
	}
	offset += fastpb.WriteDouble(buf[offset:], 3, x.GetDoubleField())
	return offset
}

func (x *Simple) fastWriteField4(buf []byte) (offset int) {
	if x.I32Field == 0 {
		return offset
	}
	offset += fastpb.WriteInt32(buf[offset:], 4, x.GetI32Field())
	return offset
}

func (x *Simple) fastWriteField5(buf []byte) (offset int) {
	if x.StringField == "" {
		return offset
	}
	offset += fastpb.WriteString(buf[offset:], 5, x.GetStringField())
	return offset
}

func (x *Simple) fastWriteField6(buf []byte) (offset int) {
	if len(x.BinaryField) == 0 {
		return offset
	}
	offset += fastpb.WriteBytes(buf[offset:], 6, x.GetBinaryField())
	return offset
}

func (x *PartialSimple) FastWrite(buf []byte) (offset int) {
	if x == nil {
		return offset
	}
	offset += x.fastWriteField1(buf[offset:])
	offset += x.fastWriteField3(buf[offset:])
	offset += x.fastWriteField6(buf[offset:])
	return offset
}

func (x *PartialSimple) fastWriteField1(buf []byte) (offset int) {
	if len(x.ByteField) == 0 {
		return offset
	}
	offset += fastpb.WriteBytes(buf[offset:], 1, x.GetByteField())
	return offset
}

func (x *PartialSimple) fastWriteField3(buf []byte) (offset int) {
	if x.DoubleField == 0 {
		return offset
	}
	offset += fastpb.WriteDouble(buf[offset:], 3, x.GetDoubleField())
	return offset
}

func (x *PartialSimple) fastWriteField6(buf []byte) (offset int) {
	if len(x.BinaryField) == 0 {
		return offset
	}
	offset += fastpb.WriteBytes(buf[offset:], 6, x.GetBinaryField())
	return offset
}

func (x *Nesting) FastWrite(buf []byte) (offset int) {
	if x == nil {
		return offset
	}
	offset += x.fastWriteField1(buf[offset:])
	offset += x.fastWriteField2(buf[offset:])
	offset += x.fastWriteField3(buf[offset:])
	offset += x.fastWriteField4(buf[offset:])
	offset += x.fastWriteField5(buf[offset:])
	offset += x.fastWriteField6(buf[offset:])
	offset += x.fastWriteField7(buf[offset:])
	offset += x.fastWriteField8(buf[offset:])
	offset += x.fastWriteField9(buf[offset:])
	offset += x.fastWriteField10(buf[offset:])
	offset += x.fastWriteField11(buf[offset:])
	offset += x.fastWriteField12(buf[offset:])
	offset += x.fastWriteField13(buf[offset:])
	offset += x.fastWriteField14(buf[offset:])
	offset += x.fastWriteField15(buf[offset:])
	return offset
}

func (x *Nesting) fastWriteField1(buf []byte) (offset int) {
	if x.String_ == "" {
		return offset
	}
	offset += fastpb.WriteString(buf[offset:], 1, x.GetString_())
	return offset
}

func (x *Nesting) fastWriteField2(buf []byte) (offset int) {
	if x.ListSimple == nil {
		return offset
	}
	for i := range x.GetListSimple() {
		offset += fastpb.WriteMessage(buf[offset:], 2, x.GetListSimple()[i])
	}
	return offset
}

func (x *Nesting) fastWriteField3(buf []byte) (offset int) {
	if x.Double == 0 {
		return offset
	}
	offset += fastpb.WriteDouble(buf[offset:], 3, x.GetDouble())
	return offset
}

func (x *Nesting) fastWriteField4(buf []byte) (offset int) {
	if x.I32 == 0 {
		return offset
	}
	offset += fastpb.WriteInt32(buf[offset:], 4, x.GetI32())
	return offset
}

func (x *Nesting) fastWriteField5(buf []byte) (offset int) {
	if len(x.ListI32) == 0 {
		return offset
	}
	offset += fastpb.WriteListPacked(buf[offset:], 5, len(x.GetListI32()),
		func(buf []byte, numTagOrKey, numIdxOrVal int32) int {
			offset := 0
			offset += fastpb.WriteInt32(buf[offset:], numTagOrKey, x.GetListI32()[numIdxOrVal])
			return offset
		})
	return offset
}

func (x *Nesting) fastWriteField6(buf []byte) (offset int) {
	if x.I64 == 0 {
		return offset
	}
	offset += fastpb.WriteInt64(buf[offset:], 6, x.GetI64())
	return offset
}

func (x *Nesting) fastWriteField7(buf []byte) (offset int) {
	if x.MapStringString == nil {
		return offset
	}
	for k, v := range x.GetMapStringString() {
		offset += fastpb.WriteMapEntry(buf[offset:], 7,
			func(buf []byte, numTagOrKey, numIdxOrVal int32) int {
				offset := 0
				offset += fastpb.WriteString(buf[offset:], numTagOrKey, k)
				offset += fastpb.WriteString(buf[offset:], numIdxOrVal, v)
				return offset
			})
	}
	return offset
}

func (x *Nesting) fastWriteField8(buf []byte) (offset int) {
	if x.SimpleStruct == nil {
		return offset
	}
	offset += fastpb.WriteMessage(buf[offset:], 8, x.GetSimpleStruct())
	return offset
}

func (x *Nesting) fastWriteField9(buf []byte) (offset int) {
	if x.MapI32I64 == nil {
		return offset
	}
	for k, v := range x.GetMapI32I64() {
		offset += fastpb.WriteMapEntry(buf[offset:], 9,
			func(buf []byte, numTagOrKey, numIdxOrVal int32) int {
				offset := 0
				offset += fastpb.WriteInt32(buf[offset:], numTagOrKey, k)
				offset += fastpb.WriteInt64(buf[offset:], numIdxOrVal, v)
				return offset
			})
	}
	return offset
}

func (x *Nesting) fastWriteField10(buf []byte) (offset int) {
	if len(x.ListString) == 0 {
		return offset
	}
	for i := range x.GetListString() {
		offset += fastpb.WriteString(buf[offset:], 10, x.GetListString()[i])
	}
	return offset
}

func (x *Nesting) fastWriteField11(buf []byte) (offset int) {
	if len(x.Binary) == 0 {
		return offset
	}
	offset += fastpb.WriteBytes(buf[offset:], 11, x.GetBinary())
	return offset
}

func (x *Nesting) fastWriteField12(buf []byte) (offset int) {
	if x.MapI64String == nil {
		return offset
	}
	for k, v := range x.GetMapI64String() {
		offset += fastpb.WriteMapEntry(buf[offset:], 12,
			func(buf []byte, numTagOrKey, numIdxOrVal int32) int {
				offset := 0
				offset += fastpb.WriteInt64(buf[offset:], numTagOrKey, k)
				offset += fastpb.WriteString(buf[offset:], numIdxOrVal, v)
				return offset
			})
	}
	return offset
}

func (x *Nesting) fastWriteField13(buf []byte) (offset int) {
	if len(x.ListI64) == 0 {
		return offset
	}
	offset += fastpb.WriteListPacked(buf[offset:], 13, len(x.GetListI64()),
		func(buf []byte, numTagOrKey, numIdxOrVal int32) int {
			offset := 0
			offset += fastpb.WriteInt64(buf[offset:], numTagOrKey, x.GetListI64()[numIdxOrVal])
			return offset
		})
	return offset
}

func (x *Nesting) fastWriteField14(buf []byte) (offset int) {
	if len(x.Byte) == 0 {
		return offset
	}
	offset += fastpb.WriteBytes(buf[offset:], 14, x.GetByte())
	return offset
}

func (x *Nesting) fastWriteField15(buf []byte) (offset int) {
	if x.MapStringSimple == nil {
		return offset
	}
	for k, v := range x.GetMapStringSimple() {
		offset += fastpb.WriteMapEntry(buf[offset:], 15,
			func(buf []byte, numTagOrKey, numIdxOrVal int32) int {
				offset := 0
				offset += fastpb.WriteString(buf[offset:], numTagOrKey, k)
				offset += fastpb.WriteMessage(buf[offset:], numIdxOrVal, v)
				return offset
			})
	}
	return offset
}

func (x *PartialNesting) FastWrite(buf []byte) (offset int) {
	if x == nil {
		return offset
	}
	offset += x.fastWriteField2(buf[offset:])
	offset += x.fastWriteField8(buf[offset:])
	offset += x.fastWriteField15(buf[offset:])
	return offset
}

func (x *PartialNesting) fastWriteField2(buf []byte) (offset int) {
	if x.ListSimple == nil {
		return offset
	}
	for i := range x.GetListSimple() {
		offset += fastpb.WriteMessage(buf[offset:], 2, x.GetListSimple()[i])
	}
	return offset
}

func (x *PartialNesting) fastWriteField8(buf []byte) (offset int) {
	if x.SimpleStruct == nil {
		return offset
	}
	offset += fastpb.WriteMessage(buf[offset:], 8, x.GetSimpleStruct())
	return offset
}

func (x *PartialNesting) fastWriteField15(buf []byte) (offset int) {
	if x.MapStringSimple == nil {
		return offset
	}
	for k, v := range x.GetMapStringSimple() {
		offset += fastpb.WriteMapEntry(buf[offset:], 15,
			func(buf []byte, numTagOrKey, numIdxOrVal int32) int {
				offset := 0
				offset += fastpb.WriteString(buf[offset:], numTagOrKey, k)
				offset += fastpb.WriteMessage(buf[offset:], numIdxOrVal, v)
				return offset
			})
	}
	return offset
}

func (x *Nesting2) FastWrite(buf []byte) (offset int) {
	if x == nil {
		return offset
	}
	offset += x.fastWriteField1(buf[offset:])
	offset += x.fastWriteField2(buf[offset:])
	offset += x.fastWriteField3(buf[offset:])
	offset += x.fastWriteField4(buf[offset:])
	offset += x.fastWriteField5(buf[offset:])
	offset += x.fastWriteField6(buf[offset:])
	offset += x.fastWriteField7(buf[offset:])
	offset += x.fastWriteField8(buf[offset:])
	offset += x.fastWriteField9(buf[offset:])
	offset += x.fastWriteField10(buf[offset:])
	offset += x.fastWriteField11(buf[offset:])
	return offset
}

func (x *Nesting2) fastWriteField1(buf []byte) (offset int) {
	if x.MapSimpleNesting == nil {
		return offset
	}
	for k, v := range x.GetMapSimpleNesting() {
		offset += fastpb.WriteMapEntry(buf[offset:], 1,
			func(buf []byte, numTagOrKey, numIdxOrVal int32) int {
				offset := 0
				offset += fastpb.WriteInt64(buf[offset:], numTagOrKey, k)
				offset += fastpb.WriteMessage(buf[offset:], numIdxOrVal, v)
				return offset
			})
	}
	return offset
}

func (x *Nesting2) fastWriteField2(buf []byte) (offset int) {
	if x.SimpleStruct == nil {
		return offset
	}
	offset += fastpb.WriteMessage(buf[offset:], 2, x.GetSimpleStruct())
	return offset
}

func (x *Nesting2) fastWriteField3(buf []byte) (offset int) {
	if len(x.Byte) == 0 {
		return offset
	}
	offset += fastpb.WriteBytes(buf[offset:], 3, x.GetByte())
	return offset
}

func (x *Nesting2) fastWriteField4(buf []byte) (offset int) {
	if x.Double == 0 {
		return offset
	}
	offset += fastpb.WriteDouble(buf[offset:], 4, x.GetDouble())
	return offset
}

func (x *Nesting2) fastWriteField5(buf []byte) (offset int) {
	if x.ListNesting == nil {
		return offset
	}
	for i := range x.GetListNesting() {
		offset += fastpb.WriteMessage(buf[offset:], 5, x.GetListNesting()[i])
	}
	return offset
}

func (x *Nesting2) fastWriteField6(buf []byte) (offset int) {
	if x.I64 == 0 {
		return offset
	}
	offset += fastpb.WriteInt64(buf[offset:], 6, x.GetI64())
	return offset
}

func (x *Nesting2) fastWriteField7(buf []byte) (offset int) {
	if x.NestingStruct == nil {
		return offset
	}
	offset += fastpb.WriteMessage(buf[offset:], 7, x.GetNestingStruct())
	return offset
}

func (x *Nesting2) fastWriteField8(buf []byte) (offset int) {
	if len(x.Binary) == 0 {
		return offset
	}
	offset += fastpb.WriteBytes(buf[offset:], 8, x.GetBinary())
	return offset
}

func (x *Nesting2) fastWriteField9(buf []byte) (offset int) {
	if x.String_ == "" {
		return offset
	}
	offset += fastpb.WriteString(buf[offset:], 9, x.GetString_())
	return offset
}

func (x *Nesting2) fastWriteField10(buf []byte) (offset int) {
	if x.SetNesting == nil {
		return offset
	}
	for i := range x.GetSetNesting() {
		offset += fastpb.WriteMessage(buf[offset:], 10, x.GetSetNesting()[i])
	}
	return offset
}

func (x *Nesting2) fastWriteField11(buf []byte) (offset int) {
	if x.I32 == 0 {
		return offset
	}
	offset += fastpb.WriteInt32(buf[offset:], 11, x.GetI32())
	return offset
}

func (x *Simple) Size() (n int) {
	if x == nil {
		return n
	}
	n += x.sizeField1()
	n += x.sizeField2()
	n += x.sizeField3()
	n += x.sizeField4()
	n += x.sizeField5()
	n += x.sizeField6()
	return n
}

func (x *Simple) sizeField1() (n int) {
	if len(x.ByteField) == 0 {
		return n
	}
	n += fastpb.SizeBytes(1, x.GetByteField())
	return n
}

func (x *Simple) sizeField2() (n int) {
	if x.I64Field == 0 {
		return n
	}
	n += fastpb.SizeInt64(2, x.GetI64Field())
	return n
}

func (x *Simple) sizeField3() (n int) {
	if x.DoubleField == 0 {
		return n
	}
	n += fastpb.SizeDouble(3, x.GetDoubleField())
	return n
}

func (x *Simple) sizeField4() (n int) {
	if x.I32Field == 0 {
		return n
	}
	n += fastpb.SizeInt32(4, x.GetI32Field())
	return n
}

func (x *Simple) sizeField5() (n int) {
	if x.StringField == "" {
		return n
	}
	n += fastpb.SizeString(5, x.GetStringField())
	return n
}

func (x *Simple) sizeField6() (n int) {
	if len(x.BinaryField) == 0 {
		return n
	}
	n += fastpb.SizeBytes(6, x.GetBinaryField())
	return n
}

func (x *PartialSimple) Size() (n int) {
	if x == nil {
		return n
	}
	n += x.sizeField1()
	n += x.sizeField3()
	n += x.sizeField6()
	return n
}

func (x *PartialSimple) sizeField1() (n int) {
	if len(x.ByteField) == 0 {
		return n
	}
	n += fastpb.SizeBytes(1, x.GetByteField())
	return n
}

func (x *PartialSimple) sizeField3() (n int) {
	if x.DoubleField == 0 {
		return n
	}
	n += fastpb.SizeDouble(3, x.GetDoubleField())
	return n
}

func (x *PartialSimple) sizeField6() (n int) {
	if len(x.BinaryField) == 0 {
		return n
	}
	n += fastpb.SizeBytes(6, x.GetBinaryField())
	return n
}

func (x *Nesting) Size() (n int) {
	if x == nil {
		return n
	}
	n += x.sizeField1()
	n += x.sizeField2()
	n += x.sizeField3()
	n += x.sizeField4()
	n += x.sizeField5()
	n += x.sizeField6()
	n += x.sizeField7()
	n += x.sizeField8()
	n += x.sizeField9()
	n += x.sizeField10()
	n += x.sizeField11()
	n += x.sizeField12()
	n += x.sizeField13()
	n += x.sizeField14()
	n += x.sizeField15()
	return n
}

func (x *Nesting) sizeField1() (n int) {
	if x.String_ == "" {
		return n
	}
	n += fastpb.SizeString(1, x.GetString_())
	return n
}

func (x *Nesting) sizeField2() (n int) {
	if x.ListSimple == nil {
		return n
	}
	for i := range x.GetListSimple() {
		n += fastpb.SizeMessage(2, x.GetListSimple()[i])
	}
	return n
}

func (x *Nesting) sizeField3() (n int) {
	if x.Double == 0 {
		return n
	}
	n += fastpb.SizeDouble(3, x.GetDouble())
	return n
}

func (x *Nesting) sizeField4() (n int) {
	if x.I32 == 0 {
		return n
	}
	n += fastpb.SizeInt32(4, x.GetI32())
	return n
}

func (x *Nesting) sizeField5() (n int) {
	if len(x.ListI32) == 0 {
		return n
	}
	n += fastpb.SizeListPacked(5, len(x.GetListI32()),
		func(numTagOrKey, numIdxOrVal int32) int {
			n := 0
			n += fastpb.SizeInt32(numTagOrKey, x.GetListI32()[numIdxOrVal])
			return n
		})
	return n
}

func (x *Nesting) sizeField6() (n int) {
	if x.I64 == 0 {
		return n
	}
	n += fastpb.SizeInt64(6, x.GetI64())
	return n
}

func (x *Nesting) sizeField7() (n int) {
	if x.MapStringString == nil {
		return n
	}
	for k, v := range x.GetMapStringString() {
		n += fastpb.SizeMapEntry(7,
			func(numTagOrKey, numIdxOrVal int32) int {
				n := 0
				n += fastpb.SizeString(numTagOrKey, k)
				n += fastpb.SizeString(numIdxOrVal, v)
				return n
			})
	}
	return n
}

func (x *Nesting) sizeField8() (n int) {
	if x.SimpleStruct == nil {
		return n
	}
	n += fastpb.SizeMessage(8, x.GetSimpleStruct())
	return n
}

func (x *Nesting) sizeField9() (n int) {
	if x.MapI32I64 == nil {
		return n
	}
	for k, v := range x.GetMapI32I64() {
		n += fastpb.SizeMapEntry(9,
			func(numTagOrKey, numIdxOrVal int32) int {
				n := 0
				n += fastpb.SizeInt32(numTagOrKey, k)
				n += fastpb.SizeInt64(numIdxOrVal, v)
				return n
			})
	}
	return n
}

func (x *Nesting) sizeField10() (n int) {
	if len(x.ListString) == 0 {
		return n
	}
	for i := range x.GetListString() {
		n += fastpb.SizeString(10, x.GetListString()[i])
	}
	return n
}

func (x *Nesting) sizeField11() (n int) {
	if len(x.Binary) == 0 {
		return n
	}
	n += fastpb.SizeBytes(11, x.GetBinary())
	return n
}

func (x *Nesting) sizeField12() (n int) {
	if x.MapI64String == nil {
		return n
	}
	for k, v := range x.GetMapI64String() {
		n += fastpb.SizeMapEntry(12,
			func(numTagOrKey, numIdxOrVal int32) int {
				n := 0
				n += fastpb.SizeInt64(numTagOrKey, k)
				n += fastpb.SizeString(numIdxOrVal, v)
				return n
			})
	}
	return n
}

func (x *Nesting) sizeField13() (n int) {
	if len(x.ListI64) == 0 {
		return n
	}
	n += fastpb.SizeListPacked(13, len(x.GetListI64()),
		func(numTagOrKey, numIdxOrVal int32) int {
			n := 0
			n += fastpb.SizeInt64(numTagOrKey, x.GetListI64()[numIdxOrVal])
			return n
		})
	return n
}

func (x *Nesting) sizeField14() (n int) {
	if len(x.Byte) == 0 {
		return n
	}
	n += fastpb.SizeBytes(14, x.GetByte())
	return n
}

func (x *Nesting) sizeField15() (n int) {
	if x.MapStringSimple == nil {
		return n
	}
	for k, v := range x.GetMapStringSimple() {
		n += fastpb.SizeMapEntry(15,
			func(numTagOrKey, numIdxOrVal int32) int {
				n := 0
				n += fastpb.SizeString(numTagOrKey, k)
				n += fastpb.SizeMessage(numIdxOrVal, v)
				return n
			})
	}
	return n
}

func (x *PartialNesting) Size() (n int) {
	if x == nil {
		return n
	}
	n += x.sizeField2()
	n += x.sizeField8()
	n += x.sizeField15()
	return n
}

func (x *PartialNesting) sizeField2() (n int) {
	if x.ListSimple == nil {
		return n
	}
	for i := range x.GetListSimple() {
		n += fastpb.SizeMessage(2, x.GetListSimple()[i])
	}
	return n
}

func (x *PartialNesting) sizeField8() (n int) {
	if x.SimpleStruct == nil {
		return n
	}
	n += fastpb.SizeMessage(8, x.GetSimpleStruct())
	return n
}

func (x *PartialNesting) sizeField15() (n int) {
	if x.MapStringSimple == nil {
		return n
	}
	for k, v := range x.GetMapStringSimple() {
		n += fastpb.SizeMapEntry(15,
			func(numTagOrKey, numIdxOrVal int32) int {
				n := 0
				n += fastpb.SizeString(numTagOrKey, k)
				n += fastpb.SizeMessage(numIdxOrVal, v)
				return n
			})
	}
	return n
}

func (x *Nesting2) Size() (n int) {
	if x == nil {
		return n
	}
	n += x.sizeField1()
	n += x.sizeField2()
	n += x.sizeField3()
	n += x.sizeField4()
	n += x.sizeField5()
	n += x.sizeField6()
	n += x.sizeField7()
	n += x.sizeField8()
	n += x.sizeField9()
	n += x.sizeField10()
	n += x.sizeField11()
	return n
}

func (x *Nesting2) sizeField1() (n int) {
	if x.MapSimpleNesting == nil {
		return n
	}
	for k, v := range x.GetMapSimpleNesting() {
		n += fastpb.SizeMapEntry(1,
			func(numTagOrKey, numIdxOrVal int32) int {
				n := 0
				n += fastpb.SizeInt64(numTagOrKey, k)
				n += fastpb.SizeMessage(numIdxOrVal, v)
				return n
			})
	}
	return n
}

func (x *Nesting2) sizeField2() (n int) {
	if x.SimpleStruct == nil {
		return n
	}
	n += fastpb.SizeMessage(2, x.GetSimpleStruct())
	return n
}

func (x *Nesting2) sizeField3() (n int) {
	if len(x.Byte) == 0 {
		return n
	}
	n += fastpb.SizeBytes(3, x.GetByte())
	return n
}

func (x *Nesting2) sizeField4() (n int) {
	if x.Double == 0 {
		return n
	}
	n += fastpb.SizeDouble(4, x.GetDouble())
	return n
}

func (x *Nesting2) sizeField5() (n int) {
	if x.ListNesting == nil {
		return n
	}
	for i := range x.GetListNesting() {
		n += fastpb.SizeMessage(5, x.GetListNesting()[i])
	}
	return n
}

func (x *Nesting2) sizeField6() (n int) {
	if x.I64 == 0 {
		return n
	}
	n += fastpb.SizeInt64(6, x.GetI64())
	return n
}

func (x *Nesting2) sizeField7() (n int) {
	if x.NestingStruct == nil {
		return n
	}
	n += fastpb.SizeMessage(7, x.GetNestingStruct())
	return n
}

func (x *Nesting2) sizeField8() (n int) {
	if len(x.Binary) == 0 {
		return n
	}
	n += fastpb.SizeBytes(8, x.GetBinary())
	return n
}

func (x *Nesting2) sizeField9() (n int) {
	if x.String_ == "" {
		return n
	}
	n += fastpb.SizeString(9, x.GetString_())
	return n
}

func (x *Nesting2) sizeField10() (n int) {
	if x.SetNesting == nil {
		return n
	}
	for i := range x.GetSetNesting() {
		n += fastpb.SizeMessage(10, x.GetSetNesting()[i])
	}
	return n
}

func (x *Nesting2) sizeField11() (n int) {
	if x.I32 == 0 {
		return n
	}
	n += fastpb.SizeInt32(11, x.GetI32())
	return n
}

var fieldIDToName_Simple = map[int32]string{
	1: "ByteField",
	2: "I64Field",
	3: "DoubleField",
	4: "I32Field",
	5: "StringField",
	6: "BinaryField",
}

var fieldIDToName_PartialSimple = map[int32]string{
	1: "ByteField",
	3: "DoubleField",
	6: "BinaryField",
}

var fieldIDToName_Nesting = map[int32]string{
	1:  "String_",
	2:  "ListSimple",
	3:  "Double",
	4:  "I32",
	5:  "ListI32",
	6:  "I64",
	7:  "MapStringString",
	8:  "SimpleStruct",
	9:  "MapI32I64",
	10: "ListString",
	11: "Binary",
	12: "MapI64String",
	13: "ListI64",
	14: "Byte",
	15: "MapStringSimple",
}

var fieldIDToName_PartialNesting = map[int32]string{
	2:  "ListSimple",
	8:  "SimpleStruct",
	15: "MapStringSimple",
}

var fieldIDToName_Nesting2 = map[int32]string{
	1:  "MapSimpleNesting",
	2:  "SimpleStruct",
	3:  "Byte",
	4:  "Double",
	5:  "ListNesting",
	6:  "I64",
	7:  "NestingStruct",
	8:  "Binary",
	9:  "String_",
	10: "SetNesting",
	11: "I32",
}
