package utils

import (
	"bytes"
	"encoding/json"
	"io"

	jsonnet "github.com/strickyak/jsonnet_cgo"
	"k8s.io/client-go/pkg/util/yaml"
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
}
