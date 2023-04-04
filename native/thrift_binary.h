#include "native.h"
#include "map.h"
#include "thrift.h"

#ifndef THRIFT_BINARY_H
#define THRIFT_BINARY_H

#define TB_DATA_COUNT_LENGTH 4

#define TB_THRIFT_VERSION_1 0x80010000ul

TIMPL_IENC(tb_state *, tb_ienc, tb_menc);
// tb_ienc tb_get_iencoder();

uint64_t tb_write_map_n(tb_state *self, ttype key, ttype val, size_t n);
uint64_t tb_write_map_begin(tb_state *self, ttype key, ttype val, size_t *backp);
uint64_t tb_write_map_end(tb_state *self, size_t back_off, size_t vsize);

uint64_t tb_write_list_n(tb_state *self, ttype elem, size_t size);
uint64_t tb_write_list_begin(tb_state *self, ttype elem, size_t *backp);
uint64_t tb_write_list_end(tb_state *self, size_t back_off, size_t vsize);


uint64_t tb_write_set_n(tb_state *self, ttype elem, size_t size);
uint64_t tb_write_set_begin(tb_state *self, ttype elem, size_t *backp);
uint64_t tb_write_set_end(tb_state *self, size_t back_off, size_t vsize);

uint64_t tb_write_default_or_empty(tb_state *self, const tFieldDesc *field, long p);

uint64_t j2t2_fsm_tb_exec(J2TStateMachine *self, _GoSlice *buf, const _GoString *src, uint64_t flag);

#endif // THRIFT_COMPACT_H
