// Copyright 2018 The ksonnet authors
//
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package jsonnet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-jsonnet"
	"github.com/google/go-jsonnet/ast"
	"github.com/ksonnet/ksonnet/pkg/log"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"k8s.io/apimachinery/pkg/util/yaml"
)

type makeVMFn func() *jsonnet.VM
type evaluateSnippetFn func(vm *jsonnet.VM, name, snippet string) (string, error)

// VMOpt is an option for configuring VM.
type VMOpt func(*VM)

// VM is a ksonnet wrapper for the jsonnet VM.
type VM struct {
	// UseMemoryImporter forces the vm to use a memory importer rather than the
	// file import.
	UseMemoryImporter bool
	Fs                afero.Fs

	jPaths   []string
	extCodes map[string]string
	extVars  map[string]string
	tlaCodes map[string]string
	tlaVars  map[string]string

	makeVMFn          makeVMFn
	evaluateSnippetFn evaluateSnippetFn
}

// NewVM creates an instance of VM.
func NewVM(opts ...VMOpt) *VM {
	vm := &VM{
		jPaths:   make([]string, 0),
		extCodes: make(map[string]string),
		extVars:  make(map[string]string),
		tlaCodes: make(map[string]string),
		tlaVars:  make(map[string]string),

		makeVMFn:          jsonnet.MakeVM,
		evaluateSnippetFn: evaluateSnippet,
	}

	for _, opt := range opts {
		opt(vm)
	}

	return vm
}

// AddJPath adds JPaths to the jsonnet VM.
func (vm *VM) AddJPath(paths ...string) {
	vm.jPaths = append(vm.jPaths, paths...)
}

// ExtCode adds ExtCode to the jsonnet VM.
func (vm *VM) ExtCode(key, value string) {
	vm.extCodes[key] = value
}

// ExtVar adds ExtVar to the jsonnet VM.
func (vm *VM) ExtVar(key, value string) {
	vm.extVars[key] = value
}

// TLACode adds TLACode to the jsonnet VM.
func (vm *VM) TLACode(key, value string) {
	vm.tlaCodes[key] = value
}

// TLAVar adds TLAVar to the jsonnet VM.
func (vm *VM) TLAVar(key, value string) {
	vm.tlaVars[key] = value
}

func evaluateSnippet(vm *jsonnet.VM, name, snippet string) (string, error) {
	return vm.EvaluateSnippet(name, snippet)
}

// EvaluateSnippet evaluates a jsonnet snippet.
func (vm *VM) EvaluateSnippet(name, snippet string) (string, error) {
	now := time.Now()

	fields := logrus.Fields{
		"name": name,
	}

	if log.VerbosityLevel >= 2 {
		fields["jPaths"] = strings.Join(vm.jPaths, ", ")
		fields["snippet"] = snippet
	}

	jvm := jsonnet.MakeVM()
	jvm.ErrorFormatter.SetMaxStackTraceSize(40)
	registerNativeFuncs(jvm)
	importer, err := vm.createImporter()
	if err != nil {
		return "", errors.Wrap(err, "create jsonnet importer")
	}
	jvm.Importer(importer)

	for k, v := range vm.extCodes {
		jvm.ExtCode(k, v)
		if log.VerbosityLevel >= 2 {
			key := fmt.Sprintf("extCode#%s", k)
			fields[key] = v
		}
	}

	for k, v := range vm.extVars {
		jvm.ExtVar(k, v)
		if log.VerbosityLevel >= 2 {
			key := fmt.Sprintf("extVar#%s", k)
			fields[key] = v
		}
	}

	for k, v := range vm.tlaCodes {
		jvm.TLACode(k, v)
		if log.VerbosityLevel >= 2 {
			key := fmt.Sprintf("tlaCode#%s", k)
			fields[key] = v
		}
	}

	for k, v := range vm.tlaVars {
		jvm.TLAVar(k, v)
		if log.VerbosityLevel >= 2 {
			key := fmt.Sprintf("tlaVar#%s", k)
			fields[key] = v
		}
	}

	defer func() {
		fields["elapsed"] = time.Since(now)
		logrus.WithFields(fields).Debug("jsonnet evaluate snippet")
	}()

	return vm.evaluateSnippetFn(jvm, name, snippet)
}

func (vm *VM) createImporter() (jsonnet.Importer, error) {
	if !vm.UseMemoryImporter {
		return &jsonnet.FileImporter{
			JPaths: vm.jPaths,
		}, nil
	}

	if vm.Fs == nil {
		return nil, errors.New("unable to use memory importer without fs")
	}

	importer := &jsonnet.MemoryImporter{
		Data: make(map[string]string),
	}

	for _, jPath := range vm.jPaths {
		fis, err := afero.ReadDir(vm.Fs, jPath)
		if err != nil {
			return nil, err
		}

		for _, fi := range fis {
			if fi.IsDir() {
				continue
			}

			s, err := vm.readString(filepath.Join(jPath, fi.Name()))
			if err != nil {
				return nil, err
			}

			importer.Data[fi.Name()] = s
		}
	}

	return importer, nil
}

func (vm *VM) readString(path string) (string, error) {
	var b []byte

	b, err := afero.ReadFile(vm.Fs, path)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func registerNativeFuncs(vm *jsonnet.VM) {
	// NOTE: jsonnet native functions can only pass primitive
	// types, so some functions json-encode the arg.  These
	// "*FromJson" functions will be replaced by regular native
	// version when jsonnet is able to support this.

	vm.NativeFunction(
		&jsonnet.NativeFunction{
			Name:   "parseJson",
			Params: ast.Identifiers{"json"},
			Func:   parseJSON,
		})

	vm.NativeFunction(
		&jsonnet.NativeFunction{
			Name:   "parseYaml",
			Params: ast.Identifiers{"yaml"},
			Func:   parseYAML,
		})

	vm.NativeFunction(
		&jsonnet.NativeFunction{
			Name:   "escapeStringRegex",
			Params: ast.Identifiers{"str"},
			Func:   escapeStringRegex,
		})

	vm.NativeFunction(
		&jsonnet.NativeFunction{
			Name:   "regexMatch",
			Params: ast.Identifiers{"regex", "string"},
			Func:   regexMatch,
		})

	vm.NativeFunction(
		&jsonnet.NativeFunction{
			Name:   "regexSubst",
			Params: ast.Identifiers{"regex", "src", "repl"},
			Func:   regexSubst,
		})
}

func regexSubst(data []interface{}) (interface{}, error) {
	regex, src, repl := data[0].(string), data[1].(string), data[2].(string)

	r, err := regexp.Compile(regex)
	if err != nil {
		return "", err
	}
	return r.ReplaceAllString(src, repl), nil
}

func regexMatch(s []interface{}) (interface{}, error) {
	return regexp.MatchString(s[0].(string), s[1].(string))
}

func escapeStringRegex(s []interface{}) (interface{}, error) {
	return regexp.QuoteMeta(s[0].(string)), nil
}

func parseYAML(dataString []interface{}) (interface{}, error) {
	data := []byte(dataString[0].(string))
	ret := []interface{}{}
	d := yaml.NewYAMLToJSONDecoder(bytes.NewReader(data))
	for {
		var doc interface{}
		if err := d.Decode(&doc); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		ret = append(ret, doc)
	}
	return ret, nil
}

func parseJSON(dataString []interface{}) (res interface{}, err error) {
	data := []byte(dataString[0].(string))
	err = json.Unmarshal(data, &res)
	return
}
