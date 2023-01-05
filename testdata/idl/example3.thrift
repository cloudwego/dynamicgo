include "base.thrift"
include "ref.thrift"

namespace go example3

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
    7: string String (api.cookie="inner_string", api.path="inner_string", api.query="inner_string", api.header="inner_string"),
    8: list<i32> ListInt32,
    9: map<string, string> MapStringString,
    10: set<i32> SetInt32,
    11: FOO Foo,
    12: map<i32, string> MapInt32String,
    13: binary Binary,
    14: map<byte, string> MapInt8String,
    15: map<i16, string> MapInt16String,
    16: map<i64, string> MapInt64String,

    18: list<InnerBase> ListInnerBase,
    19: map<string, InnerBase> MapStringInnerBase,

    20: string InnerQuery (api.query = "inner_query"),

    255: base.Base Base,
}

struct ExampleReq {
    1: optional string Msg (go.tag="json:\"msg\""),
    2: optional double Cookie (api.cookie = "cookie"),
    3: required string Path (api.path = "path"),
    4: optional list<string> Query (api.query = "query"),
    5: optional bool Header (api.header = "heeader"),
    6: i64 Code (api.js_conv = "", go.tag="json:\"code_code\""),
    7: InnerBase InnerBase,
    8: string RawUri (api.raw_uri = ""),

    32767: double Subfix,

    255: base.Base Base
}

struct ExampleResp {
    1: string Msg (api.body = "msg"),
    2: optional double Cookie (api.cookie = "cookie"),
    3: required i32 Status (api.http_code = "status"),
    4: optional bool Header (api.header = "heeader"),
    6: i64 Code (api.js_conv = "", go.tag="json:\"code_code\""),
    32767: double Subfix,
    7: InnerBase InnerBase,

    255: base.BaseResp BaseResp
}

exception Exception {
    1: i32 code
    255: string msg
}

struct ExampleError {
    1: map<InnerBase, InnerBase> MapInnerBaseInnerBase
    2: binary Base64
    3: required string Query (api.query = "query"),
    4: string Header (api.header = "header"),
    5: i32 Q2 (api.query = "q2"),
}

struct ExampleErrorResp {
    2: i64 Int64
    4: string Xjson (agw.body_dynamic = ""),
}

struct ExampleInt2Float {
    1: i32 Int32 (agw.js_conv = ""),
    2: double Float64 (api.js_conv = ""),
    3: string String (go.tag="json:\"中文\"", api.js_conv = "")
    4: i64 Int64
    32767: double Subfix
}

struct JSONObject {
    1: string A (go.tag="json:\"a\""),
    2: i64 B (go.tag="json:\"b\""),
}

struct ExampleJSONString {
    1: JSONObject Query (api.query = "query"),
    2: required list<string> Query2 (api.query = "query2"),
    3: required JSONObject Header (api.header = "header"),
    4: required map<i32,string> Header2 (api.header = "header2"),
    5: JSONObject Cookie (api.cookie = "cookie"),
    6: required set<i32> Cookie2 (api.cookie = "cookie2"),
}

struct ExamplePartial {
    1: string Msg (api.body = "msxxg", go.tag="json:\"msg\""),
}

struct ExamplePartial2 {
    1: string Msg (api.body = "msxxg", go.tag="json:\"msg\""),
    2: optional double Cookie (api.cookie = "cookie"),
    3: required i32 Status (api.http_code = "status"),
    4: optional bool Header (api.header = "heeader"),
    6: i64 Code (api.js_conv = "", go.tag="json:\"code_code\""),
    32767: double Subfix,
    7: InnerBasePartial InnerBase,

    255: base.BaseResp BaseResp
}

struct InnerBasePartial {
    1: bool Bool,
    255: base.Base Base,
}

struct ExampleFallback {
    2: string Msg (api.query = "A", api.path = "B"),
    3: string Heeader (api.header = "heeader"),
}

struct InnerCode {
    1: i64 C1 (api.body = "code"),
    2: i16 C2 (go.tag="json:\"code\""),
    3: list<InnerCode> C3,
}

struct ExampleApiBody {
    1: i64 Code (api.body = "code"),
    2: i16 Code2 (go.tag="json:\"code\""),
    3: InnerCode InnerCode,
}

struct InnerJSON {
    1: string A (go.tag="json:\"a\""),
    2: i64 B (go.tag="json:\"b\""),
    3: double inner_form 
}

struct ExamplePostForm {
    1: string Query (api.query = "query"),
    2: string Form (api.form = "form"),
    3: InnerJSON JSON,
}

struct InnerStruct {
    1: string InnerJSON (go.tag="json:\"inner_json\"", agw.body_dynamic = ""),
    2: required string Must,
}

struct ExampleDynamicStruct {
    1: required string Query (api.query = "query"),
    2: string JSON (go.tag="json:\"json\"", agw.body_dynamic = ""),
    3: InnerStruct InnerStruct (go.tag="json:\"inner_struct\""),
}

struct ExampleBase64Binary {
    1: binary Binary,
    2: binary Binary2 (api.header = "Binary2"),
}

struct ExampleDefaultValue {
    1: string A = "hello",
    2: i32 B = ref.FOO.A,
    3: double C = 1.2 (api.header = "c"),
    4: string D = ref.ConstString (api.cookie = "d"),
}

struct ExampleOptionalDefaultValue {
    1: optional string A = "hello",
    2: required i32 B = ref.FOO.A,
    3: optional double C = 1.2 (api.header = "c"),
    4: required string D = ref.ConstString (api.cookie = "d"),
    5: optional string E,
    6: optional string F (api.header = "f"),
}

struct ExampleNoBodyStruct {
    1: NoBodyStruct NoBodyStruct (agw.source = "not_body_struct"),
}

struct NoBodyStruct {
    1: optional i32 A,
    2: optional i32 B (api.query = "b"),
    3: optional i32 C = 1 (api.header = "c"),
}

service ExampleService {
    ExampleResp ExampleMethod(1: ExampleReq req)throws(1: Exception err) (api.post = "/example/set"),
    ExampleErrorResp ErrorMethod(1: ExampleError req) (api.get = "/example/get"),
    ExampleInt2Float Int2FloatMethod(1: ExampleInt2Float req),
    ExampleJSONString JSONStringMethod(1: ExampleJSONString req),
    ExamplePartial PartialMethod(1: ExamplePartial2 req),
    ExampleFallback FallbackMethod(1: ExampleFallback req),
    ExampleApiBody ApiBodyMethod(1: ExampleApiBody req),
    ExamplePostForm PostFormMethod(1: ExamplePostForm req),
    ExampleDynamicStruct DynamicStructMethod(1: ExampleDynamicStruct req),
    ExampleBase64Binary Base64BinaryMethod(1: ExampleBase64Binary req),
    ExampleDefaultValue DefaultValueMethod(1: ExampleDefaultValue req),
    ExampleOptionalDefaultValue OptionalDefaultValueMethod(1: ExampleOptionalDefaultValue req),
    ExampleNoBodyStruct NoBodyStructMethod(1: ExampleNoBodyStruct req),
}