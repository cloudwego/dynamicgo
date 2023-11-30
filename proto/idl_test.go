package proto

import (
	"context"
	"fmt"
	"testing"
)

func TestProtoFromContent(t *testing.T) {
	// change / to \\ on windows in path for right mapkey in includes
	// because `filepath.Join(importPath, path)` in SourceResolver.FindFileByPath will change \\ to /
	path := "a/b/main.proto"
	content := `
	syntax = "proto3";
	package pb3;
	option go_package = "pb/main";

	import "x.proto";
	import "../y.proto";

	message Main {
		string name = 1;
	}

	service TestService2 {
		rpc ExampleMethod(A) returns (B);
	}
	`

	includes := map[string]string{
		"a/b/x.proto": `
		syntax = "proto3";
		package pb3;
		option go_package = "pb/a";
		message A {
			string name = 1;
		}
		`,
		"a/y.proto": `
		syntax = "proto3";
		package pb3;
		option go_package = "pb/b";
		message B {
			string name = 1;
		}
		`,
	}

	importDirs := []string{"a/b/"}
	opts := Options{}
	svc, err := opts.NewDesccriptorFromContent(context.Background(), path, content, includes, importDirs...)
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
