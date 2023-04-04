#include "native.h"
#include "map.h"
#include "thrift.h"

#ifndef THRIFT_COMPACT_H
#define THRIFT_COMPACT_H

// #define TTC_MAX_STRUCT_DEPTH 8

#define TC_VARINT32_BUF_SZ  (size_t)5
#define TC_VARINT64_BUF_SZ  (size_t)9

#define TTC_PROTOCOL_ID         0x082
#define TTC_VERSION             1
#define TTC_VERSION_MASK        0x1f
#define TTC_TYPE_MASK           0x0E0
#define TTC_TYPE_BITS           0x07
#define TTC_TYPE_SHIFT_AMOUNT   5

typedef uint8_t tcompacttype;

#define TTC_EINVAL          (tcompacttype)0xff
#define TTC_STOP            (tcompacttype)TTYPE_STOP
#define TTC_BOOLEAN_TRUE    (tcompacttype)0x01
#define TTC_BOOLEAN_FALSE   (tcompacttype)0x02
#define TTC_BYTE            (tcompacttype)0x03
#define TTC_I16             (tcompacttype)0x04
#define TTC_I32             (tcompacttype)0x05
#define TTC_I64             (tcompacttype)0x06
#define TTC_DOUBLE          (tcompacttype)0x07
#define TTC_BINARY          (tcompacttype)0x08
#define TTC_LIST            (tcompacttype)0x09
#define TTC_SET             (tcompacttype)0x0A
#define TTC_MAP             (tcompacttype)0x0B
#define TTC_STRUCT          (tcompacttype)0x0C
#define TTC_LAST            TTC_STRUCT+1

#define TTC_I64ToZigzag(v)  (int64_t)( (v << 1) ^ (v >> 63) )
#define TTC_ZigzagToI64(v)  (int64_t)((uint64_t)v>>1) ^ -((uint64_t)v & 1)

#define TTC_I32ToZigzag(v)  (int32_t)( (v << 1) ^ (v >> 31) )
#define TTC_ZigzagToI32(v)  (int32_t)((uint32_t)v>>1) ^ -((uint32_t)v & 1)

// u16 MSB [key u8  ----|----  elem u8]
#define TC_CNTR_U16_VAL2(key,elem) (((uint8_t)(key&0xff)<<8)|((uint8_t)elem&0xff))
// u16 MSB [ignored ----|----  elem u8]
#define TC_CNTR_U16_VAL1(elem) TC_CNTR_U16_VAL2(0, elem)
#define TC_CNTR_RPUSH(q,v)                      \
    do                                          \
    {                                           \
        size_t off = pq->len;                   \
        if (off > pq->cap)                      \
            WRAP_ERR0(ERR_OOM_CWBS, off+1);     \
        pq->buf[pq->len++] = v;                 \
    } while (0)
#define TC_CNTR_RPOP(q)     \
    pval = q->len > 0       \
        ? q->buf[--q->len]  \
        : 0



TIMPL_IENC(tc_state *, tc_ienc, tc_menc);
// tc_ienc tc_get_iencoder();

uint64_t tc_write_message_begin(tc_state *self, const _GoString name, int32_t type, int32_t seq);
uint64_t tc_write_message_end(tc_state *self);

uint64_t tc_write_struct_begin(tc_state *self);
uint64_t tc_write_struct_end(tc_state *self);

uint64_t tc_write_field_begin(tc_state *self, ttype type, int16_t id);
uint64_t tc_write_field_end(tc_state *self);
uint64_t tc_write_field_stop(tc_state *self);

uint64_t tc_write_map_n(tc_state *self, ttype key, ttype val, size_t size);
uint64_t tc_write_map_begin(tc_state *self, ttype key, ttype val, size_t *size);
uint64_t tc_write_map_end(tc_state *self, size_t back_off, size_t vsize);

uint64_t tc_write_list_n(tc_state *self, ttype elem, size_t size);
uint64_t tc_write_list_begin(tc_state *self, ttype elem, size_t *backp);
uint64_t tc_write_list_end(tc_state *self, size_t back_off, size_t vsize);

uint64_t tc_write_set_n(tc_state *self, ttype elem, size_t size);
uint64_t tc_write_set_begin(tc_state *self, ttype elem, size_t *backp);
uint64_t tc_write_set_end(tc_state *self, size_t back_off, size_t vsize);

uint64_t tc_write_bool(tc_state *self, bool v);
static inline uint64_t tc_write_byte(tc_state *self, char b);
uint64_t tc_write_i16(tc_state *self, int16_t v);
uint64_t tc_write_i32(tc_state *self, int32_t v);
uint64_t tc_write_i64(tc_state *self, int64_t v);
uint64_t tc_write_double(tc_state *self, double v);
uint64_t tc_write_string(tc_state *self, const char *v, size_t n);
uint64_t tc_write_binary(tc_state *self, const _GoSlice v);

uint64_t j2t2_fsm_tc_exec(J2TStateMachine *self, _GoSlice *buf, const _GoString *src, uint64_t flag);

#endif // THRIFT_COMPACT_H