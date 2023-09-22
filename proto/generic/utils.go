package generic

import (
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/cloudwego/dynamicgo/proto"
	"github.com/stretchr/testify/require"
)

var bytesType = reflect.TypeOf([]byte{})

const (
	none int = 0
	b2s  int = 1
	s2b  int = 2
)

func toInterface(v interface{}) interface{} {
	return toInterface2(v, false, none)
}

func toInterface2(v interface{}, fieldId bool, byte2string int) interface{} {
	vt := reflect.ValueOf(v)
	if vt.Kind() == reflect.Ptr {
		if vt.IsNil() {
			return nil
		}
		vt = vt.Elem()
	}
	if k := vt.Kind(); k == reflect.Slice || k == reflect.Array {
		if vt.Type() == bytesType {
			if byte2string == b2s {
				return string(vt.Bytes())
			} else {
				return vt.Bytes()
			}
		}
		var r = make([]interface{}, 0, vt.Len())
		for i := 0; i < vt.Len(); i++ {
			vv := toInterface2(vt.Index(i).Interface(), fieldId, byte2string)
			if vv != nil {
				r = append(r, vv)
			}
		}
		return r
	} else if k == reflect.Map {
		if kt := vt.Type().Key().Kind(); kt == reflect.String {
			var r = make(map[string]interface{}, vt.Len())
			for _, k := range vt.MapKeys() {
				vv := toInterface2(vt.MapIndex(k).Interface(), fieldId, byte2string)
				if vv != nil {
					r[k.String()] = vv
				}
			}
			return r
		} else if kt == reflect.Int || kt == reflect.Int8 || kt == reflect.Int16 || kt == reflect.Int32 || kt == reflect.Int64 {
			var r = make(map[int]interface{}, vt.Len())
			for _, k := range vt.MapKeys() {
				vv := toInterface2(vt.MapIndex(k).Interface(), fieldId, byte2string)
				if vv != nil {
					r[int(k.Int())] = vv
				}
			}
			return r
		} else {
			var r = make(map[interface{}]interface{}, vt.Len())
			for _, k := range vt.MapKeys() {
				kv := toInterface2(k.Interface(), fieldId, byte2string)
				vv := toInterface2(vt.MapIndex(k).Interface(), fieldId, byte2string)
				if kv != nil && vv != nil {
					switch t := kv.(type) {
					case map[string]interface{}:
						r[&t] = vv
					case map[int]interface{}:
						r[&t] = vv
					case map[interface{}]interface{}:
						r[&t] = vv
					case []interface{}:
						r[&t] = vv
					default:
						r[kv] = vv
					}
				}
			}
			return r
		}
	} else if k == reflect.Struct {
		var r interface{}
		if fieldId {
			r = map[proto.FieldNumber]interface{}{}
		} else {
			r = map[int]interface{}{}
		}
		for i := 0; i < vt.NumField(); i++ {
			field := vt.Type().Field(i)
			if field.Name == "state" || field.Name == "unknownFields" || field.Name == "sizeCache" {
				continue
			}
			tag := field.Tag.Get("protobuf")
			ts := strings.Split(tag, ",")
			id := i
			if len(ts) > 1 {
				id, _ = strconv.Atoi(ts[1])
			}
			vv := toInterface2(vt.Field(i).Interface(), fieldId, byte2string)
			if vv != nil {
				if fieldId {
					r.(map[proto.FieldNumber]interface{})[proto.FieldNumber(id)] = vv
				} else {
					r.(map[int]interface{})[int(id)] = vv
				}
			}
		}
		return r
	} else if k == reflect.Int || k == reflect.Int8 || k == reflect.Int16 || k == reflect.Int32 || k == reflect.Int64 {
		return int(vt.Int())
	} else if k == reflect.String {
		if byte2string == s2b {
			return []byte(vt.String())
		} else {
			return vt.String()
		}
	}
	return vt.Interface()
}

func DeepEqual(exp interface{}, act interface{}) bool {
	switch ev := exp.(type) {
	case map[int]interface{}:
		av, ok := act.(map[int]interface{})
		if !ok {
			return false
		}
		for k, v := range ev {
			vv, ok := av[k]
			if !ok {
				return false
			}
			if !DeepEqual(v, vv) {
				return false
			}
		}
		return true
	case map[string]interface{}:
		av, ok := act.(map[string]interface{})
		if !ok {
			return false
		}
		for k, v := range ev {
			vv, ok := av[k]
			if !ok {
				return false
			}
			if !DeepEqual(v, vv) {
				return false
			}
		}
		return true
	case map[interface{}]interface{}:
		av, ok := act.(map[interface{}]interface{})
		if !ok {
			return false
		}
		if len(ev) == 0 {
			return true
		}
		erv := reflect.ValueOf(ev)
		arv := reflect.ValueOf(av)
		eks := erv.MapKeys()
		aks := arv.MapKeys()
		isPointer := eks[0].Elem().Kind() == reflect.Ptr
		if !isPointer {
			for k, v := range ev {
				vv, ok := av[k]
				if !ok {
					return false
				}
				if !DeepEqual(v, vv) {
					return false
				}
			}
		} else {
			for _, ek := range eks {
				found := false
				for _, ak := range aks {
					if DeepEqual(ek.Elem().Elem().Interface(), ak.Elem().Elem().Interface()) {
						found = true
						evv := erv.MapIndex(ek)
						avv := arv.MapIndex(ak)
						if !DeepEqual(evv.Interface(), avv.Interface()) {
							return false
						}
					}
					if !found {
						return false
					}
				}
			}
		}
		return true
	case []interface{}:
		av, ok := act.([]interface{})
		if !ok {
			return false
		}
		for i, v := range ev {
			vv := av[i]
			if !DeepEqual(v, vv) {
				return false
			}
		}
		return true
	default:
		return reflect.DeepEqual(exp, act)
	}
}

func TestDeepEqual(t *testing.T) {
	a := map[interface{}]interface{}{
		float64(0.1): "A",
		float64(0.2): "B",
		float64(0.3): "C",
		float64(0.4): "D",
		float64(0.5): "E",
		float64(0.6): "F",
		float64(0.7): "G",
		float64(0.8): "H",
		float64(0.9): "I",
	}
	b := map[interface{}]interface{}{
		float64(0.4): "D",
		float64(0.8): "H",
		float64(0.7): "G",
		float64(0.5): "E",
		float64(0.6): "F",
		float64(0.9): "I",
		float64(0.2): "B",
		float64(0.1): "A",
		float64(0.3): "C",
	}
	for i := 0; i < 10; i++ {
		require.Equal(t, a, b)
	}
	require.True(t, DeepEqual(a, b))
}