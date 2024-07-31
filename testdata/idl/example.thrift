include "base.thrift"
include "deep/deep.ref.thrift"
include "ref.thrift"
namespace go example

struct ExampleReq {
    1: optional string Msg,
    3: i32 InnerBase,
    255: required base.Base Base,
    32767: double Subfix,
}

struct ExampleToSnakeCase {
    1: optional string Msg,
    2: list<string> req_list,
    3: i32 InnerBase,
    255: required base.Base Base,
}

struct ExampleResp {
    1: optional string Msg,
    2: required string required_field
    32767: base.BaseResp BaseResp,
}

exception Exception {
    1: i32 code
    255: string msg
}

struct A {
    // 1: A self
    // 1024: self_ref.A a
    1: string self,
    2:FOO foo
}

enum FOO {
    B,
    A,
}

struct ExampleDefaultValue {
    1: string String = "default"
    2: i32 Int = 1
    3: double Double = 1.1
    4: bool Bool = true
    5: list<string> List = ["a", "b"]
    6: map<string, string> Map = {"a": "b"}
    7: set<string> Set = ["a", "b"]
    8: string ConstString = deep.ref.ConstString
    9: FOO Enum = FOO.A
}

struct DeepRef {
    1: deep.ref.TestStruct DeepRef
}

service ExampleService {
    DeepRef ExampleDeepRef(1: DeepRef req)
    ExampleResp ExampleMethod(1: ExampleReq req)throws(1: Exception err),
    A Foo(1: A req)
    string Ping(1: string msg)
    oneway void Oneway(1: string msg)
    void Void(1: string msg)
    void ExampleToSnakeCase(1: ExampleToSnakeCase req) (agw.to_snake = "")
    ExampleDefaultValue ExampleDefaultValue(1: ExampleDefaultValue req)
}
