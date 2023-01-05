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

// #include <stdio.h>
#include "../native.c"
#include "xprintf.h"
#include "example.h"

J2TStateMachine j2t = {};
GoSlice buf = (GoSlice){
    .buf = (char[1024]){},
    .len = 0,
    .cap = 1024,
};
char DBUF[800] = {};

int main()
{
    // FILE *fp = NULL;
    // fp = fopen("../testdata/twitter.json", "r");
    // char data[1000000];
    // size_t rn = fread(data, 1, 1000000, fp);
    // if (rn < 0)
    // {
    //     printf("read err: %zu\n", rn);
    //     return 0;
    // }

    const char *data = EXAMPLE_JSON;
    size_t rn = sizeof(EXAMPLE_JSON);

    StateMachine sm = {};
    GoString s = {};
    s = (GoString){
        .buf = data,
        .len = rn,
    };

    long p = 0;
    long q = skip_one(&s, &p, &sm);

    xprintf("q: %d, p: %d\n", q, p);

    INIT_BASE_DESC();
    INIT_EXAMPLE_DESC();

    j2t.reqs_cache = (GoSlice){
        .buf = (char[72]){},
        .len = 0,
        .cap = 72,
    };
    j2t.key_cache = (GoSlice){
        .buf = (char[EXMAPLE_MAX_KEY_LEN]){},
        .len = 0,
        .cap = EXMAPLE_MAX_KEY_LEN,
    };
    j2t.sp = 1;
    j2t.vt[0] = (J2TState){
        .st = J2T_VAL,
        .jp = 0,
        .td = &EXAMPLE_DESC,
    };
    j2t.jt = (JsonState){
        .dbuf = DBUF,
        .dcap = 800,
    };

    xprintf("sizof JsonState:%d, StateMachine:%d, J2TStateMachine:%d, J2TExtra:%d\n", sizeof(JsonState), sizeof(StateMachine), sizeof(J2TStateMachine), sizeof(J2TExtra));
    xprintf("offsetof: j2t.jt:%d, j2t.reqs_cache:%d, j2t.vt:%d, vt.jp:%d, vt.td:%d, vt.ex:%d\n", __offsetof(J2TStateMachine, jt), __offsetof(J2TStateMachine, reqs_cache), __offsetof(J2TStateMachine, vt), __offsetof(J2TState, jp), __offsetof(J2TState, td), __offsetof(J2TState, ex));

    // for (int i = 0; i < 800; i++)
    // {
    //     xprintf("%d", j2t.jt.dbuf[i]);
    // }
    xprintf("j2t_fsm_exec: %d\nbuf:%l\n", j2t_fsm_exec(&j2t, &buf, &s, 0), &buf);
    return 0;
}