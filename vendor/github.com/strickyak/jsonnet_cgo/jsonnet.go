/*
jsonnet is a simple Go wrapper for the JSonnet VM.

See http://jsonnet.org/
*/
package jsonnet

// By Henry Strickland <@yak.net:strick>
// Made self-contained by Marko Mikulicic <mkm@bitnami.com>

/*
#include <memory.h>
#include <string.h>
#include <stdio.h>
#include "bridge.h"
#cgo CXXFLAGS: -std=c++0x -O3
*/
import "C"

import (
	"errors"
	"unsafe"
)

type ImportCallback func(base, rel string) (result string, path string, err error)

type VM struct {
	guts           *C.struct_JsonnetVm
	importCallback ImportCallback
}

//export go_call_import
func go_call_import(vmPtr unsafe.Pointer, base, rel *C.char, pathPtr **C.char, okPtr *C.int) *C.char {
	vm := (*VM)(vmPtr)
	result, path, err := vm.importCallback(C.GoString(base), C.GoString(rel))
	if err != nil {
		*okPtr = C.int(0)
		return C.CString(err.Error())
	}
	*pathPtr = C.CString(path)
	*okPtr = C.int(1)
	return C.CString(result)
}

// Evaluate a file containing Jsonnet code, return a JSON string.
func Version() string {
	return C.GoString(C.jsonnet_version())
}

// Create a new Jsonnet virtual machine.
func Make() *VM {
	vm := &VM{guts: C.jsonnet_make()}
	return vm
}

// Complement of Make().
func (vm *VM) Destroy() {
	C.jsonnet_destroy(vm.guts)
	vm.guts = nil
}

// Evaluate a file containing Jsonnet code, return a JSON string.
func (vm *VM) EvaluateFile(filename string) (string, error) {
	var e C.int
	z := C.GoString(C.jsonnet_evaluate_file(vm.guts, C.CString(filename), &e))
	if e != 0 {
		return "", errors.New(z)
	}
	return z, nil
}

// Evaluate a string containing Jsonnet code, return a JSON string.
func (vm *VM) EvaluateSnippet(filename, snippet string) (string, error) {
	var e C.int
	z := C.GoString(C.jsonnet_evaluate_snippet(vm.guts, C.CString(filename), C.CString(snippet), &e))
	if e != 0 {
		return "", errors.New(z)
	}
	return z, nil
}

// Override the callback used to locate imports.
func (vm *VM) ImportCallback(f ImportCallback) {
	vm.importCallback = f
	C.jsonnet_import_callback(vm.guts, C.JsonnetImportCallbackPtr(C.CallImport_cgo), unsafe.Pointer(vm))
}

// Bind a Jsonnet external var to the given value.
func (vm *VM) ExtVar(key, val string) {
	C.jsonnet_ext_var(vm.guts, C.CString(key), C.CString(val))
}

// Bind a Jsonnet external var to the given Jsonnet code.
func (vm *VM) ExtCode(key, val string) {
	C.jsonnet_ext_code(vm.guts, C.CString(key), C.CString(val))
}

// Bind a Jsonnet top-level argument to the given value.
func (vm *VM) TlaVar(key, val string) {
	C.jsonnet_tla_var(vm.guts, C.CString(key), C.CString(val))
}

// Bind a Jsonnet top-level argument to the given Jsonnet code.
func (vm *VM) TlaCode(key, val string) {
	C.jsonnet_tla_code(vm.guts, C.CString(key), C.CString(val))
}

// Set the maximum stack depth.
func (vm *VM) MaxStack(v uint) {
	C.jsonnet_max_stack(vm.guts, C.uint(v))
}

// Set the number of lines of stack trace to display (0 for all of them).
func (vm *VM) MaxTrace(v uint) {
	C.jsonnet_max_trace(vm.guts, C.uint(v))
}

// Set the number of objects required before a garbage collection cycle is allowed.
func (vm *VM) GcMinObjects(v uint) {
	C.jsonnet_gc_min_objects(vm.guts, C.uint(v))
}

// Run the garbage collector after this amount of growth in the number of objects.
func (vm *VM) GcGrowthTrigger(v float64) {
	C.jsonnet_gc_growth_trigger(vm.guts, C.double(v))
}

// Expect a string as output and don't JSON encode it.
func (vm *VM) StringOutput(v bool) {
	if v {
		C.jsonnet_string_output(vm.guts, C.int(1))
	} else {
		C.jsonnet_string_output(vm.guts, C.int(0))
	}
}

// Add to the default import callback's library search path.
func (vm *VM) JpathAdd(path string) {
	C.jsonnet_jpath_add(vm.guts, C.CString(path))
}

/* The following are not implemented because they are trivial to implement in Go on top of the
 * existing API by parsing and post-processing the JSON output by regular evaluation.
 *
 * jsonnet_evaluate_file_multi
 * jsonnet_evaluate_snippet_multi
 * jsonnet_evaluate_file_stream
 * jsonnet_evaluate_snippet_stream
 */
