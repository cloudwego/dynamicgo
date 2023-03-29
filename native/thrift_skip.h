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

int64_t tb_skip(skipbuf_t *st, const char *s, int64_t n, uint8_t t);