# Trim Package

`trim` 是 `dynamicgo` 中的一个核心的数据处理包，主要提供对象级别的数据裁剪（Pruning / Fetch）与按需赋值（Assign）的能力。其设计初衷是能以一套统一的 DSL (描述符) 高效处理嵌套层次较深的复杂 Go 结构体（特别是通过 Protobuf 和 Thrift 自动生成的结构化数据），实现按需保留所需字段、并将其逆向赋值组回的功能。

## 1. 核心功能及概念

`trim` 包围绕三个核心抽象展开：

1.  **Descriptor（描述符）**：通过树状结构描述所需保留或访问的特定路径与字段（类似于 GraphQL 的选择树）。
2.  **Fetcher（提取器）**：顺着 Descriptor 描述的路径，以反射的方式将源（结构体 / Map / List 等）中特定数据抽取出来。
3.  **Assigner（赋值器）**：提取过程的逆向应用——根据 Descriptor 的树状结构，将零散结构化数据（如通过 Fetch 得到的嵌套 Map/Slice）自动精准映射赋值到目标强类型 Go 对象中，并专门处理了 Protobuf 的未知字段。

---

## 2. 核心原理与实现细节

### 2.1 Descriptor (模型映射描述)
位于 `desc.go`。
- **设计**：由特定的类型种类 `Kind` (支持 `Leaf` / `Struct` / `StrMap` / `List`) 和一组 `Children` 字段组成。例如描述提取字典的键或对应的 ID 字段。
- **循环引用支持**：提供了 `Normalize()` 方法（此方法被调用后 Descriptor 才会变为可用状态）。在其内部会对具有指向自身（或者内部构成环）的图状描述符网络进行缓存和解析，防脱轨或无尽循环。
- **序列化监控**：内置 `String()` 和 `MarshalJSON()` 支持以文本格式探查复杂 Descriptor 的全貌。

### 2.2 Fetcher (数据裁剪/提取)
位于 `fetch.go`。
- **底层机制**：基于 `reflect` 实施深层对象的遍历。
- **高速缓存**：对于 Thrift 或 Protobuf 结构体，`Fetcher` 不会每次都动态解析 struct tag。它全局维护了一个 `fieldCache` 字典 (`sync.Map`)，一次解析字段与 Tag 之后（包括解析 Protobuf Name、Thrift Field ID 等），就将结构体的索引记录到缓存，之后以 O(1) 的开销命中字段的 `reflect.Value`。
- **未知字段提取**：内置对 thrift `_unknownFields` 等未知域字段的读取兼容，以防止在降级/动态处理时漏掉非强类型的字段。可以使用 `SetThriftUnknownFieldName()` 灵活重定义配置。

### 2.3 Assigner (按需赋值映射)
位于 `assign.go`。
- **尽力而为（try-best mode）策略**：赋值器遇到单条字段的类型不匹配或缺失时，不仅不会立刻中断进程，还会将其记录进一个 `errorCollector` 中并继续执行，最后通过 `MultiErrors` 打包返回所有的错误（或者可通过 `AssignOptions.DisallowNotDefined` 提供更严苛的拦截策略）。
- **Protobuf Unknown Fields 兜底**：如果在源数据中提取到了不在实际 Go `struct` 定义内的字段（例如高版本的 protobuf 数据赋值降级到低版本的 Go 结构体）：
  - 若相应的未知字段名存在于 Descriptor 的元数据内，Assigner 会将数据序列化为 protobuf Binary wire 结构并放至目标结构体的 `XXX_unrecognized` 字段。
  - 对于未分类（Unkeyed）字面量，支持注入至 `XXX_NoUnkeyedLiteral`（也可以由 `SetPB***FieldName` 函数重新指定）。
- **零分配和对象池化（对象复用）**：为了在递归过程中降低性能损耗，在栈内跟踪 (Path stack / frames) 广泛使用了资源池(`sync.Pool`)复用，减少了不必要的 GC 开销。

---

## 3. 使用示例

```go
package main

import (
	"fmt"
	"github.com/cloudwego/dynamicgo/trim"
)

// 以一个生成的 Protobuf 消息结构体为例
type User struct {
	ID   int    `protobuf:"varint,1,req,name=id"`
	Name string `protobuf:"bytes,2,opt,name=name"`
}

func main() {
	// 1. 构建数据抓取描述符 Descriptor: 这里意图仅保留 ID
	desc := &trim.Descriptor{
		Kind: trim.TypeKind_Struct,
		Type: "User",
		Children: []trim.Field{
			{Name: "id", ID: 1, Desc: &trim.Descriptor{Kind: trim.TypeKind_Leaf}},
		},
	}
	// WARNING: 构建完毕必须调用 Normalize() 处理树的搜索索引
	desc.Normalize()

	// 原始需要被裁剪的数据
	user := &User{ID: 10086, Name: "Alice"}

	// 2. 使用 FetchAny (裁剪)
	fetcher := trim.Fetcher{}
	fetchedResult, err := fetcher.FetchAny(desc, user)
	if err != nil {
		panic(err)
	}
	
	// fetchedResult 中目前仅包含要求的 ID 字段了
	// 输出: Fetch 结果: map[id:10086]
	fmt.Printf("Fetch 结果: %+v\n", fetchedResult)

	// 3. 使用 AssignAny (重新映射到空数据体)
	newUser := &User{}
	assigner := trim.Assigner{}
	
	// 从前一级裁切获得的 object (如 map) 填入到 newUser 结构中
	err = assigner.AssignAny(desc, fetchedResult, newUser)
	if err != nil {
		panic(err)
	}
	
	// newUser 被赋予了部分值
	// 输出: Assign 结果: &{ID:10086 Name:}
	fmt.Printf("Assign 结果: %+v\n", newUser)
}
```