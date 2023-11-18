package proto

import (
	"context"
	"fmt"
	"testing"
)

func TestProtoFromContent(t *testing.T) {
	// TODO
	path := "main.proto"
	content := `
	syntax = "proto3";
	package pb3;
	option go_package = "pb/main";

	import "x.proto";
	import "y.proto";

	message Main {
		string name = 1;
	}

	service TestService2 {
		rpc ExampleMethod(A) returns (B);
	}
	`

	includes := map[string]string{
		path: content,
		"x.proto": `
		syntax = "proto3";
		package pb3;
		option go_package = "pb/a";
		message A {
			string name = 1;
		}
		`,
		"y.proto": `
		syntax = "proto3";
		package pb3;
		option go_package = "pb/b";
		message B {
			string name = 1;
		}
		`,
	}


	opts := Options{}
	svc, err := opts.NewDesccriptorFromContent(context.Background(), path, includes)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%#v\n", svc)

}

func TestProtoFromPath(t *testing.T) {
	opts := Options{}
	importDirs := []string{"../testdata/idl/"}
	svc, err := opts.NewDescriptorFromPath(context.Background(), "example2.proto", importDirs...)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%#v\n", svc)
}