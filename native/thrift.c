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
#include "thrift.h"
#include "test/xprintf.h"

inline uint64_t buf_malloc(_GoSlice *buf, size_t size)
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

inline uint64_t bm_malloc_reqs(_GoSlice *cache, ReqBitMap src, ReqBitMap *copy, long p)
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


uint64_t j2t2_write_unset_fields(vt_ienc *ienc, const tStructDesc *st, ReqBitMap reqs, uint64_t flags, long p)
{
    J2TStateMachine *self = VT_J2TSM(ienc->base);
    _GoSlice *buf = ienc->base.resv1;
    void *tproto = VT_TC_STATE(ienc->base);

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
                    J2T_ZERO(ienc->method.write_field_begin(tproto, f->type->type, f->ID));
                    J2T_ZERO(ienc->extra.write_default_or_empty(tproto, f, p));
                    J2T_ZERO(ienc->method.write_field_end(tproto));
                }
                v >>= 1;
            }
        }
        s++;
    }

    return 0;
}

uint64_t j2t2_number(vt_ienc *ienc, const tTypeDesc *desc, const _GoString *src, long *p, JsonState *ret)
{
    void *tproto = VT_TC_STATE(ienc->base);
    long s = *p;
    vnumber(src, p, ret);
    xprintf("[j2t2_number] p:%d, ret.vt:%d, ret.iv:%d, ret.dv:%f\n", s, ret->vt, ret->iv, ret->dv);
    if (ret->vt < 0)
        WRAP_ERR(-ret->vt, s);
    // FIXME: check double-integer overflow
    switch (desc->type)
    {
    case TTYPE_I08:
        return ienc->method.write_byte(tproto,
            ret->vt == V_INTEGER
                ? (uint8_t)ret->iv
                : (uint8_t)ret->dv);
    case TTYPE_I16:
        return ienc->method.write_i16(tproto,
            ret->vt == V_INTEGER
                ? (int16_t)ret->iv
                : (int16_t)ret->dv);
    case TTYPE_I32:
        return ienc->method.write_i32(tproto,
            ret->vt == V_INTEGER
                ? (int32_t)ret->iv
                : (int32_t)ret->dv);
    case TTYPE_I64:
        return ienc->method.write_i64(tproto,
            ret->vt == V_INTEGER
                ? (int64_t)ret->iv
                : (int64_t)ret->dv);
    case TTYPE_DOUBLE:
        return ienc->method.write_double(tproto,
            ret->vt == V_DOUBLE
                ? (double)ret->dv
                : (double)((uint64_t)ret->iv));
    }
    WRAP_ERR2(ERR_DISMATCH_TYPE, desc->type, V_INTEGER);
}

uint64_t j2t2_string(vt_ienc *ienc, const _GoString *src, long *p, uint64_t flags)
{
    _GoSlice *buf = ienc->base.resv1;
    void *tproto = VT_TC_STATE(ienc->base);

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
        J2T_ZERO(ienc->extra.write_data_count(tproto, n));
        // unescape string
        size_t o = buf->len;
        ep = -1;
        J2T_ZERO( buf_malloc(buf, n) );
        xprintf("[j2t_string] unquote o:%d, n:%d, ep:%d, flags:%d\n", o, n, ep, flags);
        ssize_t l = unquote(&src->buf[s], n, &buf->buf[o], &ep, flags);
        xprintf("[j2t_string] unquote end, l:%d\n", l);
        if (l < 0)
            WRAP_ERR(-l, s);
        buf->len = o + l;
    }
    else
        J2T_ZERO(ienc->method.write_string(tproto, &src->buf[0], src->len));
    return 0;
}

uint64_t j2t2_binary(vt_ienc *ienc, const _GoString *src, long *p, uint64_t flags)
{
    _GoSlice *buf = ienc->base.resv1;
    void *tproto = VT_TC_STATE(ienc->base);

    long s = *p;
    int64_t ep;
    ssize_t e = advance_string(src, s, &ep);
    J2T_XERR(e, s);
    *p = e;
    size_t n = e - s - 1;
    xprintf("[j2t_binary]s:%d, p:%d, ep:%d, n:%d\n", s, *p, ep, n);
    // write length
    J2T_ZERO( ienc->extra.write_data_count(tproto, n) );
    // write content
    ssize_t l = b64decode(buf, &src->buf[s], n, 0);
    xprintf("[j2t_binary] base64 endm l:%d\n", l);
    if (l < 0)
        WRAP_ERR(ERR_DECODE_BASE64, -l - 1);
    return 0;
}

uint64_t j2t2_map_key(vt_ienc *ienc, const char *sp, size_t n, const tTypeDesc *dc, JsonState *js, long p)
{
    void *tproto = VT_TC_STATE(ienc->base);
    xprintf("[j2t2_map_key]\n");
    switch (dc->type)
    {
    case TTYPE_STRING:
    {
        J2T_ZERO(ienc->method.write_string(tproto, sp, n));
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

tFieldDesc *j2t2_find_field_key(vt_ienc *ienc, const _GoString *key, const tStructDesc *st)
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

uint64_t j2t2_read_key(vt_ienc *ienc, const _GoString *src, long *p, const char **spp, size_t *knp)
{
    J2TStateMachine *self = VT_J2TSM(ienc->base);
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

uint64_t j2t2_field_vm(vt_ienc *ienc, const _GoString *src, long *p, J2TState *vt)
{
    J2TStateMachine *self = VT_J2TSM(ienc->base);
    _GoSlice *buf = ienc->base.resv1;
    void *tproto = VT_TC_STATE(ienc->base);

    tFieldDesc *f = vt->ex.ef.f;
    xprintf("[j2t_field_vm] f->ID:%d, f->type->type:%d, p:%d \n", f->ID, f->type->type, *p);
    // if it is inlined value-mapping, write field tag first
    if (f->vm <= VM_INLINE_MAX)
    {
        J2T_ZERO( ienc->method.write_field_begin(tproto, f->type->type, f->ID) );
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
                    J2T_ZERO(ienc->extra.write_default_or_empty(tproto, f, *p));
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
                J2T_ZERO(ienc->method.write_string(tproto, &src->buf[0], *p - s));
                return 0;
            }
            case TTYPE_I64:
            {
                int64_t v = self->jt.vt == V_INTEGER
                    ? (int64_t)self->jt.iv
                    : (int64_t)self->jt.dv;
                J2T_ZERO(ienc->method.write_i64(tproto, v));
                break;
            }
            case TTYPE_I32:
            {
                xprintf("[j2t2_field_vm] TTYPE_I32\n");
                int32_t v = self->jt.vt == V_INTEGER
                    ? (int32_t)self->jt.iv
                    : (int32_t)self->jt.dv;
                J2T_ZERO(ienc->method.write_i32(tproto, v));
                break;
            }
            case TTYPE_I16:
            {
                int16_t v = self->jt.vt == V_INTEGER
                    ? (int16_t)self->jt.iv
                    : (int16_t)self->jt.dv;
                J2T_ZERO(ienc->method.write_i16(tproto, v));
                break;
            }
            case TTYPE_I08:
            {
                char v = self->jt.vt == V_INTEGER
                    ? (char)self->jt.iv
                    : (char)self->jt.dv;
                J2T_ZERO(ienc->method.write_byte(tproto, v));
                break;
            }
            case TTYPE_DOUBLE:
            {
                double v = self->jt.vt == V_INTEGER
                    ? (double)((uint64_t)self->jt.iv)
                    : (double)self->jt.dv;
                J2T_ZERO(ienc->method.write_double(tproto, v));
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

#define j2t2_key(obj0)                                                                                                  \
    do                                                                                                                  \
    {                                                                                                                   \
        const char *sp;                                                                                                 \
        size_t kn;                                                                                                      \
        J2T_STORE(j2t2_read_key(ienc, src, p, &sp, &kn));                                                              \
        if (dc->type == TTYPE_MAP)                                                                                      \
        {                                                                                                               \
            unwindPos = buf->len;                                                                                       \
            J2T_STORE(j2t2_map_key(ienc, sp, kn, dc->key, &self->jt, *p));                                             \
            if (obj0)                                                                                                   \
            {                                                                                                           \
                vt->ex.ec.size = 0;                                                                                     \
                J2T_PUSH(self, J2T_ELEM, *p, dc->elem);                                                                 \
            }                                                                                                           \
            else                                                                                                        \
            {                                                                                                           \
                J2T_REPL(self, J2T_ELEM, *p, dc->elem);                                                                 \
            }                                                                                                           \
        }                                                                                                               \
        else                                                                                                            \
        {                                                                                                               \
            J2TExtra *pex;                                                                                              \
            if (obj0)                                                                                                   \
            {                                                                                                           \
                pex = &vt->ex;                                                                                          \
            }                                                                                                           \
            else                                                                                                        \
            {                                                                                                           \
                pex = &self->vt[self->sp - 2].ex;                                                                       \
            }                                                                                                           \
            _GoString tmp_str = (_GoString){.buf = sp, .len = kn};                                                      \
            tFieldDesc *f = j2t2_find_field_key(ienc, &tmp_str, dc->st);                                               \
            if (f == NULL)                                                                                              \
            {                                                                                                           \
                xprintf("[J2T_KEY] unknown field %s\n", &tmp_str);                                                      \
                if ((flag & F_ALLOW_UNKNOWN) == 0)                                                                      \
                {                                                                                                       \
                    WRAP_ERR(ERR_UNKNOWN_FIELD, kn);                                                                    \
                }                                                                                                       \
                if (obj0)                                                                                               \
                {                                                                                                       \
                    J2T_PUSH(self, J2T_ELEM | STATE_SKIP, *p, NULL);                                                    \
                }                                                                                                       \
                else                                                                                                    \
                {                                                                                                       \
                    J2T_REPL(self, J2T_ELEM | STATE_SKIP, *p, NULL);                                                    \
                }                                                                                                       \
            }                                                                                                           \
            else                                                                                                        \
            {                                                                                                           \
                if unlikely ((flag & F_ENABLE_HM) && (f->http_mappings.len != 0) && !bm_is_set(pex->es.reqs, f->ID))    \
                {                                                                                                       \
                    xprintf("[J2T_KEY] skip field %s\n", &tmp_str);                                                     \
                    if (obj0)                                                                                           \
                    {                                                                                                   \
                        J2T_PUSH(self, J2T_ELEM | STATE_SKIP, *p, f->type);                                             \
                    }                                                                                                   \
                    else                                                                                                \
                    {                                                                                                   \
                        J2T_REPL(self, J2T_ELEM | STATE_SKIP, *p, f->type);                                             \
                    }                                                                                                   \
                }                                                                                                       \
                else                                                                                                    \
                {                                                                                                       \
                    J2TExtra ex;                                                                                        \
                    ex.ef.f = f;                                                                                        \
                    uint64_t vm = STATE_VM;                                                                             \
                    if ((flag & F_ENABLE_VM) == 0 || f->vm == VM_NONE)                                                  \
                    {                                                                                                   \
                        vm = STATE_FIELD;                                                                               \
                        unwindPos = buf->len;                                                                           \
                        lastField = f;                                                                                  \
                        J2T_STORE(ienc->method.write_field_begin(VT_TC_STATE(ienc->base), f->type->type, f->ID));                \
                    }                                                                                                   \
                    xprintf("[J2T_KEY] vm: %d\n", vm);                                                                  \
                    if (obj0)                                                                                           \
                    {                                                                                                   \
                        J2T_PUSH_EX(self, J2T_ELEM | vm, *p, f->type, ex);                                              \
                    }                                                                                                   \
                    else                                                                                                \
                    {                                                                                                   \
                        J2T_REPL_EX(self, J2T_ELEM | vm, *p, f->type, ex);                                              \
                    }                                                                                                   \
                }                                                                                                       \
                bm_set_req(pex->es.reqs, f->ID, REQ_OPTIONAL);                                                          \
            }                                                                                                           \
        }                                                                                                               \
    } while (0)


inline uint64_t j2t2_fsm_exec(vt_ienc *ienc, const _GoString *src, uint64_t flag)
{
    #define J2TSTATE_CUR    self->vt[self->sp - 1]
    #define J2TSTATE_NEXT   self->vt[self->sp]

    J2TStateMachine *self = VT_J2TSM(ienc->base);
    _GoSlice *buf = VT_OUTBUF(ienc->base);

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
                J2T_ZERO( ienc->method.write_list_end(VT_TC_STATE(ienc->base), vt->ex.ec.bp, 0) );
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
                J2T_ZERO( ienc->method.write_list_end(VT_TC_STATE(ienc->base), vt->ex.ec.bp, vt->ex.ec.size) );
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
                    J2T_STORE(j2t2_write_unset_fields(ienc, vt->ex.es.sd, vt->ex.es.reqs, flag, *p - 1));
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
                        J2T_STORE(ienc->method.write_struct_end(VT_TC_STATE(ienc->base)));
                    }
                }
                else
                {
                    J2T_STORE( ienc->method.write_struct_end(VT_TC_STATE(ienc->base)) );
                }
                J2T_DROP(self);
                continue;
            }

            //  the quote of the first key
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
                    J2T_STORE(j2t2_write_unset_fields(ienc, vt->ex.es.sd, vt->ex.es.reqs, flag, *p - 1));
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
                        J2T_STORE(ienc->method.write_struct_end(VT_TC_STATE(ienc->base)));
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
                    ienc->method.write_map_end(VT_TC_STATE(ienc->base), vt->ex.ec.bp, vt->ex.ec.size);
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
            J2T_STORE_NEXT(j2t2_field_vm(ienc, src, p, vt));
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
            J2T_STORE_NEXT(j2t2_number(ienc, dc, src, p, &self->jt));
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
            J2T_STORE_NEXT(ienc->method.write_bool(VT_TC_STATE(ienc->base), true));
            break;
        }

        case 'f':
        {
            xprintf("[J2T2_VAL] false\n");
            long s = *p;
            J2T_XERR(advance_dword(src, p, 1, *p - 1, VS_ALSE), s);
            J2T_EXP(TTYPE_BOOL, dc->type);
            J2T_STORE_NEXT(ienc->method.write_bool(VT_TC_STATE(ienc->base), false));
            break;
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
            // pass write-back address bp to v.ex.ec.bp
            J2T_STORE_NEXT( ienc->method.write_list_begin(VT_TC_STATE(ienc->base), dc->elem->type, &v->ex.ec.bp) );
            // pass bp to state
            J2T_PUSH_EX(self, J2T_ARR_0, *p, dc, v->ex);
            break;
        }
        
        case '{':
        {
            xprintf("[J2T2_VAL] object\n");
            J2T_EXP2(TTYPE_STRUCT, TTYPE_MAP, dc->type);
            J2TState *v = &J2TSTATE_NEXT;
            if (dc->type == TTYPE_STRUCT)
            {
                xprintf("[J2T2_VAL] object struct name: %s\n", &dc->st->name);
                ienc->method.write_struct_begin(VT_TC_STATE(ienc->base));
                v->ex.es.sd = dc->st;
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
                v->ex.ec.size = 0;                
                // pass write-back address bp to v.ex.ec.bp
                J2T_STORE_NEXT( ienc->method.write_map_begin(VT_TC_STATE(ienc->base), dc->key->type, dc->elem->type, &v->ex.ec.bp) );
                J2T_PUSH(self, J2T_OBJ_0, *p, dc);
            }
            break;
        }

        case '"':
        {
            xprintf("[J2T2_VAL] string\n");
            if (dc->type == TTYPE_STRING)
            {
                if unlikely ((flag & F_NO_BASE64) == 0 && memeq(dc->name.buf, "binary", 6))
                {
                    J2T_STORE_NEXT(j2t2_binary(ienc, src, p, 0));
                }
                else
                {
                    J2T_STORE_NEXT(j2t2_string(ienc, src, p, 0));
                }
            }
            else if ((flag & F_ENABLE_I2S) != 0 && (dc->type == TTYPE_I64 || dc->type == TTYPE_I32 || dc->type == TTYPE_I16 || dc->type == TTYPE_BYTE))
            {
                J2T_STORE_NEXT(j2t2_number(ienc, dc, src, p, &self->jt));
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
            break;
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