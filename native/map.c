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
#include "memops.h"
#include "map.h"
#include "test/xprintf.h"

uint32_t hash_DJB32(const GoString str)
{
    uint32_t hash = 5381ul;
    for (int i = 0; i < str.len; i++)
    {
        hash = ((hash << 5) + hash + (uint32_t)str.buf[i]);
    }
    return hash;
}

void *hm_get(const HashMap *self, const GoString *key)
{
    uint32_t h = hash_DJB32(*key);
    uint32_t p = h % self->N;
    Entry s = self->b[p];

    while (s.hash != 0)
    {
        if (s.hash == h && s.key.len == key->len && memeq(key->buf, s.key.buf, key->len))
        {
            // xprintf("s.key:%s, s.val:%x\n", s.key.buf, s.val);
            return s.val;
        }
        else
        {
            p = (p + 1) % self->N;
            s = self->b[p];
        }
    }
    return NULL;
}

void hm_set(HashMap *self, const GoString *key, void *val)
{
    uint32_t h = hash_DJB32(*key);
    uint32_t p = h % self->N;
    Entry *s = &self->b[p];

    while (s->hash != 0)
    {
        p = (p + 1) % self->N;
        s = &self->b[p];
    }

    // xprintf("h:%d, p:%d, key:%s, val:%x\n", h, p, key, val);
    s->val = val;
    s->hash = h;
    s->key = *key;
    self->c++;
}

const size_t SIZE_SIZE_T = sizeof(size_t);
const size_t SIZE_TRIE_NODE = sizeof(TrieNode);
const size_t SIZE_PAIR = sizeof(Pair);

uint8_t ascii2int(const unsigned char c)
{
    if (c < CHAR_POINT)
    {
        return (uint8_t)(255 + c - CHAR_POINT);
    }
    return (uint8_t)(c - CHAR_POINT);
}

void *trie_get(const TrieTree *self, const GoString *key)
{
    // xprintf("[trie_get] key:%s\n", key);
    if (key->len == 0)
    {
        return self->empty;
    }

    size_t l = key->len - 1;
    GoSlice fs = self->node.index;
    GoSlice ls = *self->node.leaves;

    for (int s = 0; s < self->positions.len; s++)
    {
        size_t i = *(size_t *)(self->positions.buf + s * SIZE_SIZE_T);
        // xprintf("[trie_get] l:%d, position:%d\n", l, i);
        if (i > l)
        {
            i = l;
        }
        uint8_t j = ascii2int(key->buf[i]);
        if (j > fs.len)
        {
            return NULL;
        }
        // xprintf("[trie_get] j:%d, fs.len:%d\n", j, fs.len);
        TrieNode *fp = (TrieNode *)(fs.buf + j * SIZE_TRIE_NODE);
        if (fp->leaves == NULL)
        {
            return NULL;
        }
        fs = fp->index;
        ls = *fp->leaves;
        // xprintf("[trie_get] fp->leaves[0]:%s\n", &((Pair *)(ls.buf))->key);
    }
    // xprintf("[trie_get] ls.len:%d\n", l, ls.len);

    for (int i = 0; i < ls.len; i++)
    {
        Pair *v = (Pair *)(ls.buf + i * SIZE_PAIR);
        if (key->len == v->key.len && memeq(key->buf, v->key.buf, key->len))
        {
            return v->val;
        }
    }
    return NULL;
}

bool bm_is_set(ReqBitMap b, tid id)
{
    uint64_t v = b.buf[id / BSIZE_INT64];
    return (v & (1ull << (id % BSIZE_INT64))) != 0;
}

void bm_set_req(ReqBitMap b, tid id, req_em req)
{
    uint64_t *p = &b.buf[id / BSIZE_INT64];
    switch (req)
    {
    case REQ_DEFAULT:
    case REQ_REQUIRED:
        *p |= (uint64_t)(1ull << (id % BSIZE_INT64));
        break;
    case REQ_OPTIONAL:
        *p &= ~(uint64_t)(1ull << (id % BSIZE_INT64));
        break;
    default:
        break;
    }
}