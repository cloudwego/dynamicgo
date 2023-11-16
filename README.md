# Dynamic-Go
Dynamically operating data for Go. Aiming at reducing serialization/deserializtion process thus it can be fast as much as possible.

## Introduction
Dynamic-Go for Thrift protocol: [introduction.md](introduction.md).

Dynamic-Go for Protobuf protocol: [introduction.md](./proto/INTRODUCTION.md)

## Usage
### thrift
Thrift IDL parser and message operators. It can parse thrift IDL in runtime and handle thrift data in generic way.
[DOC](thrift/README.md)

#### thrift/generic
Reflection APIs to search, modify, deserialize, serialize thrift value **with or without** runtime type descriptor.
[DOC](thrift/generic/README.md)

#### thrift/base 
The meta data about message transportation, including caller, address, log-id, etc. It is mainly used for `conv` (protocol convertion) modules.

#### thrift/annotation 
Built-in implementation of thrid-party annotations, see [thrift_idl_annotation_standards](https://www.cloudwego.io/docs/kitex/tutorials/advanced-feature/generic-call/thrift_idl_annotation_standards/). They are mainly used for `conv` (protocol convertion) modules. 

### http
Http request/response wrapper interfaces. They are mainly used to pass http values on `http<>thrift` conversion. 
[DOC](http/README.md)

### conv
Protocol convertors. Based on reflecting ability of `thrift`, `json` and `protobuf`(WIP) modules, it can convert message from one protocol into another. 

#### conv/j2t
Convert JSON value or JSON-body HTTP request into thrift message.
[DOC](conv/j2t/README.md)

#### conv/t2j
Convert thrift message to JSON value or JSON-body HTTP response.
[DOC](conv/t2j/README.md)


## Requirements
- Go 1.16~1.20
- Amd64 arch CPU
