//go:build sharedlib && go1.20
// +build sharedlib,go1.20

package native

import _ "unsafe"

// https: //cs.opensource.google/go/go/+/master:src/runtime/runtime1.go;l=343;drc=91a40f43b629ac9237967f3faf0733de268ea652
// just set env GODEBUG=cgocheck=0 OR since the run-time already check that
// before this package init() is running
// we can use this hack to disable cgocheck at run-time.

//go:linkname dbgvars runtime.dbgvars
var dbgvars []struct {
	name  string
	value *int32 // control target
}
var _init bool

func init() {
	if !_init {
		for _, dv := range dbgvars {
			if dv.name == "cgocheck" {
				println("disable godebug cgocheck")
				*dv.value = 0
			}
		}
		_init = true
	}
}
