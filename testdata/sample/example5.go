package sample

import (
	"strconv"

	"github.com/cloudwego/dynamicgo/testdata/kitex_gen/example5"
)

func GetThriftInnerBase5(width int, depth int) *example5.InnerBase {
	th := &example5.InnerBase{}
	th.Bool = true
	th.Int32 = 123
	th.Int64 = -456
	th.Double = 3.14159
	th.String_ = "hello"
	th.ListString = []string{"a", "b"}
	th.MapStringString = map[string]string{"k": "v"}
	th.SetInt32_ = []int32{1, 2, 3}
	th.MapInt32String = map[int32]string{1: "x"}
	th.Binary = []byte{0x00, 0xFF, 0x10}
	th.MapInt64String = map[int64]string{7: "vv"}

	// recurse set depth and width
	assignInnerBase(th, width, depth)
	return th
}

func assignInnerBase(inner *example5.InnerBase, width int, depth int) {
	if width > 0 && depth > 0 {
		for i := 0; i < width; i++ {
			inner2 := example5.NewInnerBase()
			inner2.Bool = true
			assignInnerBase(inner2, width, depth-1)
			inner.ListInnerBase = append(inner.ListInnerBase, inner2)
			if inner.MapStringInnerBase == nil {
				inner.MapStringInnerBase = map[string]*example5.InnerBase{}
			}
			inner.MapStringInnerBase["inner"+strconv.Itoa(i)] = inner2
		}
	}
}
