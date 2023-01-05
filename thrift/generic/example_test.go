package generic

import (
	"fmt"

	"github.com/cloudwego/dynamicgo/thrift"
)

var opts = &Options{
	UseNativeSkip: true,
}

func ExampleValue_SetByPath() {
	// pack root value
	desc := getExampleDesc()
	data := getExampleData()
	v := NewValue(desc, data)

	// pack insert value
	d := desc.Struct().FieldByKey("Base").Type().Struct().FieldByKey("Extra").Type().Elem()
	p := thrift.NewBinaryProtocol([]byte{})
	exp := "中文"
	p.WriteString(exp)
	buf := p.RawBuf()
	vv := NewValue(d, buf)

	// insert value
	ps := []Path{NewPathFieldName("Base"), NewPathFieldName("Extra"), NewPathStrKey("b")}
	exist, err2 := v.SetByPath(vv, ps...)
	if err2 != nil {
		panic(err2)
	}
	println(exist) // false

	// check inserted value
	s2 := v.GetByPath(ps...)
	if s2.Error() != "" {
		panic(s2.Error())
	}
	f2, _ := s2.String()
	println(f2) // 中文
}

func ExampleValue_SetMany() {
	// make root value
	desc := getExampleDesc()
	data := getExampleData()
	d1 := desc.Struct().FieldByKey("Msg").Type()
	d2 := desc.Struct().FieldByKey("Subfix").Type()
	v := NewValue(desc, data)

	// make insert values
	p := thrift.NewBinaryProtocol([]byte{})
	e1 := "test1"
	p.WriteString(e1)
	v1 := NewValue(d1, p.RawBuf())
	p = thrift.NewBinaryProtocol([]byte{})
	e2 := float64(-255.0001)
	p.WriteDouble(e2)
	v2 := NewValue(d2, p.RawBuf())
	v3 := v.GetByPath(NewPathFieldName("Base"))

	// pack insert pathes and values
	ps := []PathNode{
		{
			Path: NewPathFieldId(1),
			Node: v1.Node,
		},
		{
			Path: NewPathFieldId(32767),
			Node: v2.Node,
		},
		{
			Path: NewPathFieldId(255),
			Node: v3.Node,
		},
	}

	// insert values
	err := v.SetMany(ps, opts)
	if err != nil {
		panic(err)
	}
	any, err := v.Interface(opts)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v", any)

	// check inserted values
	ps2 := []PathNode{
		{
			Path: NewPathFieldId(1),
		},
		{
			Path: NewPathFieldId(32767),
		},
		{
			Path: NewPathFieldId(255),
		},
	}
	if err := v.GetMany(ps2, opts); err != nil {
		panic(err)
	}
	any0, err := ps2[2].Node.Interface(opts)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v", any0)
}

func ExampleValue_MarshalTo() {
	// make full value
	desc := getExampleDesc()
	data := getExampleData()
	v := NewValue(desc, data)

	// print full value
	full, err := NewNode(thrift.STRUCT, data).Interface(opts)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v", full)

	// get partial descriptor
	pdesc := getExamplePartialDesc()
	// cut full value to partial value
	out, err := v.MarshalTo(pdesc, opts)
	if err != nil {
		panic(err)
	}

	// print partial value
	partial, err := NewNode(thrift.STRUCT, out).Interface(opts)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v", partial)
}

func ExamplePathNode_Load() {
	// make root PathNode
	data := getExampleData()
	root := PathNode{
		Node: NewNode(thrift.STRUCT, data),
	}

	// load first level children
	err := root.Load(false, opts)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v", root)

	// load all level children
	err = root.Load(true, opts)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v", root)

	// reuse PathNode memory
	reuse := pathNodePool.Get().(*PathNode)
	root.Node = NewNode(thrift.STRUCT, data)
	err = root.Load(true, opts)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%#v", root)
	reuse.ResetValue()
	pathNodePool.Put(reuse)

}
