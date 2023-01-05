namespace go null

struct NullStruct {
    1: i64 Null1,
    2: optional i64 Null2,
    3: required i64 Null3,
    4: list<i64> Null4,
    5: map<string,i64> Null5,
    6: map<i64,i64> Null6
}

service NullService {
    NullStruct NullTest(1: NullStruct req)
}