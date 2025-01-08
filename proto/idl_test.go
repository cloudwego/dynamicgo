package proto

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cloudwego/dynamicgo/meta"
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
		rpc ExampleClientStreamingMethod(stream A) returns (B);
		rpc ExampleServerStreamingMethod(A) returns (stream B);
		rpc ExampleBidirectionalStreamingMethod(stream A) returns (stream B);
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

	mtdDsc := svc.LookupMethodByName("ExampleMethod")
	if mtdDsc.isClientStreaming || mtdDsc.isServerStreaming {
		t.Fatal("must be unary streaming")
	}
	mtdDsc = svc.LookupMethodByName("ExampleClientStreamingMethod")
	if !mtdDsc.IsClientStreaming() || mtdDsc.IsServerStreaming() {
		t.Fatal("must be client streaming")
	}
	mtdDsc = svc.LookupMethodByName("ExampleServerStreamingMethod")
	if mtdDsc.IsClientStreaming() || !mtdDsc.IsServerStreaming() {
		t.Fatal("must be server streaming")
	}
	mtdDsc = svc.LookupMethodByName("ExampleBidirectionalStreamingMethod")
	if !mtdDsc.IsClientStreaming() || !mtdDsc.isServerStreaming {
		t.Fatal("must be bidirectional streaming")
	}
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

func TestIsCombinedServices(t *testing.T) {
	opts := Options{}
	importDirs := []string{"../testdata/idl/"}
	svc, err := opts.NewDescriptorFromPath(context.Background(), "example2.proto", importDirs...)
	if err != nil {
		t.Fatal(err)
	}
	if svc.IsCombinedServices() {
		t.Fatal("must not be a combined service")
	}

	opts.ParseServiceMode = meta.CombineServices
	svc, err = opts.NewDescriptorFromPath(context.Background(), "example2.proto", importDirs...)
	if err != nil {
		t.Fatal(err)
	}
	if !svc.IsCombinedServices() {
		t.Fatal("must be a combined service")
	}
	fmt.Printf("%#v\n", svc)
}

func TestParsePackageName(t *testing.T) {
	opts := Options{}
	importDirs := []string{"../testdata/idl/"}
	svc, err := opts.NewDescriptorFromPath(context.Background(), "basic_example.proto", importDirs...)
	require.Nil(t, err, err)
	require.Equal(t, "pb3", svc.PackageName(), svc.PackageName())
}
