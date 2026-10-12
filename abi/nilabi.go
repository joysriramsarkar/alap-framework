//go:build cgo
// +build cgo

// Package abi implements the C ABI boundary for the NilX Runtime.
// It allows embedding NilX in Android (JNI/NDK), iOS (Swift/Metal), and desktop apps.
package abi

/*
#include "nilabi.h"
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"unsafe"
)

// Runtime handles cross the C ABI as an opaque pointer. A handle is really an
// integer index (not a Go pointer), so it is boxed in C-allocated memory. That
// way the Go side never converts a uintptr back into an unsafe.Pointer, a
// conversion the Go runtime explicitly forbids.

//export nilx_runtime_create
func nilx_runtime_create() C.NilContext {
	id := CreateRuntime()
	box := C.malloc(C.size_t(unsafe.Sizeof(uintptr(0))))
	*(*C.uintptr_t)(box) = C.uintptr_t(id)
	return C.NilContext(box)
}

//export nilx_runtime_destroy
func nilx_runtime_destroy(ctx C.NilContext) {
	if ctx == nil {
		return
	}
	DestroyRuntime(handleOf(ctx))
	C.free(unsafe.Pointer(ctx))
}

//export nilx_runtime_run
func nilx_runtime_run(ctx C.NilContext, nabc *C.uint8_t, nabc_len C.size_t) C.NilResult {
	data := C.GoBytes(unsafe.Pointer(nabc), C.int(nabc_len))
	err := ExecuteBytecode(handleOf(ctx), data)
	var res C.NilResult
	if err != nil {
		res.ok = C.bool(false)
		res.err.code = C.int32_t(-1)
		res.err.message = C.CString(err.Error())
		return res
	}

	res.ok = C.bool(true)
	return res
}

//export nilx_runtime_dispatch_touch
func nilx_runtime_dispatch_touch(ctx C.NilContext, x C.float, y C.float) C.bool {
	handled := DispatchTouchEvent(handleOf(ctx), float64(x), float64(y))
	return C.bool(handled)
}

//export nilx_runtime_get_ui_json
func nilx_runtime_get_ui_json(ctx C.NilContext) *C.char {
	jsonStr, err := GetUIRootJSON(handleOf(ctx))
	if err != nil {
		return C.CString("{}")
	}
	return C.CString(jsonStr)
}

// handleOf unwraps the opaque C handle into the registered Go runtime ID.
func handleOf(ctx C.NilContext) uintptr {
	if ctx == nil {
		return 0
	}
	return uintptr(*(*C.uintptr_t)(unsafe.Pointer(ctx)))
}
