/**
 * Copyright 2023 CloudWeGo Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package base

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	"github.com/apache/thrift/lib/go/thrift"

	"github.com/cloudwego/kitex/pkg/protocol/bthrift"
)

// unused protection
var (
	_ = fmt.Formatter(nil)
	_ = (*bytes.Buffer)(nil)
	_ = (*strings.Builder)(nil)
	_ = reflect.Type(nil)
	_ = thrift.TProtocol(nil)
	_ = bthrift.BinaryWriter(nil)
)

func (p *TrafficEnv) FastRead(buf []byte) (int, error) {
	var err error
	var offset int
	var l int
	var fieldTypeId thrift.TType
	var fieldId int16
	_, l, err = bthrift.Binary.ReadStructBegin(buf)
	offset += l
	if err != nil {
		goto ReadStructBeginError
	}

	for {
		_, fieldTypeId, fieldId, l, err = bthrift.Binary.ReadFieldBegin(buf[offset:])
		offset += l
		if err != nil {
			goto ReadFieldBeginError
		}
		if fieldTypeId == thrift.STOP {
			break
		}
		switch fieldId {
		case 1:
			if fieldTypeId == thrift.BOOL {
				l, err = p.FastReadField1(buf[offset:])
				offset += l
				if err != nil {
					goto ReadFieldError
				}
			} else {
				l, err = bthrift.Binary.Skip(buf[offset:], fieldTypeId)
				offset += l
				if err != nil {
					goto SkipFieldError
				}
			}
		case 2:
			if fieldTypeId == thrift.STRING {
				l, err = p.FastReadField2(buf[offset:])
				offset += l
				if err != nil {
					goto ReadFieldError
				}
			} else {
				l, err = bthrift.Binary.Skip(buf[offset:], fieldTypeId)
				offset += l
				if err != nil {
					goto SkipFieldError
				}
			}
		default:
			l, err = bthrift.Binary.Skip(buf[offset:], fieldTypeId)
			offset += l
			if err != nil {
				goto SkipFieldError
			}
		}

		l, err = bthrift.Binary.ReadFieldEnd(buf[offset:])
		offset += l
		if err != nil {
			goto ReadFieldEndError
		}
	}
	l, err = bthrift.Binary.ReadStructEnd(buf[offset:])
	offset += l
	if err != nil {
		goto ReadStructEndError
	}

	return offset, nil
ReadStructBeginError:
	return offset, thrift.PrependError(fmt.Sprintf("%T read struct begin error: ", p), err)
ReadFieldBeginError:
	return offset, thrift.PrependError(fmt.Sprintf("%T read field %d begin error: ", p, fieldId), err)
ReadFieldError:
	return offset, thrift.PrependError(fmt.Sprintf("%T read field %d '%s' error: ", p, fieldId, fieldIDToName_TrafficEnv[fieldId]), err)
SkipFieldError:
	return offset, thrift.PrependError(fmt.Sprintf("%T field %d skip type %d error: ", p, fieldId, fieldTypeId), err)
ReadFieldEndError:
	return offset, thrift.PrependError(fmt.Sprintf("%T read field end error", p), err)
ReadStructEndError:
	return offset, thrift.PrependError(fmt.Sprintf("%T read struct end error: ", p), err)
}

func (p *TrafficEnv) FastReadField1(buf []byte) (int, error) {
	offset := 0

	if v, l, err := bthrift.Binary.ReadBool(buf[offset:]); err != nil {
		return offset, err
	} else {
		offset += l

		p.Open = v

	}
	return offset, nil
}

func (p *TrafficEnv) FastReadField2(buf []byte) (int, error) {
	offset := 0

	if v, l, err := bthrift.Binary.ReadString(buf[offset:]); err != nil {
		return offset, err
	} else {
		offset += l

		p.Env = v

	}
	return offset, nil
}

// for compatibility
func (p *TrafficEnv) FastWrite(buf []byte) int {
	offset := 0
	offset += bthrift.Binary.WriteStructBegin(buf[offset:], "TrafficEnv")
	if p != nil {
		offset += p.fastWriteField1(buf[offset:])
		offset += p.fastWriteField2(buf[offset:])
	}
	offset += bthrift.Binary.WriteFieldStop(buf[offset:])
	offset += bthrift.Binary.WriteStructEnd(buf[offset:])
	return offset
}

func (p *TrafficEnv) BLength() int {
	l := 0
	l += bthrift.Binary.StructBeginLength("TrafficEnv")
	if p != nil {
		l += p.field1Length()
		l += p.field2Length()
	}
	l += bthrift.Binary.FieldStopLength()
	l += bthrift.Binary.StructEndLength()
	return l
}

func (p *TrafficEnv) fastWriteField1(buf []byte) int {
	offset := 0
	offset += bthrift.Binary.WriteFieldBegin(buf[offset:], "Open", thrift.BOOL, 1)
	offset += bthrift.Binary.WriteBool(buf[offset:], p.Open)

	offset += bthrift.Binary.WriteFieldEnd(buf[offset:])
	return offset
}

func (p *TrafficEnv) fastWriteField2(buf []byte) int {
	offset := 0
	offset += bthrift.Binary.WriteFieldBegin(buf[offset:], "Env", thrift.STRING, 2)
	offset += bthrift.Binary.WriteString(buf[offset:], p.Env)

	offset += bthrift.Binary.WriteFieldEnd(buf[offset:])
	return offset
}

func (p *TrafficEnv) field1Length() int {
	l := 0
	l += bthrift.Binary.FieldBeginLength("Open", thrift.BOOL, 1)
	l += bthrift.Binary.BoolLength(p.Open)

	l += bthrift.Binary.FieldEndLength()
	return l
}

func (p *TrafficEnv) field2Length() int {
	l := 0
	l += bthrift.Binary.FieldBeginLength("Env", thrift.STRING, 2)
	l += bthrift.Binary.StringLength(p.Env)

	l += bthrift.Binary.FieldEndLength()
	return l
}

func (p *Base) FastRead(buf []byte) (int, error) {
	var err error
	var offset int
	var l int
	var fieldTypeId thrift.TType
	var fieldId int16
	_, l, err = bthrift.Binary.ReadStructBegin(buf)
	offset += l
	if err != nil {
		goto ReadStructBeginError
	}

	for {
		_, fieldTypeId, fieldId, l, err = bthrift.Binary.ReadFieldBegin(buf[offset:])
		offset += l
		if err != nil {
			goto ReadFieldBeginError
		}
		if fieldTypeId == thrift.STOP {
			break
		}
		switch fieldId {
		case 1:
			if fieldTypeId == thrift.STRING {
				l, err = p.FastReadField1(buf[offset:])
				offset += l
				if err != nil {
					goto ReadFieldError
				}
			} else {
				l, err = bthrift.Binary.Skip(buf[offset:], fieldTypeId)
				offset += l
				if err != nil {
					goto SkipFieldError
				}
			}
		case 2:
			if fieldTypeId == thrift.STRING {
				l, err = p.FastReadField2(buf[offset:])
				offset += l
				if err != nil {
					goto ReadFieldError
				}
			} else {
				l, err = bthrift.Binary.Skip(buf[offset:], fieldTypeId)
				offset += l
				if err != nil {
					goto SkipFieldError
				}
			}
		case 3:
			if fieldTypeId == thrift.STRING {
				l, err = p.FastReadField3(buf[offset:])
				offset += l
				if err != nil {
					goto ReadFieldError
				}
			} else {
				l, err = bthrift.Binary.Skip(buf[offset:], fieldTypeId)
				offset += l
				if err != nil {
					goto SkipFieldError
				}
			}
		case 4:
			if fieldTypeId == thrift.STRING {
				l, err = p.FastReadField4(buf[offset:])
				offset += l
				if err != nil {
					goto ReadFieldError
				}
			} else {
				l, err = bthrift.Binary.Skip(buf[offset:], fieldTypeId)
				offset += l
				if err != nil {
					goto SkipFieldError
				}
			}
		case 5:
			if fieldTypeId == thrift.STRUCT {
				l, err = p.FastReadField5(buf[offset:])
				offset += l
				if err != nil {
					goto ReadFieldError
				}
			} else {
				l, err = bthrift.Binary.Skip(buf[offset:], fieldTypeId)
				offset += l
				if err != nil {
					goto SkipFieldError
				}
			}
		case 6:
			if fieldTypeId == thrift.MAP {
				l, err = p.FastReadField6(buf[offset:])
				offset += l
				if err != nil {
					goto ReadFieldError
				}
			} else {
				l, err = bthrift.Binary.Skip(buf[offset:], fieldTypeId)
				offset += l
				if err != nil {
					goto SkipFieldError
				}
			}
		default:
			l, err = bthrift.Binary.Skip(buf[offset:], fieldTypeId)
			offset += l
			if err != nil {
				goto SkipFieldError
			}
		}

		l, err = bthrift.Binary.ReadFieldEnd(buf[offset:])
		offset += l
		if err != nil {
			goto ReadFieldEndError
		}
	}
	l, err = bthrift.Binary.ReadStructEnd(buf[offset:])
	offset += l
	if err != nil {
		goto ReadStructEndError
	}

	return offset, nil
ReadStructBeginError:
	return offset, thrift.PrependError(fmt.Sprintf("%T read struct begin error: ", p), err)
ReadFieldBeginError:
	return offset, thrift.PrependError(fmt.Sprintf("%T read field %d begin error: ", p, fieldId), err)
ReadFieldError:
	return offset, thrift.PrependError(fmt.Sprintf("%T read field %d '%s' error: ", p, fieldId, fieldIDToName_Base[fieldId]), err)
SkipFieldError:
	return offset, thrift.PrependError(fmt.Sprintf("%T field %d skip type %d error: ", p, fieldId, fieldTypeId), err)
ReadFieldEndError:
	return offset, thrift.PrependError(fmt.Sprintf("%T read field end error", p), err)
ReadStructEndError:
	return offset, thrift.PrependError(fmt.Sprintf("%T read struct end error: ", p), err)
}

func (p *Base) FastReadField1(buf []byte) (int, error) {
	offset := 0

	if v, l, err := bthrift.Binary.ReadString(buf[offset:]); err != nil {
		return offset, err
	} else {
		offset += l

		p.LogID = v

	}
	return offset, nil
}

func (p *Base) FastReadField2(buf []byte) (int, error) {
	offset := 0

	if v, l, err := bthrift.Binary.ReadString(buf[offset:]); err != nil {
		return offset, err
	} else {
		offset += l

		p.Caller = v

	}
	return offset, nil
}

func (p *Base) FastReadField3(buf []byte) (int, error) {
	offset := 0

	if v, l, err := bthrift.Binary.ReadString(buf[offset:]); err != nil {
		return offset, err
	} else {
		offset += l

		p.Addr = v

	}
	return offset, nil
}

func (p *Base) FastReadField4(buf []byte) (int, error) {
	offset := 0

	if v, l, err := bthrift.Binary.ReadString(buf[offset:]); err != nil {
		return offset, err
	} else {
		offset += l

		p.Client = v

	}
	return offset, nil
}

func (p *Base) FastReadField5(buf []byte) (int, error) {
	offset := 0
	p.TrafficEnv = NewTrafficEnv()
	if l, err := p.TrafficEnv.FastRead(buf[offset:]); err != nil {
		return offset, err
	} else {
		offset += l
	}
	return offset, nil
}

func (p *Base) FastReadField6(buf []byte) (int, error) {
	offset := 0

	_, _, size, l, err := bthrift.Binary.ReadMapBegin(buf[offset:])
	offset += l
	if err != nil {
		return offset, err
	}
	p.Extra = make(map[string]string, size)
	for i := 0; i < size; i++ {
		var _key string
		if v, l, err := bthrift.Binary.ReadString(buf[offset:]); err != nil {
			return offset, err
		} else {
			offset += l

			_key = v

		}

		var _val string
		if v, l, err := bthrift.Binary.ReadString(buf[offset:]); err != nil {
			return offset, err
		} else {
			offset += l

			_val = v

		}

		p.Extra[_key] = _val
	}
	if l, err := bthrift.Binary.ReadMapEnd(buf[offset:]); err != nil {
		return offset, err
	} else {
		offset += l
	}
	return offset, nil
}

// for compatibility
func (p *Base) FastWrite(buf []byte) int {
	offset := 0
	offset += bthrift.Binary.WriteStructBegin(buf[offset:], "Base")
	if p != nil {
		offset += p.fastWriteField1(buf[offset:])
		offset += p.fastWriteField2(buf[offset:])
		offset += p.fastWriteField3(buf[offset:])
		offset += p.fastWriteField4(buf[offset:])
		offset += p.fastWriteField5(buf[offset:])
		offset += p.fastWriteField6(buf[offset:])
	}
	offset += bthrift.Binary.WriteFieldStop(buf[offset:])
	offset += bthrift.Binary.WriteStructEnd(buf[offset:])
	return offset
}

func (p *Base) BLength() int {
	l := 0
	l += bthrift.Binary.StructBeginLength("Base")
	if p != nil {
		l += p.field1Length()
		l += p.field2Length()
		l += p.field3Length()
		l += p.field4Length()
		l += p.field5Length()
		l += p.field6Length()
	}
	l += bthrift.Binary.FieldStopLength()
	l += bthrift.Binary.StructEndLength()
	return l
}

func (p *Base) fastWriteField1(buf []byte) int {
	offset := 0
	offset += bthrift.Binary.WriteFieldBegin(buf[offset:], "LogID", thrift.STRING, 1)
	offset += bthrift.Binary.WriteString(buf[offset:], p.LogID)

	offset += bthrift.Binary.WriteFieldEnd(buf[offset:])
	return offset
}

func (p *Base) fastWriteField2(buf []byte) int {
	offset := 0
	offset += bthrift.Binary.WriteFieldBegin(buf[offset:], "Caller", thrift.STRING, 2)
	offset += bthrift.Binary.WriteString(buf[offset:], p.Caller)

	offset += bthrift.Binary.WriteFieldEnd(buf[offset:])
	return offset
}

func (p *Base) fastWriteField3(buf []byte) int {
	offset := 0
	offset += bthrift.Binary.WriteFieldBegin(buf[offset:], "Addr", thrift.STRING, 3)
	offset += bthrift.Binary.WriteString(buf[offset:], p.Addr)

	offset += bthrift.Binary.WriteFieldEnd(buf[offset:])
	return offset
}

func (p *Base) fastWriteField4(buf []byte) int {
	offset := 0
	offset += bthrift.Binary.WriteFieldBegin(buf[offset:], "Client", thrift.STRING, 4)
	offset += bthrift.Binary.WriteString(buf[offset:], p.Client)

	offset += bthrift.Binary.WriteFieldEnd(buf[offset:])
	return offset
}

func (p *Base) fastWriteField5(buf []byte) int {
	offset := 0
	if p.IsSetTrafficEnv() {
		offset += bthrift.Binary.WriteFieldBegin(buf[offset:], "TrafficEnv", thrift.STRUCT, 5)
		offset += p.TrafficEnv.FastWrite(buf[offset:])
		offset += bthrift.Binary.WriteFieldEnd(buf[offset:])
	}
	return offset
}

func (p *Base) fastWriteField6(buf []byte) int {
	offset := 0
	if p.IsSetExtra() {
		offset += bthrift.Binary.WriteFieldBegin(buf[offset:], "Extra", thrift.MAP, 6)
		offset += bthrift.Binary.WriteMapBegin(buf[offset:], thrift.STRING, thrift.STRING, len(p.Extra))
		for k, v := range p.Extra {

			offset += bthrift.Binary.WriteString(buf[offset:], k)

			offset += bthrift.Binary.WriteString(buf[offset:], v)

		}
		offset += bthrift.Binary.WriteMapEnd(buf[offset:])
		offset += bthrift.Binary.WriteFieldEnd(buf[offset:])
	}
	return offset
}

func (p *Base) field1Length() int {
	l := 0
	l += bthrift.Binary.FieldBeginLength("LogID", thrift.STRING, 1)
	l += bthrift.Binary.StringLength(p.LogID)

	l += bthrift.Binary.FieldEndLength()
	return l
}

func (p *Base) field2Length() int {
	l := 0
	l += bthrift.Binary.FieldBeginLength("Caller", thrift.STRING, 2)
	l += bthrift.Binary.StringLength(p.Caller)

	l += bthrift.Binary.FieldEndLength()
	return l
}

func (p *Base) field3Length() int {
	l := 0
	l += bthrift.Binary.FieldBeginLength("Addr", thrift.STRING, 3)
	l += bthrift.Binary.StringLength(p.Addr)

	l += bthrift.Binary.FieldEndLength()
	return l
}

func (p *Base) field4Length() int {
	l := 0
	l += bthrift.Binary.FieldBeginLength("Client", thrift.STRING, 4)
	l += bthrift.Binary.StringLength(p.Client)

	l += bthrift.Binary.FieldEndLength()
	return l
}

func (p *Base) field5Length() int {
	l := 0
	if p.IsSetTrafficEnv() {
		l += bthrift.Binary.FieldBeginLength("TrafficEnv", thrift.STRUCT, 5)
		l += p.TrafficEnv.BLength()
		l += bthrift.Binary.FieldEndLength()
	}
	return l
}

func (p *Base) field6Length() int {
	l := 0
	if p.IsSetExtra() {
		l += bthrift.Binary.FieldBeginLength("Extra", thrift.MAP, 6)
		l += bthrift.Binary.MapBeginLength(thrift.STRING, thrift.STRING, len(p.Extra))
		for k, v := range p.Extra {

			l += bthrift.Binary.StringLength(k)

			l += bthrift.Binary.StringLength(v)

		}
		l += bthrift.Binary.MapEndLength()
		l += bthrift.Binary.FieldEndLength()
	}
	return l
}

func (p *BaseResp) FastRead(buf []byte) (int, error) {
	var err error
	var offset int
	var l int
	var fieldTypeId thrift.TType
	var fieldId int16
	_, l, err = bthrift.Binary.ReadStructBegin(buf)
	offset += l
	if err != nil {
		goto ReadStructBeginError
	}

	for {
		_, fieldTypeId, fieldId, l, err = bthrift.Binary.ReadFieldBegin(buf[offset:])
		offset += l
		if err != nil {
			goto ReadFieldBeginError
		}
		if fieldTypeId == thrift.STOP {
			break
		}
		switch fieldId {
		case 1:
			if fieldTypeId == thrift.STRING {
				l, err = p.FastReadField1(buf[offset:])
				offset += l
				if err != nil {
					goto ReadFieldError
				}
			} else {
				l, err = bthrift.Binary.Skip(buf[offset:], fieldTypeId)
				offset += l
				if err != nil {
					goto SkipFieldError
				}
			}
		case 2:
			if fieldTypeId == thrift.I32 {
				l, err = p.FastReadField2(buf[offset:])
				offset += l
				if err != nil {
					goto ReadFieldError
				}
			} else {
				l, err = bthrift.Binary.Skip(buf[offset:], fieldTypeId)
				offset += l
				if err != nil {
					goto SkipFieldError
				}
			}
		case 3:
			if fieldTypeId == thrift.MAP {
				l, err = p.FastReadField3(buf[offset:])
				offset += l
				if err != nil {
					goto ReadFieldError
				}
			} else {
				l, err = bthrift.Binary.Skip(buf[offset:], fieldTypeId)
				offset += l
				if err != nil {
					goto SkipFieldError
				}
			}
		default:
			l, err = bthrift.Binary.Skip(buf[offset:], fieldTypeId)
			offset += l
			if err != nil {
				goto SkipFieldError
			}
		}

		l, err = bthrift.Binary.ReadFieldEnd(buf[offset:])
		offset += l
		if err != nil {
			goto ReadFieldEndError
		}
	}
	l, err = bthrift.Binary.ReadStructEnd(buf[offset:])
	offset += l
	if err != nil {
		goto ReadStructEndError
	}

	return offset, nil
ReadStructBeginError:
	return offset, thrift.PrependError(fmt.Sprintf("%T read struct begin error: ", p), err)
ReadFieldBeginError:
	return offset, thrift.PrependError(fmt.Sprintf("%T read field %d begin error: ", p, fieldId), err)
ReadFieldError:
	return offset, thrift.PrependError(fmt.Sprintf("%T read field %d '%s' error: ", p, fieldId, fieldIDToName_BaseResp[fieldId]), err)
SkipFieldError:
	return offset, thrift.PrependError(fmt.Sprintf("%T field %d skip type %d error: ", p, fieldId, fieldTypeId), err)
ReadFieldEndError:
	return offset, thrift.PrependError(fmt.Sprintf("%T read field end error", p), err)
ReadStructEndError:
	return offset, thrift.PrependError(fmt.Sprintf("%T read struct end error: ", p), err)
}

func (p *BaseResp) FastReadField1(buf []byte) (int, error) {
	offset := 0

	if v, l, err := bthrift.Binary.ReadString(buf[offset:]); err != nil {
		return offset, err
	} else {
		offset += l

		p.StatusMessage = v

	}
	return offset, nil
}

func (p *BaseResp) FastReadField2(buf []byte) (int, error) {
	offset := 0

	if v, l, err := bthrift.Binary.ReadI32(buf[offset:]); err != nil {
		return offset, err
	} else {
		offset += l

		p.StatusCode = v

	}
	return offset, nil
}

func (p *BaseResp) FastReadField3(buf []byte) (int, error) {
	offset := 0

	_, _, size, l, err := bthrift.Binary.ReadMapBegin(buf[offset:])
	offset += l
	if err != nil {
		return offset, err
	}
	p.Extra = make(map[string]string, size)
	for i := 0; i < size; i++ {
		var _key string
		if v, l, err := bthrift.Binary.ReadString(buf[offset:]); err != nil {
			return offset, err
		} else {
			offset += l

			_key = v

		}

		var _val string
		if v, l, err := bthrift.Binary.ReadString(buf[offset:]); err != nil {
			return offset, err
		} else {
			offset += l

			_val = v

		}

		p.Extra[_key] = _val
	}
	if l, err := bthrift.Binary.ReadMapEnd(buf[offset:]); err != nil {
		return offset, err
	} else {
		offset += l
	}
	return offset, nil
}

// for compatibility
func (p *BaseResp) FastWrite(buf []byte) int {
	offset := 0
	offset += bthrift.Binary.WriteStructBegin(buf[offset:], "BaseResp")
	if p != nil {
		offset += p.fastWriteField2(buf[offset:])
		offset += p.fastWriteField1(buf[offset:])
		offset += p.fastWriteField3(buf[offset:])
	}
	offset += bthrift.Binary.WriteFieldStop(buf[offset:])
	offset += bthrift.Binary.WriteStructEnd(buf[offset:])
	return offset
}

func (p *BaseResp) BLength() int {
	l := 0
	l += bthrift.Binary.StructBeginLength("BaseResp")
	if p != nil {
		l += p.field1Length()
		l += p.field2Length()
		l += p.field3Length()
	}
	l += bthrift.Binary.FieldStopLength()
	l += bthrift.Binary.StructEndLength()
	return l
}

func (p *BaseResp) fastWriteField1(buf []byte) int {
	offset := 0
	offset += bthrift.Binary.WriteFieldBegin(buf[offset:], "StatusMessage", thrift.STRING, 1)
	offset += bthrift.Binary.WriteString(buf[offset:], p.StatusMessage)

	offset += bthrift.Binary.WriteFieldEnd(buf[offset:])
	return offset
}

func (p *BaseResp) fastWriteField2(buf []byte) int {
	offset := 0
	offset += bthrift.Binary.WriteFieldBegin(buf[offset:], "StatusCode", thrift.I32, 2)
	offset += bthrift.Binary.WriteI32(buf[offset:], p.StatusCode)

	offset += bthrift.Binary.WriteFieldEnd(buf[offset:])
	return offset
}

func (p *BaseResp) fastWriteField3(buf []byte) int {
	offset := 0
	if p.IsSetExtra() {
		offset += bthrift.Binary.WriteFieldBegin(buf[offset:], "Extra", thrift.MAP, 3)
		offset += bthrift.Binary.WriteMapBegin(buf[offset:], thrift.STRING, thrift.STRING, len(p.Extra))
		for k, v := range p.Extra {

			offset += bthrift.Binary.WriteString(buf[offset:], k)

			offset += bthrift.Binary.WriteString(buf[offset:], v)

		}
		offset += bthrift.Binary.WriteMapEnd(buf[offset:])
		offset += bthrift.Binary.WriteFieldEnd(buf[offset:])
	}
	return offset
}

func (p *BaseResp) field1Length() int {
	l := 0
	l += bthrift.Binary.FieldBeginLength("StatusMessage", thrift.STRING, 1)
	l += bthrift.Binary.StringLength(p.StatusMessage)

	l += bthrift.Binary.FieldEndLength()
	return l
}

func (p *BaseResp) field2Length() int {
	l := 0
	l += bthrift.Binary.FieldBeginLength("StatusCode", thrift.I32, 2)
	l += bthrift.Binary.I32Length(p.StatusCode)

	l += bthrift.Binary.FieldEndLength()
	return l
}

func (p *BaseResp) field3Length() int {
	l := 0
	if p.IsSetExtra() {
		l += bthrift.Binary.FieldBeginLength("Extra", thrift.MAP, 3)
		l += bthrift.Binary.MapBeginLength(thrift.STRING, thrift.STRING, len(p.Extra))
		for k, v := range p.Extra {

			l += bthrift.Binary.StringLength(k)

			l += bthrift.Binary.StringLength(v)

		}
		l += bthrift.Binary.MapEndLength()
		l += bthrift.Binary.FieldEndLength()
	}
	return l
}
