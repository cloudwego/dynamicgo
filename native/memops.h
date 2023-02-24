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

#ifndef MEMOPS_H
#define MEMOPS_H

#include "native.h"

void memcpy2(char *dp, const char *sp, size_t nb)
{
    for (int i = 0; i < nb; i++)
    {
        dp[i] = sp[i];
    }
}

bool memeq(const void *dp, const void *sp, size_t nb)
{
    for (int i = 0; i < nb; i++)
    {
        if (*(char *)(dp + i) != *(char *)(sp + i))
        {
            return false;
        }
    };
    return true;
}

#endif