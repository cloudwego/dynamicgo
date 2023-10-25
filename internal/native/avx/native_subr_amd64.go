// +build !noasm !appengine
// Code generated by asm2asm, DO NOT EDIT.

package avx

import (
	`github.com/bytedance/sonic/loader`
)

const (
    _entry__hm_get = 27776
    _entry__j2t_fsm_exec = 39584
    _entry__advance_ns = 12672
    _entry__fsm_exec = 19376
    _entry__validate_string = 21568
    _entry__utf8_validate = 22912
    _entry__advance_string = 14880
    _entry__skip_number = 18160
    _entry__j2t_number = 34384
    _entry__vnumber = 15872
    _entry__atof_eisel_lemire64 = 10448
    _entry__atof_native = 12000
    _entry__decimal_to_f64 = 10816
    _entry__right_shift = 49920
    _entry__left_shift = 49424
    _entry__j2t_string = 35040
    _entry__unquote = 6864
    _entry__b64decode = 24208
    _entry__j2t_field_vm = 37840
    _entry__tb_write_default_or_empty = 31200
    _entry__j2t_write_unset_fields = 33488
    _entry__j2t_find_field_key = 36784
    _entry__tb_skip = 47920
    _entry__tb_write_i64 = 29328
    _entry__trie_get = 28592
)

const (
    _stack__hm_get = 16
    _stack__j2t_fsm_exec = 592
    _stack__advance_ns = 16
    _stack__fsm_exec = 208
    _stack__validate_string = 120
    _stack__utf8_validate = 32
    _stack__advance_string = 64
    _stack__skip_number = 48
    _stack__j2t_number = 296
    _stack__vnumber = 240
    _stack__atof_eisel_lemire64 = 32
    _stack__atof_native = 136
    _stack__decimal_to_f64 = 80
    _stack__right_shift = 8
    _stack__left_shift = 24
    _stack__j2t_string = 160
    _stack__unquote = 88
    _stack__b64decode = 160
    _stack__j2t_field_vm = 312
    _stack__tb_write_default_or_empty = 56
    _stack__j2t_write_unset_fields = 176
    _stack__j2t_find_field_key = 32
    _stack__tb_skip = 48
    _stack__tb_write_i64 = 8
    _stack__trie_get = 32
)

const (
    _size__hm_get = 464
    _size__j2t_fsm_exec = 7816
    _size__advance_ns = 688
    _size__fsm_exec = 1416
    _size__validate_string = 1344
    _size__utf8_validate = 416
    _size__advance_string = 944
    _size__skip_number = 924
    _size__j2t_number = 616
    _size__vnumber = 1536
    _size__atof_eisel_lemire64 = 368
    _size__atof_native = 608
    _size__decimal_to_f64 = 1184
    _size__right_shift = 400
    _size__left_shift = 496
    _size__j2t_string = 848
    _size__unquote = 2272
    _size__b64decode = 2832
    _size__j2t_field_vm = 1684
    _size__tb_write_default_or_empty = 1092
    _size__j2t_write_unset_fields = 864
    _size__j2t_find_field_key = 752
    _size__tb_skip = 968
    _size__tb_write_i64 = 64
    _size__trie_get = 304
)

var (
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
    _pcsp__advance_ns = [][2]uint32{
        {1, 0},
        {4, 8},
        {659, 16},
        {660, 8},
        {661, 0},
        {685, 16},
        {686, 8},
        {688, 0},
    }
    _pcsp__fsm_exec = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {12, 40},
        {13, 48},
        {1220, 88},
        {1224, 48},
        {1225, 40},
        {1227, 32},
        {1229, 24},
        {1231, 16},
        {1233, 8},
        {1234, 0},
        {1416, 88},
    }
    _pcsp__validate_string = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {12, 40},
        {13, 48},
        {643, 88},
        {647, 48},
        {648, 40},
        {650, 32},
        {652, 24},
        {654, 16},
        {656, 8},
        {657, 0},
        {1341, 88},
    }
    _pcsp__utf8_validate = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {409, 32},
        {410, 24},
        {412, 16},
        {414, 8},
        {416, 0},
    }
    _pcsp__advance_string = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {12, 40},
        {13, 48},
        {552, 64},
        {556, 48},
        {557, 40},
        {559, 32},
        {561, 24},
        {563, 16},
        {565, 8},
        {566, 0},
        {931, 64},
    }
    _pcsp__skip_number = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {12, 40},
        {849, 48},
        {850, 40},
        {852, 32},
        {854, 24},
        {856, 16},
        {858, 8},
        {859, 0},
        {924, 48},
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
        {1144, 56},
        {1148, 48},
        {1149, 40},
        {1151, 32},
        {1153, 24},
        {1155, 16},
        {1157, 8},
        {1158, 0},
        {1169, 56},
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
        {1684, 88},
        {1688, 48},
        {1689, 40},
        {1691, 32},
        {1693, 24},
        {1695, 16},
        {1697, 8},
        {1698, 0},
        {2270, 88},
    }
    _pcsp__b64decode = [][2]uint32{
        {1, 0},
        {4, 8},
        {6, 16},
        {8, 24},
        {10, 32},
        {12, 40},
        {13, 48},
        {2813, 160},
        {2817, 48},
        {2818, 40},
        {2820, 32},
        {2822, 24},
        {2824, 16},
        {2826, 8},
        {2828, 0},
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
)

var Funcs = []loader.CFunc{
    {"__native_entry__", 0, 67, 0, nil},
    {"_hm_get", _entry__hm_get, _size__hm_get, _stack__hm_get, _pcsp__hm_get},
    {"_j2t_fsm_exec", _entry__j2t_fsm_exec, _size__j2t_fsm_exec, _stack__j2t_fsm_exec, _pcsp__j2t_fsm_exec},
    {"_advance_ns", _entry__advance_ns, _size__advance_ns, _stack__advance_ns, _pcsp__advance_ns},
    {"_fsm_exec", _entry__fsm_exec, _size__fsm_exec, _stack__fsm_exec, _pcsp__fsm_exec},
    {"_validate_string", _entry__validate_string, _size__validate_string, _stack__validate_string, _pcsp__validate_string},
    {"_utf8_validate", _entry__utf8_validate, _size__utf8_validate, _stack__utf8_validate, _pcsp__utf8_validate},
    {"_advance_string", _entry__advance_string, _size__advance_string, _stack__advance_string, _pcsp__advance_string},
    {"_skip_number", _entry__skip_number, _size__skip_number, _stack__skip_number, _pcsp__skip_number},
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
    {"_tb_skip", _entry__tb_skip, _size__tb_skip, _stack__tb_skip, _pcsp__tb_skip},
    {"_tb_write_i64", _entry__tb_write_i64, _size__tb_write_i64, _stack__tb_write_i64, _pcsp__tb_write_i64},
    {"_trie_get", _entry__trie_get, _size__trie_get, _stack__trie_get, _pcsp__trie_get},
}
