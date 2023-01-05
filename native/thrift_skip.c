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
#include "test/xprintf.h"

#define ETAG -1
#define EEOF -2
#define ESTACK -3
#define MAX_STACK 1024

#define T_bool 2
#define T_i8 3
#define T_double 4
#define T_i16 6
#define T_i32 8
#define T_i64 10
#define T_string 11
#define T_struct 12
#define T_map 13
#define T_set 14
#define T_list 15
#define T_list_elem 0xfe
#define T_map_pair 0xff

typedef struct
{
    uint8_t t;
    uint8_t k;
    uint8_t v;
    uint32_t n;
} skipbuf_t;

static const char WireTags[256] = {
    [T_bool] = 1,
    [T_i8] = 1,
    [T_double] = 1,
    [T_i16] = 1,
    [T_i32] = 1,
    [T_i64] = 1,
    [T_string] = 1,
    [T_struct] = 1,
    [T_map] = 1,
    [T_set] = 1,
    [T_list] = 1,
};

static const int8_t SkipSizeFixed[256] = {
    [T_bool] = 1,
    [T_i8] = 1,
    [T_double] = 8,
    [T_i16] = 2,
    [T_i32] = 4,
    [T_i64] = 8,
};

static inline int64_t u32be(const char *s)
{
    return __builtin_bswap32(*(const uint32_t *)s);
}

static inline int64_t u16be(const char *s)
{
    return __builtin_bswap16(*(const uint16_t *)s);
}

static inline char stpop(skipbuf_t *s, int64_t *p)
{
    if (s[*p].n == 0)
    {
        (*p)--;
        return 1;
    }
    else
    {
        s[*p].n--;
        return 0;
    }
}

static inline char stadd(skipbuf_t *s, int64_t *p, uint8_t t)
{
    if (++*p >= MAX_STACK)
    {
        return 0;
    }
    else
    {
        s[*p].t = t;
        s[*p].n = 0;
        return 1;
    }
}

static inline void mvbuf(const char **s, int64_t *n, int64_t *r, int64_t nb)
{
    *n -= nb;
    *r += nb;
    *s += nb;
}

int64_t tb_skip(skipbuf_t *st, const char *s, int64_t n, uint8_t t)
{
    int64_t nb;
    int64_t rv = 0;
    int64_t sp = 0;

    /* initialize the stack */
    st->n = 0;
    st->t = t;

    /* run until drain */
    while (sp >= 0)
    {
        xprintf("[T_%d] sp:%d, rv:%d, st:{t:%d, k:%d, e:%d, n:%d}\n", st[sp].t, sp, rv, st[sp].t, st[sp].k, st[sp].v, st[sp].n);
        // xprintf("[T_%d] sp:%d, rv:%d\n", st[sp].t, sp, rv);
        switch (st[sp].t)
        {
        default:
        {
            return ETAG;
        }

        /* simple fixed types */
        case T_bool:
        case T_i8:
        case T_double:
        case T_i16:
        case T_i32:
        case T_i64:
        {
            if ((nb = SkipSizeFixed[st[sp].t]) > n)
            {
                return EEOF;
            }
            else
            {
                stpop(st, &sp);
                mvbuf(&s, &n, &rv, nb);
                break;
            }
        }

        /* strings & binaries */
        case T_string:
        {
            if (n < 4)
            {
                return EEOF;
            }
            else if ((nb = u32be(s) + 4) > n)
            {
                return EEOF;
            }
            else
            {
                stpop(st, &sp);
                mvbuf(&s, &n, &rv, nb);
                break;
            }
        }

        /* structs */
        case T_struct:
        {
            int64_t nf;
            uint8_t vt;

            /* must have at least 1 byte */
            if (n < 1)
            {
                return EEOF;
            }

            /* check for end of tag */
            if ((vt = *s) == 0)
            {
                stpop(st, &sp);
                mvbuf(&s, &n, &rv, 1);
                continue;
            }

            xprintf("[T_struct] rv:%d, ft:%d, fid:%d\n", rv + 3, vt, u16be(s + 1));

            /* check for tag value */
            if (!(WireTags[vt]))
            {
                return ETAG;
            }

            /* fast-path for primitive fields */
            if ((nf = SkipSizeFixed[vt]) != 0)
            {
                if (n < nf + 3)
                {
                    return EEOF;
                }
                else
                {
                    mvbuf(&s, &n, &rv, nf + 3);
                    continue;
                }
            }

            /* must have more than 3 bytes (fields cannot have a size of zero), also skip the field ID cause we don't care */
            if (n <= 3)
            {
                return EEOF;
            }
            else if (!stadd(st, &sp, vt))
            {
                return ESTACK;
            }
            else
            {
                mvbuf(&s, &n, &rv, 3);
                break;
            }
        }

        /* maps */
        case T_map:
        {
            int64_t np;
            uint8_t kt;
            uint8_t vt;

            /* must have at least 6 bytes */
            if (n < 6)
            {
                return EEOF;
            }

            /* get the element type and count */
            kt = s[0];
            vt = s[1];
            np = u32be(s + 2);
            xprintf("[T_map] header rv:%d, kt:%d, vt:%d, np:%d\n", rv + 6, kt, vt, np);

            /* check for tag value */
            if (!(WireTags[kt] && WireTags[vt]))
            {
                return ETAG;
            }

            /* empty map */
            if (np == 0)
            {
                stpop(st, &sp);
                mvbuf(&s, &n, &rv, 6);
                continue;
            }

            /* check for fixed key and value */
            int64_t nk = SkipSizeFixed[kt];
            int64_t nv = SkipSizeFixed[vt];

            /* fast path for fixed key and value */
            if (nk != 0 && nv != 0)
            {
                if ((nb = np * (nk + nv) + 6) > n)
                {
                    return EEOF;
                }
                else
                {
                    stpop(st, &sp);
                    mvbuf(&s, &n, &rv, nb);
                    continue;
                }
            }

            /* set to parse the map pairs */
            st[sp].k = kt;
            st[sp].v = vt;
            st[sp].t = T_map_pair;
            st[sp].n = np * 2 - 1;
            mvbuf(&s, &n, &rv, 6);
            break;
        }

        /* map pairs */
        case T_map_pair:
        {
            uint8_t kt = st[sp].k;
            uint8_t vt = st[sp].v;

            /* there are keys pending */
            if (!stpop(st, &sp))
            {
                if ((st[sp].n & 1) == 0)
                {
                    xprintf("[T_map] key rv:%d, i:%d\n", rv, st[sp].n / 2);
                    vt = kt;
                }
                else
                {
                    xprintf("[T_map] elem rv:%d, i:%d\n", rv, st[sp].n / 2);
                }
            }
            else
            {
                xprintf("[T_map] elem rv:%d, i:%d\n", rv, st[sp].n / 2);
            }
            /* push the element onto stack */
            if (stadd(st, &sp, vt))
            {
                break;
            }
            else
            {
                return ESTACK;
            }
        }

        /* sets and lists */
        case T_set:
        case T_list:
        {
            int64_t nv;
            int64_t nt;
            uint8_t et;

            /* must have at least 5 bytes */
            if (n < 5)
            {
                return EEOF;
            }

            /* get the element type and count */
            et = s[0];
            nv = u32be(s + 1);
            xprintf("[T_list] header rv:%d, et:%d, np:%d\n", rv + 5, et, nv);

            /* check for tag value */
            if (!(WireTags[et]))
            {
                return ETAG;
            }

            /* empty sequence */
            if (nv == 0)
            {
                stpop(st, &sp);
                mvbuf(&s, &n, &rv, 5);
                continue;
            }

            /* fast path for fixed types */
            if ((nt = SkipSizeFixed[et]) != 0)
            {
                if ((nb = nv * nt + 5) > n)
                {
                    return EEOF;
                }
                else
                {
                    xprintf("[T_list] elem rv:%d, i:%d\n", rv + 5, st[sp].n);
                    stpop(st, &sp);
                    mvbuf(&s, &n, &rv, nb);
                    continue;
                }
            }

            /* set to parse the elements */
            st[sp].t = T_list_elem;
            st[sp].v = et;
            st[sp].n = nv - 1;
            mvbuf(&s, &n, &rv, 5);
            break;
        }

        /* list elem */
        case T_list_elem:
        {
            uint8_t et = st[sp].v;
            xprintf("[T_list] elem rv:%d, i:%d\n", rv, st[sp].n);
            stpop(st, &sp);
            /* push the element onto stack */
            if (stadd(st, &sp, et))
            {
                break;
            }
            else
            {
                return ESTACK;
            }
        }
        }
    }

    /* all done */
    return rv;
}