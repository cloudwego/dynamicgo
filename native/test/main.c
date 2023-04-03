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

// #include <stdio.h>
#include "../native.c"
#include "xprintf.h"

#include "../thrift_compact.h"
#include <stdint.h>
// #include "example.h"

J2TStateMachine j2t = {};
_GoSlice buf = (_GoSlice){
    .buf = (char[1024]){},
    .len = 0,
    .cap = 1024,
};
char DBUF[800] = {};

uint64_t test_tcompact()
{
    char buf_c[1000];
    int16_t lfst_c[64];
    uint16_t cwbs_c[64];

    _GoSlice buf = { .buf = &buf_c[0], .len = 0, .cap = sizeof(buf_c) };
    Int16Slice lfst = { .buf = &lfst_c[0], .len = 0, .cap = sizeof(lfst_c)/sizeof(int16_t) };
    Uint16Slice cwbs = { .buf = &cwbs_c[0], .len = 0, .cap = sizeof(cwbs_c)/sizeof(uint16_t)};

    tc_state tc = { 
        .buf = &buf,
        .last_field_id = 0,
        .last_field_id_stack = lfst,
        .pending_field_write = {
            .valid = false,
        },
        .container_write_back_stack = cwbs,
    };
    tc_ienc ienc = tc_get_iencoder((tc_get_iencoder_arg){
        .j2tsm = 0, .outbuf = &buf, .tc = &tc,
    });
    void *tproto = VT_TSTATE(ienc.base);


    char rpc_name[] = "GetFavoriteMethod";
    
    J2T_ZERO(ienc.method->write_message_begin(tproto,
        (_GoString){ .buf = &rpc_name[0], .len = sizeof(rpc_name)-1 },
        TMESSAGE_TYPE_CALL,
        0));

    // BEGIN ARGS STRUCT
    J2T_ZERO(ienc.method->write_struct_begin(tproto));

        J2T_ZERO(ienc.method->write_field_begin(tproto, TTYPE_STRUCT, 1));
            J2T_ZERO(ienc.method->write_struct_begin(tproto));

                J2T_ZERO(ienc.method->write_field_begin(tproto, TTYPE_I32, 1));
                J2T_ZERO(ienc.method->write_i32(tproto, 7749));
                J2T_ZERO(ienc.method->write_field_end(tproto));
            
                J2T_ZERO(ienc.method->write_field_begin(tproto, TTYPE_LIST, 1));
            xprintf("---- GoSlice: %l\n", &buf);
                size_t back_off;
                J2T_ZERO(ienc.method->write_list_begin(tproto, TTYPE_I32, &back_off));
                #define LIST_N 10
                for (int i = 0; i < LIST_N; i++)
                    J2T_ZERO(ienc.method->write_i32(tproto, i));
                J2T_ZERO(ienc.method->write_list_end(tproto, back_off, LIST_N));
            xprintf("---- GoSlice: %l\n", &buf);
                J2T_ZERO(ienc.method->write_field_end(tproto));

            J2T_ZERO(ienc.method->write_field_stop(tproto));
            J2T_ZERO(ienc.method->write_struct_end(tproto));
        J2T_ZERO(ienc.method->write_field_end(tproto));
    
    J2T_ZERO(ienc.method->write_field_stop(tproto));
    J2T_ZERO(ienc.method->write_struct_end(tproto));
    // END ARGS STRUCT
    J2T_ZERO(ienc.method->write_message_end(tproto));

    xprintf("GoSlice: %l\n", &buf);
    return 0;
}

int main()
{
    uint64_t res;
    
    res = test_tcompact();
    xprintf("test_tcompact: %d\n", res);

    // FILE *fp = NULL;
    // fp = fopen("../testdata/twitter.json", "r");
    // char data[1000000];
    // size_t rn = fread(data, 1, 1000000, fp);
    // if (rn < 0)
    // {
    //     printf("read err: %zu\n", rn);
    //     return 0;
    // }

    // const char *data = EXAMPLE_JSON;
    // size_t rn = sizeof(EXAMPLE_JSON);

    // StateMachine sm = {};
    // _GoString s = {};
    // s = (_GoString){
    //     .buf = data,
    //     .len = rn,
    // };

    // long p = 0;
    // long q = skip_one(&s, &p, &sm);

    // xprintf("q: %d, p: %d\n", q, p);

    // INIT_BASE_DESC();
    // INIT_EXAMPLE_DESC();

    // j2t.reqs_cache = (_GoSlice){
    //     .buf = (char[72]){},
    //     .len = 0,
    //     .cap = 72,
    // };
    // j2t.key_cache = (_GoSlice){
    //     .buf = (char[EXMAPLE_MAX_KEY_LEN]){},
    //     .len = 0,
    //     .cap = EXMAPLE_MAX_KEY_LEN,
    // };
    // j2t.sp = 1;
    // j2t.vt[0] = (J2TState){
    //     .st = J2T_VAL,
    //     .jp = 0,
    //     .td = &EXAMPLE_DESC,
    // };
    // j2t.jt = (JsonState){
    //     .dbuf = DBUF,
    //     .dcap = 800,
    // };

    xprintf("sizof JsonState:%d, "
        "StateMachine:%d, "
        "J2TStateMachine:%d, "
        "J2TExtra:%d\n", 
        sizeof(JsonState), 
        sizeof(StateMachine), 
        sizeof(J2TStateMachine), 
        sizeof(J2TExtra));
    xprintf("offsetof: j2t.jt:%d, "
        "j2t.reqs_cache:%d, "
        "j2t.vt:%d, "
        "vt.jp:%d, "
        "vt.td:%d, "
        "vt.ex:%d\n", 
        __builtin_offsetof(J2TStateMachine, jt), 
        __builtin_offsetof(J2TStateMachine, reqs_cache), 
        __builtin_offsetof(J2TStateMachine, vt), 
        __builtin_offsetof(J2TState, jp), 
        __builtin_offsetof(J2TState, td), 
        __builtin_offsetof(J2TState, ex));

    // for (int i = 0; i < 800; i++)
    // {
    //     xprintf("%d", j2t.jt.dbuf[i]);
    // }
    // xprintf("j2t_fsm_exec: %d\nbuf:%l\n", j2t_fsm_exec(&j2t, &buf, &s, 0), &buf);
    return 0;
}