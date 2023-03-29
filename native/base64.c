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

#include <stdint.h>
#include <immintrin.h>
#include <sys/types.h>
#include "native.h"

#define MODE_URL 1
#define MODE_RAW 2
#define MODE_AVX2 4

#define as_m32v(v) (*(uint32_t *)(v))
#define as_m64v(v) (*(uint64_t *)(v))

#define as_m128p(v) ((__m128i *)(v))
#define as_m256p(v) ((__m256i *)(v))

#define as_m8c(v) ((const uint8_t *)(v))
#define as_m128c(v) ((const __m128i *)(v))
#define as_m256c(v) ((const __m256i *)(v))

/** Exported Functions **/

void b64encode(_GoSlice *out, const _GoSlice *src, int mode);
ssize_t b64decode(_GoSlice *out, const char *src, size_t nb, int mode);

/** Encoder Helper Functions **/

static const char TabEncodeCharsetStd[64] = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";
static const char TabEncodeCharsetURL[64] = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_";

static const uint8_t VecEncodeShuffles[32] = {
    1,
    0,
    2,
    1,
    4,
    3,
    5,
    4,
    7,
    6,
    8,
    7,
    10,
    9,
    11,
    10,
    1,
    0,
    2,
    1,
    4,
    3,
    5,
    4,
    7,
    6,
    8,
    7,
    10,
    9,
    11,
    10,
};

static const uint8_t VecEncodeCharsetStd[32] = {
    'a' - 26,
    '0' - 52,
    '0' - 52,
    '0' - 52,
    '0' - 52,
    '0' - 52,
    '0' - 52,
    '0' - 52,
    '0' - 52,
    '0' - 52,
    '0' - 52,
    '+' - 62,
    '/' - 63,
    'A',
    0,
    0,
    'a' - 26,
    '0' - 52,
    '0' - 52,
    '0' - 52,
    '0' - 52,
    '0' - 52,
    '0' - 52,
    '0' - 52,
    '0' - 52,
    '0' - 52,
    '0' - 52,
    '+' - 62,
    '/' - 63,
    'A',
    0,
    0,
};

static const uint8_t VecEncodeCharsetURL[32] = {
    'a' - 26,
    '0' - 52,
    '0' - 52,
    '0' - 52,
    '0' - 52,
    '0' - 52,
    '0' - 52,
    '0' - 52,
    '0' - 52,
    '0' - 52,
    '0' - 52,
    '-' - 62,
    '_' - 63,
    'A',
    0,
    0,
    'a' - 26,
    '0' - 52,
    '0' - 52,
    '0' - 52,
    '0' - 52,
    '0' - 52,
    '0' - 52,
    '0' - 52,
    '0' - 52,
    '0' - 52,
    '0' - 52,
    '-' - 62,
    '_' - 63,
    'A',
    0,
    0,
};

static inline __m256i encode_avx2(__m128i v0, __m128i v1, const uint8_t *tab)
{
    __m256i vv = _mm256_set_m128i(v1, v0);
    __m256i sh = _mm256_loadu_si256(as_m256c(VecEncodeShuffles));
    __m256i in = _mm256_shuffle_epi8(vv, sh);
    __m256i t0 = _mm256_and_si256(in, _mm256_set1_epi32(0x0fc0fc00));
    __m256i t1 = _mm256_mulhi_epu16(t0, _mm256_set1_epi32(0x04000040));
    __m256i t2 = _mm256_and_si256(in, _mm256_set1_epi32(0x003f03f0));
    __m256i t3 = _mm256_mullo_epi16(t2, _mm256_set1_epi32(0x01000010));
    __m256i vi = _mm256_or_si256(t1, t3);
    __m256i s0 = _mm256_cmpgt_epi8(_mm256_set1_epi8(26), vi);
    __m256i s1 = _mm256_and_si256(_mm256_set1_epi8(13), s0);
    __m256i s2 = _mm256_loadu_si256(as_m256c(tab));
    __m256i r0 = _mm256_subs_epu8(vi, _mm256_set1_epi8(51));
    __m256i r1 = _mm256_or_si256(r0, s1);
    __m256i r2 = _mm256_shuffle_epi8(s2, r1);
    __m256i r3 = _mm256_add_epi8(vi, r2);
    return r3;
}

/** Function Implementations **/

void b64encode(_GoSlice *out, const _GoSlice *src, int mode)
{
    char *ob = out->buf + out->len;
    char *op = out->buf + out->len;
    const char *ip = src->buf;
    const char *ie = src->buf + src->len;
    const char *st = TabEncodeCharsetStd;
    const uint8_t *vt = VecEncodeCharsetStd;

    /* check for empty string */
    if (src->len == 0)
    {
        return;
    }

    /* check for URL encoding */
    if (mode & MODE_URL)
    {
        st = TabEncodeCharsetURL;
        vt = VecEncodeCharsetURL;
    }

#if USE_AVX2
    /* SIMD 24 bytes loop, but the SIMD instruction will load 4 bytes
     * past the end, so it's safe only if there are 28 bytes or more left */
    while ((ip <= ie - 28) && (mode & MODE_AVX2) != 0)
    {
        __m128i v0 = _mm_loadu_si128(as_m128c(ip));
        __m128i v1 = _mm_loadu_si128(as_m128c(ip + 12));
        __m256i vv = encode_avx2(v0, v1, vt);

        /* store the result, and advance buffer pointers */
        _mm256_storeu_si256(as_m256p(op), vv);
        op += 32;
        ip += 24;
    }

    /* can do one more 24 bytes round, but needs special handling */
    if ((ip <= ie - 24) && (mode & MODE_AVX2) != 0)
    {
        __m128i v0 = _mm_loadu_si128(as_m128c(ip));
        __m128i v1 = _mm_loadu_si128(as_m128c(ip + 8));
        __m128i v2 = _mm_srli_si128(v1, 4);
        __m256i vv = encode_avx2(v0, v2, vt);

        /* store the result, and advance buffer pointers */
        _mm256_storeu_si256(as_m256p(op), vv);
        op += 32;
        ip += 24;
    }
#endif

    /* no more bytes */
    if (ip == ie)
    {
        out->len += op - ob;
        return;
    }

    /* handle the remaining bytes with scalar code (with 4 bytes load) */
    while (ip <= ie - 4)
    {
        uint32_t v0 = __builtin_bswap32(*(const uint32_t *)ip);
        uint8_t v1 = (v0 >> 26) & 0x3f;
        uint8_t v2 = (v0 >> 20) & 0x3f;
        uint8_t v3 = (v0 >> 14) & 0x3f;
        uint8_t v4 = (v0 >> 8) & 0x3f;

        /* encode the characters, and move to next block */
        ip += 3;
        *op++ = st[v1];
        *op++ = st[v2];
        *op++ = st[v3];
        *op++ = st[v4];
    }

    /* load the last bytes */
    size_t dp = ie - ip;
    uint32_t v0 = (uint32_t)(uint8_t)ip[0] << 16;

#define B2 v0 |= (uint32_t)(uint8_t)ip[2]
#define B1 v0 |= (uint32_t)(uint8_t)ip[1] << 8

#define R4 *op++ = st[(v0 >> 0) & 0x3f]
#define R3 *op++ = st[(v0 >> 6) & 0x3f]
#define R2 *op++ = st[(v0 >> 12) & 0x3f]
#define R1 *op++ = st[(v0 >> 18) & 0x3f]

#define NB                   \
    {                        \
        out->len += op - ob; \
    }
#define PD                          \
    {                               \
        if ((mode & MODE_RAW) == 0) \
        {                           \
            *op++ = '=';            \
        }                           \
    }

    /* encode the last few bytes */
    switch (dp)
    {
    case 3:
        B2;
        B1;
        R1;
        R2;
        R3;
        R4;
        NB;
        break;
    case 2:
        B1;
        R1;
        R2;
        R3;
        PD;
        NB;
        break;
    case 1:
        R1;
        R2;
        PD;
        PD;
        NB;
        break;
    default:
        NB;
        break;
    }

#undef PD
#undef NB
#undef R1
#undef R2
#undef R3
#undef R4
#undef B1
#undef B2
}

/** Decoder Helper Functions **/

static const uint8_t VecPacking[32] = {
    2, 1, 0, 6, 5, 4, 10, 9, 8, 14, 13, 12, 128, 128, 128, 128,
    2, 1, 0, 6, 5, 4, 10, 9, 8, 14, 13, 12, 128, 128, 128, 128};

static const uint8_t VecDecodeBits[32] = {
    0x01, 0x02, 0x04, 0x08, 0x10, 0x20, 0x40, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
    0x01, 0x02, 0x04, 0x08, 0x10, 0x20, 0x40, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00};

static const uint8_t VecDecodeTableStd[128] = {
    0x00, 0x00, 0x13, 0x04, 0xbf, 0xbf, 0xb9, 0xb9, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
    0x00, 0x00, 0x13, 0x04, 0xbf, 0xbf, 0xb9, 0xb9, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
    0xa8, 0xf8, 0xf8, 0xf8, 0xf8, 0xf8, 0xf8, 0xf8, 0xf8, 0xf8, 0xf0, 0x54, 0x50, 0x50, 0x50, 0x54,
    0xa8, 0xf8, 0xf8, 0xf8, 0xf8, 0xf8, 0xf8, 0xf8, 0xf8, 0xf8, 0xf0, 0x54, 0x50, 0x50, 0x50, 0x54,
    0x2f, 0x2f, 0x2f, 0x2f, 0x2f, 0x2f, 0x2f, 0x2f, 0x2f, 0x2f, 0x2f, 0x2f, 0x2f, 0x2f, 0x2f, 0x2f,
    0x2f, 0x2f, 0x2f, 0x2f, 0x2f, 0x2f, 0x2f, 0x2f, 0x2f, 0x2f, 0x2f, 0x2f, 0x2f, 0x2f, 0x2f, 0x2f,
    0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10,
    0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10, 0x10};

static const uint8_t VecDecodeTableURL[128] = {
    0x00,
    0x00,
    0x11,
    0x04,
    0xbf,
    0xbf,
    0xb9,
    0xb9,
    0x00,
    0x00,
    0x00,
    0x00,
    0x00,
    0x00,
    0x00,
    0x00,
    0x00,
    0x00,
    0x11,
    0x04,
    0xbf,
    0xbf,
    0xb9,
    0xb9,
    0x00,
    0x00,
    0x00,
    0x00,
    0x00,
    0x00,
    0x00,
    0x00,
    0xa8,
    0xf8,
    0xf8,
    0xf8,
    0xf8,
    0xf8,
    0xf8,
    0xf8,
    0xf8,
    0xf8,
    0xf0,
    0x50,
    0x50,
    0x54,
    0x50,
    0x70,
    0xa8,
    0xf8,
    0xf8,
    0xf8,
    0xf8,
    0xf8,
    0xf8,
    0xf8,
    0xf8,
    0xf8,
    0xf0,
    0x50,
    0x50,
    0x54,
    0x50,
    0x70,
    0x5f,
    0x5f,
    0x5f,
    0x5f,
    0x5f,
    0x5f,
    0x5f,
    0x5f,
    0x5f,
    0x5f,
    0x5f,
    0x5f,
    0x5f,
    0x5f,
    0x5f,
    0x5f,
    0x5f,
    0x5f,
    0x5f,
    0x5f,
    0x5f,
    0x5f,
    0x5f,
    0x5f,
    0x5f,
    0x5f,
    0x5f,
    0x5f,
    0x5f,
    0x5f,
    0x5f,
    0x5f,
    0xe0,
    0xe0,
    0xe0,
    0xe0,
    0xe0,
    0xe0,
    0xe0,
    0xe0,
    0xe0,
    0xe0,
    0xe0,
    0xe0,
    0xe0,
    0xe0,
    0xe0,
    0xe0,
    0xe0,
    0xe0,
    0xe0,
    0xe0,
    0xe0,
    0xe0,
    0xe0,
    0xe0,
    0xe0,
    0xe0,
    0xe0,
    0xe0,
    0xe0,
    0xe0,
    0xe0,
    0xe0,
};

static const uint8_t VecDecodeCharsetStd[256] = {
    0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
    0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
    0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 62, 0xff, 0xff, 0xff, 63,
    52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
    0xff, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14,
    15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 0xff, 0xff, 0xff, 0xff, 0xff,
    0xff, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40,
    41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 0xff, 0xff, 0xff, 0xff, 0xff,
    0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
    0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
    0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
    0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
    0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
    0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
    0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
    0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff};

static const uint8_t VecDecodeCharsetURL[256] = {
    0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
    0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
    0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 62, 0xff, 0xff,
    52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
    0xff, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14,
    15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 0xff, 0xff, 0xff, 0xff, 63,
    0xff, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40,
    41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 0xff, 0xff, 0xff, 0xff, 0xff,
    0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
    0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
    0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
    0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
    0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
    0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
    0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
    0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff};

static inline void memcopy_24(char *dp, const uint8_t *sp)
{
    *(uint64_t *)(dp + 0) = *(const uint64_t *)(sp + 0);
    *(uint64_t *)(dp + 8) = *(const uint64_t *)(sp + 8);
    *(uint64_t *)(dp + 16) = *(const uint64_t *)(sp + 16);
}

static inline __m256i decode_avx2(__m256i v0, int *pos, const uint8_t *tab)
{
    __m256i v1 = _mm256_srli_epi32(v0, 4);
    __m256i vl = _mm256_and_si256(v0, _mm256_set1_epi8(0x0f));
    __m256i vh = _mm256_and_si256(v1, _mm256_set1_epi8(0x0f));
    __m256i st = _mm256_loadu_si256(as_m256c(tab));
    __m256i mt = _mm256_loadu_si256(as_m256c(tab + 32));
    __m256i et = _mm256_loadu_si256(as_m256c(tab + 64));
    __m256i rt = _mm256_loadu_si256(as_m256c(tab + 96));
    __m256i pt = _mm256_loadu_si256(as_m256c(VecPacking));
    __m256i bt = _mm256_loadu_si256(as_m256c(VecDecodeBits));
    __m256i sh = _mm256_shuffle_epi8(st, vh);
    __m256i eq = _mm256_cmpeq_epi8(v0, et);
    __m256i sv = _mm256_blendv_epi8(sh, rt, eq);
    __m256i bm = _mm256_shuffle_epi8(mt, vl);
    __m256i bv = _mm256_shuffle_epi8(bt, vh);
    __m256i mr = _mm256_and_si256(bm, bv);
    __m256i nm = _mm256_cmpeq_epi8(mr, _mm256_setzero_si256());
    __m256i sr = _mm256_add_epi8(v0, sv);
    __m256i r0 = _mm256_and_si256(sr, _mm256_set1_epi8(0x3f));
    __m256i r1 = _mm256_maddubs_epi16(r0, _mm256_set1_epi32(0x01400140));
    __m256i r2 = _mm256_madd_epi16(r1, _mm256_set1_epi32(0x00011000));
    __m256i r3 = _mm256_shuffle_epi8(r2, pt);
    __m256i r4 = _mm256_permutevar8x32_epi32(r3, _mm256_setr_epi32(0, 1, 2, 4, 5, 6, 3, 7));
    int64_t mp = _mm256_movemask_epi8(nm);
    int32_t np = __builtin_ctzll(mp | 0xffffffff00000000);
    return (*pos = np), r4;
}

/* Return 0 if success, otherwise return the error position + 1 */
static inline int64_t decode_block(
    const uint8_t *ie,
    const uint8_t **ipp,
    char **opp,
    const uint8_t *tab,
    int mode)
{
    int nb = 0;
    uint32_t v0 = 0;

    /* buffer pointers */
    char *op = *opp;
    const uint8_t *ip = *ipp;

    /* load up to 4 characters */
    while (nb < 4 && ip < ie)
    {
        uint8_t id;
        uint8_t ch = *ip;

        /* skip new lines */
        if (ch == '\r' || ch == '\n')
        {
            ip++;
            continue;
        }

        /* lookup the index, and check for invalid characters */
        if ((id = tab[ch]) == 0xff)
        {
            break;
        }

        /* move to next character */
        ip++;
        nb++;
        v0 = (v0 << 6) | id;
    }

    /* never ends with 1 characer */
    if (nb == 1)
    {
        return ip - *ipp + 1;
    }

#define P2() \
    {        \
        E2() \
        P1() \
        P1() \
    }
#define P1()                  \
    {                         \
        if (*ip++ != '=')     \
            return ip - *ipp; \
    } // ip has been added 1
#define E2()                      \
    {                             \
        if (ip >= ie - 1)         \
            return ip - *ipp + 1; \
    }
#define R1()                        \
    {                               \
        if ((mode & MODE_RAW) == 0) \
            return ip - *ipp + 1;   \
    }

#define align_val()          \
    {                        \
        v0 <<= 6 * (4 - nb); \
    }
#define parse_eof()               \
    {                             \
        if (ip < ie)              \
            return ip - *ipp + 1; \
    }
#define check_pad()       \
    {                     \
        if (ip == ie)     \
            R1()          \
        else if (nb == 3) \
            P1()          \
        else              \
            P2()          \
    }

    /* not enough characters, can either be EOF or paddings or illegal characters */
    if (nb < 4)
    {
        check_pad()
            parse_eof()
                align_val()
    }

#undef check_pad
#undef parse_eof
#undef align_val

#undef R1
#undef E2
#undef P1
#undef P2

    /* decode into output */
    switch (nb)
    {
    case 4:
        op[2] = (v0 >> 0) & 0xff;
    case 3:
        op[1] = (v0 >> 8) & 0xff;
    case 2:
        op[0] = (v0 >> 16) & 0xff;
    }

    /* update the pointers */
    *ipp = ip;
    *opp = op + nb - 1;
    return 0;
}

ssize_t b64decode(_GoSlice *out, const char *src, size_t nb, int mode)
{
    int ep;
    __m256i vv;
    int64_t dv;
    uint8_t buf[32] = {0};

    /* check for empty input */
    if (nb == 0)
    {
        return 0;
    }

    /* output buffer */
    char *ob = out->buf + out->len;
    char *op = out->buf + out->len;
    char *oe = out->buf + out->cap;

    /* input buffer */
    const uint8_t *dt = VecDecodeTableStd;
    const uint8_t *st = VecDecodeCharsetStd;
    const uint8_t *ib = (const uint8_t *)src;
    const uint8_t *ip = (const uint8_t *)src;
    const uint8_t *ie = (const uint8_t *)src + nb;

    /* check for URL encoding */
    if (mode & MODE_URL)
    {
        dt = VecDecodeTableURL;
        st = VecDecodeCharsetURL;
    }

#if USE_AVX2
    /* decode every 32 bytes, the final round should be handled separately, because the
     * SIMD instruction performs 32-byte store, and it might store past the end of the
     * output buffer */
    while ((ip <= ie - 32) && (mode & MODE_AVX2) != 0)
    {
        vv = _mm256_loadu_si256(as_m256c(ip));
        vv = decode_avx2(vv, &ep, dt);

        /* check for invalid characters (or '=' paddings) */
        if (ep < 32)
        {
            if ((dv = decode_block(ie, &ip, &op, st, mode)) != 0)
            {
                return ib - ip - dv;
            }
            else
            {
                continue;
            }
        }

        /* check for store boundary, perform the last 24-byte store if needed */
        if (op <= oe - 32)
        {
            _mm256_storeu_si256(as_m256p(op), vv);
        }
        else
        {
            _mm256_storeu_si256(as_m256p(buf), vv);
            memcopy_24(op, buf);
        }

        /* move to next block */
        ip += 32;
        op += 24;
    }
#endif
    /* handle the remaining bytes with scalar code (8 byte loop) */
    while (ip <= ie - 8 && op <= oe - 8)
    {
        uint8_t v0 = st[ip[0]];
        uint8_t v1 = st[ip[1]];
        uint8_t v2 = st[ip[2]];
        uint8_t v3 = st[ip[3]];
        uint8_t v4 = st[ip[4]];
        uint8_t v5 = st[ip[5]];
        uint8_t v6 = st[ip[6]];
        uint8_t v7 = st[ip[7]];

        /* check for invalid bytes */
        if ((v0 | v1 | v2 | v3 | v4 | v5 | v6 | v7) == 0xff)
        {
            if ((dv = decode_block(ie, &ip, &op, st, mode)) != 0)
            {
                return ib - ip - dv;
            }
            else
            {
                continue;
            }
        }

        /* construct the characters */
        uint64_t vv = __builtin_bswap64(
            ((uint64_t)v0 << 58) |
            ((uint64_t)v1 << 52) |
            ((uint64_t)v2 << 46) |
            ((uint64_t)v3 << 40) |
            ((uint64_t)v4 << 34) |
            ((uint64_t)v5 << 28) |
            ((uint64_t)v6 << 22) |
            ((uint64_t)v7 << 16));

        /* store the result, and move to next block */
        as_m64v(op) = vv;
        ip += 8;
        op += 6;
    }

    /* handle the remaining bytes with scalar code (4 byte loop) */
    while (ip <= ie - 4 && op <= oe - 4)
    {
        uint8_t v0 = st[ip[0]];
        uint8_t v1 = st[ip[1]];
        uint8_t v2 = st[ip[2]];
        uint8_t v3 = st[ip[3]];

        /* check for invalid bytes */
        if ((v0 | v1 | v2 | v3) == 0xff)
        {
            if ((dv = decode_block(ie, &ip, &op, st, mode)) != 0)
            {
                return ib - ip - dv;
            }
            else
            {
                continue;
            }
        }

        /* construct the characters */
        uint32_t vv = __builtin_bswap32(
            ((uint32_t)v0 << 26) |
            ((uint32_t)v1 << 20) |
            ((uint32_t)v2 << 14) |
            ((uint32_t)v3 << 8));

        /* store the result, and move to next block */
        as_m32v(op) = vv;
        ip += 4;
        op += 3;
    }

    /* decode the last few bytes */
    while (ip < ie)
    {
        if ((dv = decode_block(ie, &ip, &op, st, mode)) != 0)
        {
            return ib - ip - dv;
        }
    }

    /* update the result length */
    out->len += op - ob;
    return op - ob;
}