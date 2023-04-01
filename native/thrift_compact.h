#include "native.h"
#include "map.h"
#include "thrift.h"

#ifndef THRIFT_COMPACT_H
#define THRIFT_COMPACT_H

// #define TTC_MAX_STRUCT_DEPTH 8

#define TTC_PROTOCOL_ID 0x082
#define TTC_VERSION 1
#define TTC_VERSION_MASK 0x1f
#define TTC_TYPE_MASK 0x0E0
#define TTC_TYPE_BITS 0x07
#define TTC_TYPE_SHIFT_AMOUNT 5

typedef uint8_t tcompacttype;

#define TTC_EINVAL (tcompacttype)0xff
#define TTC_STOP (tcompacttype)TTYPE_STOP
#define TTC_BOOLEAN_TRUE (tcompacttype)0x01
#define TTC_BOOLEAN_FALSE (tcompacttype)0x02
#define TTC_BYTE (tcompacttype)0x03
#define TTC_I16 (tcompacttype)0x04
#define TTC_I32 (tcompacttype)0x05
#define TTC_I64 (tcompacttype)0x06
#define TTC_DOUBLE (tcompacttype)0x07
#define TTC_BINARY (tcompacttype)0x08
#define TTC_LIST (tcompacttype)0x09
#define TTC_SET (tcompacttype)0x0A
#define TTC_MAP (tcompacttype)0x0B
#define TTC_STRUCT (tcompacttype)0x0C
#define TTC_LAST TTC_STRUCT

#define TTC_I64ToZigzag(v) (int64_t)( (v << 1) ^ (v >> 63) )
#define TTC_ZigzagToI64(v) (int64_t)((uint64_t)v>>1) ^ -((uint64_t)v & 1)

#define TTC_I32ToZigzag(v) (int32_t)( (v << 1) ^ (v >> 31) )
#define TTC_ZigzagToI32(v) (int32_t)((uint32_t)v>>1) ^ -((uint32_t)v & 1)

#define AS_TC(p) ((tc_state *)p)


typedef struct {
    _GoSlice *buf;
    struct {
        int16_t id;
        ttype type;
        bool valid;
    } pending_field_write;
    int16_t last_field_id;
    Int16Slice last_field_id_stack;
    Uint16Slice container_write_back_stack;
} tc_state;


typedef struct {
    uint64_t (*write_default_or_empty)(tc_state *, const tFieldDesc *, long);
    uint64_t (*write_data_count)(tc_state *, size_t n);
    size_t (*write_data_count_max_length)(tc_state *);
} tc_ienc_extra;

typedef struct {
    // resv0 = J2TStateMachine
    // resv1 = buf (output buffer / GoSlice)
    // resv2 = unk
    // resv3 = tc_state
    vt_base base;

    uint64_t (*write_message_begin)(tc_state *, const _GoString, int32_t, int32_t);
    uint64_t (*write_message_end)(tc_state *);
    uint64_t (*write_struct_begin)(tc_state *);
    uint64_t (*write_struct_end)(tc_state *);
    uint64_t (*write_field_begin)(tc_state *, ttype, int16_t);
    uint64_t (*write_field_end)(tc_state *);
    uint64_t (*write_field_stop)(tc_state *);
    
    uint64_t (*write_map_n)(tc_state *, ttype, ttype, size_t);
    uint64_t (*write_map_begin)(tc_state *, ttype, ttype, size_t *);
    uint64_t (*write_map_end)(tc_state *, size_t, size_t);

    uint64_t (*write_list_n)(tc_state *, ttype, size_t);
    uint64_t (*write_list_begin)(tc_state *, ttype, size_t *);
    uint64_t (*write_list_end)(tc_state *, size_t, size_t);
    
    uint64_t (*write_set_n)(tc_state *, ttype, size_t);
    uint64_t (*write_set_begin)(tc_state *, ttype, size_t *);
    uint64_t (*write_set_end)(tc_state *, size_t, size_t);


    uint64_t (*write_bool)(tc_state *, bool);
    uint64_t (*write_byte)(tc_state *, char);
    uint64_t (*write_i16)(tc_state *, int16_t);
    uint64_t (*write_i32)(tc_state *, int32_t);
    uint64_t (*write_i64)(tc_state *, int64_t);
    uint64_t (*write_double)(tc_state *, double);
    uint64_t (*write_string)(tc_state *, const char *, size_t);
    uint64_t (*write_binary)(tc_state *, const _GoSlice);

    tc_ienc_extra extra;
} tc_ienc;


tc_ienc tc_get_iencoder();

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

#endif