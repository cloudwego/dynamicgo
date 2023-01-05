# Dynamic-Go
Dynamically operating data for Go. Aiming at reducing serialization/deserializtion process thus it can be fast as much as possible.

## Introduction
See [introduction.md](introduction.md).

## Usage
### Thrift
Thrift message operators. It can parse thrift IDL in runtime and handle thrift data in generic way.

#### thrift
 Thrift IDL parser, runtime type descriptor and binary-encoding codec. 
 [DOC](thrift/README.md)

#### thrift/generic
 Reflection APIs to search, modify, deserialize, serialize thrift value **with or without** runtime type descriptor.
[DOC](thrift/generic/README.md)

#### thrift/base 
The meta data about message transportation, including caller, address, log-id, etc. It is mostly used for `conv` (protocol convertion) modules.

#### thrift/annotation 
Built-in implementation of thrid-party annotations, see https://www.cloudwego.io/docs/kitex/tutorials/advanced-feature/generic-call/thrift_idl_annotation_standards/. They are mostly used for `conv` (protocol convertion) modules. 

### JSON
JSON value operators. It can search, modify, deserialize, serialize without any type descriptor.
[DOC](json/README.md)

WARN: original codes are imported from https://github.com/bytedance/sonic/ast. It may be optimized and refactored in future.

### Conv
Protocol convertors. Based on reflecting ability of `thrift`, `json` and `protobuf`(WIP) modules, it can convert message from one protocol into another. 

#### j2t
Convert JSON value or JSON-body HTTP request into thrift message.
[DOC](conv/j2t/README.md)

#### t2j
Convert thrift message to JSON value or JSON-body HTTP response.
[DOC](conv/t2j/README.md)