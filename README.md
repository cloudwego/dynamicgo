# Dynamic-Go
Dynamically operating data for Go. Aiming at reducing serialization/deserializtion process thus it can be fast as much as possible.

## Introduction
Dynamic-Go for Thrift protocol: [introduction.md](introduction.md).

Dynamic-Go for Protobuf protocol: [introduction.md](./proto/INTRODUCTION.md)

## Usage

[![GoDoc](https://godoc.org/github.com/cloudwego/dynamicgo?status.svg)](https://pkg.go.dev/github.com/cloudwego/dynamicgo?tab=doc)

### thrift
Thrift IDL parser and message operators. It can parse thrift IDL in runtime and handle thrift data in generic way.

#### thrift/generic
Reflection APIs to search, modify, deserialize, serialize thrift value **with or without** runtime type descriptor.

#### thrift/base 
The meta data about message transportation, including caller, address, log-id, etc. It is mainly used for `conv` (protocol convertion) modules.

#### thrift/annotation 
Built-in implementation of thrid-party annotations, see [thrift_idl_annotation_standards](https://www.cloudwego.io/docs/kitex/tutorials/advanced-feature/generic-call/thrift_idl_annotation_standards/). They are mainly used for `conv` (protocol convertion) modules. 

### proto
Protobuf IDL parser and message operators. It can parse protobuf IDL in runtime and handle protobuf data in generic way.

#### proto/generic
Reflection APIs to search, modify, deserialize, serialize protobuf value **with or without** runtime descriptor.

#### proto/protowire
Protobuf data encode and decode APIs. It parses and formats the low-level raw wire encoding. It is modified from Protobuf official code [`encoding/protowire`](https://pkg.go.dev/google.golang.org/protobuf/encoding/protowire).

#### proto/binary
BinaryProtocol tool for Protobuf Protocol. It can read, wirte and skip fields directly on binary data of protobuf message.

### http
Http request/response wrapper interfaces. They are mainly used to pass http values on `http<>thrift` conversion. 

### conv
Protocol convertors. Based on reflecting ability of `thrift`, `json` and `protobuf` modules, it can convert message from one protocol into another. 

#### conv/j2t
Convert JSON value or JSON-body HTTP request into thrift message.

#### conv/t2j
Convert thrift message to JSON value or JSON-body HTTP response.

#### conv/j2p
Convert JSON value into protobuf message.

#### conv/p2j
Convert protobuf message into JSON value.


