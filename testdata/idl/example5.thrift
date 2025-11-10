include "base.thrift"

namespace go example5

enum FOO {
    A = 1;
}

struct InnerBase {
    1: bool Bool,
    4: i32 Int32,
    5: i64 Int64,
    6: double Double,
    7: string String,
    8: list<string> ListString,
    9: map<string, string> MapStringString,
    10: set<i32> SetInt32,
    12: map<i32, string> MapInt32String,
    13: binary Binary,
    16: map<i64, string> MapInt64String,
    18: list<InnerBase> ListInnerBase,
    19: map<string, InnerBase> MapStringInnerBase,

    255: base.Base Base,
}

service ExampleService {
    InnerBase ExampleMethod(1: InnerBase req)
}
