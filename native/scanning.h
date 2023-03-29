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

#include "native.h"

#ifndef SCANNING_H
#define SCANNING_H

#define FSM_VAL 0
#define FSM_ARR 1
#define FSM_OBJ 2
#define FSM_KEY 3
#define FSM_ELEM 4
#define FSM_ARR_0 5
#define FSM_OBJ_0 6

#define FSM_DROP(v) (v)->sp--
#define FSM_REPL(v, t) (v)->vt[(v)->sp - 1] = (t)

#define FSM_CHAR(c)            \
    do                         \
    {                          \
        if (ch != (c))         \
            return -ERR_INVAL; \
    } while (0)
#define FSM_XERR(v)   \
    do                \
    {                 \
        long r = (v); \
        if (r < 0)    \
            return r; \
    } while (0)

#define VALID_DEFAULT 0 // basic validate, except JSON string.
#define VALID_FULL 1    // also validate JSON string, including control chars or invalid UTF-8.

#define check_bits(mv)                            \
    if (unlikely((v = mv & (mv - 1)) != 0))       \
    {                                             \
        return -(sp - ss + __builtin_ctz(v) + 1); \
    }

#define check_sidx(iv)     \
    if (likely(iv == -1))  \
    {                      \
        iv = sp - ss - 1;  \
    }                      \
    else                   \
    {                      \
        return -(sp - ss); \
    }

#define check_vidx(iv, mv)                             \
    if (mv != 0)                                       \
    {                                                  \
        if (likely(iv == -1))                          \
        {                                              \
            iv = sp - ss + __builtin_ctz(mv);          \
        }                                              \
        else                                           \
        {                                              \
            return -(sp - ss + __builtin_ctz(mv) + 1); \
        }                                              \
    }

#define FSM_INIT(self, v) \
    self->sp = 1;         \
    self->vt[0] = v;

#define FSM_PUSH(self, v)         \
    if (self->sp >= MAX_RECURSE)  \
    {                             \
        return -ERR_RECURSE_MAX;  \
    }                             \
    else                          \
    {                             \
        self->vt[self->sp++] = v; \
        long r = 0;               \
    }

#define FSM_ZERO(e)    \
    do                 \
    {                  \
        int re = e;    \
        if (re != 0)   \
        {              \
            return re; \
        }              \
    } while (0)

char advance_ns(const _GoString *src, long *p);

int64_t advance_dword(const _GoString *src, long *p, long dec, int64_t ret, uint32_t val);

#endif // SCANNING_H