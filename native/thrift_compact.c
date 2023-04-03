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

#include "memops.h"
#include "native.h"
#include "scanning.h"
#include "map.h"
#include "thrift.h"
#include "thrift_compact.h"
#include "test/xprintf.h"
#include <stdbool.h>
#include <stddef.h>

static inline tcompacttype ttype2ttc(ttype type);
static inline ttype ttc2ttype(tcompacttype ttc);

static inline uint64_t _tc_buf_malloc(_GoSlice *buf, size_t n)
{
    return buf_malloc(buf, n);
}

static inline uint64_t tc_buf_malloc(tc_state *self, size_t n)
{
    return _tc_buf_malloc(self->buf, n);
}

static inline uint64_t __tc_write_varint32(_GoSlice *obuf, int32_t n)
{
    char buf[TC_VARINT32_BUF_SZ];
    size_t idx = 0;
    while (1)
    {
        if (unlikely((n & ~0x7F) == 0)) 
        {
            buf[idx++] = (char)(n);
            break;
        }
        else
        {
            buf[idx++] = (char)((n & 0x7F) | 0x80);
            n = (int32_t)((uint32_t)(n) >> 7);
        }
    }
    size_t off = obuf->len;
    J2T_ZERO( _tc_buf_malloc(obuf, idx) );
    memcpy2(&obuf->buf[off], (char *)&buf[0], idx);
    __builtin_memcpy(&obuf->buf[off], &buf[0], idx);
    return 0;
}
uint64_t _tc_write_varint32(tc_state *self, int32_t n)
{
    return __tc_write_varint32(self->buf, n);
}

static inline uint64_t __tc_write_varint64(_GoSlice *obuf, int64_t n)
{
    char buf[TC_VARINT64_BUF_SZ];
    size_t idx = 0;
    while (1)
    {
        if (unlikely((n & ~0x7F) == 0))
        {
            buf[idx++] = (char)(n);
            break;
        }
        else
        {
            buf[idx++] = (char)((n & 0x7F) | 0x80);
            n = (int64_t)((uint64_t)(n) >> 7);
        }
    }
    size_t off = obuf->len;
    J2T_ZERO( _tc_buf_malloc(obuf, idx) );
    memcpy2(&obuf->buf[off], (char *)&buf[0], idx);
    return 0;
}
uint64_t _tc_write_varint64(tc_state *self, int64_t n)
{
    return __tc_write_varint64(self->buf, n);
}

static inline uint64_t _tc_write_data_count(_GoSlice *buf, size_t n)
{
    return __tc_write_varint32(buf, n);
}

static inline uint64_t tc_write_data_count(tc_state *self, size_t n)
{
    return _tc_write_data_count(self->buf, n);
}

size_t tc_write_data_count_max_length(tc_state *self)
{
    return TC_VARINT32_BUF_SZ;
}

uint64_t _tc_write_field(tc_state *self, ttype type, int64_t id, uint8_t type_override)
{
    int16_t delta = id-self->last_field_id;
    tcompacttype type_to_write = type_override == 0xFF
        ? ttype2ttc(type)
        : (tcompacttype)type_override;
    if (id > self->last_field_id && delta <= 0x0F)
    {
        // write_byte( field_id_delta | type )
        J2T_ZERO( tc_write_byte(self, (delta << 4) | (type_to_write)) );
    }
    else
    {
        // write_byte, write_i16
        J2T_ZERO( tc_write_byte(self, type_to_write) );
        J2T_ZERO( tc_write_i16(self, id) );
    }
    self->last_field_id = id;
    return 0;
}


static inline uint64_t __tc_write_byte(_GoSlice *buf, char v)
{
    size_t n = buf->len;
    J2T_ZERO( _tc_buf_malloc(buf, 1) );
    buf->buf[n] = v;
    return 0;
}
static inline uint64_t _tc_write_byte(tc_state *self, char v)
{
    return __tc_write_byte(self->buf, v);
}
uint64_t tc_write_byte(tc_state *self, char v)
{
    return _tc_write_byte(self, v);
}

uint64_t tc_write_bool(tc_state *self, bool value)
{
    tcompacttype vb = value ? TTC_BOOLEAN_TRUE : TTC_BOOLEAN_FALSE;
    if (self->pending_field_write.valid) {    
        self->pending_field_write.valid = false;
        return _tc_write_field(self, 
            self->pending_field_write.type, 
            self->pending_field_write.id, 
            vb);
    }
    // not a part of a field so just write the value.
    return tc_write_byte(self, (char)vb);
}

uint64_t tc_write_i16(tc_state *self, int16_t v)
{   
    return _tc_write_varint32(self, TTC_I32ToZigzag((int32_t)v));
}
uint64_t tc_write_i32(tc_state *self, int32_t v)
{
    return _tc_write_varint32(self, TTC_I32ToZigzag(v));
}
uint64_t tc_write_i64(tc_state *self, int64_t v)
{
    return _tc_write_varint64(self, TTC_I64ToZigzag(v));
}
uint64_t tc_write_double(tc_state *self, double v)
{
    size_t off = self->buf->len;
    J2T_ZERO( tc_buf_malloc(self, 8) );
    *(uint64_t *)(&self->buf->buf[off]) = *(uint64_t *)(&v);
    return 0;
}
uint64_t tc_write_string(tc_state *self, const char *v, size_t n)
{
    J2T_ZERO( tc_write_data_count(self, n));
    size_t off = self->buf->len;
    J2T_ZERO( tc_buf_malloc(self, n) );
    memcpy2(&self->buf->buf[off], v, n);
    return 0;
}
uint64_t tc_write_binary(tc_state *self, const _GoSlice v)
{
    J2T_ZERO( tc_write_data_count(self, v.len) );
    size_t off = self->buf->len;
    J2T_ZERO( tc_buf_malloc(self, v.len) );
    memcpy2(&self->buf->buf[off], v.buf, v.len);
    return 0;
}


uint64_t tc_write_message_begin(tc_state *self, const _GoString name, int32_t type, int32_t seq)
{
    size_t n = self->buf->len;
    J2T_ZERO( tc_buf_malloc(self, 2) );
    char *buf = self->buf->buf;
    buf[n] = TTC_PROTOCOL_ID;
    buf[n+1] = (TTC_VERSION & TTC_VERSION_MASK)
        | ((type << TTC_TYPE_SHIFT_AMOUNT) & TTC_TYPE_MASK);
    J2T_ZERO( _tc_write_varint32(self, seq) );
    return tc_write_string(self, name.buf, name.len);
}
uint64_t tc_write_message_end(tc_state *self)
{ return 0; }


uint64_t tc_write_struct_begin(tc_state *self)
{ 
    Int16Slice *st = &self->last_field_id_stack;
    size_t off = st->len;
    if (off > st->cap)
        WRAP_ERR0(ERR_OOM_LFID, off+1);
    st->buf[off] = self->last_field_id;
    st->len++;
    self->last_field_id = 0;
    return 0;
}
uint64_t tc_write_struct_end(tc_state *self)
{
    Int16Slice *st = &self->last_field_id_stack;
    self->last_field_id = st->len > 0
        ? st->buf[--st->len]
        : 0;
    return 0;
}


uint64_t tc_write_field_begin(tc_state *self, ttype type, int16_t id)
{
    if (type == TTYPE_BOOL)
    {
        // setup pending write for boolean
        self->pending_field_write.valid = true;
        self->pending_field_write.type = type;
        self->pending_field_write.id = id;
        return 0;
    }
    return _tc_write_field(self, type, id, 0xFF);
}
uint64_t tc_write_field_end(tc_state *self)
{
    return 0;
}
uint64_t tc_write_field_stop(tc_state *self)
{
    return tc_write_byte(self, TTC_STOP);
}

static inline uint64_t _tc_write_map_n(_GoSlice *buf, ttype key, ttype val, size_t size)
{
    if (size == 0)
        return __tc_write_byte(buf, 0);
    J2T_ZERO( _tc_write_data_count(buf, size) );
    return __tc_write_byte(buf, 
        (char)( (ttype2ttc(key)<<4) | (ttype2ttc(val)) ));
}
uint64_t tc_write_map_n(tc_state *self, ttype key, ttype val, size_t size)
{
    return _tc_write_map_n(self->buf, key, val, size);
}

uint64_t tc_write_map_begin(tc_state *self, ttype key, ttype val, size_t *backp)
{
    Uint16Slice *pq = &self->container_write_back_stack;
    *backp = self->buf->len;

    // store ttype
    TC_CNTR_RPUSH(pq, TC_CNTR_U16_VAL2(key, val));

    // get reserved bytes count
    // varint32 max + 1 byte (key, elem type)
    size_t resv_count = tc_write_data_count_max_length(self) + 1;
    return tc_buf_malloc(self, resv_count);
}
uint64_t tc_write_map_end(tc_state *self, size_t back_off, size_t vsize)
{ 
    Uint16Slice *pq = &self->container_write_back_stack;
    uint16_t pval;
    TC_CNTR_RPOP(pq);

    char *hdr = &self->buf->buf[back_off];
    // varint32 max + 1
    size_t resv_count = tc_write_data_count_max_length(self) + 1;
    size_t wr = 0;
    char *cur = &self->buf->buf[self->buf->len];
    size_t cur_sz = self->buf->len;
    size_t mov_sz = cur_sz-(back_off+resv_count);
    ttype key, elem;
    
    key = (pval&0xff00)>>8;
    elem = pval&0x00ff;

    if likely (vsize == 0)
    {
        // write_byte 0
        *hdr = 0;
        wr = 1;
    }
    else
    {
        _GoSlice tmp = { .buf = hdr, .len = 0, .cap = resv_count };
        J2T_ZERO( _tc_write_map_n(&tmp, key, elem, vsize) );
        wr = tmp.len;
    }
    
    // hdr[0..wr] written
    // Compacting written bytes
    char *dsp = &hdr[wr];
    char *ssp = &hdr[resv_count];
    size_t delta = resv_count - wr;
    
    memcpy2(dsp, ssp, mov_sz);
    self->buf->len =  cur_sz - delta;
    return 0;
}

static inline uint64_t _tc_write_collection_n(_GoSlice *buf, ttype elem, size_t size)
{
    if (size <= 14)
        return __tc_write_byte(buf,
            (char)( ((int32_t)size << 4) | ((int32_t)ttype2ttc(elem)) ));
    J2T_ZERO( __tc_write_byte(buf, 0xf0 | (char)(ttype2ttc(elem))) );
    return __tc_write_varint32(buf, size);
}

static inline uint64_t tc_write_collection_n(tc_state *self, ttype elem, size_t size)
{
    return _tc_write_collection_n(self->buf, elem, size);
}
static inline uint64_t tc_write_collection_begin(tc_state *self, ttype elem, size_t *backp)
{
    Uint16Slice *pq = &self->container_write_back_stack;
    *backp = self->buf->len;

    // store ttype
    TC_CNTR_RPUSH(pq, TC_CNTR_U16_VAL1(elem));

    // get reserved bytes count
    // 1 + varint32 max
    size_t resv_count = 1 + tc_write_data_count_max_length(self);
    return tc_buf_malloc(self, resv_count);
}
static inline uint64_t tc_write_collection_end(tc_state *self, size_t back_off, size_t vsize)
{
    Uint16Slice *pq = &self->container_write_back_stack;
    uint16_t pval;
    TC_CNTR_RPOP(pq);

    char *hdr = &self->buf->buf[back_off];
    // 1 + varint32 max
    size_t resv_count = 1 + tc_write_data_count_max_length(self);
    size_t wr = 0;
    char *cur = &self->buf->buf[self->buf->len];
    size_t cur_sz = self->buf->len;
    size_t mov_sz = cur_sz-(back_off+resv_count);
    ttype elem;

    elem = pval&0x00ff;

    _GoSlice tmp = { .buf = hdr, .len = 0, .cap = resv_count };
    J2T_ZERO( _tc_write_collection_n(&tmp, elem, vsize) );
    wr = tmp.len;

    // Compacting written bytes
    char *dsp = &hdr[wr];
    char *ssp = &hdr[resv_count];
    size_t delta = resv_count - wr;

    memcpy2(dsp, ssp, mov_sz);
    self->buf->len = cur_sz - delta;
    return 0;
}

uint64_t tc_write_list_n(tc_state *self, ttype elem, size_t size)
{
    return tc_write_collection_n(self, elem, size);
}
uint64_t tc_write_list_begin(tc_state *self, ttype elem, size_t *backp)
{
    return tc_write_collection_begin(self, elem, backp);
}
uint64_t tc_write_list_end(tc_state *self, size_t back_off, size_t vsize)
{
    return tc_write_collection_end(self, back_off, vsize);
}
uint64_t tc_write_set_n(tc_state *self, ttype elem, size_t size)
{
    return tc_write_collection_n(self, elem, size);
}
uint64_t tc_write_set_begin(tc_state *self, ttype elem, size_t *backp)
{
    return tc_write_collection_begin(self, elem, backp);
}
uint64_t tc_write_set_end(tc_state *self, size_t back_off, size_t vsize)
{
    return tc_write_collection_end(self, back_off, vsize);
}


uint64_t tc_write_default_or_empty(tc_state *self, const tFieldDesc *field, long p)
{
    _GoSlice *buf = self->buf;
    if (field->default_value != NULL)
    {
        xprintf("[tc_write_default_or_empty] json_val: %s\n", &field->default_value->json_val);
        size_t n = field->default_value->thrift_compact.len;
        size_t l = buf->len;
        J2T_ZERO(tc_buf_malloc(self, n));
        memcpy2(&buf->buf[l],
            field->default_value->thrift_compact.buf, n);
        return 0;
    }
    tTypeDesc *desc = field->type;
    switch (desc->type)
    {
    case TTYPE_BOOL:
        return tc_write_bool(self, false);
    case TTYPE_BYTE:
        return tc_write_byte(self, 0);
    case TTYPE_I16:
        return tc_write_i16(self, 0);
    case TTYPE_I32:
        return tc_write_i32(self, 0);
    case TTYPE_I64:
        return tc_write_i64(self, 0);
    case TTYPE_DOUBLE:
        return tc_write_double(self, 0);
    case TTYPE_STRING:
        return tc_write_string(self, NULL, 0);
    case TTYPE_LIST:
    case TTYPE_SET:
        J2T_ZERO( tc_write_collection_n(self, desc->elem->type, 0) );
        return 0;
    case TTYPE_MAP:
        J2T_ZERO( tc_write_map_n(self,  desc->key->type, desc->elem->type, 0));
        return 0;
    case TTYPE_STRUCT:
        J2T_ZERO( tc_write_struct_begin(self) );
        tFieldDesc **f = desc->st->ids.buf_all;
        size_t l = desc->st->ids.len_all;
        for (int i = 0; i < l; i++)
        {
            if (f[i]->required == REQ_OPTIONAL)
                continue;
            J2T_ZERO( tc_write_field_begin(self, f[i]->type->type, f[i]->ID) );
            J2T_ZERO( tc_write_default_or_empty(self, f[i], p) );
            J2T_ZERO( tc_write_field_end(self) );
        }
        J2T_ZERO( tc_write_field_stop(self) );
        J2T_ZERO( tc_write_struct_end(self) );
        return 0;
    default:
        WRAP_ERR_POS(ERR_UNSUPPORT_THRIFT_TYPE, desc->type, p);
    }
}

static const tc_menc_extra TC_IENC_M_EXTRA = {
    tc_write_default_or_empty,
    tc_write_data_count,
    tc_write_data_count_max_length,
};
static const tc_menc TC_IENC_M = {
    tc_write_message_begin,
    tc_write_message_end,
    tc_write_struct_begin,
    tc_write_struct_end,
    tc_write_field_begin,
    tc_write_field_end,
    tc_write_field_stop,
    tc_write_map_n,
    tc_write_map_begin,
    tc_write_map_end,
    tc_write_list_n,
    tc_write_list_begin,
    tc_write_list_end,
    tc_write_set_n,
    tc_write_set_begin,
    tc_write_set_end,
    
    tc_write_bool,
    tc_write_byte,
    tc_write_i16,
    tc_write_i32,
    tc_write_i64,
    tc_write_double,
    tc_write_string,
    tc_write_binary,
};

static const tc_ienc TC_IENCODER = {
    .base = {},
    .method = &TC_IENC_M,
    .extra = &TC_IENC_M_EXTRA,
};

typedef struct { 
    J2TStateMachine *j2tsm; 
    _GoSlice *outbuf; 
    tc_state *tc;
} tc_get_iencoder_arg;

// resv0: VT_J2TSM      = J2TStateMachine
// resv1: VT_OUTBUF     = out buffer
// resv2 = unk
// resv3: VT_TSTATE     = tc_state
static __always_inline
tc_ienc tc_get_iencoder(tc_get_iencoder_arg arg)
{
    tc_ienc vt = TC_IENCODER;
    vt.base.resv2 = NULL;
    VT_J2TSM(vt.base)    = arg.j2tsm;
    VT_OUTBUF(vt.base)   = arg.outbuf;
    VT_TSTATE(vt.base)   = arg.tc;
    return vt; // copy table
}

uint64_t j2t2_fsm_tc_exec(J2TStateMachine *self, _GoSlice *buf, const _GoString *src, uint64_t flag)
{    
    tc_state tc = {
        .buf = buf,
        .last_field_id = 0,
        .last_field_id_stack = self->tc_last_field_id,
        .pending_field_write = {
            .valid = false,
        },
        .container_write_back_stack = self->tc_container_write_back,
    };
    
    // Get vTable
    tc_ienc ienc = tc_get_iencoder((tc_get_iencoder_arg){
        .j2tsm  = self,
        .outbuf = buf,
        .tc     = &tc,
    });

    return j2t2_fsm_exec((vt_ienc *)&ienc, src, flag);
}

// ttype2tcc convert ttype to tcompacttype
static
inline
// __always_inline
tcompacttype ttype2ttc(ttype type)
{
    if (type >= TTYPE_LAST)
        return TTYPE_EINVAL;
    switch (type)
    {
    case TTYPE_STOP: 
        return TTC_STOP;
    case TTYPE_BOOL:
        return TTC_BOOLEAN_TRUE;
    case TTYPE_BYTE:
        return TTC_BYTE;
    case TTYPE_I16:
        return TTC_I16;
    case TTYPE_I32:
        return TTC_I32;
    case TTYPE_I64:
        return TTC_I64;
    case TTYPE_DOUBLE:
        return TTC_DOUBLE;
    case TTYPE_STRING:
        return TTC_BINARY;
    case TTYPE_LIST:
        return TTC_LIST;
    case TTYPE_SET:
        return TTC_SET;
    case TTYPE_MAP:
        return TTC_MAP;
    case TTYPE_STRUCT:
        return TTC_STRUCT;
    }
    __builtin_unreachable();
}

// ttc2ttype convert tcompacttype to ttype
static
inline
// __always_inline
ttype ttc2ttype(tcompacttype ttc)
{
    if (ttc >= TTC_LAST)
        return TTC_EINVAL;
    switch (ttc)
    {
    case TTC_STOP: 
        return TTYPE_STOP;
    case TTC_BOOLEAN_TRUE:
    case TTC_BOOLEAN_FALSE:
        return TTYPE_BOOL;
    case TTC_BYTE:
        return TTYPE_BYTE;
    case TTC_I16:
        return TTYPE_I16;
    case TTC_I32:
        return TTYPE_I32;
    case TTC_I64:
        return TTYPE_I64;
    case TTC_DOUBLE:
        return TTYPE_DOUBLE;
    case TTC_BINARY:
        // binary/string
        return TTYPE_STRING;
    case TTC_LIST:
        return TTYPE_LIST;
    case TTC_SET:
        return TTYPE_SET;
    case TTC_MAP:
        return TTYPE_MAP;
    case TTC_STRUCT:
        return TTYPE_STRUCT;
    }
    __builtin_unreachable();
}
