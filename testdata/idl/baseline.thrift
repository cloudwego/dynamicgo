namespace go baseline

struct Simple {
    1: byte ByteField 
    2: i64 I64Field (api.js_conv = "")
    3: double DoubleField
    4: i32 I32Field 
    5: string StringField,
    6: binary BinaryField
}

struct PartialSimple {
     1: byte ByteField 
     3: double DoubleField
     6: binary BinaryField
}

struct Nesting {
    1: string String (api.header = "String")
    2: list<Simple> ListSimple
    3: double Double (api.path = "double")
    4: i32 I32 (api.http_code = "", api.body = "I32")
    5: list<i32> ListI32 (api.query = "ListI32")
    6: i64 I64
    7: map<string, string> MapStringString
    8: Simple SimpleStruct
    9: map<i32, i64> MapI32I64
    10: list<string> ListString
    11: binary Binary
    12: map<i64, string> MapI64String
    13: list<i64> ListI64 (api.cookie = "list_i64"),
    14: byte Byte
    15: map<string, Simple> MapStringSimple
}

struct PartialNesting {
    2: list<PartialSimple> ListSimple
    8: PartialSimple SimpleStruct
    15: map<string, PartialSimple> MapStringSimple
}

struct Nesting2 {
    1: map<Simple, Nesting> MapSimpleNesting
    2: Simple SimpleStruct
    3: byte Byte
    4: double Double
    5: list<Nesting> ListNesting
    6: i64 I64
    7: Nesting NestingStruct
    8: binary Binary
    9: string String
    10: set<Nesting> SetNesting
    11: i32 I32
}

service BaselineService {
    Simple SimpleMethod(1: Simple req) (api.post = "/simple")
    PartialSimple PartialSimpleMethod(1: PartialSimple req)
    Nesting NestingMethod(1: Nesting req) (api.post = "/nesting")
    PartialNesting PartialNestingMethod(1: PartialNesting req)
    Nesting2 Nesting2Method(1: Nesting2 req) (api.post = "/nesting2")
}
