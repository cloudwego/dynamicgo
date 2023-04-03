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

#include <stdint.h>
#include "memops.h"
#include "native.h"
#include "thrift_binary.h"
#include "test/xprintf.h"
#include <stdbool.h>

static inline uint64_t _tb_buf_malloc(_GoSlice *buf, size_t n)
{
    return buf_malloc(buf, n);
}

static inline uint64_t tb_buf_malloc(tb_state *self, size_t n)
{
    return _tb_buf_malloc(self->buf, n);
}

static inline uint64_t _tb_write_data_count(_GoSlice *buf, size_t n)
{
    size_t off = buf->len;
    J2T_ZERO( _tb_buf_malloc(buf, 4) );
    *(int32_t *)(&buf->buf[off]) = __builtin_bswap32(n);
    return 0;
}

static inline uint64_t tb_write_data_count(tb_state *self, size_t n)
{
    return _tb_write_data_count(self->buf, n);
}

size_t tb_write_data_count_max_length(tb_state *self)
{
    return TB_DATA_COUNT_LENGTH;
}

static inline uint64_t __tb_write_byte(_GoSlice *buf, char v)
{
    size_t off = buf->len;
    J2T_ZERO( _tb_buf_malloc(buf, 1) );
    buf->buf[off] = v;
    return 0;
}
static inline uint64_t _tb_write_byte(tb_state *self, char v)
{
    return __tb_write_byte(self->buf, v);
}
static inline uint64_t tb_write_byte(tb_state *self, char v)
{
    return _tb_write_byte(self, v);
}

uint64_t tb_write_bool(tb_state *self, bool v)
{
    return tb_write_byte(self, v ? 1 : 0);
}

uint64_t tb_write_i16(tb_state *self, int16_t v)
{
    size_t off = self->buf->len;
    J2T_ZERO( tb_buf_malloc(self, 2) );
    *(int16_t *)(&self->buf->buf[off]) = __builtin_bswap16(v);
    return 0;
}

uint64_t tb_write_i32(tb_state *self, int32_t v)
{
    size_t off = self->buf->len;
    J2T_ZERO( tb_buf_malloc(self, 4) );
    *(int32_t *)(&self->buf->buf[off]) = __builtin_bswap32(v);
    return 0;
}

uint64_t tb_write_i64(tb_state *self, int64_t v)
{
    size_t off = self->buf->len;
    J2T_ZERO( tb_buf_malloc(self, 8) );
    *(int64_t *)(&self->buf->buf[off]) = __builtin_bswap64(v);
    return 0;
}

uint64_t tb_write_double(tb_state *self, double v)
{
    uint64_t f = *(uint64_t *)&v;
    return tb_write_i64(self, f);
}

uint64_t tb_write_string(tb_state *self, const char *v, size_t n)
{
    // _GoSlice *buf = self->buf;
    // size_t s = buf->len;
    size_t off; // = buf->len;
    J2T_ZERO( tb_write_data_count(self, n) );
    // xprintf("[tb_write_string] siz_enc=[%x,%x,%x,%x]\n", buf->buf[off], buf->buf[off + 1], buf->buf[off + 2], buf->buf[off + 3]);
    // s = buf->len;
    off = self->buf->len;
    J2T_ZERO( tb_buf_malloc(self, n) );
    memcpy2(&self->buf->buf[off], v, n);
    return 0;
}

uint64_t tb_write_binary(tb_state *self, const _GoSlice v)
{
    size_t off;
    J2T_ZERO( tb_write_data_count(self, v.len) );
    off = self->buf->len;
    J2T_ZERO( tb_buf_malloc(self, v.len) );
    memcpy2(&self->buf->buf[off], v.buf, v.len);
    return 0;
}

uint64_t tb_write_message_begin(tb_state *self, const _GoString name, int32_t type, int32_t seq)
{
    uint32_t version = TB_THRIFT_VERSION_1 | (uint32_t)type;
    J2T_ZERO( tb_write_i32(self, version) );
    J2T_ZERO( tb_write_string(self, name.buf, name.len) );
    return tb_write_i32(self, seq);
}
uint64_t tb_write_message_end(tb_state *self)
{ return 0; }

uint64_t tb_write_struct_begin(tb_state *self)
{ return 0; }
uint64_t tb_write_struct_end(tb_state *self)
{
    J2T_ZERO( tb_write_byte(self, TTYPE_STOP) );
    xprintf("[tb_write_struct_end]:%x\n", self->buf->buf[self->buf->len - 1]);
    return 0;
}

uint64_t tb_write_field_begin(tb_state *self, ttype type, int16_t id)
{
    J2T_ZERO( tb_write_byte(self, type) );
    return tb_write_i16(self, id);
}
uint64_t tb_write_field_end(tb_state *self)
{ return 0; }
uint64_t tb_write_field_stop(tb_state *self)
{ return 0; }

uint64_t tb_write_map_n(tb_state *self, ttype key, ttype val, size_t n)
{
    J2T_ZERO( tb_write_byte(self, key) );
    J2T_ZERO( tb_write_byte(self, val) );
    J2T_ZERO( tb_write_i32(self, n) );
    return 0;
}
uint64_t tb_write_map_begin(tb_state *self, ttype key, ttype val, size_t *backp)
{
    J2T_ZERO( tb_write_byte(self, key) );
    J2T_ZERO( tb_write_byte(self, val) );
    *backp = self->buf->len;
    J2T_ZERO( tb_buf_malloc(self, 4) );
    return 0;
}
uint64_t tb_write_map_end(tb_state *self, size_t back_off, size_t vsize)
{
    xprintf("[tb_write_map_end]: %d, bp pos: %d\n", vsize, back_off);
    _GoSlice tmp = {
        .buf = &self->buf->buf[back_off],
        .len = 0,
        .cap = 4,
    };
    J2T_ZERO( _tb_write_data_count(&tmp, vsize) );
    return 0;
}

static inline uint64_t _tb_write_collection_n(_GoSlice *buf, ttype elem, int32_t size)
{
    J2T_ZERO( __tb_write_byte(buf, elem) );
    J2T_ZERO( _tb_write_data_count(buf, size) );
    return 0;
}
static inline uint64_t tb_write_collection_n(tb_state *self, ttype elem, size_t size)
{
    return _tb_write_collection_n(self->buf, elem, size);
}
static inline uint64_t tb_write_collection_begin(tb_state *self, ttype elem, size_t *backp)
{
    J2T_ZERO( tb_write_byte(self, elem) );
    *backp = self->buf->len;
    J2T_ZERO( tb_buf_malloc(self, 4) );
    return 0;
}
static inline uint64_t tb_write_collection_end(tb_state *self, size_t back_off, size_t vsize)
{
    xprintf("[tb_write_list_end]: %d, bp pos: %d\n", vsize, back_off);
    _GoSlice tmp = {
        .buf = &self->buf->buf[back_off],
        .len = 0,
        .cap = 4,
    };
    J2T_ZERO( _tb_write_data_count(&tmp, vsize));
    return 0;
}

uint64_t tb_write_list_n(tb_state *self, ttype elem, size_t size)
{
    return tb_write_collection_n(self, elem, size);
}
uint64_t tb_write_list_begin(tb_state *self, ttype elem, size_t *backp)
{
    return tb_write_collection_begin(self, elem, backp);
}
uint64_t tb_write_list_end(tb_state *self, size_t back_off, size_t vsize)
{
    return tb_write_collection_end(self, back_off, vsize);
}

uint64_t tb_write_set_n(tb_state *self, ttype elem, size_t size)
{
    return tb_write_collection_n(self, elem, size);
}
uint64_t tb_write_set_begin(tb_state *self, ttype elem, size_t *backp)
{
    return tb_write_collection_begin(self, elem, backp);
}
uint64_t tb_write_set_end(tb_state *self, size_t back_off, size_t vsize)
{
    return tb_write_collection_end(self, back_off, vsize);
}

uint64_t tb_write_default_or_empty(tb_state *self, const tFieldDesc *field, long p)
{
    if (field->default_value != NULL)
    {
        xprintf("[tb_write_default_or_empty] json_val: %s\n", &field->default_value->json_val);
        size_t n = field->default_value->thrift_binary.len;
        size_t l = self->buf->len;
        J2T_ZERO(tb_buf_malloc(self, n));
        memcpy2(&self->buf->buf[l], 
            field->default_value->thrift_binary.buf, n);
        return 0;
    }
    tTypeDesc *desc = field->type;
    switch (desc->type)
    {
    case TTYPE_BOOL:
        return tb_write_bool(self, false);
    case TTYPE_BYTE:
        return tb_write_byte(self, 0);
    case TTYPE_I16:
        return tb_write_i16(self, 0);
    case TTYPE_I32:
        return tb_write_i32(self, 0);
    case TTYPE_I64:
        return tb_write_i64(self, 0);
    case TTYPE_DOUBLE:
        return tb_write_double(self, 0);
    case TTYPE_STRING:
        return tb_write_string(self, NULL, 0);
    case TTYPE_LIST: // fallthrough
    case TTYPE_SET:
        J2T_ZERO( tb_write_collection_n(self, desc->elem->type, 0) );
        return 0;
    case TTYPE_MAP:
        J2T_ZERO( tb_write_map_n(self, desc->key->type, desc->elem->type, 0) );
        return 0;
    case TTYPE_STRUCT:
        J2T_ZERO( tb_write_struct_begin(self) );
        tFieldDesc **f = desc->st->ids.buf_all;
        size_t l = desc->st->ids.len_all;
        for (int i = 0; i < l; i++)
        {
            if (f[i]->required == REQ_OPTIONAL)
                continue;
            J2T_ZERO( tb_write_field_begin(self, f[i]->type->type, f[i]->ID) );
            J2T_ZERO( tb_write_default_or_empty(self, f[i], p));
            J2T_ZERO( tb_write_field_end(self) );
        }
        J2T_ZERO( tb_write_field_stop(self) );
        J2T_ZERO( tb_write_struct_end(self) );
        return 0;
    default:
        WRAP_ERR_POS(ERR_UNSUPPORT_THRIFT_TYPE, desc->type, p);
    }
}

static const tb_menc_extra TB_IENC_M_EXTRA = {
    tb_write_default_or_empty,
    tb_write_data_count,
    tb_write_data_count_max_length,
};
static const tb_menc TB_IENC_M = {
    tb_write_message_begin,
    tb_write_message_end,
    tb_write_struct_begin,
    tb_write_struct_end,
    tb_write_field_begin,
    tb_write_field_end,
    tb_write_field_stop,
    tb_write_map_n,
    tb_write_map_begin,
    tb_write_map_end,
    tb_write_list_n,
    tb_write_list_begin,
    tb_write_list_end,
    tb_write_set_n,
    tb_write_set_begin,
    tb_write_set_end,

    tb_write_bool,
    tb_write_byte,
    tb_write_i16,
    tb_write_i32,
    tb_write_i64,
    tb_write_double,
    tb_write_string,
    tb_write_binary,
};

static const tb_ienc TB_IENCODER = {
    .base = {},
    .method = &TB_IENC_M,
    .extra = &TB_IENC_M_EXTRA,
};

typedef struct { 
    J2TStateMachine *j2tsm; 
    _GoSlice *outbuf; 
    tb_state *tb;
} tb_get_iencoder_arg;

// resv0: VT_J2TSM      = J2TStateMachine
// resv1: VT_OUTBUF     = out buffer
// resv2 = unk
// resv3: VT_TC_STATE   = tc_state
static __always_inline
tb_ienc tb_get_iencoder(tb_get_iencoder_arg arg)
{
    tb_ienc vt = TB_IENCODER;
    vt.base.resv2 = NULL;
    VT_J2TSM(vt.base)    = arg.j2tsm;
    VT_OUTBUF(vt.base)   = arg.outbuf;
    VT_TB_STATE(vt.base) = arg.tb;
    return vt; // copy table
}

uint64_t j2t2_fsm_tb_exec(J2TStateMachine *self, _GoSlice *buf, const _GoString *src, uint64_t flag)
{
    tb_state tb = {
        .buf = buf,
    };
    
    // Get vTable
    tb_ienc ienc = tb_get_iencoder((tb_get_iencoder_arg){
        .j2tsm  = self,
        .outbuf = buf,
        .tb     = &tb,
    });

    return j2t2_fsm_exec((vt_ienc *)&ienc, src, flag);
}