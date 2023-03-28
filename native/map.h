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

#ifndef MAP_H
#define MAP_H

#define CHAR_POINT (unsigned char)'.'

typedef struct
{
    uint32_t hash;
    GoString key;
    void *val;
} Entry;

typedef struct
{
    Entry *b;
    uint32_t c;
    uint32_t N;
} HashMap;

typedef struct
{
    void *val;
    GoString key;
} Pair;

typedef struct
{
    GoSlice *leaves;
    GoSlice index;
} TrieNode;

typedef struct
{
    size_t count;
    GoSlice positions;
    void *empty;
    TrieNode node;
} TrieTree;

void hm_set(HashMap *self, const GoString *key, void *val);
void *hm_get(const HashMap *self, const GoString *key);
void *trie_get(const TrieTree *self, const GoString *key);

typedef uint16_t tid;

typedef int8_t req_em;

#define REQ_OPTIONAL 0
#define REQ_DEFAULT 1
#define REQ_REQUIRED 2
#define BSIZE_INT64 64
#define SIZE_INT64 8
#define BSIZE_REQ 2
#define RequirenessInt64Count \
    (BSIZE_INT64 / BSIZE_REQ)

#define RequirenessMaskLeft \
    (BSIZE_INT64 - BSIZE_REQ)

#define RequirenessInt64Count \
    (BSIZE_INT64 / BSIZE_REQ)

#define Int64Mask 0xffffffffffffffffull

#define RequiredMask 0xaaaaaaaaaaaaaaaaull

#define DefaultMask \
    (RequiredMask >> 1)

#define BM_HAS_REQUIRED(v) \
    (v & RequiredMask) != 0

#define BM_HAS_DEFAULT(v) \
    (v & DefaultMask) != 0

typedef struct
{
    uint64_t *buf;
    size_t len;
} ReqBitMap;

bool bm_is_set(ReqBitMap b, tid id);
void bm_set_req(ReqBitMap b, tid id, req_em req);

#endif // MAP_H