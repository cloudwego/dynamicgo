/**
 * Copyright 2022 CloudWeGo Authors.
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

#include "../thrift.h"
#include "xprintf.h"
#include "../map.h"

#ifndef EXAMPLE_H
#define EXAMPLE_H

const char EXAMPLE_JSON[] = "{\"Msg\":\"中文\",\"Base\":{\"LogID\":\"a\",\"Caller\":\"b\",\"Addr\":\"c\",\"Client\":\"d\"}}";

const size_t EXMAPLE_MAX_KEY_LEN = 12;

tTypeDesc STRING_DESC = {
    .type = TTYPE_STRING,
    .name = (GoString){.buf = "string", .len = 6},
};

tTypeDesc INT32_DESC = {
    .type = TTYPE_I32,
    .name = (GoString){.buf = "i32", .len = 3},
};

tTypeDesc EXTRA_DESC = {
    .type = TTYPE_I32,
    .name = (GoString){.buf = "map", .len = 3},
    .key = &STRING_DESC,
    .elem = &STRING_DESC,
};

tTypeDesc BASE_DESC = {
    .is_base = true,
    .type = TTYPE_STRUCT,
    .name = (GoString){.buf = "struct", .len = 6},
    .st = &(tStructDesc){
        .name = (GoString){.buf = "Base", .len = 4},
        .ids = (GoSlice){
            .buf = (char *)(&(tFieldDesc *[7]){}), //TODO
            .cap = 7,
            .len = 7,
        },
        .names = (tFieldNameMap){
            .trie = &(TrieTree){
                .positions = (GoSlice){
                    .buf = (char *)(&(size_t[1]){1}), // o,a,d,l,r,x
                    .cap = 1,
                    .len = 1,
                },
                .node = (TrieNode){
                    .index = (GoSlice){
                        .buf = (char *)(&(TrieNode[127]){}), //TODO: o,a,d,l,r,x
                        .cap = 127,
                        .len = 127,
                    },
                },
            },
            .hash = &(HashMap){
                .N = 24, .b = (Entry[24]){}, //TODO
            },
        },
        .reqs = &(ReqBitMap){
            .buf = (uint64_t[1]){0x1540000000000000ull},
            .len = 1,
        },
    },
};

tFieldDesc BASE_FIELD_1 = {
    .required = REQ_DEFAULT,
    .ID = 1,
    .name = (GoString){.buf = "LogID", .len = 5},
    .alias = (GoString){.buf = "log_id", .len = 5},
    .type = &STRING_DESC,
};

tFieldDesc BASE_FIELD_2 = {
    .required = REQ_DEFAULT,
    .ID = 2,
    .name = (GoString){.buf = "Caller", .len = 6},
    .alias = (GoString){.buf = "caller", .len = 6},
    .type = &STRING_DESC,
};

tFieldDesc BASE_FIELD_3 = {
    .required = REQ_DEFAULT,
    .ID = 3,
    .name = (GoString){.buf = "Addr", .len = 4},
    .alias = (GoString){.buf = "addr", .len = 4},
    .type = &STRING_DESC,
};

tFieldDesc BASE_FIELD_4 = {
    .required = REQ_DEFAULT,
    .ID = 4,
    .name = (GoString){.buf = "Client", .len = 6},
    .alias = (GoString){.buf = "Client", .len = 6},
    .type = &STRING_DESC,
};

tFieldDesc BASE_FIELD_5 = {
    .required = REQ_OPTIONAL,
    .ID = 5,
    .name = (GoString){.buf = "TrafficEnv", .len = 10},
    .alias = (GoString){.buf = "traffic_env", .len = 11},
    .type = &STRING_DESC,
};

tFieldDesc BASE_FIELD_6 = {
    .required = REQ_OPTIONAL,
    .ID = 6,
    .name = (GoString){.buf = "Extra", .len = 5},
    .alias = (GoString){.buf = "extra", .len = 5},
    .type = &EXTRA_DESC,
};

tTypeDesc EXAMPLE_DESC = {
    .name = (GoString){.buf = "struct", .len = 6},
    .type = TTYPE_STRUCT,
    .st = &(tStructDesc){
        .name = (GoString){.buf = "ExampleReq", .len = 10},
        .ids = (GoSlice){
            .buf = (char *)(&(tFieldDesc *[256]){}), //TODO
            .cap = 256,
            .len = 256,
        },
        .names = (tFieldNameMap){
            .trie = &(TrieTree){
                .positions = (GoSlice){
                    .buf = (char *)(&(size_t[1]){0}), // M,I,B
                    .cap = 1,
                    .len = 1,
                },
                .node = (TrieNode){
                    .index = (GoSlice){
                        .buf = (char *)(&(TrieNode[127]){}), //TODO
                        .cap = 127,
                        .len = 127,
                    },
                },
            },
            .hash = &(HashMap){
                .N = 12, .b = (Entry[12]){}, //TODO
            },
        },
        .reqs = &(ReqBitMap){
            .buf = (uint64_t[8]){0x100000000000000ull, 0x0ull, 0x0ull, 0x0ull, 0x0ull, 0x0ull, 0x0ull, 0x2ull},
            .len = 8,
        },
    },
};

tFieldDesc EXAMPLE_FIELD_1 = {
    .required = REQ_OPTIONAL,
    .ID = 1,
    .name = (GoString){.buf = "Msg", .len = 3},
    .alias = (GoString){.buf = "msg", .len = 3},
    .type = &STRING_DESC,
};

tFieldDesc EXAMPLE_FIELD_3 = {
    .required = REQ_DEFAULT,
    .ID = 3,
    .name = (GoString){.buf = "InnerBase", .len = 9},
    .alias = (GoString){.buf = "inner_base", .len = 10},
    .type = &INT32_DESC,
};

tFieldDesc EXAMPLE_FIELD_255 = {
    .required = REQ_REQUIRED,
    .ID = 255,
    .name = (GoString){.buf = "Base", .len = 4},
    .alias = (GoString){.buf = "base", .len = 4},
    .type = &BASE_DESC,
};

TrieNode EXAMPLE_TN_1 = {
    .leaves = &(GoSlice){
        .buf = (char *)&(Pair){
            .key = (GoString){.buf = "Msg", .len = 3},
            .val = &EXAMPLE_FIELD_1,
        },
        .len = 1,
        .cap = 1,
    },
};

TrieNode EXAMPLE_TN_3 = {
    .leaves = &(GoSlice){
        .buf = (char *)&(Pair){
            .key = (GoString){.buf = "InnerBase", .len = 9},
            .val = &EXAMPLE_FIELD_3,
        },
        .len = 1,
        .cap = 1,
    },
};

TrieNode EXAMPLE_TN_255 = {
    .leaves = &(GoSlice){
        .buf = (char *)&(Pair){
            .key = (GoString){.buf = "Base", .len = 4},
            .val = &EXAMPLE_FIELD_255,
        },
        .len = 1,
        .cap = 1,
    },
};

#define SIZE_GO_SLICE sizeof(GoSlice)

TrieNode BASE_TN_1 = {
    .leaves = &(GoSlice){
        .buf = (char *)&(Pair){
            .key = (GoString){.buf = "LogID", .len = 5},
            .val = &BASE_FIELD_1,
        },
        .len = 1,
        .cap = 1,
    },
};

TrieNode BASE_TN_2 = {
    .leaves = &(GoSlice){
        .buf = (char *)&(Pair){
            .key = (GoString){.buf = "Caller", .len = 6},
            .val = &BASE_FIELD_2,
        },
        .len = 1,
        .cap = 1,
    },
};

TrieNode BASE_TN_3 = {
    .leaves = &(GoSlice){
        .buf = (char *)&(Pair){
            .key = (GoString){.buf = "Addr", .len = 4},
            .val = &BASE_FIELD_3,
        },
        .len = 1,
        .cap = 1,
    },
};

TrieNode BASE_TN_4 = {
    .leaves = &(GoSlice){
        .buf = (char *)&(Pair){
            .key = (GoString){.buf = "Client", .len = 6},
            .val = &BASE_FIELD_4,
        },
        .len = 1,
        .cap = 1,
    },
};

TrieNode BASE_TN_5 = {
    .leaves = &(GoSlice){
        .buf = (char *)&(Pair){
            .key = (GoString){.buf = "TrafficEnv", .len = 10},
            .val = &BASE_FIELD_5,
        },
        .len = 1,
        .cap = 1,
    },
};

TrieNode BASE_TN_6 = {
    .leaves = &(GoSlice){
        .buf = (char *)&(Pair){
            .key = (GoString){.buf = "Extra", .len = 5},
            .val = &BASE_FIELD_6,
        },
        .len = 1,
        .cap = 1,
    },
};

#define INIT_BASE_DESC()                                                     \
    do                                                                       \
    {                                                                        \
        tFieldDesc **p0 = (tFieldDesc **)BASE_DESC.st->ids.buf;              \
        HashMap *hm = BASE_DESC.st->names.hash;                              \
        p0[1] = &BASE_FIELD_1;                                               \
        hm_set(hm, &(p0[1]->name), p0[1]);                                   \
        p0[2] = &BASE_FIELD_2;                                               \
        hm_set(hm, &(p0[2]->name), p0[2]);                                   \
        p0[3] = &BASE_FIELD_3;                                               \
        hm_set(hm, &(p0[3]->name), p0[3]);                                   \
        p0[4] = &BASE_FIELD_4;                                               \
        hm_set(hm, &(p0[4]->name), p0[4]);                                   \
        p0[5] = &BASE_FIELD_5;                                               \
        hm_set(hm, &(p0[5]->name), p0[5]);                                   \
        p0[6] = &BASE_FIELD_6;                                               \
        hm_set(hm, &(p0[6]->name), p0[6]);                                   \
        BASE_DESC.st->names.hash = NULL;                                     \
        TrieNode *p1 = (TrieNode *)BASE_DESC.st->names.trie->node.index.buf; \
        p1['o' - CHAR_POINT] = BASE_TN_1;                                    \
        p1['a' - CHAR_POINT] = BASE_TN_2;                                    \
        p1['d' - CHAR_POINT] = BASE_TN_3;                                    \
        p1['l' - CHAR_POINT] = BASE_TN_4;                                    \
        p1['r' - CHAR_POINT] = BASE_TN_5;                                    \
        p1['x' - CHAR_POINT] = BASE_TN_6;                                    \
    } while (0)

#define INIT_EXAMPLE_DESC()                                                     \
    do                                                                          \
    {                                                                           \
        tFieldDesc **p0 = (tFieldDesc **)EXAMPLE_DESC.st->ids.buf;              \
        HashMap *hm = EXAMPLE_DESC.st->names.hash;                              \
        p0[1] = &EXAMPLE_FIELD_1;                                               \
        hm_set(hm, &(p0[1]->name), p0[1]);                                      \
        p0[3] = &EXAMPLE_FIELD_3;                                               \
        hm_set(hm, &(p0[3]->name), p0[3]);                                      \
        p0[255] = &EXAMPLE_FIELD_255;                                           \
        hm_set(hm, &(p0[255]->name), p0[255]);                                  \
        EXAMPLE_DESC.st->names.hash = NULL;                                     \
        TrieNode *p1 = (TrieNode *)EXAMPLE_DESC.st->names.trie->node.index.buf; \
        p1['M' - CHAR_POINT] = EXAMPLE_TN_1;                                    \
        p1['I' - CHAR_POINT] = EXAMPLE_TN_3;                                    \
        p1['B' - CHAR_POINT] = EXAMPLE_TN_255;                                  \
    } while (0)

#endif // SCANNING_H