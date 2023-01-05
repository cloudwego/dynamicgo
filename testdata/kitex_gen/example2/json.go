package example2

import (
	"encoding/json"
	"strconv"
)

func (v *BasePartial) MarshalJSON() ([]byte, error) {
	if v == nil || v.TrafficEnv == nil {
		return []byte("null"), nil
	}
	b, e := json.Marshal(v.TrafficEnv)
	return []byte(strconv.Quote(string(b))), e
}
