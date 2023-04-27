//go:build !sharedlib
// +build !sharedlib

//
// Copyright 2023 CloudWeGo Authors·
//
// Licensed under the Apache License, Version 2·0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

#include "go_asm.h"
#include "funcdata.h"
#include "textflag.h"

TEXT ·Quote(SB), NOSPLIT, $0 - 48
    CMPB github·com∕cloudwego∕dynamicgo∕internal∕cpu·HasAVX2(SB), $0
    JE   2(PC)
    JMP  github·com∕cloudwego∕dynamicgo∕internal∕native∕avx2·__quote(SB)
    JMP  github·com∕cloudwego∕dynamicgo∕internal∕native∕avx·__quote(SB)

TEXT ·Unquote(SB), NOSPLIT, $0 - 48
    CMPB github·com∕cloudwego∕dynamicgo∕internal∕cpu·HasAVX2(SB), $0
    JE   2(PC)
    JMP  github·com∕cloudwego∕dynamicgo∕internal∕native∕avx2·__unquote(SB)
    JMP  github·com∕cloudwego∕dynamicgo∕internal∕native∕avx·__unquote(SB)

TEXT ·HTMLEscape(SB), NOSPLIT, $0 - 40
    CMPB github·com∕cloudwego∕dynamicgo∕internal∕cpu·HasAVX2(SB), $0
    JE   2(PC)
    JMP  github·com∕cloudwego∕dynamicgo∕internal∕native∕avx2·__html_escape(SB)
    JMP  github·com∕cloudwego∕dynamicgo∕internal∕native∕avx·__html_escape(SB)

TEXT ·Value(SB), NOSPLIT, $0 - 48
    CMPB github·com∕cloudwego∕dynamicgo∕internal∕cpu·HasAVX2(SB), $0
    JE   2(PC)
    JMP  github·com∕cloudwego∕dynamicgo∕internal∕native∕avx2·__value(SB)
    JMP  github·com∕cloudwego∕dynamicgo∕internal∕native∕avx·__value(SB)

TEXT ·SkipOne(SB), NOSPLIT, $0 - 32
    CMPB github·com∕cloudwego∕dynamicgo∕internal∕cpu·HasAVX2(SB), $0
    JE   2(PC)
    JMP  github·com∕cloudwego∕dynamicgo∕internal∕native∕avx2·__skip_one(SB)
    JMP  github·com∕cloudwego∕dynamicgo∕internal∕native∕avx·__skip_one(SB)

TEXT ·ValidateOne(SB), NOSPLIT, $0 - 32
    CMPB github·com∕cloudwego∕dynamicgo∕internal∕cpu·HasAVX2(SB), $0
    JE   2(PC)
    JMP  github·com∕cloudwego∕dynamicgo∕internal∕native∕avx2·__validate_one(SB)
    JMP  github·com∕cloudwego∕dynamicgo∕internal∕native∕avx·__validate_one(SB)

TEXT ·I64toa(SB), NOSPLIT, $0 - 32
    CMPB github·com∕cloudwego∕dynamicgo∕internal∕cpu·HasAVX2(SB), $0
    JE   2(PC)
    JMP  github·com∕cloudwego∕dynamicgo∕internal∕native∕avx2·__i64toa(SB)
    JMP  github·com∕cloudwego∕dynamicgo∕internal∕native∕avx·__i64toa(SB)

TEXT ·U64toa(SB), NOSPLIT, $0 - 32
    CMPB github·com∕cloudwego∕dynamicgo∕internal∕cpu·HasAVX2(SB), $0
    JE   2(PC)
    JMP  github·com∕cloudwego∕dynamicgo∕internal∕native∕avx2·__u64toa(SB)
    JMP  github·com∕cloudwego∕dynamicgo∕internal∕native∕avx·__u64toa(SB)

TEXT ·F64toa(SB), NOSPLIT, $0 - 32
    CMPB github·com∕cloudwego∕dynamicgo∕internal∕cpu·HasAVX2(SB), $0
    JE   2(PC)
    JMP  github·com∕cloudwego∕dynamicgo∕internal∕native∕avx2·__f64toa(SB)
    JMP  github·com∕cloudwego∕dynamicgo∕internal∕native∕avx·__f64toa(SB)

TEXT ·J2T_FSM(SB), NOSPLIT, $0 - 40
    PCDATA $0, $-2
    CMPB github·com∕cloudwego∕dynamicgo∕internal∕cpu·HasAVX2(SB), $0
    JE   2(PC)
    JMP  github·com∕cloudwego∕dynamicgo∕internal∕native∕avx2·__j2t_fsm_exec(SB)
    JMP  github·com∕cloudwego∕dynamicgo∕internal∕native∕avx·__j2t_fsm_exec(SB)

TEXT ·TBSkip(SB), NOSPLIT, $0 - 40
    PCDATA $0, $-2
    CMPB github·com∕cloudwego∕dynamicgo∕internal∕cpu·HasAVX2(SB), $0
    JE   2(PC)
    JMP  github·com∕cloudwego∕dynamicgo∕internal∕native∕avx2·__tb_skip(SB)
    JMP  github·com∕cloudwego∕dynamicgo∕internal∕native∕avx·__tb_skip(SB)
