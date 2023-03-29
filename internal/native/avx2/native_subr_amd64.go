// +build !noasm !appengine
// Code generated by asm2asm, DO NOT EDIT.

package avx2

import (
	`github.com/bytedance/sonic/loader`
)

var Stubs = []loader.GoC{
    {"_f64toa", &_subr__f64toa, &__f64toa},
    {"_fsm_exec", &_subr__fsm_exec, &__fsm_exec},
    {"_hm_get", &_subr__hm_get, &__hm_get},
    {"_html_escape", &_subr__html_escape, &__html_escape},
    {"_i64toa", &_subr__i64toa, &__i64toa},
    {"_j2t_fsm_exec", &_subr__j2t_fsm_exec, &__j2t_fsm_exec},
    {"_lspace", &_subr__lspace, &__lspace},
    {"_quote", &_subr__quote, &__quote},
    {"_skip_array", &_subr__skip_array, &__skip_array},
    {"_skip_object", &_subr__skip_object, &__skip_object},
    {"_skip_one", &_subr__skip_one, &__skip_one},
    {"_tb_skip", &_subr__tb_skip, &__tb_skip},
    {"_tb_write_i64", &_subr__tb_write_i64, &__tb_write_i64},
    {"_trie_get", &_subr__trie_get, &__trie_get},
    {"_u64toa", &_subr__u64toa, &__u64toa},
    {"_unquote", &_subr__unquote, &__unquote},
    {"_validate_one", &_subr__validate_one, &__validate_one},
    {"_value", &_subr__value, &__value},
    {"_vnumber", &_subr__vnumber, &__vnumber},
    {"_vsigned", &_subr__vsigned, &__vsigned},
    {"_vstring", &_subr__vstring, &__vstring},
    {"_vunsigned", &_subr__vunsigned, &__vunsigned},
}

var Funcs = []loader.CFunc{
    {"__native_entry__", 0, 67, 0, nil},
    {"_f64toa", _entry__f64toa, _size__f64toa, _stack__f64toa, _pcsp__f64toa},
    {"_format_significand", _entry__format_significand, _size__format_significand, _stack__format_significand, _pcsp__format_significand},
    {"_format_integer", _entry__format_integer, _size__format_integer, _stack__format_integer, _pcsp__format_integer},
    {"_fsm_exec", _entry__fsm_exec, _size__fsm_exec, _stack__fsm_exec, _pcsp__fsm_exec},
    {"_advance_ns", _entry__advance_ns, _size__advance_ns, _stack__advance_ns, _pcsp__advance_ns},
    {"_validate_string", _entry__validate_string, _size__validate_string, _stack__validate_string, _pcsp__validate_string},
    {"_utf8_validate", _entry__utf8_validate, _size__utf8_validate, _stack__utf8_validate, _pcsp__utf8_validate},
    {"_advance_string", _entry__advance_string, _size__advance_string, _stack__advance_string, _pcsp__advance_string},
    {"_skip_number", _entry__skip_number, _size__skip_number, _stack__skip_number, _pcsp__skip_number},
    {"_hm_get", _entry__hm_get, _size__hm_get, _stack__hm_get, _pcsp__hm_get},
    {"_html_escape", _entry__html_escape, _size__html_escape, _stack__html_escape, _pcsp__html_escape},
    {"_i64toa", _entry__i64toa, _size__i64toa, _stack__i64toa, _pcsp__i64toa},
    {"_u64toa", _entry__u64toa, _size__u64toa, _stack__u64toa, _pcsp__u64toa},
    {"_j2t_fsm_exec", _entry__j2t_fsm_exec, _size__j2t_fsm_exec, _stack__j2t_fsm_exec, _pcsp__j2t_fsm_exec},
    {"_j2t_number", _entry__j2t_number, _size__j2t_number, _stack__j2t_number, _pcsp__j2t_number},
    {"_vnumber", _entry__vnumber, _size__vnumber, _stack__vnumber, _pcsp__vnumber},
    {"_atof_eisel_lemire64", _entry__atof_eisel_lemire64, _size__atof_eisel_lemire64, _stack__atof_eisel_lemire64, _pcsp__atof_eisel_lemire64},
    {"_atof_native", _entry__atof_native, _size__atof_native, _stack__atof_native, _pcsp__atof_native},
    {"_decimal_to_f64", _entry__decimal_to_f64, _size__decimal_to_f64, _stack__decimal_to_f64, _pcsp__decimal_to_f64},
    {"_right_shift", _entry__right_shift, _size__right_shift, _stack__right_shift, _pcsp__right_shift},
    {"_left_shift", _entry__left_shift, _size__left_shift, _stack__left_shift, _pcsp__left_shift},
    {"_j2t_string", _entry__j2t_string, _size__j2t_string, _stack__j2t_string, _pcsp__j2t_string},
    {"_unquote", _entry__unquote, _size__unquote, _stack__unquote, _pcsp__unquote},
    {"_b64decode", _entry__b64decode, _size__b64decode, _stack__b64decode, _pcsp__b64decode},
    {"_j2t_field_vm", _entry__j2t_field_vm, _size__j2t_field_vm, _stack__j2t_field_vm, _pcsp__j2t_field_vm},
    {"_tb_write_default_or_empty", _entry__tb_write_default_or_empty, _size__tb_write_default_or_empty, _stack__tb_write_default_or_empty, _pcsp__tb_write_default_or_empty},
    {"_j2t_write_unset_fields", _entry__j2t_write_unset_fields, _size__j2t_write_unset_fields, _stack__j2t_write_unset_fields, _pcsp__j2t_write_unset_fields},
    {"_j2t_find_field_key", _entry__j2t_find_field_key, _size__j2t_find_field_key, _stack__j2t_find_field_key, _pcsp__j2t_find_field_key},
    {"_lspace", _entry__lspace, _size__lspace, _stack__lspace, _pcsp__lspace},
    {"_quote", _entry__quote, _size__quote, _stack__quote, _pcsp__quote},
    {"_skip_array", _entry__skip_array, _size__skip_array, _stack__skip_array, _pcsp__skip_array},
    {"_skip_object", _entry__skip_object, _size__skip_object, _stack__skip_object, _pcsp__skip_object},
    {"_skip_one", _entry__skip_one, _size__skip_one, _stack__skip_one, _pcsp__skip_one},
    {"_tb_skip", _entry__tb_skip, _size__tb_skip, _stack__tb_skip, _pcsp__tb_skip},
    {"_tb_write_i64", _entry__tb_write_i64, _size__tb_write_i64, _stack__tb_write_i64, _pcsp__tb_write_i64},
    {"_trie_get", _entry__trie_get, _size__trie_get, _stack__trie_get, _pcsp__trie_get},
    {"_validate_one", _entry__validate_one, _size__validate_one, _stack__validate_one, _pcsp__validate_one},
    {"_value", _entry__value, _size__value, _stack__value, _pcsp__value},
    {"_vsigned", _entry__vsigned, _size__vsigned, _stack__vsigned, _pcsp__vsigned},
    {"_vstring", _entry__vstring, _size__vstring, _stack__vstring, _pcsp__vstring},
    {"_vunsigned", _entry__vunsigned, _size__vunsigned, _stack__vunsigned, _pcsp__vunsigned},
}

var (
    _subr__f64toa       uintptr
    _subr__fsm_exec     uintptr
    _subr__hm_get       uintptr
    _subr__html_escape  uintptr
    _subr__i64toa       uintptr
    _subr__j2t_fsm_exec uintptr
    _subr__lspace       uintptr
    _subr__quote        uintptr
    _subr__skip_array   uintptr
    _subr__skip_object  uintptr
    _subr__skip_one     uintptr
    _subr__tb_skip      uintptr
    _subr__tb_write_i64 uintptr
    _subr__trie_get     uintptr
    _subr__u64toa       uintptr
    _subr__unquote      uintptr
    _subr__validate_one uintptr
    _subr__value        uintptr
    _subr__vnumber      uintptr
    _subr__vsigned      uintptr
    _subr__vstring      uintptr
    _subr__vunsigned    uintptr
)

const (
    _stack__f64toa = 80
    _stack__format_significand = 24
    _stack__format_integer = 16
    _stack__fsm_exec = 208
    _stack__advance_ns = 32
    _stack__validate_string = 120
    _stack__utf8_validate = 32
    _stack__advance_string = 48
    _stack__skip_number = 48
    _stack__hm_get = 16
    _stack__html_escape = 72
    _stack__i64toa = 16
    _stack__u64toa = 8
    _stack__j2t_fsm_exec = 592
    _stack__j2t_number = 296
    _stack__vnumber = 240
    _stack__atof_eisel_lemire64 = 32
    _stack__atof_native = 136
    _stack__decimal_to_f64 = 80
    _stack__right_shift = 8
    _stack__left_shift = 24
    _stack__j2t_string = 144
    _stack__unquote = 72
    _stack__b64decode = 152
    _stack__j2t_field_vm = 312
    _stack__tb_write_default_or_empty = 56
    _stack__j2t_write_unset_fields = 176
    _stack__j2t_find_field_key = 32
    _stack__lspace = 8
    _stack__quote = 56
    _stack__skip_array = 216
    _stack__skip_object = 216
    _stack__skip_one = 216
    _stack__tb_skip = 48
    _stack__tb_write_i64 = 8
    _stack__trie_get = 32
    _stack__validate_one = 216
    _stack__value = 312
    _stack__vsigned = 16
    _stack__vstring = 104
    _stack__vunsigned = 8
)

const (
    _entry__f64toa = 800
    _entry__format_significand = 54464
    _entry__format_integer = 3648
    _entry__fsm_exec = 23456
    _entry__advance_ns = 16064
    _entry__validate_string = 25696
    _entry__utf8_validate = 26848
    _entry__advance_string = 18464
    _entry__skip_number = 21808
    _entry__hm_get = 33280
    _entry__html_escape = 11040
    _entry__i64toa = 4080
    _entry__u64toa = 4192
    _entry__j2t_fsm_exec = 45088
    _entry__j2t_number = 39888
    _entry__vnumber = 19280
    _entry__atof_eisel_lemire64 = 13104
    _entry__atof_native = 15248
    _entry__decimal_to_f64 = 13536
    _entry__right_shift = 55424
    _entry__left_shift = 54928
    _entry__j2t_string = 40544
    _entry__unquote = 8368
    _entry__b64decode = 29104
    _entry__j2t_field_vm = 43344
    _entry__tb_write_default_or_empty = 36704
    _entry__j2t_write_unset_fields = 38992
    _entry__j2t_find_field_key = 42288
    _entry__lspace = 224
    _entry__quote = 5584
    _entry__skip_array = 25408
    _entry__skip_object = 25472
    _entry__skip_one = 23408
    _entry__tb_skip = 53424
    _entry__tb_write_i64 = 34832
    _entry__trie_get = 34096
    _entry__validate_one = 27552
    _entry__value = 17040
    _entry__vsigned = 20816
    _entry__vstring = 18256
    _entry__vunsigned = 21136
)

const (
    _size__f64toa = 2848
    _size__format_significand = 464
    _size__format_integer = 432
    _size__fsm_exec = 1416
    _size__advance_ns = 832
    _size__validate_string = 1152
    _size__utf8_validate = 492
    _size__advance_string = 768
    _size__skip_number = 1360
    _size__hm_get = 464
    _size__html_escape = 2064
    _size__i64toa = 48
    _size__u64toa = 1248
    _size__j2t_fsm_exec = 7816
    _size__j2t_number = 616
    _size__vnumber = 1536
    _size__atof_eisel_lemire64 = 368
    _size__atof_native = 624
    _size__decimal_to_f64 = 1712
    _size__right_shift = 400
    _size__left_shift = 496
    _size__j2t_string = 848
    _size__unquote = 2480
    _size__b64decode = 3440
    _size__j2t_field_vm = 1684
    _size__tb_write_default_or_empty = 1092
    _size__j2t_write_unset_fields = 864
    _size__j2t_find_field_key = 752
    _size__lspace = 544
    _size__quote = 2736
    _size__skip_array = 48
    _size__skip_object = 48
    _size__skip_one = 48
    _size__tb_skip = 968
    _size__tb_write_i64 = 64
    _size__trie_get = 304
    _size__validate_one = 48
    _size__value = 712
    _size__vsigned = 320
    _size__vstring = 144
    _size__vunsigned = 336
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
        {2788, 56},
        {2792, 48},
        {2793, 40},
        {2795, 32},
        {2797, 24},
        {2799, 16},
        {2801, 8},
        {2805, 0},
        {2843, 56},
    }
    _pcsp__format_significand = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {452, 24},
        {453, 16},
        {455, 8},
        {457, 0},
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
    _pcsp__fsm_exec = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {12, 40},
        {13, 48},
        {1190, 88},
        {1194, 48},
        {1195, 40},
        {1197, 32},
        {1199, 24},
        {1201, 16},
        {1203, 8},
        {1204, 0},
        {1416, 88},
    }
    _pcsp__advance_ns = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {793, 32},
        {794, 24},
        {796, 16},
        {798, 8},
        {799, 0},
        {831, 32},
    }
    _pcsp__validate_string = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {12, 40},
        {13, 48},
        {521, 88},
        {525, 48},
        {526, 40},
        {528, 32},
        {530, 24},
        {532, 16},
        {534, 8},
        {538, 0},
        {1139, 88},
    }
    _pcsp__utf8_validate = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {481, 32},
        {482, 24},
        {484, 16},
        {486, 8},
        {492, 0},
    }
    _pcsp__advance_string = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {12, 40},
        {332, 48},
        {333, 40},
        {335, 32},
        {337, 24},
        {339, 16},
        {341, 8},
        {345, 0},
        {757, 48},
    }
    _pcsp__skip_number = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {12, 40},
        {1274, 48},
        {1275, 40},
        {1277, 32},
        {1279, 24},
        {1281, 16},
        {1283, 8},
        {1287, 0},
        {1360, 48},
    }
    _pcsp__hm_get = [][2]uint32{
        {1, 0},
        {4, 8},
        {387, 16},
        {388, 8},
        {390, 0},
        {450, 16},
        {451, 8},
        {452, 0},
        {458, 16},
        {459, 8},
        {461, 0},
    }
    _pcsp__html_escape = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {12, 40},
        {13, 48},
        {2045, 72},
        {2049, 48},
        {2050, 40},
        {2052, 32},
        {2054, 24},
        {2056, 16},
        {2058, 8},
        {2063, 0},
    }
    _pcsp__i64toa = [][2]uint32{
        {14, 0},
        {34, 8},
        {36, 0},
    }
    _pcsp__u64toa = [][2]uint32{
        {1, 0},
        {161, 8},
        {162, 0},
        {457, 8},
        {458, 0},
        {758, 8},
        {759, 0},
        {1225, 8},
        {1227, 0},
    }
    _pcsp__j2t_fsm_exec = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {12, 40},
        {13, 48},
        {6456, 280},
        {6463, 48},
        {6464, 40},
        {6466, 32},
        {6468, 24},
        {6470, 16},
        {6472, 8},
        {6476, 0},
        {7816, 280},
    }
    _pcsp__j2t_number = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {12, 40},
        {13, 48},
        {598, 56},
        {602, 48},
        {603, 40},
        {605, 32},
        {607, 24},
        {609, 16},
        {611, 8},
        {616, 0},
    }
    _pcsp__vnumber = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {12, 40},
        {13, 48},
        {787, 104},
        {791, 48},
        {792, 40},
        {794, 32},
        {796, 24},
        {798, 16},
        {800, 8},
        {801, 0},
        {1531, 104},
    }
    _pcsp__atof_eisel_lemire64 = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {292, 32},
        {293, 24},
        {295, 16},
        {297, 8},
        {298, 0},
        {362, 32},
    }
    _pcsp__atof_native = [][2]uint32{
        {1, 0},
        {4, 8},
        {587, 56},
        {591, 8},
        {593, 0},
    }
    _pcsp__decimal_to_f64 = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {12, 40},
        {13, 48},
        {1673, 56},
        {1677, 48},
        {1678, 40},
        {1680, 32},
        {1682, 24},
        {1684, 16},
        {1686, 8},
        {1690, 0},
        {1702, 56},
    }
    _pcsp__right_shift = [][2]uint32{
        {1, 0},
        {318, 8},
        {319, 0},
        {387, 8},
        {388, 0},
        {396, 8},
        {398, 0},
    }
    _pcsp__left_shift = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {363, 24},
        {364, 16},
        {366, 8},
        {367, 0},
        {470, 24},
        {471, 16},
        {473, 8},
        {474, 0},
        {486, 24},
    }
    _pcsp__j2t_string = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {12, 40},
        {13, 48},
        {638, 72},
        {642, 48},
        {643, 40},
        {645, 32},
        {647, 24},
        {649, 16},
        {651, 8},
        {655, 0},
        {834, 72},
    }
    _pcsp__unquote = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {12, 40},
        {13, 48},
        {79, 72},
        {83, 48},
        {84, 40},
        {86, 32},
        {88, 24},
        {90, 16},
        {92, 8},
        {96, 0},
        {2464, 72},
    }
    _pcsp__b64decode = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {12, 40},
        {13, 48},
        {3408, 152},
        {3412, 48},
        {3413, 40},
        {3415, 32},
        {3417, 24},
        {3419, 16},
        {3421, 8},
        {3426, 0},
    }
    _pcsp__j2t_field_vm = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {12, 40},
        {13, 48},
        {343, 72},
        {347, 48},
        {348, 40},
        {350, 32},
        {352, 24},
        {354, 16},
        {356, 8},
        {357, 0},
        {654, 72},
        {658, 48},
        {659, 40},
        {661, 32},
        {663, 24},
        {665, 16},
        {667, 8},
        {671, 0},
        {1684, 72},
    }
    _pcsp__tb_write_default_or_empty = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {12, 40},
        {13, 48},
        {1011, 56},
        {1015, 48},
        {1016, 40},
        {1018, 32},
        {1020, 24},
        {1022, 16},
        {1024, 8},
        {1028, 0},
        {1092, 56},
    }
    _pcsp__j2t_write_unset_fields = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {12, 40},
        {13, 48},
        {835, 120},
        {839, 48},
        {840, 40},
        {842, 32},
        {844, 24},
        {846, 16},
        {848, 8},
        {850, 0},
    }
    _pcsp__j2t_find_field_key = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {721, 32},
        {722, 24},
        {724, 16},
        {726, 8},
        {727, 0},
        {738, 32},
    }
    _pcsp__lspace = [][2]uint32{
        {1, 0},
        {480, 8},
        {481, 0},
        {488, 8},
        {489, 0},
        {504, 8},
        {505, 0},
        {512, 8},
        {514, 0},
    }
    _pcsp__quote = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {12, 40},
        {13, 48},
        {2687, 56},
        {2691, 48},
        {2692, 40},
        {2694, 32},
        {2696, 24},
        {2698, 16},
        {2700, 8},
        {2704, 0},
        {2731, 56},
    }
    _pcsp__skip_array = [][2]uint32{
        {1, 0},
        {30, 8},
        {36, 0},
    }
    _pcsp__skip_object = [][2]uint32{
        {1, 0},
        {30, 8},
        {36, 0},
    }
    _pcsp__skip_one = [][2]uint32{
        {1, 0},
        {32, 8},
        {38, 0},
    }
    _pcsp__tb_skip = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {12, 40},
        {944, 48},
        {945, 40},
        {947, 32},
        {949, 24},
        {951, 16},
        {953, 8},
        {954, 0},
        {968, 48},
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
        {286, 32},
        {287, 24},
        {289, 16},
        {291, 8},
        {293, 0},
    }
    _pcsp__validate_one = [][2]uint32{
        {1, 0},
        {35, 8},
        {41, 0},
    }
    _pcsp__value = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {11, 40},
        {698, 72},
        {702, 40},
        {703, 32},
        {705, 24},
        {707, 16},
        {709, 8},
        {712, 0},
    }
    _pcsp__vsigned = [][2]uint32{
        {1, 0},
        {4, 8},
        {109, 16},
        {110, 8},
        {111, 0},
        {122, 16},
        {123, 8},
        {124, 0},
        {260, 16},
        {261, 8},
        {262, 0},
        {266, 16},
        {267, 8},
        {268, 0},
        {306, 16},
        {307, 8},
        {308, 0},
        {316, 16},
        {317, 8},
        {319, 0},
    }
    _pcsp__vstring = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {11, 40},
        {105, 56},
        {109, 40},
        {110, 32},
        {112, 24},
        {114, 16},
        {116, 8},
        {118, 0},
    }
    _pcsp__vunsigned = [][2]uint32{
        {1, 0},
        {65, 8},
        {66, 0},
        {77, 8},
        {78, 0},
        {98, 8},
        {99, 0},
        {257, 8},
        {258, 0},
        {299, 8},
        {300, 0},
        {308, 8},
        {309, 0},
        {316, 8},
        {318, 0},
    }
)
