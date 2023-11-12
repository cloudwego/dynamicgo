package generic

import (
	"fmt"
	"math"
	"testing"

	"github.com/cloudwego/dynamicgo/proto"
	"github.com/cloudwego/dynamicgo/proto/binary"
	"github.com/stretchr/testify/require"
)

var opts = &Options{
	UseNativeSkip:      false,
	MapStructById:      true,
	ClearDirtyValues:   true,
	CastStringAsBinary: false,
}

func TestExampleValue_GetByPath(t *testing.T) {
	desc := getExample2Desc()
	data := getExample2Data()
	v := NewRootValue(desc, data)

	ps := []Path{NewPathFieldName("InnerBase2"), NewPathFieldName("Base"), NewPathFieldName("Extra"), NewPathStrKey("1b")}
	s := v.GetByPath(ps...)
	if s.Error() != "" {
		panic(s.Error())
	}
	str, _ := s.String()
	fmt.Println(str) // "aaa"
}

func ExampleValue_GetMany() {
	desc := getExample2Desc()
	data := getExample2Data()
	v := NewRootValue(desc, data)

	ps := []PathNode{
		{Path: NewPathFieldId(1)},
		{Path: NewPathFieldId(3)},
		{Path: NewPathFieldId(32767)},
	}

	err := v.GetMany(ps, opts)
	if err != nil {
		panic(err)
	}
}

func ExampleValue_SetByPath() {
	desc := getExample2Desc()
	data := getExample2Data()
	v := NewRootValue(desc, data)

	// d := (*desc).Fields().ByName("InnerBase2").Message().Fields().ByName("Base").Message().Fields().ByName("Extra")
	p := binary.NewBinaryProtol([]byte{})
	exp := "中文"
	p.WriteString(exp)
	buf := p.RawBuf()
	// vv := NewValue(&d, buf)
	vv := NewNode(proto.STRING, buf)

	ps := []Path{NewPathFieldName("InnerBase2"), NewPathFieldName("Base"), NewPathFieldName("Extra"), NewPathStrKey("b")}
	exist, err2 := v.SetByPath(vv, ps...)
	if err2 != nil {
		panic(err2)
	}
	fmt.Println(exist) // false

	// check inserted value
	s2 := v.GetByPath(ps...)
	if s2.Error() != "" {
		panic(s2.Error())
	}
	f2, _ := s2.String()
	fmt.Println(f2) // 中文
}

func TestExampleValue_SetMany(t *testing.T) {
	desc := getExample2Desc()
	data := getExample2Data()
	root := NewRootValue(desc, data)


	exp1 := int32(-1024)
	n1 := NewNodeInt32(exp1)
	address := []int{}
	pathes := []Path{}
	err := root.SetMany([]PathNode{
		{
			Path: NewPathFieldId(2),
			Node: n1,
		},
	}, opts, &root, address, pathes...)

	require.Nil(t, err)
	v1 := root.GetByPath(NewPathFieldName("A"))
	act1, err := v1.Int()
	fmt.Println(act1) // -1024
	

	expmin := int32(math.MinInt32)
	expmax := int32(math.MaxInt32)
	nmin := NewNodeInt32(expmin)
	nmax := NewNodeInt32(expmax)
	PathExampleListInt32 = []Path{NewPathFieldName("InnerBase2"), NewPathFieldId(proto.FieldNumber(8))}
	
	vv, listInt2root := root.GetByPathWithAddress(PathExampleListInt32...)
	l2, err := vv.Len()
	fmt.Println(l2) // 6
	if err != nil {
		panic(err)
	}

	
	path2root := []Path{NewPathFieldName("InnerBase2"), NewPathFieldId(proto.FieldNumber(8)), NewPathIndex(1024)}
	address2root := append(listInt2root, 0)

	err = vv.SetMany([]PathNode{
		{
			Path: NewPathIndex(1024),
			Node: nmin,
		},
		{
			Path: NewPathIndex(1024),
			Node: nmax,
		},
	}, opts, &root, address2root, path2root...)
	
	if err != nil {
		panic(err)
	}

	v21 := vv.GetByPath(NewPathIndex(6))
	act21, err := v21.Int()
	if err != nil {
		panic(err)
	}

	fmt.Println(act21) // -2147483648

	v22 := vv.GetByPath(NewPathIndex(7))
	act22, err := v22.Int()
	if err != nil {
		panic(err)
	}

	fmt.Println(act22) // 2147483647

	ll2, err := vv.Len()
	if err != nil {
		panic(err)
	}
	
	fmt.Println(ll2) // 8
}

func TestExampleValue_MarshalTo(t *testing.T) {
	desc := getExample2Desc()
	data := getExample2Data()
	v := NewRootValue(desc, data)

	partial := getExamplePartialDesc()

	out, err := v.MarshalTo(partial, opts)
	if err != nil {
		panic(err)
	}

	partMsg, err := NewRootValue(partial, out).Interface(opts)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v", partMsg)
}


func TestExamplePathNode_Load(* testing.T) {
	desc := getExample2Desc()
	data := getExample2Data()

	root := PathNode{
		Node: NewNode(proto.MESSAGE, data),
	}
	
	// load first level children
	err := root.Load(false, opts, desc)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v", root)

	// load all level children
	err = root.Load(true, opts, desc)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v", root)


	// reuse PathNode memory
	reuse := pathNodePool.Get().(*PathNode)
	reuse.Node = NewNode(proto.MESSAGE, data)
	err = root.Load(true, opts, desc)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v", reuse)
	reuse.ResetValue()
	pathNodePool.Put(reuse)
}

func ExampleValue_MarshalAll() {
	desc := getExample2Desc()
	data := getExample2Data()
	v := NewRootValue(desc, data)
	p := PathNode{
		Node: v.Node,
	}

	if err := p.Load(true, opts, desc); err != nil {
		panic(err)
	}

	out, err := p.Marshal(opts)
	if err != nil {
		panic(err)
	}

	msg, err := NewRootValue(desc, out).Interface(opts)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v", msg)
}

func TestExamplePathNode_MarshalMany(t *testing.T) {
	desc := getExample2Desc()
	data := getExample2Data()

	v := NewRootValue(desc, data)

	ps := []PathNode{
		{Path: NewPathFieldId(1)},
		// {Path: NewPathFieldId(3)},
		{Path: NewPathFieldId(32767)},
	}

	err := v.GetMany(ps, opts)
	if err != nil {
		panic(err)
	}
	
	n := PathNode{
		Path: NewPathFieldId(1), // just used path type for message flag, id is not used
		Node: v.Node,
		Next: ps,
	}

	buf, err := n.Marshal(opts)

	msg, err := NewRootValue(desc, buf).Interface(opts)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v", msg)
		
}