namespace go skip

struct TestListMap {
    1: list<map<string, string>> listMap
}

service SkipTest {
    TestListMap TestListMap(1: TestListMap req)
}
