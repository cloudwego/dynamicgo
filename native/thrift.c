/*
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

#include "native.h"
#include "map.h"
#include "scanning.h"
#include "memops.h"
#include "thrift.h"
#include <stdint.h>
#include "test/xprintf.h"

uint64_t buf_malloc(GoSlice *buf, size_t size)
{
    if (size == 0)
    {
        return 0;
    }
    size_t d = buf->len + size;
    if (d > buf->cap)
    {
        WRAP_ERR0(ERR_OOM_BUF, d - buf->cap);
    }
    buf->len = d;
    return 0;
}

uint64_t tb_write_byte(GoSlice *buf, char v)
{
    size_t s = buf->len;
    J2T_ZERO(buf_malloc(buf, 1));
    buf->buf[s] = v;
    return 0;
}

uint64_t tb_write_bool(GoSlice *buf, bool v)
{
    if (v)
    {
        return tb_write_byte(buf, 1);
    }
    return tb_write_byte(buf, 0);
}

uint64_t tb_write_i16(GoSlice *buf, int16_t v)
{
    size_t s = buf->len;
    J2T_ZERO(buf_malloc(buf, 2));
    *(int16_t *)(&buf->buf[s]) = __builtin_bswap16(v);
    return 0;
}

uint64_t tb_write_i32(GoSlice *buf, int32_t v)
{
    size_t s = buf->len;
    J2T_ZERO(buf_malloc(buf, 4));
    *(int32_t *)(&buf->buf[s]) = __builtin_bswap32(v);
    return 0;
}

uint64_t tb_write_i64(GoSlice *buf, int64_t v)
{
    size_t s = buf->len;
    J2T_ZERO(buf_malloc(buf, 8));
    *(int64_t *)(&buf->buf[s]) = __builtin_bswap64(v);
    return 0;
}

uint64_t tb_write_double(GoSlice *buf, double v)
{
    uint64_t f = *(uint64_t *)(&v);
    return tb_write_i64(buf, f);
}

uint64_t tb_write_string(GoSlice *buf, const char *v, size_t n)
{
    size_t s = buf->len;
    J2T_ZERO(buf_malloc(buf, n + 4));
    *(int32_t *)(&buf->buf[s]) = __builtin_bswap32(n);
    xprintf("[tb_write_string] %c%c%c%c\n", buf->buf[s], buf->buf[s + 1], buf->buf[s + 2], buf->buf[s + 3]);
    memcpy2(&buf->buf[s + 4], v, n);
    return 0;
}

uint64_t tb_write_binary(GoSlice *buf, const GoSlice v)
{
    size_t s = buf->len;
    J2T_ZERO(buf_malloc(buf, v.len + 4));
    *(int32_t *)(&buf->buf[s]) = __builtin_bswap32(v.len);
    memcpy2(&buf->buf[s + 4], v.buf, v.len);
    return 0;
}

void tb_write_struct_begin(GoSlice *buf)
{
    return;
}

uint64_t tb_write_struct_end(GoSlice *buf)
{
    J2T_ZERO(tb_write_byte(buf, TTYPE_STOP));
    xprintf("[tb_write_struct_end]:%c\n", buf->buf[buf->len - 1]);
    return 0;
}

uint64_t tb_write_field_begin(GoSlice *buf, ttype type, int16_t id)
{
    J2T_ZERO(tb_write_byte(buf, type));
    return tb_write_i16(buf, id);
}

uint64_t tb_write_map_n(GoSlice *buf, ttype key, ttype val, int32_t n)
{
    J2T_ZERO(tb_write_byte(buf, key));
    J2T_ZERO(tb_write_byte(buf, val));
    J2T_ZERO(tb_write_i32(buf, n));
    return 0;
}

uint64_t tb_write_map_begin(GoSlice *buf, ttype key, ttype val, size_t *back)
{
    J2T_ZERO(tb_write_byte(buf, key));
    J2T_ZERO(tb_write_byte(buf, val));
    *back = buf->len;
    J2T_ZERO(buf_malloc(buf, 4));

    return 0;
}

void tb_write_map_end(GoSlice *buf, size_t back, size_t size)
{
    xprintf("[tb_write_map_end]: %d, bp pos: %d\n", size, back);
    *(int32_t *)(&buf->buf[back]) = __builtin_bswap32(size);
    return;
}

uint64_t tb_write_list_begin(GoSlice *buf, ttype elem, size_t *back)
{
    J2T_ZERO(tb_write_byte(buf, elem));
    *back = buf->len;
    J2T_ZERO(buf_malloc(buf, 4));
    return 0;
}

void tb_write_list_end(GoSlice *buf, size_t back, size_t size)
{
    xprintf("[tb_write_list_end]: %d, bp pos: %d\n", size, back);
    *(int32_t *)(&buf->buf[back]) = __builtin_bswap32(size);
    return;
}

uint64_t tb_write_list_n(GoSlice *buf, ttype elem, int32_t size)
{
    J2T_ZERO(tb_write_byte(buf, elem));
    J2T_ZERO(tb_write_i32(buf, size));
    return 0;
}

uint64_t tb_write_default_or_empty(GoSlice *buf, const tFieldDesc *field, long p)
{
    if (field->default_value != NULL)
    {
        xprintf("[tb_write_default_or_empty] json_val: %s\n", &field->default_value->json_val);
        size_t n = field->default_value->thrift_binary.len;
        size_t l = buf->len;
        J2T_ZERO(buf_malloc(buf, n));
        memcpy2(&buf->buf[l], field->default_value->thrift_binary.buf, n);
        return 0;
    }
    tTypeDesc *desc = field->type;
    switch (desc->type)
    {
    case TTYPE_BOOL:
        return tb_write_bool(buf, false);
    case TTYPE_BYTE:
        return tb_write_byte(buf, 0);
    case TTYPE_I16:
        return tb_write_i16(buf, 0);
    case TTYPE_I32:
        return tb_write_i32(buf, 0);
    case TTYPE_I64:
        return tb_write_i64(buf, 0);
    case TTYPE_DOUBLE:
        return tb_write_double(buf, 0);
    case TTYPE_STRING:
        return tb_write_string(buf, NULL, 0);
    case TTYPE_LIST:
    case TTYPE_SET:
        J2T_ZERO(tb_write_list_n(buf, desc->elem->type, 0));
        return 0;
    case TTYPE_MAP:
        J2T_ZERO(tb_write_map_n(buf, desc->key->type, desc->elem->type, 0));
        return 0;
    case TTYPE_STRUCT:
        tb_write_struct_begin(buf);
        tFieldDesc **f = desc->st->ids.buf_all;
        size_t l = desc->st->ids.len_all;
        for (int i = 0; i < l; i++)
        {
            if (f[i]->required == REQ_OPTIONAL)
            {
                continue;
            }
            J2T_ZERO(tb_write_field_begin(buf, f[i]->type->type, f[i]->ID));
            J2T_ZERO(tb_write_default_or_empty(buf, f[i], p));
        }
        J2T_ZERO(tb_write_struct_end(buf));
        return 0;
    default:
        WRAP_ERR_POS(ERR_UNSUPPORT_THRIFT_TYPE, desc->type, p);
    }
}

uint64_t tb_write_message_begin(GoSlice *buf, const GoString name, int32_t type, int32_t seq)
{
    uint32_t version = THRIFT_VERSION_1 | (uint32_t)type;
    J2T_ZERO(tb_write_i32(buf, version));
    J2T_ZERO(tb_write_string(buf, name.buf, name.len));
    return tb_write_i32(buf, seq);
}

void tb_write_message_end(GoSlice *buf)
{
    return;
}

uint64_t bm_malloc_reqs(GoSlice *cache, ReqBitMap src, ReqBitMap *copy, long p)
{
    xprintf("[bm_malloc_reqs oom] malloc, cache.len:%d, cache.cap:%d, src.len:%d\n", cache->len, cache->cap, src.len);
    size_t n = src.len * SIZE_INT64;
    size_t d = cache->len + n;
    if unlikely (d > cache->cap)
    {
        xprintf("[bm_malloc_reqs] oom d:%d, cap:%d\n", d, cache->cap);
        WRAP_ERR0(ERR_OOM_BM, (d - cache->cap));
    }

    copy->buf = (uint64_t *)&cache->buf[cache->len];
    copy->len = src.len;
    memcpy2((char *)copy->buf, (char *)src.buf, n);
    cache->len = d;
    // xprintf("[bm_malloc_reqs] copy:%n, src:%n \n", copy, &src);

    return 0;
}

void bm_free_reqs(GoSlice *cache, size_t len)
{
    xprintf("[bm_free_reqs] free, cache.len: %d, src.len:%d", cache->len, len);
    cache->len -= (len * SIZE_INT64);
}

uint64_t j2t_write_unset_fields(J2TStateMachine *self, GoSlice *buf, const tStructDesc *st, ReqBitMap reqs, uint64_t flags, long p)
{
    bool wr_enabled = flags & F_WRITE_REQUIRE;
    bool wd_enabled = flags & F_WRITE_DEFAULT;
    bool wo_enabled = flags & F_WRITE_OPTIONAL;
    bool tb_enabled = flags & F_TRACE_BACK;
    uint64_t *s = reqs.buf;
    // xprintf("[j2t_write_unset_fields] reqs: %n \n", &reqs);
    for (int i = 0; i < reqs.len; i++)
    {
        uint64_t v = *s;
        for (int j = 0; v != 0 && j < BSIZE_INT64; j++)
        {
            if ((v % 2) != 0)
            {
                tid id = i * BSIZE_INT64 + j;
                if (id >= st->ids.len)
                {
                    continue;
                }
                tFieldDesc *f = (st->ids.buf)[id];
                // xprintf("[j2t_write_unset_fields] field:%d, f.name:%s\n", (int64_t)id, &f->name);

                // NOTICE: if traceback is enabled AND (field is required OR current state is root layer),
                // return field id to traceback field value from http-mapping in Go.
                if (tb_enabled && (f->required == REQ_REQUIRED || self->sp == 1))
                {
                    if (self->field_cache.cap <= self->field_cache.len)
                    {
                        WRAP_ERR0(ERR_OOM_FIELD, p);
                    }
                    self->field_cache.buf[self->field_cache.len++] = f->ID;
                }
                else if (!wr_enabled && f->required == REQ_REQUIRED)
                {
                    WRAP_ERR_POS(ERR_NULL_REQUIRED, id, p);
                }
                else if ((wr_enabled && f->required == REQ_REQUIRED) || (wd_enabled && f->required == REQ_DEFAULT) || (wo_enabled && f->required == REQ_OPTIONAL))
                {
                    J2T_ZERO(tb_write_field_begin(buf, f->type->type, f->ID));
                    J2T_ZERO(tb_write_default_or_empty(buf, f, p));
                }
            }
            v >>= 1;
        }
        s++;
    }
    return 0;
}

uint64_t j2t_number(GoSlice *buf, const tTypeDesc *desc, const GoString *src, long *p, JsonState *ret)
{
    long s = *p;
    vnumber(src, p, ret);
    xprintf("[j2t_number] p:%d, ret.vt:%d, ret.iv:%d, ret.dv:%f\n", s, ret->vt, ret->iv, ret->dv);
    if (ret->vt < 0)
    {
        WRAP_ERR(-ret->vt, s);
    }
    // FIXME: check double-integer overflow
    switch (desc->type)
    {
    case TTYPE_I08:
        if (ret->vt == V_INTEGER)
        {
            return tb_write_byte(buf, (uint8_t)ret->iv);
        }
        else
        {
            return tb_write_byte(buf, (uint8_t)ret->dv);
        }
    case TTYPE_I16:
        if (ret->vt == V_INTEGER)
        {
            return tb_write_i16(buf, (int16_t)ret->iv);
        }
        else
        {
            return tb_write_i16(buf, (int16_t)ret->dv);
        }
    case TTYPE_I32:
        if (ret->vt == V_INTEGER)
        {
            return tb_write_i32(buf, (int32_t)ret->iv);
        }
        else
        {
            return tb_write_i32(buf, (int32_t)ret->dv);
        }
    case TTYPE_I64:
        if (ret->vt == V_INTEGER)
        {
            return tb_write_i64(buf, (int64_t)ret->iv);
        }
        else
        {
            return tb_write_i64(buf, (int64_t)ret->dv);
        }
    case TTYPE_DOUBLE:
        if (ret->vt == V_DOUBLE)
        {
            return tb_write_double(buf, (double)ret->dv);
        }
        else
        {
            return tb_write_double(buf, (double)((uint64_t)ret->iv));
        }
    }
    WRAP_ERR2(ERR_DISMATCH_TYPE, desc->type, V_INTEGER);
}

uint64_t j2t_string(GoSlice *buf, const GoString *src, long *p, uint64_t flags)
{
    long s = *p;
    int64_t ep;
    ssize_t e = advance_string(src, s, &ep);
    J2T_XERR(e, s);
    *p = e;
    size_t n = e - s - 1;
    xprintf("[j2t_string]s:%d, p:%d, ep:%d, n:%d\n", s, e, ep, n);
    if unlikely (ep >= s && ep < e)
    {
        char *sp = &buf->buf[buf->len];
        J2T_ZERO(buf_malloc(buf, 4));
        // unescape string
        size_t o = buf->len;
        ep = -1;
        J2T_ZERO(buf_malloc(buf, n));
        xprintf("[j2t_string] unquote o:%d, n:%d, ep:%d, flags:%d\n", o, n, ep, flags);
        ssize_t l = unquote(&src->buf[s], n, &buf->buf[o], &ep, flags);
        xprintf("[j2t_string] unquote end, l:%d\n", l);
        if (l < 0)
        {
            WRAP_ERR(-l, s);
        }
        buf->len = o + l;
        *(int32_t *)sp = __builtin_bswap32(l);
    }
    else
    {
        J2T_ZERO(tb_write_string(buf, &src->buf[s], n));
    }
    return 0;
}

uint64_t j2t_binary(GoSlice *buf, const GoString *src, long *p, uint64_t flags)
{
    long s = *p;
    int64_t ep;
    ssize_t e = advance_string(src, s, &ep);
    J2T_XERR(e, s);
    *p = e;
    size_t n = e - s - 1;
    xprintf("[j2t_binary]s:%d, p:%d, ep:%d, n:%d\n", s, *p, ep, n);
    char *back = &buf->buf[buf->len];
    J2T_ZERO(buf_malloc(buf, 4));
    ssize_t l = b64decode(buf, &src->buf[s], n, 0);
    xprintf("[j2t_binary] base64 endm l:%d\n", l);
    if (l < 0)
    {
        WRAP_ERR(ERR_DECODE_BASE64, -l - 1);
    }
    *(int32_t *)(back) = __builtin_bswap32(l);
    return 0;
}

uint64_t j2t_map_key(GoSlice *buf, const char *sp, size_t n, const tTypeDesc *dc, JsonState *js, long p)
{
    xprintf("[j2t_map_key]\n");
    switch (dc->type)
    {
    case TTYPE_STRING:
    {
        J2T_ZERO(tb_write_string(buf, sp, n));
        break;
    }
    case TTYPE_I08:
    case TTYPE_I16:
    case TTYPE_I32:
    case TTYPE_I64:
    case TTYPE_DOUBLE:
    {
        GoString tmp = (GoString){.buf = sp, .len = n};
        long p = 0;
        J2T_ZERO(j2t_number(buf, dc, &tmp, &p, js));
        break;
    }
    default:
        WRAP_ERR_POS(ERR_UNSUPPORT_THRIFT_TYPE, dc->type, p);
    }
    return 0;
}

tFieldDesc *j2t_find_field_key(const GoString *key, const tStructDesc *st)
{
    xprintf("[j2t_find_field_key] key:%s\t", key);
    if unlikely (st->names.hash != NULL)
    {
        tFieldDesc *ret = (tFieldDesc *)hm_get(st->names.hash, key);
        xprintf("hash field:%x\n", ret);
        return ret;
    }
    else
    {
        tFieldDesc *ret = (tFieldDesc *)trie_get(st->names.trie, key);
        xprintf("trie field:%x\n", ret);
        return ret;
    }
}

uint64_t j2t_read_key(J2TStateMachine *self, const GoString *src, long *p, const char **spp, size_t *knp)
{
    long s = *p;
    int64_t ep = -1;
    ssize_t e = advance_string(src, s, &ep);
    J2T_XERR(e, s);
    *p = e;
    size_t kn = e - s - 1;
    const char *sp = &src->buf[s];
    xprintf("[j2t_read_key] s:%d, e:%d, kn:%d, ep:%d\n", s, e, kn, ep);
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
            xprintf("[j2t_read_key] unquote end l:%d\n", l);
            if unlikely (l < 0)
            {
                WRAP_ERR(-l, s);
            }
            sp = dp;
            kn = l;
        }
    }
    *spp = sp;
    *knp = kn;
    xprintf("[j2t_read_key]sp:%c, kn:%d\n", *sp, kn);
    return 0;
}

uint64_t j2t_field_vm(J2TStateMachine *self, GoSlice *buf, const GoString *src, long *p, J2TState *vt)
{
    tFieldDesc *f = vt->ex.ef.f;
    xprintf("[j2t_field_vm] f->ID:%d, f->type->type:%d, p:%d \n", f->ID, f->type->type, *p);
    // if it is inlined value-mapping, write field tag first
    if (f->vm <= VM_INLINE_MAX)
    {
        J2T_ZERO(tb_write_field_begin(buf, f->type->type, f->ID));
        switch (f->vm)
        {
        case VM_JSCONV:
        {
            char ch = src->buf[*p - 1];
            if (ch == '"')
            {
                // If the field is a string and value is also a string, we just handle it normally.
                if (f->type->type == TTYPE_STRING)
                {
                    J2T_ZERO(j2t_string(buf, src, p, 0));
                    return 0;
                }
                if (src->buf[*p] == '"')
                {
                    // empty string
                    J2T_ZERO(tb_write_default_or_empty(buf, f, *p));
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
                // back to begin of number
                *p -= 1;
            }

            // handle it as number
            long s = *p;
            vnumber(src, p, &self->jt);
            if (self->jt.vt != V_INTEGER && self->jt.vt != V_DOUBLE)
            {
                WRAP_ERR(ERR_NUMBER_FMT, self->jt.vt);
            }
            xprintf("[j2t_field_vm] jt.vt:%d, jt.iv:%d, jt.dv:%d\n", self->jt.vt, self->jt.iv, self->jt.dv);
            switch (f->type->type)
            {
            case TTYPE_STRING:
            {
                J2T_ZERO(tb_write_string(buf, &src->buf[s], *p - s));
                return 0;
            }
            case TTYPE_I64:
            {
                int64_t v = 0;
                if (self->jt.vt == V_INTEGER)
                {
                    v = (int64_t)self->jt.iv;
                }
                else
                {
                    v = (int64_t)self->jt.dv;
                }
                J2T_ZERO(tb_write_i64(buf, v));
                break;
            }
            case TTYPE_I32:
            {
                xprintf("[j2t_field_vm] TTYPE_I32\n");
                int32_t v = 0;
                if (self->jt.vt == V_INTEGER)
                {
                    v = (int32_t)self->jt.iv;
                }
                else
                {
                    v = (int32_t)self->jt.dv;
                }
                J2T_ZERO(tb_write_i32(buf, v));
                break;
            }
            case TTYPE_I16:
            {
                int16_t v = 0;
                if (self->jt.vt == V_INTEGER)
                {
                    v = (int16_t)self->jt.iv;
                }
                else
                {
                    v = (int16_t)self->jt.dv;
                }
                J2T_ZERO(tb_write_i16(buf, v));
            }
            case TTYPE_I08:
            {
                char v = 0;
                if (self->jt.vt == V_INTEGER)
                {
                    v = (char)self->jt.iv;
                }
                else
                {
                    v = (char)self->jt.dv;
                }
                J2T_ZERO(tb_write_byte(buf, v));
                break;
            }
            case TTYPE_DOUBLE:
            {
                double v = 0;
                if (self->jt.vt == V_INTEGER)
                {
                    v = (double)((uint64_t)self->jt.iv);
                }
                else
                {
                    v = (double)self->jt.dv;
                }
                J2T_ZERO(tb_write_double(buf, v));
                break;
            }
            default:
                WRAP_ERR(ERR_UNSUPPORT_THRIFT_TYPE, f->type->type);
            }

            // for int/float field and string value, we need to skip the quote.
            if (ch == '"')
            {
                if (src->buf[*p] != '"')
                {
                    WRAP_ERR(ERR_INVAL, src->buf[*p]);
                }
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
        {
            WRAP_ERR(-r, s);
        }
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

#define j2t_key(obj0)                                                                                                \
    do                                                                                                               \
    {                                                                                                                \
        const char *sp;                                                                                              \
        size_t kn;                                                                                                   \
        J2T_STORE(j2t_read_key(self, src, p, &sp, &kn));                                                             \
        if (dc->type == TTYPE_MAP)                                                                                   \
        {                                                                                                            \
            unwindPos = buf->len;                                                                                    \
            J2T_STORE(j2t_map_key(buf, sp, kn, dc->key, &self->jt, *p));                                             \
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
            GoString tmp_str = (GoString){.buf = sp, .len = kn};                                                     \
            tFieldDesc *f = j2t_find_field_key(&tmp_str, dc->st);                                                    \
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
                    J2TExtra ex = {};                                                                                \
                    ex.ef.f = f;                                                                                     \
                    uint64_t vm = STATE_VM;                                                                          \
                    if ((flag & F_ENABLE_VM) == 0 || f->vm == VM_NONE)                                               \
                    {                                                                                                \
                        vm = STATE_FIELD;                                                                            \
                        unwindPos = buf->len;                                                                        \
                        lastField = f;                                                                               \
                        J2T_STORE(tb_write_field_begin(buf, f->type->type, f->ID));                                  \
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

uint64_t j2t_fsm_exec(J2TStateMachine *self, GoSlice *buf, const GoString *src, uint64_t flag)
{
    if (self->sp <= 0)
    {
        return 0;
    }
    // load json position from last state
    long pv = self->vt[self->sp - 1].jp;
    long *p = &pv;

    // for 'null' value backtrace
    bool null_val = false;
    size_t unwindPos = 0;
    tFieldDesc *lastField = NULL;

    /* run until no more nested values */
    while (self->sp)
    {
        if unlikely (self->sp >= MAX_RECURSE)
        {
            WRAP_ERR(ERR_RECURSE_MAX, self->sp);
        }
        // load varialbes for current state
        J2TState *vt = &self->vt[self->sp - 1];
        J2TState bt = *vt;
        const tTypeDesc *dc = bt.td;
        uint64_t st = bt.st;
        size_t wp = buf->len;

        // advance to next char
        const char ch = advance_ns(src, p);
        DEBUG_PRINT_STATE(0, self->sp - 1, *p);

        /* check for special types */
        switch (J2T_ST(st))
        {

        default:
        {
            xprintf("[J2T_VAL] drop: %d\n", self->sp);
            J2T_DROP(self);
            break;
        }

        /* arrays, first element */
        case J2T_ARR_0:
        {
            xprintf("[J2T_ARR_0] 0\n");
            switch (ch)
            {
            case ']':
            {
                xprintf("[J2T_ARR_0] end\n");
                tb_write_list_end(buf, vt->ex.ec.bp, 0);
                J2T_DROP(self);
                continue;
            }
            default:
            {
                xprintf("[J2T_ARR_0] start\n");
                vt->ex.ec.size = 0;
                // set parent state as J2T_ARR
                vt->st = J2T_ARR;
                *p -= 1;
                J2T_PUSH(self, J2T_VAL, *p, dc->elem);
                continue;
            }
            }
        }

        /* arrays */
        case J2T_ARR:
        {
            xprintf("[J2T_ARR] 0\n");
            switch (ch)
            {
            case ']':
            {
                xprintf("[J2T_ARR] end\n");
                if (!null_val)
                {
                    // count last element
                    vt->ex.ec.size += 1;
                }
                else
                {
                    // last element is null, so don't count it
                    null_val = false;
                }
                // since the encoding of list and set, we call only one type's func here.
                // TODO: add thrift type choose for corresponding func
                tb_write_list_end(buf, vt->ex.ec.bp, vt->ex.ec.size);
                J2T_DROP(self);
                continue;
            }
            case ',':
                xprintf("[J2T_ARR] add\n");
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
            default:
                WRAP_ERR2(ERR_INVAL, ch, J2T_ARR);
            }
        }

        /* objects, first pair */
        case J2T_OBJ_0:
        {
            xprintf("[J2T_OBJ_0] 0\n");
            switch (ch)
            {
            default:
            {
                WRAP_ERR2(ERR_INVAL, ch, J2T_OBJ_0);
            }

            /* empty object */
            case '}':
            {
                xprintf("[J2T_OBJ_0] end\n");
                if (dc->type == TTYPE_STRUCT)
                {
                    // check if all required/default fields got written
                    J2T_STORE(j2t_write_unset_fields(self, buf, vt->ex.es.sd, vt->ex.es.reqs, flag, *p - 1));
                    bm_free_reqs(&self->reqs_cache, vt->ex.es.reqs.len);
                    // http-mapping enalbed and some fields remain not written
                    if (flag & F_ENABLE_HM && self->field_cache.len > 0)
                    { // return back to go handler
                        xprintf("[J2T_OBJ_0] fallback http-mapping pos: %d\n", *p);
                        WRAP_ERR0(ERR_HM_END, *p);
                        // NOTICE: should tb_write_struct_end(), then J2T_DROP(self) in Go handler
                    }
                    else
                    {
                        J2T_STORE(tb_write_struct_end(buf));
                    }
                }
                else
                {
                    tb_write_map_end(buf, vt->ex.ec.bp, 0);
                }
                J2T_DROP(self);
                continue;
            }

            /* the quote of the first key */
            case '"':
            {
                xprintf("[J2T_OBJ_0] start\n");
                // set parent state
                vt->st = J2T_OBJ;
                j2t_key(true);
                continue;
            }
            }
        }

        /* objects */
        case J2T_OBJ:
        {
            xprintf("[J2T_OBJ] 0\n");
            switch (ch)
            {
            case '}':
                xprintf("[J2T_OBJ] end\n");
                // update backtrace position, in case of OOM return
                bt.jp = *p - 1;
                if (dc->type == TTYPE_STRUCT)
                {
                    // last field value is null, so must unwind written field header
                    if (null_val)
                    {
                        null_val = false;
                        bm_set_req(vt->ex.es.reqs, lastField->ID, lastField->required);
                        buf->len = unwindPos;
                    }
                    // check if all required/default fields got written
                    J2T_STORE(j2t_write_unset_fields(self, buf, vt->ex.es.sd, vt->ex.es.reqs, flag, *p - 1));
                    bm_free_reqs(&self->reqs_cache, vt->ex.es.reqs.len);
                    // http-mapping enalbed and some fields remain not written
                    if (flag & F_ENABLE_HM && self->field_cache.len != 0)
                    {
                        // return back to go handler
                        xprintf("[J2T_OBJ] fallback http-mapping pos: %d\n", *p);
                        WRAP_ERR0(ERR_HM_END, *p);
                        // NOTICE: should tb_write_struct_end(), then J2T_DROP(self) in Go handler
                    }
                    else
                    {
                        J2T_STORE(tb_write_struct_end(buf));
                    }
                }
                else
                {
                    if (!null_val)
                    {
                        vt->ex.ec.size += 1;
                    }
                    else
                    {
                        // last element value is null, so must rewind to written key
                        null_val = false;
                        buf->len = unwindPos;
                    }
                    tb_write_map_end(buf, vt->ex.ec.bp, vt->ex.ec.size);
                }
                J2T_DROP(self);
                continue;
            case ',':
                xprintf("[J2T_OBJ] add\n");
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
            default:
                WRAP_ERR2(ERR_INVAL, ch, J2T_OBJ);
            }
        }

            /* object keys */
        case J2T_KEY:
        {
            xprintf("[J2T_KEY] 0\n");
            J2T_CHAR('"', J2T_KEY);
            j2t_key(false);
            continue;
        }

        /* object element */
        case J2T_ELEM:
        {
            xprintf("[J2T_ELEM] 0\n");
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
            xprintf("[J2T_VAL] value mapping\n");
            J2T_STORE_NEXT(j2t_field_vm(self, buf, src, p, vt));
            continue;
        }
        // normal value
        xprintf("[J2T_VAL] switch\n");
        /* simple values */
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
            xprintf("[J2T_VAL] number\n");
            *p -= 1;
            J2T_STORE_NEXT(j2t_number(buf, dc, src, p, &self->jt));
            break;
        }
        case 'n':
        {
            xprintf("[J2T_VAL] null\n");
            long s = *p;
            J2T_XERR(advance_dword(src, p, 1, *p - 1, VS_NULL), s);
            if (IS_STATE_FIELD(st) && unlikely(vt->ex.ef.f->required == REQ_REQUIRED))
            {
                WRAP_ERR(ERR_NULL_REQUIRED, vt->ex.ef.f->ID);
            }
            null_val = true;
            break;
        }
        case 't':
        {
            xprintf("[J2T_VAL] true\n");
            long s = *p;
            J2T_XERR(advance_dword(src, p, 1, *p - 1, VS_TRUE), s);
            J2T_EXP(TTYPE_BOOL, dc->type);
            J2T_STORE_NEXT(tb_write_bool(buf, true));
            break;
        }
        case 'f':
        {
            xprintf("[J2T_VAL] false\n");
            long s = *p;
            J2T_XERR(advance_dword(src, p, 0, *p - 1, VS_ALSE), s);
            J2T_EXP(TTYPE_BOOL, dc->type);
            J2T_STORE_NEXT(tb_write_bool(buf, false));
            break;
        }
        case '[':
        {
            xprintf("[J2T_VAL] array\n");
            J2T_EXP2(TTYPE_LIST, TTYPE_SET, dc->type);
            J2TState *v = &self->vt[self->sp];
            // TODO: since the encoding of list and set, we call only one type's func here.
            // add thrift type choose for corresponding func
            v->ex.ec.size = 0;
            xprintf("[J2T_ARRAY]ex size: %d, vt: %d, vt.ex.ec: %d, vt.ex.ec.bp: %d, vt.ex.ec.size: %d\n", sizeof(J2TExtra), vt, &vt->ex.ec, &vt->ex.ec.bp, &vt->ex.ec.size);
            J2T_STORE_NEXT(tb_write_list_begin(buf, dc->elem->type, &v->ex.ec.bp)); // pass write-back address bp to v.ex.ec.bp
            J2T_PUSH_EX(self, J2T_ARR_0, *p, dc, v->ex);                            // pass bp to state
            break;
        }
        case '{':
        {
            xprintf("[J2T_VAL] object\n");
            J2T_EXP2(TTYPE_STRUCT, TTYPE_MAP, dc->type);
            J2TState *v = &self->vt[self->sp];
            if (dc->type == TTYPE_STRUCT)
            {
                xprintf("[J2T_VAL] object struct name: %s\n", &dc->st->name);
                tb_write_struct_begin(buf);
                J2TExtra_Struct es = {.sd = dc->st};
                v->ex.es = es;
                J2T_PUSH(self, J2T_OBJ_0, *p, dc);
                J2T_STORE(bm_malloc_reqs(&self->reqs_cache, dc->st->reqs, &v->ex.es.reqs, *p));
                if ((flag & F_ENABLE_HM) != 0 && dc->st->hms.len > 0)
                {
                    xprintf("[J2T_VAL] HTTP_MAPPING begin pos: %d\n", *p - 1);
                    WRAP_ERR0(ERR_HM, *p - 1);
                }
            }
            else
            {
                J2TExtra_Cont ec = {};
                v->ex.ec = ec;
                J2T_STORE_NEXT(tb_write_map_begin(buf, dc->key->type, dc->elem->type, &v->ex.ec.bp)); // pass write-back address bp to v.ex.ec.bp
                J2T_PUSH(self, J2T_OBJ_0, *p, dc);
            }
            break;
        }
        case '"':
        {
            xprintf("[J2T_VAL] string\n");
            if (dc->type == TTYPE_STRING)
            {
                if unlikely ((flag & F_NO_BASE64) == 0 && memeq(dc->name.buf, "binary", 6))
                {
                    J2T_STORE_NEXT(j2t_binary(buf, src, p, 0));
                }
                else
                {
                    J2T_STORE_NEXT(j2t_string(buf, src, p, 0));
                }
            }
            else if ((flag & F_ENABLE_I2S) != 0 && (dc->type == TTYPE_I64 || dc->type == TTYPE_I32 || dc->type == TTYPE_I16 || dc->type == TTYPE_BYTE))
            {
                J2T_STORE_NEXT(j2t_number(buf, dc, src, p, &self->jt));
                long x = *p;
                if (x >= src->len)
                {
                    WRAP_ERR(ERR_EOF, J2T_VAL);
                }
                // must end with a comma
                if (src->buf[x] != '"')
                {
                    WRAP_ERR2(ERR_INVAL, src->buf[x], J2T_VAL);
                }
                *p += 1;
            }
            else
            {
                WRAP_ERR2(ERR_DISMATCH_TYPE, dc->type, TTYPE_STRING);
            }
            break;
        }
        case 0:
            WRAP_ERR(ERR_EOF, J2T_VAL);
        default:
            WRAP_ERR2(ERR_INVAL, ch, J2T_VAL);
        }
    }

    xprintf("[J2T] return\n");
    /* all done */
    return 0;
}