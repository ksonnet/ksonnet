// Copyright 2017 The kubecfg authors
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

package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"regexp"

	jsonnet "github.com/strickyak/jsonnet_cgo"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func resolveImage(resolver Resolver, image string) (string, error) {
	n, err := ParseImageName(image)
	if err != nil {
		return "", err
	}

	if err := resolver.Resolve(&n); err != nil {
		return "", err
	}

	return n.String(), nil
}

// RegisterNativeFuncs adds kubecfg's native jsonnet functions to provided VM
func RegisterNativeFuncs(vm *jsonnet.VM, resolver Resolver) {
	vm.NativeCallback("parseJson", []string{"json"}, func(data []byte) (res interface{}, err error) {
		err = json.Unmarshal(data, &res)
		return
	})

	vm.NativeCallback("parseYaml", []string{"yaml"}, func(data []byte) ([]interface{}, error) {
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
	})

	vm.NativeCallback("resolveImage", []string{"image"}, func(image string) (string, error) {
		return resolveImage(resolver, image)
	})

	vm.NativeCallback("escapeStringRegex", []string{"str"}, func(s string) (string, error) {
		return regexp.QuoteMeta(s), nil
	})

	vm.NativeCallback("regexMatch", []string{"regex", "string"}, regexp.MatchString)

	vm.NativeCallback("regexSubst", []string{"regex", "src", "repl"}, func(regex, src, repl string) (string, error) {
		r, err := regexp.Compile(regex)
		if err != nil {
			return "", err
		}
		return r.ReplaceAllString(src, repl), nil
	})
}
