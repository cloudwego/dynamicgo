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
    if (n == 0)
        return 0;
    size_t req = buf->len + n;
    if (req > buf->cap)
        WRAP_ERR0(ERR_OOM_BUF, req - buf->cap);
    buf->len = req;
    return 0;
}

static inline uint64_t tc_buf_malloc(tc_state *self, size_t n)
{
    _GoSlice *buf = self->buf;
    return _tc_buf_malloc(buf, n);
}

#define TC_VARINT32_BUF_SZ (size_t)4
#define TC_VARINT64_BUF_SZ (size_t)9

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
    if (off >= st->cap)
        return ERR_OOM_LFID;
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

// u16 MSB [key u8  ----|----  elem u8]
#define TC_CNTR_U16_VAL2(key,elem) (((uint8_t)(key&0xff)<<8)|((uint8_t)elem&0xff))
// u16 MSB [ignored ----|----  elem u8]
#define TC_CNTR_U16_VAL1(elem) TC_CNTR_U16_VAL2(0, elem)
#define TC_CNTR_RPUSH(q,v)                      \
    do                                          \
    {                                           \
        if (pq->len >= pq->cap)                 \
            WRAP_ERR0(ERR_OOM_CWBS, pq->cap);   \
        pq->buf[pq->len++] = v;                 \
    } while (0)
#define TC_CNTR_RPOP(q)     \
    pval = q->len > 0       \
        ? q->buf[q->len--]  \
        : 0

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
        size_t n = field->default_value->thrift_binary.len;
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
        J2T_ZERO( tc_write_field_end(self) );
        J2T_ZERO( tc_write_struct_end(self) );
        return 0;
    default:
        WRAP_ERR_POS(ERR_UNSUPPORT_THRIFT_TYPE, desc->type, p);
    }
}

static const tc_ienc tc_iencoder = {
    .base = {},

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

    .extra = {
        tc_write_default_or_empty,
        tc_write_data_count,
        tc_write_data_count_max_length,
    },
};

tc_ienc tc_get_iencoder()
{
    return tc_iencoder; // copy table
}


uint64_t tc_bm_malloc_reqs(_GoSlice *cache, ReqBitMap src, ReqBitMap *copy, long p)
{
    size_t n = src.len * SIZE_INT64;
    size_t d = cache->len + n;
    if unlikely (d > cache->cap)
    {
        xprintf("bm_malloc_reqs oom, d:%d, cap:%d\n", d, cache->cap);
        WRAP_ERR0(ERR_OOM_BM, (d - cache->cap) * SIZE_INT64);
    }

    copy->buf = (uint64_t *)&cache->buf[cache->len];
    copy->len = src.len;
    memcpy2((char *)copy->buf, (char *)src.buf, n);
    cache->len = d;
    // xprintf("[bm_malloc_reqs] copy:%n, src:%n \n", copy, &src);

    return 0;
}

uint64_t j2t2_write_unset_fields(tc_ienc *ienc, const tStructDesc *st, ReqBitMap reqs, uint64_t flags, long p)
{
    J2TStateMachine *self = ienc->base.resv0;
    _GoSlice *buf = ienc->base.resv1;
    tc_state *tc = ienc->base.resv3;

    bool wr_enabled = flags & F_WRITE_REQUIRE;
    bool wd_enabled = flags & F_WRITE_DEFAULT;
    bool wo_enabled = flags & F_WRITE_OPTIONAL;
    bool tb_enabled = flags & F_TRACE_BACK;
    uint64_t *s = reqs.buf;
    for (int i = 0; i < reqs.len; i++)
    {
        uint64_t v = *s;
        for (int j = 0; v != 0 && j < BSIZE_INT64; j++)
        {
            if ((v % 2) != 0)
            {
                tid id = i * BSIZE_INT64 + j;
                tFieldDesc *f = (st->ids.buf)[id];
                // xprintf("[j2t_write_unset_fields] field:%d, f.name:%s\n", (int64_t)id, &f->name);

                // NOTICE: if traceback is enabled AND (field is required OR current state is root layer),
                // return field id to traceback field value from http-mapping in Go.
                if (tb_enabled && (f->required == REQ_REQUIRED || self->sp == 1))
                {
                    if (self->field_cache.len >= self->field_cache.cap)
                        WRAP_ERR0(ERR_OOM_FIELD, p);
                    self->field_cache.buf[self->field_cache.len++] = f->ID;
                }
                else if (!wr_enabled && f->required == REQ_REQUIRED)
                {
                    WRAP_ERR_POS(ERR_NULL_REQUIRED, id, p);
                }
                else if ((wr_enabled && f->required == REQ_REQUIRED) || (wd_enabled && f->required == REQ_DEFAULT) || (wo_enabled && f->required == REQ_OPTIONAL))
                {
                    J2T_ZERO(tc_write_field_begin(tc, f->type->type, f->ID));
                    J2T_ZERO(tc_write_default_or_empty(tc, f, p));
                    J2T_ZERO(tc_write_field_end(tc));
                }
                v >>= 1;
            }
        }
        s++;
    }

    return 0;
}

uint64_t j2t2_number(tc_ienc *ienc, const tTypeDesc *desc, const _GoString *src, long *p, JsonState *ret)
{
    tc_state *tc = ienc->base.resv3;
    long s = *p;
    vnumber(src, p, ret);
    xprintf("[j2t2_number] p:%d, ret.vt:%d, ret.iv:%d, ret.dv:%f\n", s, ret->vt, ret->iv, ret->dv);
    if (ret->vt < 0)
        WRAP_ERR(-ret->vt, s);
    // FIXME: check double-integer overflow
    switch (desc->type)
    {
    case TTYPE_I08:
        return ienc->write_byte(tc,
            ret->vt == V_INTEGER
                ? (uint8_t)ret->iv
                : (uint8_t)ret->dv);
    case TTYPE_I16:
        return ienc->write_i16(tc,
            ret->vt == V_INTEGER
                ? (int16_t)ret->iv
                : (int16_t)ret->dv);
    case TTYPE_I32:
        return ienc->write_i32(tc,
            ret->vt == V_INTEGER
                ? (int32_t)ret->iv
                : (int32_t)ret->dv);
    case TTYPE_I64:
        return ienc->write_i64(tc,
            ret->vt == V_INTEGER
                ? (int64_t)ret->iv
                : (int64_t)ret->dv);
    case TTYPE_DOUBLE:
        return ienc->write_double(tc,
            ret->vt == V_DOUBLE
                ? (double)ret->dv
                : (double)((uint64_t)ret->iv));
    }
    WRAP_ERR2(ERR_DISMATCH_TYPE, desc->type, V_INTEGER);
}

uint64_t j2t2_string(tc_ienc *ienc, const _GoString *src, long *p, uint64_t flags)
{
    _GoSlice *buf = ienc->base.resv1;
    tc_state *tc = ienc->base.resv3;

    long s = *p;
    int64_t ep;
    ssize_t e = advance_string(src, s, &ep);
    J2T_XERR(e, s);
    *p = e;
    size_t n = e - s - 1;
    xprintf("[j2t_string]s:%d, p:%d, ep:%d, n:%d\n", s, e, ep, n);
    if unlikely (ep >= s && ep < e)
    {
        // write length
        J2T_ZERO(ienc->extra.write_data_count(tc, n));
        // unescape string
        size_t o = buf->len;
        ep = -1;
        J2T_ZERO(_tc_buf_malloc(buf, n));
        xprintf("[j2t_string] unquote o:%d, n:%d, ep:%d, flags:%d\n", o, n, ep, flags);
        ssize_t l = unquote(&src->buf[s], n, &buf->buf[o], &ep, flags);
        xprintf("[j2t_string] unquote end, l:%d\n", l);
        if (l < 0)
            WRAP_ERR(-l, s);
        buf->len = o + l;
    }
    else
        J2T_ZERO(ienc->write_string(tc,&src->buf[0], src->len));
    return 0;
}

uint64_t j2t2_binary(tc_ienc *ienc, const _GoString *src, long *p, uint64_t flags)
{
    _GoSlice *buf = ienc->base.resv1;
    tc_state *tc = ienc->base.resv3;

    long s = *p;
    int64_t ep;
    ssize_t e = advance_string(src, s, &ep);
    J2T_XERR(e, s);
    *p = e;
    size_t n = e - s - 1;
    xprintf("[j2t_binary]s:%d, p:%d, ep:%d, n:%d\n", s, *p, ep, n);
    // write length
    J2T_ZERO( ienc->extra.write_data_count(tc, n) );
    // write content
    ssize_t l = b64decode(buf, &src->buf[s], n, 0);
    xprintf("[j2t_binary] base64 endm l:%d\n", l);
    if (l < 0)
        WRAP_ERR(ERR_DECODE_BASE64, -l - 1);
    return 0;
}

uint64_t j2t2_map_key(tc_ienc *ienc, const char *sp, size_t n, const tTypeDesc *dc, JsonState *js, long p)
{
    tc_state *tc = ienc->base.resv3;
    xprintf("[j2t2_map_key]\n");
    switch (dc->type)
    {
    case TTYPE_STRING:
    {
        J2T_ZERO(ienc->write_string(tc, sp, n));
        break;
    }
    case TTYPE_I08:
    case TTYPE_I16:
    case TTYPE_I32:
    case TTYPE_I64:
    case TTYPE_DOUBLE:
    {
        _GoString tmp = { .buf = sp, .len = n };
        long p = 0;
        J2T_ZERO(j2t2_number(ienc, dc, &tmp, &p, js));
        break;
    }
    default:
        WRAP_ERR_POS(ERR_UNSUPPORT_THRIFT_TYPE, dc->type, p);
    }
    return 0;
}

tFieldDesc *j2t2_find_field_key(tc_ienc *ienc, const _GoString *key, const tStructDesc *st)
{
    tFieldDesc *ret;
    xprintf("[j2t2_find_field_key] key:%s\t", key);
    if unlikely (st->names.hash != NULL)
    {
        ret = (tFieldDesc *)hm_get(st->names.hash, key);
        xprintf("hash field:%x\n", ret);
    }
    else
    {
        ret = (tFieldDesc *)trie_get(st->names.trie, key);
        xprintf("hash field:%x\n", ret);
    }
    return ret;
}

uint64_t j2t2_read_key(tc_ienc *ienc, const _GoString *src, long *p, const char **spp, size_t *knp)
{
    J2TStateMachine *self = ienc->base.resv0;
    long s = *p;
    int64_t ep = -1;
    ssize_t e = advance_string(src, s, &ep);
    J2T_XERR(e, s);
    *p = e;
    size_t kn = e - s - 1;
    const char *sp = &src->buf[s];
    xprintf("[j2t2_read_key] s:%d, e:%d, kn:%d, ep:%d\n", s, e, kn, ep);
    if unlikely (ep >= s && ep < e)
    {
        if unlikely (kn > self->key_cache.cap)
        {
            WRAP_ERR0(ERR_OOM_KEY, kn - self->key_cache.cap);
        }
        else
        {
            ep = -1;
            char *dp = self->key_cache.buf;
            ssize_t l = unquote(sp, kn, dp, &ep, 0);
            xprintf("[j2t2_read_key] unquote end l:%d\n", l);
            if unlikely (l < 0)
                WRAP_ERR(-l,  s);
            sp = dp;
            kn = l;
        }
    }
    *spp = sp;
    *knp = kn;
    xprintf("[j2t2_read_key]sp:%c, kn:%d\n", *sp, kn);
    return 0;
}

uint64_t j2t2_field_vm(tc_ienc *ienc, const _GoString *src, long *p, J2TState *vt)
{
    J2TStateMachine *self = ienc->base.resv0;
    _GoSlice *buf = ienc->base.resv1;
    tc_state *tc = ienc->base.resv3;

    tFieldDesc *f = vt->ex.ef.f;
    xprintf("[j2t_field_vm] f->ID:%d, f->type->type:%d, p:%d \n", f->ID, f->type->type, *p);
    // if it is inlined value-mapping, write field tag first
    if (f->vm <= VM_INLINE_MAX)
    {
        J2T_ZERO( ienc->write_field_begin(tc, f->type->type, f->ID) );
        switch (f->vm)
        {
        case VM_JSCONV:
        {
            char ch = src->buf[*p - 1];
            if (ch == '"')
            {
                if (f->type->type == TTYPE_STRING)
                {
                    J2T_ZERO(j2t2_string(ienc, src, p, 0));
                    return 0;
                }
                if (src->buf[*p] == '"')
                {
                    // empty string
                    J2T_ZERO(ienc->extra.write_default_or_empty(tc, f, *p));
                    *p += 1;
                    return 0;
                }
            }
            else
            {
                // invalid number
                if (ch != '-' && (ch < '0' || ch > '9'))
                {
                    WRAP_ERR(ERR_INVAL, ch);
                }
                // back to begin of number;
                *p -= 1;
            }

            // handle it as number
            long s = *p;
            vnumber(src, p, &self->jt);
            if (self->jt.vt != V_INTEGER && self->jt.vt != V_DOUBLE)
                WRAP_ERR(ERR_NUMBER_FMT, self->jt.vt);
            xprintf("[j2t2_field_vm] jt.vt:%d, jt.iv:%d, jt.dv:%d\n", self->jt.vt, self->jt.iv, self->jt.dv);
            switch (f->type->type)
            {
            case TTYPE_STRING:
            {
                J2T_ZERO(ienc->write_string(tc, &src->buf[0], *p - s));
                return 0;
            }
            case TTYPE_I64:
            {
                int64_t v = self->jt.vt == V_INTEGER
                    ? (int64_t)self->jt.iv
                    : (int64_t)self->jt.dv;
                J2T_ZERO(ienc->write_i64(tc, v));
                break;
            }
            case TTYPE_I32:
            {
                xprintf("[j2t2_field_vm] TTYPE_I32\n");
                int32_t v = self->jt.vt == V_INTEGER
                    ? (int32_t)self->jt.iv
                    : (int32_t)self->jt.dv;
                J2T_ZERO(ienc->write_i32(tc, v));
                break;
            }
            case TTYPE_I16:
            {
                int16_t v = self->jt.vt == V_INTEGER
                    ? (int16_t)self->jt.iv
                    : (int16_t)self->jt.dv;
                J2T_ZERO(ienc->write_i16(tc, v));
                break;
            }
            case TTYPE_I08:
            {
                char v = self->jt.vt == V_INTEGER
                    ? (char)self->jt.iv
                    : (char)self->jt.dv;
                J2T_ZERO(ienc->write_byte(tc, v));
                break;
            }
            case TTYPE_DOUBLE:
            {
                double v = self->jt.vt == V_INTEGER
                    ? (double)((uint64_t)self->jt.iv)
                    : (double)self->jt.dv;
                J2T_ZERO(ienc->write_double(tc, v));
                break;
            }
            default:
                WRAP_ERR(ERR_UNSUPPORT_THRIFT_TYPE, f->type->type);
            }

            // For int/float field and string value, we need to skip the quote.
            if (ch == '"')
            {
                if (src->buf[*p] != '"')
                    WRAP_ERR(ERR_INVAL, src->buf[*p]);
                *p += 1;
            }
            return 0;
        }
        default:
            WRAP_ERR(ERR_UNSUPPORT_VM_TYPE, f->vm);
        }
    }
    else
    {
        // non-inline implementation, need return to caller
        *p -= 1;
        long s = *p;
        self->sm.sp = 0;
        long r = skip_one(src, p, &self->sm);
        if (r < 0)
            WRAP_ERR(-r, s);
        // check if the fval_cache if full
        // if (self->fval_cache.len >= self->fval_cache.cap)
        // {
        //     WRAP_ERR0(ERR_OOM_FVAL, s);
        // }
        // self->fval_cache.len += 1;
        // FieldVal *fv = (FieldVal *)&self->fval_cache.buf[self->fval_cache.len - 1];
        // record field id and value (begin and end) into FSM.FieldValueCache
        FieldVal *fv = &self->fval_cache;
        fv->id = f->ID;
        fv->start = r;
        fv->end = *p;
        WRAP_ERR0(ERR_VM_END, *p);
    }
}

#define j2t2_key(obj0)                                                                                                \
    do                                                                                                               \
    {                                                                                                                \
        const char *sp;                                                                                              \
        size_t kn;                                                                                                   \
        J2T_STORE(j2t2_read_key(&ienc, src, p, &sp, &kn));                                                             \
        if (dc->type == TTYPE_MAP)                                                                                   \
        {                                                                                                            \
            unwindPos = buf->len;                                                                                    \
            J2T_STORE(j2t2_map_key(&ienc, sp, kn, dc->key, &self->jt, *p));                                             \
            if (obj0)                                                                                                \
            {                                                                                                        \
                vt->ex.ec.size = 0;                                                                                  \
                J2T_PUSH(self, J2T_ELEM, *p, dc->elem);                                                              \
            }                                                                                                        \
            else                                                                                                     \
            {                                                                                                        \
                J2T_REPL(self, J2T_ELEM, *p, dc->elem);                                                              \
            }                                                                                                        \
        }                                                                                                            \
        else                                                                                                         \
        {                                                                                                            \
            J2TExtra *pex;                                                                                           \
            if (obj0)                                                                                                \
            {                                                                                                        \
                pex = &vt->ex;                                                                                       \
            }                                                                                                        \
            else                                                                                                     \
            {                                                                                                        \
                pex = &self->vt[self->sp - 2].ex;                                                                    \
            }                                                                                                        \
            _GoString tmp_str = (_GoString){.buf = sp, .len = kn};                                                   \
            tFieldDesc *f = j2t2_find_field_key(&ienc, &tmp_str, dc->st);                                             \
            if (f == NULL)                                                                                           \
            {                                                                                                        \
                xprintf("[J2T_KEY] unknown field %s\n", &tmp_str);                                                   \
                if ((flag & F_ALLOW_UNKNOWN) == 0)                                                                   \
                {                                                                                                    \
                    WRAP_ERR(ERR_UNKNOWN_FIELD, kn);                                                                 \
                }                                                                                                    \
                if (obj0)                                                                                            \
                {                                                                                                    \
                    J2T_PUSH(self, J2T_ELEM | STATE_SKIP, *p, NULL);                                                 \
                }                                                                                                    \
                else                                                                                                 \
                {                                                                                                    \
                    J2T_REPL(self, J2T_ELEM | STATE_SKIP, *p, NULL);                                                 \
                }                                                                                                    \
            }                                                                                                        \
            else                                                                                                     \
            {                                                                                                        \
                if unlikely ((flag & F_ENABLE_HM) && (f->http_mappings.len != 0) && !bm_is_set(pex->es.reqs, f->ID)) \
                {                                                                                                    \
                    xprintf("[J2T_KEY] skip field %s\n", &tmp_str);                                                  \
                    if (obj0)                                                                                        \
                    {                                                                                                \
                        J2T_PUSH(self, J2T_ELEM | STATE_SKIP, *p, f->type);                                          \
                    }                                                                                                \
                    else                                                                                             \
                    {                                                                                                \
                        J2T_REPL(self, J2T_ELEM | STATE_SKIP, *p, f->type);                                          \
                    }                                                                                                \
                }                                                                                                    \
                else                                                                                                 \
                {                                                                                                    \
                    J2TExtra ex;                                                                                     \
                    ex.ef.f = f;                                                                                     \
                    uint64_t vm = STATE_VM;                                                                          \
                    if ((flag & F_ENABLE_VM) == 0 || f->vm == VM_NONE)                                               \
                    {                                                                                                \
                        vm = STATE_FIELD;                                                                            \
                        unwindPos = buf->len;                                                                        \
                        lastField = f;                                                                               \
                        J2T_STORE(ienc.write_field_begin(ienc.base.resv3, f->type->type, f->ID));                   \
                    }                                                                                                \
                    xprintf("[J2T_KEY] vm: %d\n", vm);                                                               \
                    if (obj0)                                                                                        \
                    {                                                                                                \
                        J2T_PUSH_EX(self, J2T_ELEM | vm, *p, f->type, ex);                                           \
                    }                                                                                                \
                    else                                                                                             \
                    {                                                                                                \
                        J2T_REPL_EX(self, J2T_ELEM | vm, *p, f->type, ex);                                           \
                    }                                                                                                \
                }                                                                                                    \
                bm_set_req(pex->es.reqs, f->ID, REQ_OPTIONAL);                                                       \
            }                                                                                                        \
        }                                                                                                            \
    } while (0)


uint64_t j2t2_fsm_exec(J2TStateMachine *self, _GoSlice *buf, const _GoString *src, uint64_t flag)
{
    #define J2TSTATE_CUR  self->vt[self->sp - 1]
    #define J2TSTATE_NEXT self->vt[self->sp]

    // vTable
    // resv0 = J2TStateMachine
    // resv1 = out buffer
    // resv2 = unk
    // resv3 = tc_state
    tc_ienc ienc = tc_get_iencoder();
    ienc.base.resv0 = self;
    ienc.base.resv1 = buf;
    ienc.base.resv2 = NULL;
    tc_state tc = {
        .buf = buf,
        .last_field_id = 0,
        .last_field_id_stack = self->tc_last_field_id,
        .pending_field_write = {
            .valid = false,
        },
        .container_write_back_stack = self->tc_container_write_back,
    };
    ienc.base.resv3 = &tc;

    if (self->sp <= 0)
        return 0;
    // Load JSON position from last state
    long pv = J2TSTATE_CUR.jp;
    long *p = &pv;

    // for 'null' value backtrace
    bool null_val = false;
    size_t unwindPos = 0;
    tFieldDesc *lastField = NULL;

    // run until no more nested values
    while (self->sp)
    {
        if unlikely (self->sp >= MAX_RECURSE)
        {
            WRAP_ERR(ERR_RECURSE_MAX, self->sp);
        }
        // Load variables for current state
        J2TState *vt = &J2TSTATE_CUR;
        J2TState bt = *vt;
        const tTypeDesc *dc = bt.td;
        uint64_t st = bt.st;
        size_t wp = buf->len;

        // Advance to next char
        const char ch = advance_ns(src, p);
        DEBUG_PRINT_STATE(0, self->sp, *p);

        // check for special types
        switch (J2T_ST(st))
        {
        
        default:
        {
            xprintf("[J2T2_VAL] drop: %d\n", self->sp);
            J2T_DROP(self);
            break;
        }   

        // Arrays, first element
        case J2T_ARR_0:
        {
            xprintf("[J2T2_ARR_0] 0\n");
            switch (ch)
            {
            case ']':
            {
                xprintf("[J2T2_ARR_0] end\n");
                J2T_ZERO( ienc.write_list_begin(ienc.base.resv3, vt->ex.ec.bp, 0) );
                J2T_DROP(self);
                continue;
            }
            default:
            {
                xprintf("[J2T2_ARR_0] start\n");
                vt->ex.ec.size = 0;
                // Set parent state as J2T_ARR
                vt->st = J2T_ARR;
                *p -= 1;
                J2T_PUSH(self, J2T_VAL, *p, dc->elem);
                continue;
            }
            }
        }

        // Arrays
        case J2T_ARR:
        {
            xprintf("[J2T2_ARR] 0\n");
            switch (ch)
            {
            case ']':
            {
                xprintf("[J2T2_ARR] end\n");
                if (!null_val)
                {
                    // Count last elemnt
                    vt->ex.ec.size += 1;
                }
                else
                {
                    // Last element is null, so don't count it
                    null_val = false;
                }
                // Since the encoding of list and set, we call only one type's func here.
                // TODO: add Thrift type choose for corresponding func
                
                // TODO: mitigate SIZE write back.
                // ienc.write_list_end(ienc.base.resv3, vt->ex.ec.bp, vt->ex.ec.size);
                J2T_DROP(self);
                continue;
            }
            case ',':
            {
                xprintf("[J2T2_ARR] add\n");
                if (!null_val)
                {
                    vt->ex.ec.size += 1;
                }
                else
                {
                    null_val = false;
                }
                J2T_PUSH_EX(self, J2T_VAL, *p, dc->elem, vt->ex);
                continue;
            }
            default:
                WRAP_ERR2(ERR_INVAL, ch, J2T_ARR);
            }
        }

        // objects, first pair
        case J2T_OBJ_0:
        {
            xprintf("[J2T2_OBJ_0] 0\n");
            switch (ch)
            {
            default:
                WRAP_ERR2(ERR_INVAL, ch, J2T_OBJ_0);
            
            // empty object
            case '}':
            {
                xprintf("[J2T2_OBJ_0] end\n");
                if (dc->type == TTYPE_STRUCT)
                {
                    // Check if all required/default fields got written
                    J2T_STORE(j2t2_write_unset_fields(&ienc, vt->ex.es.sd, vt->ex.es.reqs, flag, *p - 1));
                    // http-mapping enabled and some fields remain not written
                    if (flag & F_ENABLE_HM && self->field_cache.len > 0)
                    {
                        // return back to Go handler
                        xprintf("[J2T2_OBJ_0] fallback http-mapping pos: %d\n", *p);
                        WRAP_ERR0(ERR_HM_END, *p);
                        // NOTICE: should tb_write_struct_end(), then J2T_DROP(self) in Go handler
                    }
                    else
                    {
                        J2T_STORE(ienc.write_struct_end(ienc.base.resv3));
                    }
                }
                else
                {
                    // TODO: mitigate SIZE write back.
                    // ienc.write_map_end
                }
                J2T_DROP(self);
                continue;
            }

            case '"':
            {
                xprintf("[J2T2_OBJ_0] start\n");
                // set parent state
                vt->st = J2T_OBJ;
                j2t2_key(true);
                continue;
            }
            }
        }

        // objects
        case J2T_OBJ:
        {
            xprintf("[J2T2_OBJ] 0\n");
            switch (ch)
            {
            case '}':
            {
                xprintf("[J2T2_OBJ] end\n");
                // Update backtrace position, in case of OOM return
                bt.jp = *p - 1;
                if (dc->type == TTYPE_STRUCT)
                {
                    // Last field value is null, so must unwind written field header
                    if (null_val)
                    {
                        null_val = false;
                        bm_set_req(vt->ex.es.reqs, lastField->ID, lastField->required);
                        buf->len = unwindPos;
                    }
                    // Check if all required/default fields got written
                    J2T_STORE(j2t2_write_unset_fields(&ienc, vt->ex.es.sd, vt->ex.es.reqs, flag, *p - 1));
                    bm_free_reqs(&self->reqs_cache, &vt->ex.es.reqs);
                    // http-mapping enabled and some fields remain not written
                    if (flag & F_ENABLE_HM && self->field_cache.len != 0)
                    {
                        // Return back to Go handler
                        xprintf("[J2T2_OBJ] fallback http-mapping pos: %d\n", *p);
                        WRAP_ERR0(ERR_HM_END, *p);
                        // NOTICE: should tb_write_struct_end(), then J2T_DROP(self) in Go handler
                    }
                    else
                        J2T_STORE(ienc.write_struct_end(ienc.base.resv3));
                }
                else
                {
                    if (!null_val)
                    {
                        vt->ex.ec.size += 1;
                    }
                    else
                    {
                        // Last element value is null, so must rewind to written key.
                        null_val = false;
                        buf->len = unwindPos;
                    }
                    // TODO: mitigate SIZE write back.
                    // ienc.write_map_end
                }
                J2T_DROP(self);
                continue;
            }
            case ',':
            {
                xprintf("[J2T2_OBJ] add\n");
                if (dc->type == TTYPE_MAP)
                {
                    if (!null_val)
                    {
                        vt->ex.ec.size += 1;
                    }
                    else
                    {
                        null_val = false;
                        buf->len = unwindPos;
                    }
                }
                else
                {
                    if (null_val)
                    {
                        null_val = false;
                        bm_set_req(vt->ex.es.reqs, lastField->ID, lastField->required);
                        buf->len = unwindPos;
                    }
                }
                J2T_PUSH(self, J2T_KEY, *p, dc);
                continue;
            }
            default:
                WRAP_ERR2(ERR_INVAL, ch, J2T_OBJ);
            }
        }

        // object keys
        case J2T_KEY:
        {
            xprintf("[J2T2_KEY] 0\n");
            J2T_CHAR('"', J2T_KEY);
            j2t2_key(false);
            continue;
        }

        // object element
        case J2T_ELEM:
        {
            xprintf("[J2T2_ELEM] 0\n");
            J2T_CHAR(':', J2T_ELEM);
            J2T_REPL(self, J2T_VAL | J2T_EX(st), *p, dc);
            continue;
        }
        }

        // skip
        if unlikely (IS_STATE_SKIP(st))
        {
            xprintf("[J2T_VAL] skip: %d\n", *p);
            *p -= 1;
            long s = *p;
            self->sm.sp = 0;
            J2T_XERR(skip_one(src, p, &self->sm), s);
            continue;
        }
        // value-mapping
        if unlikely (flag & F_ENABLE_VM && IS_STATE_VM(st))
        {
            xprintf("[J2T2_VAL] value mapping\n");
            J2T_STORE_NEXT(j2t2_field_vm(&ienc, src, p, vt));
            continue;
        }

        // normal value
        xprintf("[J2T2_VAL] switch\n");
        /* Simple values */
        switch (ch)
        {
        case '0': /* fallthrough */
        case '1': /* fallthrough */
        case '2': /* fallthrough */
        case '3': /* fallthrough */
        case '4': /* fallthrough */
        case '5': /* fallthrough */
        case '6': /* fallthrough */
        case '7': /* fallthrough */
        case '8': /* fallthrough */
        case '9': /* fallthrough */
        case '-':
        {
            xprintf("[J2T2_VAL] number\n");
            *p -= 1;
            J2T_STORE_NEXT(j2t2_number(&ienc, dc, src, p, &self->jt));
            break;
        }

        case 'n':
        {
            xprintf("[J2T2_VAL] null\n");
            long s = *p;
            J2T_XERR(advance_dword(src, p, 1, *p - 1, VS_NULL), s);
            if (IS_STATE_FIELD(st) && unlikely(vt->ex.ef.f->required))
                WRAP_ERR(ERR_NULL_REQUIRED, vt->ex.ef.f->ID);
            null_val = true;
            break;
        }

        case 't':
        {
            xprintf("[J2T2_VAL] true\n" );
            long s = *p;
            J2T_XERR(advance_dword(src, p, 1, *p - 1, VS_TRUE), s);
            J2T_EXP(TTYPE_BOOL, dc->type);
            J2T_STORE_NEXT(ienc.write_bool(ienc.base.resv3, true));
            break;
        }

        case 'f':
        {
            xprintf("[J2T2_VAL] false\n");
            long s = *p;
            J2T_XERR(advance_dword(src, p, 1, *p - 1, VS_ALSE), s);
            J2T_EXP(TTYPE_BOOL, dc->type);
            J2T_STORE_NEXT(ienc.write_bool(ienc.base.resv3, false));
        }
        
        case '[':
        {
            xprintf("[J2T2_VAL] array\n");
            J2T_EXP2(TTYPE_LIST, TTYPE_SET, dc->type);
            J2TState *v = &J2TSTATE_NEXT;
            // TODO: since the encoding of list and set, we call only one type's func here.
            // add thrift type choose for corresponding func
            v->ex.ec.size = 0;
            xprintf("[J2T_ARRAY]ex size: %d, vt: %d, vt.ex.ec: %d, vt.ex.ec.bp: %d, vt.ex.ec.size: %d\n", sizeof(J2TExtra), vt, &vt->ex.ec, &vt->ex.ec.bp, &vt->ex.ec.size);
            // TODO: check back here!
            // J2T_STORE_NEXT(ienc.write_list_begin(ienc.base.resv3, dc->elem->type, ))
            J2T_PUSH_EX(self, J2T_ARR_0, *p, dc, v->ex);
        }
        
        case '{':
        {
            xprintf("[J2T2_VAL] object\n");
            J2T_EXP2(TTYPE_STRUCT, TTYPE_MAP, dc->type);
            J2TState *v = &J2TSTATE_NEXT;
            if (dc->type == TTYPE_STRUCT)
            {
                xprintf("[J2T2_VAL] object struct name: %s\n", &dc->st->name);
                ienc.write_struct_begin(ienc.base.resv3);
            }
            else
            {
                v->ex.ec.size = 0;
                // TODO: check back here!
                // J2T_STORE_NEXT(ienc.write_map_begin(ienc.base.resv3, dc->key->type, dc->elem->type, ));
                J2T_PUSH(self, J2T_OBJ_0, *p, dc);
            }
        }

        case '"':
        {
            xprintf("[J2T2_VAL] string\n");
            if (dc->type == TTYPE_STRING)
            {
                if unlikely ((flag & F_NO_BASE64) == 0 && memeq(dc->name.buf, "binary", 6))
                {
                    J2T_STORE_NEXT(j2t2_binary(&ienc, src, p, 0));
                }
                else
                {
                    J2T_STORE_NEXT(j2t2_string(&ienc, src, p, 0));
                }
            }
            else if ((flag & F_ENABLE_I2S) != 0 && (dc->type == TTYPE_I64 || dc->type == TTYPE_I32 || dc->type == TTYPE_I16 || dc->type == TTYPE_BYTE))
            {
                J2T_STORE_NEXT(j2t2_number(&ienc, dc, src, p, &self->jt));
                long x = *p;
                if (x >= src->len)
                    WRAP_ERR(ERR_EOF, J2T_VAL);
                // must end with a double quote
                if (src->buf[x] != '"')
                    WRAP_ERR2(ERR_INVAL, src->buf[x], J2T_VAL);
                *p += 1;
            }
            else
                WRAP_ERR2(ERR_DISMATCH_TYPE, dc->type, TTYPE_STRING);
        }

        case 0:
            WRAP_ERR(ERR_EOF, J2T_VAL);
        default:
            WRAP_ERR2(ERR_INVAL, ch, J2T_VAL);

        }

    }

    xprintf("[J2T2] return\n");
    // All done
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
