include "base.thrift"
namespace go example2

enum FOO {
    A = 1;
}

struct InnerBase {
    1: bool Bool,
    2: byte Byte,
    3: i16 Int16,
    4: i32 Int32,
    5: i64 Int64,
    6: double Double,
    7: string String,
    8: list<i32> ListInt32,
    9: map<string, string> MapStringString,
    10: set<i32> SetInt32,
    11: FOO Foo,
    12: map<i32, string> MapInt32String,
    13: binary Binary,
    14: map<byte, string> MapInt8String,
    15: map<i16, string> MapInt16String,
    16: map<i64, string> MapInt64String,
    17: map<double, string> MapDoubleString,

    18: list<InnerBase> ListInnerBase,
    19: map<InnerBase, InnerBase> MapInnerBaseInnerBase,

    255: base.Base Base,
}

struct InnerBasePartial {
    1: bool Bool,
    8: list<i32> ListInt32,
    9: map<string, string> MapStringString,
    17: map<double, string> MapDoubleString,
    18: list<InnerBasePartial> ListInnerBase,
    19: map<InnerBasePartial, InnerBasePartial> MapInnerBaseInnerBase,
    
    127: map<string, string> MapStringString2
}

struct BasePartial {
    5: optional base.TrafficEnv TrafficEnv,
}

struct ExampleReq {
    1: optional string Msg,
    2: optional i32 A,
    3: InnerBase InnerBase (agw.to_snake="true"),
    255: required base.Base Base,
    32767: double Subfix,
}

struct ExampleSuper {
    1: optional string Msg,
    2: optional i32 A,
    3: InnerBase InnerBase,
    4: string Ex1,
    5: optional string Ex2,
    6: optional string Ex3,
    7: required string Ex4,
    // 8: map<BasePartial,BasePartial> MapStructStruct,
    9: SelfRef SelfRef,
    255: required base.Base Base,
    32767: double Subfix,
}

struct SelfRef {
    1: optional SelfRef self,
}

struct ExampleReqPartial {
    1: optional string Msg,
    3: InnerBasePartial InnerBase,
    255: BasePartial Base,
}

struct ExampleResp {
    1: optional string Msg,
    2: required string required_field
    255: base.BaseResp BaseResp,
}

exception Exception {
    1: i32 code
    255: string msg
}

struct A {
    1: A self
}

service ExampleService {
    ExampleResp ExampleMethod(1: ExampleReq req)throws(1: Exception err),
    A ExamplePartialMethod(1: ExampleReqPartial req),
    A ExampleSuperMethod(1: ExampleSuper req),
    A Foo(1: A req)
    string Ping(1: string msg)
    oneway void Oneway(1: string msg)
    void Void(1: string msg)
}