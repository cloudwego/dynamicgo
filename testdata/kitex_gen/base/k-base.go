// Code generated by thriftgo (0.3.15) (fastgo). DO NOT EDIT.
package base

import (
	"encoding/binary"
	"fmt"
	"unsafe"

	"github.com/cloudwego/gopkg/protocol/thrift"
)

func (p *TrafficEnv) BLength() int {
	if p == nil {
		return 1
	}
	off := 0

	// p.Open ID:1 thrift.BOOL
	off += 3
	off += 1

	// p.Env ID:2 thrift.STRING
	off += 3
	off += 4 + len(p.Env)
	return off + 1
}

func (p *TrafficEnv) FastWrite(b []byte) int { return p.FastWriteNocopy(b, nil) }

func (p *TrafficEnv) FastWriteNocopy(b []byte, w thrift.NocopyWriter) int {
	if p == nil {
		b[0] = 0
		return 1
	}
	off := 0

	// p.Open ID:1 thrift.BOOL
	b[off] = 2
	binary.BigEndian.PutUint16(b[off+1:], 1)
	off += 3
	b[off] = *((*byte)(unsafe.Pointer(&p.Open)))
	off++

	// p.Env ID:2 thrift.STRING
	b[off] = 11
	binary.BigEndian.PutUint16(b[off+1:], 2)
	off += 3
	off += thrift.Binary.WriteStringNocopy(b[off:], w, p.Env)

	b[off] = 0
	return off + 1
}

func (p *TrafficEnv) FastRead(b []byte) (off int, err error) {
	var ftyp thrift.TType
	var fid int16
	var l int
	x := thrift.BinaryProtocol{}
	for {
		ftyp, fid, l, err = x.ReadFieldBegin(b[off:])
		off += l
		if err != nil {
			goto ReadFieldBeginError
		}
		if ftyp == thrift.STOP {
			break
		}
		switch uint32(fid)<<8 | uint32(ftyp) {
		case 0x102: // p.Open ID:1 thrift.BOOL
			p.Open, l, err = x.ReadBool(b[off:])
			off += l
			if err != nil {
				goto ReadFieldError
			}
		case 0x20b: // p.Env ID:2 thrift.STRING
			p.Env, l, err = x.ReadString(b[off:])
			off += l
			if err != nil {
				goto ReadFieldError
			}
		default:
			l, err = x.Skip(b[off:], ftyp)
			off += l
			if err != nil {
				goto SkipFieldError
			}
		}
	}
	return
ReadFieldBeginError:
	return off, thrift.PrependError(fmt.Sprintf("%T read field begin error: ", p), err)
ReadFieldError:
	return off, thrift.PrependError(fmt.Sprintf("%T read field %d '%s' error: ", p, fid, fieldIDToName_TrafficEnv[fid]), err)
SkipFieldError:
	return off, thrift.PrependError(fmt.Sprintf("%T skip field %d type %d error: ", p, fid, ftyp), err)
}

func (p *Base) BLength() int {
	if p == nil {
		return 1
	}
	off := 0

	// p.LogID ID:1 thrift.STRING
	off += 3
	off += 4 + len(p.LogID)

	// p.Caller ID:2 thrift.STRING
	off += 3
	off += 4 + len(p.Caller)

	// p.Addr ID:3 thrift.STRING
	off += 3
	off += 4 + len(p.Addr)

	// p.Client ID:4 thrift.STRING
	off += 3
	off += 4 + len(p.Client)

	// p.TrafficEnv ID:5 thrift.STRUCT
	if p.TrafficEnv != nil {
		off += 3
		off += p.TrafficEnv.BLength()
	}

	// p.Extra ID:6 thrift.MAP
	if p.Extra != nil {
		off += 3
		off += 6
		for k, v := range p.Extra {
			off += 4 + len(k)
			off += 4 + len(v)
		}
	}
	return off + 1
}

func (p *Base) FastWrite(b []byte) int { return p.FastWriteNocopy(b, nil) }

func (p *Base) FastWriteNocopy(b []byte, w thrift.NocopyWriter) int {
	if p == nil {
		b[0] = 0
		return 1
	}
	off := 0

	// p.LogID ID:1 thrift.STRING
	b[off] = 11
	binary.BigEndian.PutUint16(b[off+1:], 1)
	off += 3
	off += thrift.Binary.WriteStringNocopy(b[off:], w, p.LogID)

	// p.Caller ID:2 thrift.STRING
	b[off] = 11
	binary.BigEndian.PutUint16(b[off+1:], 2)
	off += 3
	off += thrift.Binary.WriteStringNocopy(b[off:], w, p.Caller)

	// p.Addr ID:3 thrift.STRING
	b[off] = 11
	binary.BigEndian.PutUint16(b[off+1:], 3)
	off += 3
	off += thrift.Binary.WriteStringNocopy(b[off:], w, p.Addr)

	// p.Client ID:4 thrift.STRING
	b[off] = 11
	binary.BigEndian.PutUint16(b[off+1:], 4)
	off += 3
	off += thrift.Binary.WriteStringNocopy(b[off:], w, p.Client)

	// p.TrafficEnv ID:5 thrift.STRUCT
	if p.TrafficEnv != nil {
		b[off] = 12
		binary.BigEndian.PutUint16(b[off+1:], 5)
		off += 3
		off += p.TrafficEnv.FastWriteNocopy(b[off:], w)
	}

	// p.Extra ID:6 thrift.MAP
	if p.Extra != nil {
		b[off] = 13
		binary.BigEndian.PutUint16(b[off+1:], 6)
		off += 3
		b[off] = 11
		b[off+1] = 11
		binary.BigEndian.PutUint32(b[off+2:], uint32(len(p.Extra)))
		off += 6
		for k, v := range p.Extra {
			off += thrift.Binary.WriteStringNocopy(b[off:], w, k)
			off += thrift.Binary.WriteStringNocopy(b[off:], w, v)
		}
	}

	b[off] = 0
	return off + 1
}

func (p *Base) FastRead(b []byte) (off int, err error) {
	var ftyp thrift.TType
	var fid int16
	var l int
	x := thrift.BinaryProtocol{}
	for {
		ftyp, fid, l, err = x.ReadFieldBegin(b[off:])
		off += l
		if err != nil {
			goto ReadFieldBeginError
		}
		if ftyp == thrift.STOP {
			break
		}
		switch uint32(fid)<<8 | uint32(ftyp) {
		case 0x10b: // p.LogID ID:1 thrift.STRING
			p.LogID, l, err = x.ReadString(b[off:])
			off += l
			if err != nil {
				goto ReadFieldError
			}
		case 0x20b: // p.Caller ID:2 thrift.STRING
			p.Caller, l, err = x.ReadString(b[off:])
			off += l
			if err != nil {
				goto ReadFieldError
			}
		case 0x30b: // p.Addr ID:3 thrift.STRING
			p.Addr, l, err = x.ReadString(b[off:])
			off += l
			if err != nil {
				goto ReadFieldError
			}
		case 0x40b: // p.Client ID:4 thrift.STRING
			p.Client, l, err = x.ReadString(b[off:])
			off += l
			if err != nil {
				goto ReadFieldError
			}
		case 0x50c: // p.TrafficEnv ID:5 thrift.STRUCT
			p.TrafficEnv = NewTrafficEnv()
			l, err = p.TrafficEnv.FastRead(b[off:])
			off += l
			if err != nil {
				goto ReadFieldError
			}
		case 0x60d: // p.Extra ID:6 thrift.MAP
			var sz int
			_, _, sz, l, err = x.ReadMapBegin(b[off:])
			off += l
			if err != nil {
				goto ReadFieldError
			}
			p.Extra = make(map[string]string, sz)
			for i := 0; i < sz; i++ {
				var k string
				var v string
				k, l, err = x.ReadString(b[off:])
				off += l
				if err != nil {
					goto ReadFieldError
				}
				v, l, err = x.ReadString(b[off:])
				off += l
				if err != nil {
					goto ReadFieldError
				}
				p.Extra[k] = v
			}
		default:
			l, err = x.Skip(b[off:], ftyp)
			off += l
			if err != nil {
				goto SkipFieldError
			}
		}
	}
	return
ReadFieldBeginError:
	return off, thrift.PrependError(fmt.Sprintf("%T read field begin error: ", p), err)
ReadFieldError:
	return off, thrift.PrependError(fmt.Sprintf("%T read field %d '%s' error: ", p, fid, fieldIDToName_Base[fid]), err)
SkipFieldError:
	return off, thrift.PrependError(fmt.Sprintf("%T skip field %d type %d error: ", p, fid, ftyp), err)
}

func (p *BaseResp) BLength() int {
	if p == nil {
		return 1
	}
	off := 0

	// p.StatusMessage ID:1 thrift.STRING
	off += 3
	off += 4 + len(p.StatusMessage)

	// p.StatusCode ID:2 thrift.I32
	off += 3
	off += 4

	// p.Extra ID:3 thrift.MAP
	if p.Extra != nil {
		off += 3
		off += 6
		for k, v := range p.Extra {
			off += 4 + len(k)
			off += 4 + len(v)
		}
	}
	return off + 1
}

func (p *BaseResp) FastWrite(b []byte) int { return p.FastWriteNocopy(b, nil) }

func (p *BaseResp) FastWriteNocopy(b []byte, w thrift.NocopyWriter) int {
	if p == nil {
		b[0] = 0
		return 1
	}
	off := 0

	// p.StatusMessage ID:1 thrift.STRING
	b[off] = 11
	binary.BigEndian.PutUint16(b[off+1:], 1)
	off += 3
	off += thrift.Binary.WriteStringNocopy(b[off:], w, p.StatusMessage)

	// p.StatusCode ID:2 thrift.I32
	b[off] = 8
	binary.BigEndian.PutUint16(b[off+1:], 2)
	off += 3
	binary.BigEndian.PutUint32(b[off:], uint32(p.StatusCode))
	off += 4

	// p.Extra ID:3 thrift.MAP
	if p.Extra != nil {
		b[off] = 13
		binary.BigEndian.PutUint16(b[off+1:], 3)
		off += 3
		b[off] = 11
		b[off+1] = 11
		binary.BigEndian.PutUint32(b[off+2:], uint32(len(p.Extra)))
		off += 6
		for k, v := range p.Extra {
			off += thrift.Binary.WriteStringNocopy(b[off:], w, k)
			off += thrift.Binary.WriteStringNocopy(b[off:], w, v)
		}
	}

	b[off] = 0
	return off + 1
}

func (p *BaseResp) FastRead(b []byte) (off int, err error) {
	var ftyp thrift.TType
	var fid int16
	var l int
	x := thrift.BinaryProtocol{}
	for {
		ftyp, fid, l, err = x.ReadFieldBegin(b[off:])
		off += l
		if err != nil {
			goto ReadFieldBeginError
		}
		if ftyp == thrift.STOP {
			break
		}
		switch uint32(fid)<<8 | uint32(ftyp) {
		case 0x10b: // p.StatusMessage ID:1 thrift.STRING
			p.StatusMessage, l, err = x.ReadString(b[off:])
			off += l
			if err != nil {
				goto ReadFieldError
			}
		case 0x208: // p.StatusCode ID:2 thrift.I32
			p.StatusCode, l, err = x.ReadI32(b[off:])
			off += l
			if err != nil {
				goto ReadFieldError
			}
		case 0x30d: // p.Extra ID:3 thrift.MAP
			var sz int
			_, _, sz, l, err = x.ReadMapBegin(b[off:])
			off += l
			if err != nil {
				goto ReadFieldError
			}
			p.Extra = make(map[string]string, sz)
			for i := 0; i < sz; i++ {
				var k string
				var v string
				k, l, err = x.ReadString(b[off:])
				off += l
				if err != nil {
					goto ReadFieldError
				}
				v, l, err = x.ReadString(b[off:])
				off += l
				if err != nil {
					goto ReadFieldError
				}
				p.Extra[k] = v
			}
		default:
			l, err = x.Skip(b[off:], ftyp)
			off += l
			if err != nil {
				goto SkipFieldError
			}
		}
	}
	return
ReadFieldBeginError:
	return off, thrift.PrependError(fmt.Sprintf("%T read field begin error: ", p), err)
ReadFieldError:
	return off, thrift.PrependError(fmt.Sprintf("%T read field %d '%s' error: ", p, fid, fieldIDToName_BaseResp[fid]), err)
SkipFieldError:
	return off, thrift.PrependError(fmt.Sprintf("%T skip field %d type %d error: ", p, fid, ftyp), err)
}
