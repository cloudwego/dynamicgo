/**
 * Copyright 2023 CloudWeGo Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package meta

import "strconv"

// Encoding protocol enum
type Encoding int

const (
	// EncodingJSON is json protocol
	EncodingJSON Encoding = iota + 1
	// EncodingThriftBinary is thrift binary protocol
	EncodingThriftBinary
	// EncodingThriftCompact is thrift compact protocol
	EncodingThriftCompact
	// EncodingProtobuf is protobuf protocol
	EncodingProtobuf

	// text encoding, see thrift/binary.go::EncodeText()
	EncodingText
)

// String returns the string representation of Encoding
func (p Encoding) String() string {
	switch p {
	case EncodingJSON:
		return "JSON"
	case EncodingThriftCompact:
		return "ThriftCompact"
	case EncodingThriftBinary:
		return "ThriftBinary"
	case EncodingProtobuf:
		return "Protobuf"
	case EncodingText:
		return "Text"
	default:
		return "Unknown"
	}
}

// Category is the category of dynamicgo modules
type Category uint8

const (
	// dynamicgo/json
	JSON Category = 0x01
	// dynamicgo/thrift
	THRIFT Category = 0x02
	// dynamicgo/conv/j2t
	JSON2THRIFT Category = 0x11
	// dynamicgo/conv/t2j
	THRIFT2JSON Category = 0x12
	// dynamicgo/proto
	PROTOBUF Category = 0x03
	// dynamicgo/conv/j2p
	JSON2PROTOBUF Category = 0x13
	// dynamicgo/conv/p2j
	PROTOBUF2JSON Category = 0x14
)

// String returns the string representation of Category
func (ec Category) String() string {
	switch ec {
	case JSON:
		return "JSON"
	case JSON2THRIFT:
		return "JSON-TO-THRIFT"
	case THRIFT2JSON:
		return "THRIFT-TO-JSON"
	case THRIFT:
		return "THRIFT"
	case PROTOBUF:
		return "PROTOBUF"
	case JSON2PROTOBUF:
		return "JSON-TO-PROTOBUF"
	case PROTOBUF2JSON:
		return "PROTOBUF-TO-JSON"
	default:
		return "CATEGORY " + strconv.Itoa(int(ec))
	}
}

// NameCase is the case of field name
type NameCase uint8

const (
	// CaseDefault means use the original field name
	CaseDefault NameCase = iota
	// CaseSnake means use snake case
	CaseSnake
	// CaseUpperCamel means use upper camel case
	CaseUpperCamel
	// CaseLowerCamel means use lower camel case
	CaseLowerCamel
)

// MapFieldWay is the way to map an given key to field for struct descriptor
type MapFieldWay uint8

const (
	// MapFieldUseAlias means use alias to map key to field
	MapFieldUseAlias MapFieldWay = iota
	// MapFieldUseFieldName means use field name to map key to field
	MapFieldUseFieldName
	// MapFieldUseBoth means use both alias and field name to map key to field
	MapFieldUseBoth
)

// ParseFunctionMode indicates to parse only response or request for a IDL
type ParseFunctionMode int8

const (
	// ParseBoth indicates to parse both request and response for a IDL
	ParseBoth ParseFunctionMode = iota
	// ParseRequestOnly indicates to parse only request for a IDL
	ParseRequestOnly
	// ParseResponseOnly indicates to parse only response for a IDL
	ParseResponseOnly
)

// ParseServiceMode .
type ParseServiceMode int8

const (
	// LastServiceOnly forces the parser to parse only the last service definition.
	LastServiceOnly ParseServiceMode = iota

	// FirstServiceOnly forces the parser to parse only the first service definition.
	FirstServiceOnly

	// CombineServices forces the parser to combine methods of all service definitions.
	// Note that method names of the service definitions can not be duplicate.
	CombineServices
)
