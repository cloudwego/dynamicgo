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

#ifndef THRIFT_H
#define THRIFT_H

#define F_ALLOW_UNKNOWN 1ull
#define F_WRITE_DEFAULT (1ull << 1)
#define F_ENABLE_VM (1ull << 2)
#define F_ENABLE_HM (1ull << 3)
#define F_ENABLE_I2S (1ull << 4)
#define F_WRITE_REQUIRE (1ull << 5)
#define F_NO_BASE64 (1ull << 6)
#define F_WRITE_OPTIONAL (1ull << 7)
#define F_TRACE_BACK (1ull << 8)

#define THRIFT_VERSION_MASK 0xffff0000ul
#define THRIFT_VERSION_1 0x80010000ul

typedef int32_t tmsg;

#define TMESSAGE_TYPE_INVALID 0
#define TMESSAGE_TYPE_CALL 1
#define TMESSAGE_TYPE_REPLY 2
#define TMESSAGE_TYPE_EXCEPTION 3
#define TMESSAGE_TYPE_ONEWAY 4

typedef uint8_t ttype;

#define TTYPE_STOP (ttype)0
#define TTYPE_VOID (ttype)1
#define TTYPE_BOOL (ttype)2
#define TTYPE_BYTE (ttype)3
#define TTYPE_I08 (ttype)3
#define TTYPE_DOUBLE (ttype)4
#define TTYPE_I16 (ttype)6
#define TTYPE_I32 (ttype)8
#define TTYPE_I64 (ttype)10
#define TTYPE_STRING (ttype)11
#define TTYPE_UTF7 (ttype)11
#define TTYPE_STRUCT (ttype)12
#define TTYPE_MAP (ttype)13
#define TTYPE_SET (ttype)14
#define TTYPE_LIST (ttype)15
#define TTYPE_UTF8 (ttype)16
#define TTYPE_UTF16 (ttype)17

typedef uint16_t vm_em;
#define VM_NONE 0
#define VM_JSCONV 101 // NOTICE: must be same with internal/native/types.VM_JSCONV
#define VM_INLINE_MAX 255

typedef struct tTypeDesc tTypeDesc;

typedef struct tFieldDesc tFieldDesc;

typedef struct
{
    tFieldDesc **buf;
    size_t len;
    size_t cap;
    tFieldDesc **buf_all;
    size_t len_all;
    size_t cap_all;
} FieldIdMap;

typedef struct
{
    size_t max_key_len;
    GoSlice all;
    TrieTree *trie;
    HashMap *hash;
} tFieldNameMap;

typedef struct
{
    GoEface go_val;
    GoString json_val;
    GoString thrift_binary;
} tDefaultValue;

struct tFieldDesc
{
    bool is_request_base;
    bool is_response_base;
    req_em required;
    vm_em vm;
    tid ID;
    tTypeDesc *type;
    tDefaultValue *default_value;
    GoString name;
    GoString alias;
    GoIface value_mappings;
    GoSlice http_mappings;
    GoSlice annotations;
};

typedef struct
{
    tid base_id;
    GoString name;
    FieldIdMap ids;
    tFieldNameMap names;

    // WARN: Here is a `GoSlice*` pointer actually (allocated 3 QUAD memory from golang)
    // while we use `ReqBitMap*` pointer (2 QUAD memory) for purpose of conformity to J2TState.ex.es.reqs,
    // It won't affect the calculating results AS LONG AS IT IS STATICALLY USED.
    ReqBitMap reqs;
    GoSlice hms;
    GoSlice annotations;
} tStructDesc;

struct tTypeDesc
{
    ttype type;
    GoString name;
    tTypeDesc *key;
    tTypeDesc *elem;
    tStructDesc *st;
};

typedef struct
{
    size_t bp;
    size_t size;
} J2TExtra_Cont;

typedef struct
{
    const tStructDesc *sd;
    ReqBitMap reqs;
} J2TExtra_Struct;

typedef struct
{
    bool skip;
    tFieldDesc *f;
} J2TExtra_Field;

typedef union
{
    J2TExtra_Cont ec;
    J2TExtra_Struct es;
    J2TExtra_Field ef;
} J2TExtra;

typedef struct
{
    size_t st;
    size_t jp;
    const tTypeDesc *td;
    J2TExtra ex;
} J2TState;

typedef struct
{
    int32_t *buf;
    size_t len;
    size_t cap;
} Int32Slice;

typedef struct
{
    int32_t id;
    uint32_t start;
    uint32_t end;
} FieldVal;

typedef struct
{
    size_t sp;
    JsonState jt;
    GoSlice reqs_cache;
    GoSlice key_cache;
    J2TState vt[MAX_RECURSE];
    StateMachine sm;
    Int32Slice field_cache;
    FieldVal fval_cache;
} J2TStateMachine;

#define SIZE_J2TEXTRA sizeof(J2TExtra)

uint64_t tb_write_i64(GoSlice *buf, int64_t v);

uint64_t j2t_fsm_exec(J2TStateMachine *self, GoSlice *buf, const GoString *src, uint64_t flag);

#define J2T_VAL 0
#define J2T_ARR 1
#define J2T_OBJ 2
#define J2T_KEY 3
#define J2T_ELEM 4
#define J2T_ARR_0 5
#define J2T_OBJ_0 6
#define J2T_VM 16

#define J2T_ST(st) ((st)&0xfffful)
#define J2T_EX(st) ((st)&0xffff0000ul)

#define STATE_FIELD (1ull << 16)
#define STATE_SKIP (1ull << 17)
#define STATE_VM (1ull << 18)
#define IS_STATE_FIELD(st) (st & STATE_FIELD) != 0
#define IS_STATE_SKIP(st) (st & STATE_SKIP) != 0
#define IS_STATE_VM(st) (st & STATE_VM) != 0

#define ERR_WRAP_SHIFT_CODE 8
#define ERR_WRAP_SHIFT_POS 32

#define WRAP_ERR_POS(e, v, p)                                                                                                             \
    do                                                                                                                                    \
    {                                                                                                                                     \
        xprintf("[ERROR_%d]: %d\n", e, v);                                                                                                \
        return (((uint64_t)(v) << (ERR_WRAP_SHIFT_CODE + ERR_WRAP_SHIFT_POS)) | (((uint64_t)(p) << ERR_WRAP_SHIFT_CODE) | (uint8_t)(e))); \
    } while (0)

#define WRAP_ERR(e, v) WRAP_ERR_POS(e, v, *p)

#define WRAP_ERR2(e, vh, vl) WRAP_ERR(e, ((uint32_t)(vh) << ERR_WRAP_SHIFT_CODE | (uint8_t)(vl)))

#define WRAP_ERR0(e, v)                                                   \
    do                                                                    \
    {                                                                     \
        xprintf("[ERROR_%d]: %d\n", e, v);                                \
        return (((uint64_t)(v) << (ERR_WRAP_SHIFT_CODE)) | (uint8_t)(e)); \
    } while (0)

#define J2T_EXP(e, v)                       \
    if unlikely (e != v)                    \
    {                                       \
        WRAP_ERR2(ERR_DISMATCH_TYPE, v, e); \
    }

#define J2T_EXP2(e1, e2, v)                                                                                                                                     \
    if unlikely (e1 != v && e2 != v)                                                                                                                            \
    {                                                                                                                                                           \
        WRAP_ERR(ERR_DISMATCH_TYPE2, ((uint32_t)(v) << (ERR_WRAP_SHIFT_CODE + ERR_WRAP_SHIFT_CODE) | ((uint16_t)(e2) << ERR_WRAP_SHIFT_CODE) | (uint8_t)(e1))); \
    }

#define J2T_ZERO(e)           \
    do                        \
    {                         \
        uint64_t re = e;      \
        if unlikely (re != 0) \
        {                     \
            return re;        \
        }                     \
    } while (0)

#define J2T_DROP(v) (v)->sp--

#define J2T_CHAR(c, s)                  \
    do                                  \
    {                                   \
        if unlikely (ch != (c))         \
            WRAP_ERR2(ERR_INVAL, c, s); \
    } while (0)

#define J2T_XERR(v, s)            \
    do                            \
    {                             \
        int64_t r = (int64_t)(v); \
        if unlikely (r < 0)       \
            WRAP_ERR(-r, s);      \
    } while (0)

#define J2T_INIT(self, v) \
    self->sp = 1;         \
    self->vt[0] = v;

#define J2T_REPL(self, a, b, c)                 \
    do                                          \
    {                                           \
        J2TState *xp = &self->vt[self->sp - 1]; \
        xp->st = a;                             \
        xp->jp = b;                             \
        xp->td = c;                             \
        DEBUG_PRINT_STATE(1, self->sp - 1, *p); \
    } while (0)

#define J2T_REPL_EX(self, a, b, c, d)              \
    do                                             \
    {                                              \
        J2TState *xp = &self->vt[self->sp - 1];    \
        xp->st = a;                                \
        xp->jp = b;                                \
        xp->td = c;                                \
        xp->ex = d;                                \
        DEBUG_PRINT_STATE_EX(2, self->sp - 1, *p); \
    } while (0)

#define J2T_PUSH(self, a, b, c)               \
    if unlikely (self->sp >= MAX_RECURSE)     \
    {                                         \
        WRAP_ERR(ERR_RECURSE_MAX, self->sp);  \
    }                                         \
    else                                      \
    {                                         \
        J2TState *xp = &self->vt[self->sp++]; \
        xp->st = a;                           \
        xp->jp = b;                           \
        xp->td = c;                           \
    }                                         \
    DEBUG_PRINT_STATE(3, self->sp - 1, *p);

#define J2T_PUSH_EX(self, a, b, c, d)         \
    if unlikely (self->sp >= MAX_RECURSE)     \
    {                                         \
        WRAP_ERR(ERR_RECURSE_MAX, self->sp);  \
    }                                         \
    else                                      \
    {                                         \
        J2TState *xp = &self->vt[self->sp++]; \
        xp->st = a;                           \
        xp->jp = b;                           \
        xp->td = c;                           \
        xp->ex = d;                           \
    }                                         \
    DEBUG_PRINT_STATE_EX(4, self->sp - 1, *p);

#define J2T_STORE(f)                               \
    do                                             \
    {                                              \
        uint64_t ret = f;                          \
        if unlikely (ret != 0)                     \
        {                                          \
            self->vt[self->sp - 1] = bt;           \
            buf->len = wp;                         \
            DEBUG_PRINT_STATE(5, self->sp - 1, *p) \
            return ret;                            \
        }                                          \
    } while (0)

#define J2T_STORE_NEXT(f)                          \
    do                                             \
    {                                              \
        uint64_t ret = f;                          \
        if unlikely (ret != 0)                     \
        {                                          \
            self->vt[self->sp++] = bt;             \
            buf->len = wp;                         \
            DEBUG_PRINT_STATE(6, self->sp - 1, *p) \
            return ret;                            \
        }                                          \
    } while (0)

#define DEBUG_PRINT_STATE(i, sp, p) \
    xprintf("[DEBUG_PRINT_STATE_%d] STATE{sp:%d, st:%d, jp:%d, td:%s} POS:%d CHAR:%c BUF{\n\tbuf:%l,\n\tlen:%d, cap:%d}\n", i, sp, vt->st, vt->jp, vt->td == NULL ? &(GoString){} : &vt->td->name, p, ch, buf, buf->len, buf->cap);

#define DEBUG_PRINT_STATE_EX(i, sp, p) \
    xprintf("[DEBUG_PRINT_STATE_EX_%d] STATE{sp:%d, st:%d, jp:%d, td:%s, ex:[%d,%d,%d]} POS:%d CHAR:%c BUF{\n\tbuf:%l,\n\tlen:%d, cap:%d}\n", i, sp, vt->st, vt->jp, vt->td == NULL ? &(GoString){} : &vt->td->name, *(uint64_t *)(&vt->ex), *((uint64_t *)(&vt->ex) + 1), *((uint64_t *)(&vt->ex) + 2), p, ch, buf, buf->len, buf->cap);

#endif // THRIFT_H
