package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	jsonnet "github.com/strickyak/jsonnet_cgo"
	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/util/yaml"
)

// Read fetches and decodes K8s objects by path.
// TODO: Replace this with something supporting more sophisticated
// content negotiation.
func Read(vm *jsonnet.VM, path string) ([]runtime.Object, error) {
	ext := filepath.Ext(path)
	if ext == ".json" {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		return jsonReader(f)
	} else if ext == ".yaml" {
		f, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer f.Close()
		return yamlReader(f)
	} else if ext == ".jsonnet" {
		return jsonnetReader(vm, path)
	}

	return nil, fmt.Errorf("Unknown file extension: %s", path)
}

func jsonReader(r io.Reader) ([]runtime.Object, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	obj, _, err := runtime.UnstructuredJSONScheme.Decode(data, nil, nil)
	if err != nil {
		return nil, err
	}
	return []runtime.Object{obj}, nil
}

func yamlReader(r io.ReadCloser) ([]runtime.Object, error) {
	decoder := yaml.NewYAMLReader(bufio.NewReader(r))
	ret := []runtime.Object{}
	for {
		bytes, err := decoder.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		if len(bytes) == 0 {
			continue
		}
		jsondata, err := yaml.ToJSON(bytes)
		if err != nil {
			return nil, err
		}
		obj, _, err := runtime.UnstructuredJSONScheme.Decode(jsondata, nil, nil)
		if err != nil {
			return nil, err
		}
		ret = append(ret, obj)
	}
	return ret, nil
}

func jsonWalk(obj interface{}) ([]interface{}, error) {
	switch o := obj.(type) {
	case map[string]interface{}:
		if o["kind"] != nil && o["apiVersion"] != nil {
			return []interface{}{o}, nil
		}
		ret := []interface{}{}
		for _, v := range o {
			children, err := jsonWalk(v)
			if err != nil {
				return nil, err
			}
			ret = append(ret, children...)
		}
		return ret, nil
	case []interface{}:
		ret := make([]interface{}, 0, len(o))
		for _, v := range o {
			children, err := jsonWalk(v)
			if err != nil {
				return nil, err
			}
			ret = append(ret, children...)
		}
		return ret, nil
	default:
		return nil, fmt.Errorf("Unexpected object structure: %T", o)
	}
}

func jsonnetReader(vm *jsonnet.VM, path string) ([]runtime.Object, error) {
	jsonstr, err := vm.EvaluateFile(path)
	if err != nil {
		return nil, err
	}

	log.Debugf("jsonnet result is: %s", jsonstr)

	var top interface{}
	if err = json.Unmarshal([]byte(jsonstr), &top); err != nil {
		return nil, err
	}

	objs, err := jsonWalk(top)
	if err != nil {
		return nil, err
	}

	ret := make([]runtime.Object, 0, len(objs))
	for _, v := range objs {
		// TODO: Going to json and back is a bit horrible
		data, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		obj, _, err := runtime.UnstructuredJSONScheme.Decode(data, nil, nil)
		if err != nil {
			return nil, err
		}
		ret = append(ret, obj)
	}

	return ret, nil
}

// FlattenToV1 expands any List-type objects into their members, and
// cooerces everything to v1.Unstructured.  Panics if coercion
// encounters an unexpected object type.
func FlattenToV1(objs []runtime.Object) []*runtime.Unstructured {
	ret := make([]*runtime.Unstructured, 0, len(objs))
	for _, obj := range objs {
		switch o := obj.(type) {
		case *runtime.UnstructuredList:
			for _, item := range o.Items {
				ret = append(ret, item)
			}
		case *runtime.Unstructured:
			ret = append(ret, o)
		default:
			panic("Unexpected unstructured object type")
		}
	}
	return ret
}
