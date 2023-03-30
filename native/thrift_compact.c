#include "memops.h"
#include "native.h"
#include "thrift.h"
#include "thrift_compact.h"
#include "test/xprintf.h"
#include <stddef.h>

static inline tcompacttype ttype2ttc(ttype type);
static inline ttype ttc2ttype(tcompacttype ttc);


static inline uint64_t tc_buf_malloc(tc_state *self, size_t n)
{
    if (n == 0)
        return 0;
    _GoSlice *buf = self->buf;
    size_t req = self->buf->len + n;
    if (req > buf->cap)
        WRAP_ERR0(ERR_OOM_BUF, req - buf->cap);
    buf->len = req;
    return 0;
}

uint64_t _tc_write_varint32(tc_state *self, int32_t n)
{
    char buf[5];
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
    size_t off = self->buf->len;
    J2T_ZERO( tc_buf_malloc(self, idx) );
    memcpy2(&self->buf->buf[off], (char *)&buf[0], idx);
    return 0;
}

uint64_t _tc_write_varint64(tc_state *self, int64_t n)
{
    char buf[10];
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
    size_t off = self->buf->len;
    J2T_ZERO( tc_buf_malloc(self, idx) );
    memcpy2(&self->buf->buf[off], (char *)&buf[0], idx);
    return 0;
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

static inline uint64_t tc_write_byte(tc_state *self, char v)
{
    size_t n = self->buf->len;
    J2T_ZERO( tc_buf_malloc(self, 1) );
    self->buf->buf[n] = v;
    return 0;
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
    J2T_ZERO( _tc_write_varint32(self, (int32_t)n) );
    size_t off = self->buf->len;
    J2T_ZERO( tc_buf_malloc(self, n) );
    memcpy2(&self->buf->buf[off], v, n);
    return 0;
}
uint64_t tc_write_binary(tc_state *self, const _GoSlice v)
{
    J2T_ZERO( _tc_write_varint32(self, v.len) );
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
    if (off+1 > st->cap)
        return ERR_OOM_BUF;
    st->buf[off] = self->last_field_id;
    st->len++;
    self->last_field_id = 0;
    return 0;
}

uint64_t tc_write_struct_end(tc_state *self)
{
    Int16Slice *st = &self->last_field_id_stack;
    self->last_field_id = st->len > 0
        ? st->buf[st->len--]
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

uint64_t tc_write_map_begin(tc_state *self, ttype key, ttype val, size_t size)
{
    if (size == 0)
        return tc_write_byte(self, 0);
    J2T_ZERO( _tc_write_varint32(self, (int32_t)size) );
    return tc_write_byte(self, 
        (char)( (ttype2ttc(key)<<4) | (ttype2ttc(val)) ));
}
uint64_t tc_write_map_end(tc_state *self)
{ return 0; }


static inline uint64_t tc_write_collection_begin(tc_state *self, ttype elem, size_t size)
{
    if (size <= 14)
        return tc_write_byte(self,
            (char)( ((int32_t)size << 4) | ((int32_t)ttype2ttc(elem)) ));
    J2T_ZERO( tc_write_byte(self, 0xf0 | (char)(ttype2ttc(elem))) );
    return _tc_write_varint32(self, (int32_t)size);
}
uint64_t tc_write_list_begin(tc_state *self, ttype elem, size_t size)
{
    return tc_write_collection_begin(self, elem, size);
}
uint64_t tc_write_list_end(tc_state *self)
{ return 0; }
uint64_t tc_write_set_begin(tc_state *self, ttype elem, size_t size)
{
    return tc_write_collection_begin(self, elem, size);
}
uint64_t tc_write_set_end(tc_state *self)
{ return 0; }


static const tc_ienc tc_iencoder = {
    .base = {},

    tc_write_message_begin,
    tc_write_message_end,
    tc_write_struct_begin,
    tc_write_struct_end,
    tc_write_field_begin,
    tc_write_field_end,
    tc_write_field_stop,
    tc_write_map_begin,
    tc_write_map_end,
    tc_write_list_begin,
    tc_write_list_end,
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

tc_ienc tc_get_iencoder()
{
    return tc_iencoder; // copy table
}


uint64_t tc_write_default_or_empty()
{
    

    return 0;
}




// ttype2tcc convert ttype to tcompacttype
static inline tcompacttype ttype2ttc(ttype type)
{
    if (type > TTYPE_LAST)
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
static inline ttype ttc2ttype(tcompacttype ttc)
{
    if (ttc > TTC_LAST)
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
