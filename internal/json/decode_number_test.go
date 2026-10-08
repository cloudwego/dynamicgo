package json

import (
	"math"
	"testing"

	"github.com/cloudwego/dynamicgo/internal/types"
)

// An integer literal that does not fit in int64 is a valid JSON number and must
// be classified as a floating-point value, matching how numbers containing
// '.'/'e'/'E' are handled. It must not be silently reported as integer 0.
func TestDecodeIntOverflowFallsBackToDouble(t *testing.T) {
	overflows := []struct {
		in   string
		mag  float64 // expected magnitude (approx)
		sign float64 // +1 or -1
	}{
		{"9223372036854775808", 9223372036854775808, +1}, // math.MaxInt64 + 1
		{"-9223372036854775809", 9223372036854775809, -1}, // math.MinInt64 - 1
		{"99999999999999999999999", 9.999999999999999e22, +1},
		{"-99999999999999999999999", 9.999999999999999e22, -1},
	}
	for _, c := range overflows {
		got, st := DecodeValue(c.in, 0)
		if got < 0 {
			t.Fatalf("DecodeValue(%q) = %d, want non-negative position", c.in, got)
		}
		if st.Vt != types.V_DOUBLE {
			t.Errorf("DecodeValue(%q).Vt = %d, want V_DOUBLE(%d)", c.in, st.Vt, types.V_DOUBLE)
		}
		if st.Dv == 0 {
			t.Errorf("DecodeValue(%q).Dv = 0, want the overflowed magnitude, not a silent zero", c.in)
		}
		if math.Copysign(1, st.Dv) != c.sign {
			t.Errorf("DecodeValue(%q).Dv = %v, wrong sign", c.in, st.Dv)
		}
		// relative magnitude should be within a few ULP of the literal
		if math.Abs(math.Abs(st.Dv)-c.mag)/c.mag > 1e-12 {
			t.Errorf("DecodeValue(%q).Dv = %v, magnitude ~%v expected", c.in, st.Dv, c.mag)
		}
		// the consumed length must agree with SkipValue (parser desync oracle)
		sr, _ := SkipValue(c.in, 0)
		if sr != got {
			t.Errorf("number desync for %q: decode=%d skip=%d", c.in, got, sr)
		}
	}
}

// In-range integers must still decode as integers (no behavior change).
func TestDecodeInRangeIntegerStaysInteger(t *testing.T) {
	in := "9007199254740993" // 2^53+1, exactly representable as int64
	got, st := DecodeValue(in, 0)
	if got < 0 || st.Vt != types.V_INTEGER || st.Iv != 9007199254740993 {
		t.Fatalf("DecodeValue(%q) = %d,%+v, want V_INTEGER=9007199254740993", in, got, st)
	}
}
