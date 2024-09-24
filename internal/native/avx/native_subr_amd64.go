// +build !noasm !appengine
// Code generated by asm2asm, DO NOT EDIT.

package avx

import (
	`github.com/bytedance/sonic/loader`
)

const (
    _entry__f64toa = 496
    _entry__format_significand = 49312
    _entry__format_integer = 3744
    _entry__hm_get = 28656
    _entry__i64toa = 4176
    _entry__u64toa = 4448
    _entry__j2t_fsm_exec = 39920
    _entry__advance_ns = 14336
    _entry__fsm_exec = 20928
    _entry__advance_string = 16480
    _entry__validate_string = 23040
    _entry__utf8_validate = 24416
    _entry__skip_number = 19872
    _entry__j2t_number = 35600
    _entry__vnumber = 17440
    _entry__atof_eisel_lemire64 = 11216
    _entry__atof_native = 13664
    _entry__decimal_to_f64 = 11712
    _entry__left_shift = 49792
    _entry__right_shift = 50336
    _entry__j2t_string = 36192
    _entry__unquote = 7584
    _entry__unhex16_is = 9744
    _entry__j2t_binary = 37072
    _entry__b64decode = 25760
    _entry__j2t_field_vm = 38720
    _entry__tb_write_string = 30432
    _entry__tb_write_default_or_empty = 32400
    _entry__j2t_write_unset_fields = 34944
    _entry__j2t_find_field_key = 38048
    _entry__quote = 5872
    _entry__tb_skip = 48256
    _entry__tb_write_i64 = 30304
    _entry__trie_get = 29536
)

const (
    _stack__f64toa = 80
    _stack__format_significand = 24
    _stack__format_integer = 16
    _stack__hm_get = 24
    _stack__i64toa = 16
    _stack__u64toa = 8
    _stack__j2t_fsm_exec = 560
    _stack__advance_ns = 8
    _stack__fsm_exec = 224
    _stack__advance_string = 56
    _stack__validate_string = 136
    _stack__utf8_validate = 32
    _stack__skip_number = 32
    _stack__j2t_number = 296
    _stack__vnumber = 240
    _stack__atof_eisel_lemire64 = 40
    _stack__atof_native = 136
    _stack__decimal_to_f64 = 80
    _stack__left_shift = 24
    _stack__right_shift = 16
    _stack__j2t_string = 184
    _stack__unquote = 112
    _stack__unhex16_is = 8
    _stack__j2t_binary = 224
    _stack__b64decode = 152
    _stack__j2t_field_vm = 312
    _stack__tb_write_string = 8
    _stack__tb_write_default_or_empty = 56
    _stack__j2t_write_unset_fields = 160
    _stack__j2t_find_field_key = 64
    _stack__quote = 80
    _stack__tb_skip = 48
    _stack__tb_write_i64 = 8
    _stack__trie_get = 32
)

const (
    _size__f64toa = 3248
    _size__format_significand = 480
    _size__format_integer = 432
    _size__hm_get = 544
    _size__i64toa = 272
    _size__u64toa = 1376
    _size__j2t_fsm_exec = 7804
    _size__advance_ns = 608
    _size__fsm_exec = 1336
    _size__advance_string = 912
    _size__validate_string = 1376
    _size__utf8_validate = 408
    _size__skip_number = 876
    _size__j2t_number = 548
    _size__vnumber = 1600
    _size__atof_eisel_lemire64 = 400
    _size__atof_native = 608
    _size__decimal_to_f64 = 1952
    _size__left_shift = 544
    _size__right_shift = 464
    _size__j2t_string = 880
    _size__unquote = 2160
    _size__unhex16_is = 128
    _size__j2t_binary = 224
    _size__b64decode = 2064
    _size__j2t_field_vm = 1124
    _size__tb_write_string = 656
    _size__tb_write_default_or_empty = 1160
    _size__j2t_write_unset_fields = 656
    _size__j2t_find_field_key = 384
    _size__quote = 1696
    _size__tb_skip = 972
    _size__tb_write_i64 = 64
    _size__trie_get = 336
)

var (
    _pcsp__f64toa = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {12, 40},
        {13, 48},
        {3176, 56},
        {3180, 48},
        {3181, 40},
        {3183, 32},
        {3185, 24},
        {3187, 16},
        {3189, 8},
        {3193, 0},
        {3248, 56},
    }
    _pcsp__format_significand = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {468, 24},
        {469, 16},
        {471, 8},
        {473, 0},
    }
    _pcsp__format_integer = [][2]uint32{
        {1, 0},
        {4, 8},
        {412, 16},
        {413, 8},
        {414, 0},
        {423, 16},
        {424, 8},
        {426, 0},
    }
    _pcsp__hm_get = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {531, 24},
        {532, 16},
        {534, 8},
        {536, 0},
    }
    _pcsp__i64toa = [][2]uint32{
        {1, 0},
        {171, 8},
        {172, 0},
        {207, 8},
        {208, 0},
        {222, 8},
        {223, 0},
        {247, 8},
        {248, 0},
        {253, 8},
        {259, 0},
    }
    _pcsp__u64toa = [][2]uint32{
        {13, 0},
        {162, 8},
        {163, 0},
        {175, 8},
        {240, 0},
        {498, 8},
        {499, 0},
        {519, 8},
        {592, 0},
        {850, 8},
        {928, 0},
        {1374, 8},
        {1376, 0},
    }
    _pcsp__j2t_fsm_exec = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {12, 40},
        {13, 48},
        {6550, 248},
        {6557, 48},
        {6558, 40},
        {6560, 32},
        {6562, 24},
        {6564, 16},
        {6566, 8},
        {6570, 0},
        {7804, 248},
    }
    _pcsp__advance_ns = [][2]uint32{
        {1, 0},
        {573, 8},
        {574, 0},
        {580, 8},
        {581, 0},
        {593, 8},
    }
    _pcsp__fsm_exec = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {12, 40},
        {13, 48},
        {1134, 88},
        {1138, 48},
        {1139, 40},
        {1141, 32},
        {1143, 24},
        {1145, 16},
        {1147, 8},
        {1148, 0},
        {1336, 88},
    }
    _pcsp__advance_string = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {12, 40},
        {13, 48},
        {381, 56},
        {385, 48},
        {386, 40},
        {388, 32},
        {390, 24},
        {392, 16},
        {394, 8},
        {395, 0},
        {912, 56},
    }
    _pcsp__validate_string = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {12, 40},
        {13, 48},
        {603, 104},
        {607, 48},
        {608, 40},
        {610, 32},
        {612, 24},
        {614, 16},
        {616, 8},
        {617, 0},
        {1376, 104},
    }
    _pcsp__utf8_validate = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {398, 32},
        {399, 24},
        {401, 16},
        {403, 8},
        {408, 0},
    }
    _pcsp__skip_number = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {739, 32},
        {740, 24},
        {742, 16},
        {744, 8},
        {745, 0},
        {876, 32},
    }
    _pcsp__j2t_number = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {12, 40},
        {13, 48},
        {531, 56},
        {535, 48},
        {536, 40},
        {538, 32},
        {540, 24},
        {542, 16},
        {544, 8},
        {548, 0},
    }
    _pcsp__vnumber = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {12, 40},
        {13, 48},
        {145, 104},
        {149, 48},
        {150, 40},
        {152, 32},
        {154, 24},
        {156, 16},
        {158, 8},
        {159, 0},
        {1595, 104},
    }
    _pcsp__atof_eisel_lemire64 = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {322, 40},
        {323, 32},
        {325, 24},
        {327, 16},
        {329, 8},
        {330, 0},
        {394, 40},
    }
    _pcsp__atof_native = [][2]uint32{
        {1, 0},
        {4, 8},
        {596, 56},
        {600, 8},
        {602, 0},
    }
    _pcsp__decimal_to_f64 = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {12, 40},
        {13, 48},
        {1909, 56},
        {1913, 48},
        {1914, 40},
        {1916, 32},
        {1918, 24},
        {1920, 16},
        {1922, 8},
        {1926, 0},
        {1938, 56},
    }
    _pcsp__left_shift = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {402, 24},
        {403, 16},
        {405, 8},
        {406, 0},
        {414, 24},
        {415, 16},
        {417, 8},
        {418, 0},
        {539, 24},
    }
    _pcsp__right_shift = [][2]uint32{
        {1, 0},
        {4, 8},
        {436, 16},
        {437, 8},
        {438, 0},
        {446, 16},
        {447, 8},
        {448, 0},
        {456, 16},
        {457, 8},
        {459, 0},
    }
    _pcsp__j2t_string = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {12, 40},
        {13, 48},
        {643, 72},
        {647, 48},
        {648, 40},
        {650, 32},
        {652, 24},
        {654, 16},
        {656, 8},
        {660, 0},
        {869, 72},
    }
    _pcsp__unquote = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {12, 40},
        {13, 48},
        {1587, 104},
        {1591, 48},
        {1592, 40},
        {1594, 32},
        {1596, 24},
        {1598, 16},
        {1600, 8},
        {1601, 0},
        {2156, 104},
    }
    _pcsp__unhex16_is = [][2]uint32{
        {1, 0},
        {35, 8},
        {36, 0},
        {62, 8},
        {63, 0},
        {97, 8},
        {98, 0},
        {121, 8},
        {123, 0},
    }
    _pcsp__j2t_binary = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {12, 40},
        {13, 48},
        {179, 72},
        {183, 48},
        {184, 40},
        {186, 32},
        {188, 24},
        {190, 16},
        {192, 8},
        {193, 0},
        {221, 72},
    }
    _pcsp__b64decode = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {12, 40},
        {13, 48},
        {2035, 152},
        {2039, 48},
        {2040, 40},
        {2042, 32},
        {2044, 24},
        {2046, 16},
        {2048, 8},
        {2050, 0},
    }
    _pcsp__j2t_field_vm = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {12, 40},
        {13, 48},
        {342, 72},
        {346, 48},
        {347, 40},
        {349, 32},
        {351, 24},
        {353, 16},
        {355, 8},
        {356, 0},
        {673, 72},
        {677, 48},
        {678, 40},
        {680, 32},
        {682, 24},
        {684, 16},
        {686, 8},
        {687, 0},
        {777, 72},
        {781, 48},
        {782, 40},
        {784, 32},
        {786, 24},
        {788, 16},
        {790, 8},
        {791, 0},
        {1124, 72},
    }
    _pcsp__tb_write_string = [][2]uint32{
        {1, 0},
        {45, 8},
        {46, 0},
        {275, 8},
        {279, 0},
        {627, 8},
        {631, 0},
        {649, 8},
    }
    _pcsp__tb_write_default_or_empty = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {12, 40},
        {13, 48},
        {1094, 56},
        {1098, 48},
        {1099, 40},
        {1101, 32},
        {1103, 24},
        {1105, 16},
        {1107, 8},
        {1111, 0},
        {1160, 56},
    }
    _pcsp__j2t_write_unset_fields = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {12, 40},
        {13, 48},
        {571, 104},
        {575, 48},
        {576, 40},
        {578, 32},
        {580, 24},
        {582, 16},
        {584, 8},
        {585, 0},
        {652, 104},
    }
    _pcsp__j2t_find_field_key = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {349, 40},
        {350, 32},
        {352, 24},
        {354, 16},
        {356, 8},
        {357, 0},
        {361, 40},
        {362, 32},
        {364, 24},
        {366, 16},
        {368, 8},
        {374, 0},
    }
    _pcsp__quote = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {12, 40},
        {13, 48},
        {1643, 80},
        {1647, 48},
        {1648, 40},
        {1650, 32},
        {1652, 24},
        {1654, 16},
        {1656, 8},
        {1657, 0},
        {1692, 80},
    }
    _pcsp__tb_skip = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {12, 40},
        {947, 48},
        {948, 40},
        {950, 32},
        {952, 24},
        {954, 16},
        {956, 8},
        {957, 0},
        {972, 48},
    }
    _pcsp__tb_write_i64 = [][2]uint32{
        {1, 0},
        {33, 8},
        {34, 0},
        {51, 8},
        {53, 0},
    }
    _pcsp__trie_get = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {323, 32},
        {324, 24},
        {326, 16},
        {328, 8},
        {330, 0},
    }
)

var Funcs = []loader.CFunc{
    {"__native_entry__", 0, 67, 0, nil},
    {"_f64toa", _entry__f64toa, _size__f64toa, _stack__f64toa, _pcsp__f64toa},
    {"_format_significand", _entry__format_significand, _size__format_significand, _stack__format_significand, _pcsp__format_significand},
    {"_format_integer", _entry__format_integer, _size__format_integer, _stack__format_integer, _pcsp__format_integer},
    {"_hm_get", _entry__hm_get, _size__hm_get, _stack__hm_get, _pcsp__hm_get},
    {"_i64toa", _entry__i64toa, _size__i64toa, _stack__i64toa, _pcsp__i64toa},
    {"_u64toa", _entry__u64toa, _size__u64toa, _stack__u64toa, _pcsp__u64toa},
    {"_j2t_fsm_exec", _entry__j2t_fsm_exec, _size__j2t_fsm_exec, _stack__j2t_fsm_exec, _pcsp__j2t_fsm_exec},
    {"_advance_ns", _entry__advance_ns, _size__advance_ns, _stack__advance_ns, _pcsp__advance_ns},
    {"_fsm_exec", _entry__fsm_exec, _size__fsm_exec, _stack__fsm_exec, _pcsp__fsm_exec},
    {"_advance_string", _entry__advance_string, _size__advance_string, _stack__advance_string, _pcsp__advance_string},
    {"_validate_string", _entry__validate_string, _size__validate_string, _stack__validate_string, _pcsp__validate_string},
    {"_utf8_validate", _entry__utf8_validate, _size__utf8_validate, _stack__utf8_validate, _pcsp__utf8_validate},
    {"_skip_number", _entry__skip_number, _size__skip_number, _stack__skip_number, _pcsp__skip_number},
    {"_j2t_number", _entry__j2t_number, _size__j2t_number, _stack__j2t_number, _pcsp__j2t_number},
    {"_vnumber", _entry__vnumber, _size__vnumber, _stack__vnumber, _pcsp__vnumber},
    {"_atof_eisel_lemire64", _entry__atof_eisel_lemire64, _size__atof_eisel_lemire64, _stack__atof_eisel_lemire64, _pcsp__atof_eisel_lemire64},
    {"_atof_native", _entry__atof_native, _size__atof_native, _stack__atof_native, _pcsp__atof_native},
    {"_decimal_to_f64", _entry__decimal_to_f64, _size__decimal_to_f64, _stack__decimal_to_f64, _pcsp__decimal_to_f64},
    {"_left_shift", _entry__left_shift, _size__left_shift, _stack__left_shift, _pcsp__left_shift},
    {"_right_shift", _entry__right_shift, _size__right_shift, _stack__right_shift, _pcsp__right_shift},
    {"_j2t_string", _entry__j2t_string, _size__j2t_string, _stack__j2t_string, _pcsp__j2t_string},
    {"_unquote", _entry__unquote, _size__unquote, _stack__unquote, _pcsp__unquote},
    {"_unhex16_is", _entry__unhex16_is, _size__unhex16_is, _stack__unhex16_is, _pcsp__unhex16_is},
    {"_j2t_binary", _entry__j2t_binary, _size__j2t_binary, _stack__j2t_binary, _pcsp__j2t_binary},
    {"_b64decode", _entry__b64decode, _size__b64decode, _stack__b64decode, _pcsp__b64decode},
    {"_j2t_field_vm", _entry__j2t_field_vm, _size__j2t_field_vm, _stack__j2t_field_vm, _pcsp__j2t_field_vm},
    {"_tb_write_string", _entry__tb_write_string, _size__tb_write_string, _stack__tb_write_string, _pcsp__tb_write_string},
    {"_tb_write_default_or_empty", _entry__tb_write_default_or_empty, _size__tb_write_default_or_empty, _stack__tb_write_default_or_empty, _pcsp__tb_write_default_or_empty},
    {"_j2t_write_unset_fields", _entry__j2t_write_unset_fields, _size__j2t_write_unset_fields, _stack__j2t_write_unset_fields, _pcsp__j2t_write_unset_fields},
    {"_j2t_find_field_key", _entry__j2t_find_field_key, _size__j2t_find_field_key, _stack__j2t_find_field_key, _pcsp__j2t_find_field_key},
    {"_quote", _entry__quote, _size__quote, _stack__quote, _pcsp__quote},
    {"_tb_skip", _entry__tb_skip, _size__tb_skip, _stack__tb_skip, _pcsp__tb_skip},
    {"_tb_write_i64", _entry__tb_write_i64, _size__tb_write_i64, _stack__tb_write_i64, _pcsp__tb_write_i64},
    {"_trie_get", _entry__trie_get, _size__trie_get, _stack__trie_get, _pcsp__trie_get},
}
