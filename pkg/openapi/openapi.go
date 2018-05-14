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

package openapi

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/validate"
	"github.com/ksonnet/ksonnet-lib/ksonnet-gen/kubespec"
	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var (
	errUnsupportedDefinition = errors.New("unsupported definition")
)

// ValidateAgainstSchema validates a document against the schema.
func ValidateAgainstSchema(a app.App, obj *unstructured.Unstructured, envName string) []error {
	v := newValidateAgainstSchema()
	return v.run(a, obj, envName)
}

type validateAgainstSchema struct {
	definitionName func(*unstructured.Unstructured) (string, error)
	loadSchema     func(app.App, string, string) (*spec.Schema, error)
	validate       func(*spec.Schema, interface{}, strfmt.Registry) error
}

func newValidateAgainstSchema() *validateAgainstSchema {
	return &validateAgainstSchema{
		definitionName: definitionName,
		loadSchema:     loadSchema,
		validate:       validate.AgainstSchema,
	}
}

func (v *validateAgainstSchema) run(a app.App, obj *unstructured.Unstructured, envName string) []error {
	name, err := v.definitionName(obj)
	if err != nil {
		if err == errUnsupportedDefinition {
			return nil
		}

		return []error{err}
	}

	schema, err := v.loadSchema(a, name, envName)
	if err != nil {
		return []error{err}
	}

	if err := v.validate(schema, obj.Object, strfmt.Default); err != nil {
		return []error{err}
	}

	return nil
}

func definitionName(obj *unstructured.Unstructured) (string, error) {
	apiVersion, ok := obj.Object["apiVersion"].(string)
	if !ok {
		return "", errors.New("object does not have apiVersion")
	}

	kind, ok := obj.Object["kind"].(string)
	if !ok {
		return "", errors.New("object does not have kind")
	}

	var name string
	parts := strings.Split(apiVersion, "/")
	switch len(parts) {
	default:
		return "", errors.Errorf("unknown apiVersion %q", apiVersion)
	case 1:
		name = fmt.Sprintf("io.k8s.api.core.%s.%s", parts[0], kind)
	case 2:
		if strings.Contains(parts[0], ".") {
			logrus.WithFields(logrus.Fields{
				"kind":       kind,
				"apiVersion": parts[0],
			}).Warn("ksonnet currently does not support CRDs")

			return "", errUnsupportedDefinition
		}
		name = fmt.Sprintf("io.k8s.api.%s.%s.%s", parts[0], parts[1], kind)
	}

	return name, nil
}

func loadSchema(a app.App, name, envName string) (*spec.Schema, error) {
	libPath, err := a.LibPath(envName)
	if err != nil {
		return nil, err
	}

	schemaPath := filepath.Join(libPath, "swagger.json")
	apiSpec, _, err := kubespec.Import(schemaPath)
	if err != nil {
		return nil, err
	}

	schema, ok := apiSpec.Definitions[name]
	if !ok {
		return nil, errors.Errorf("unable to find definition for %s", name)
	}

	options := &spec.ExpandOptions{
		RelativeBase: schemaPath,
	}
	if err := spec.ExpandSchemaWithBasePath(&schema, nil, options); err != nil {
		return nil, err
	}

	return &schema, nil
}
