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

package actions

import (
	"fmt"

	"github.com/ksonnet/ksonnet/client"
	"github.com/ksonnet/ksonnet/metadata/app"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const (
	OptionAPIObjects     = "api-objects"
	OptionApp            = "app"
	OptionClientConfig   = "client-config"
	OptionComponentNames = "component-names"
	OptionCreate         = "create"
	OptionDryRun         = "dry-run"
	OptionEnvName        = "env-name"
	OptionGcTag          = "gc-tag"
	OptionSkipGc         = "skip-gc"
)

type missingOptionError struct {
	name string
}

func newMissingOptionError(name string) *missingOptionError {
	return &missingOptionError{
		name: name,
	}
}

func (e *missingOptionError) Error() string {
	return fmt.Sprintf("missing required %s option", e.name)
}

type invalidOptionError struct {
	name string
}

func newInvalidOptionError(name string) *invalidOptionError {
	return &invalidOptionError{
		name: name,
	}
}

func (e *invalidOptionError) Error() string {
	return fmt.Sprintf("invalid type for option %s", e.name)
}

type optionLoader struct {
	err error
	m   map[string]interface{}
}

func newOptionLoader(m map[string]interface{}) *optionLoader {
	return &optionLoader{
		m: m,
	}
}

func (o *optionLoader) loadBool(name string) bool {
	i := o.load(name)
	if i == nil {
		return false
	}

	a, ok := i.(bool)
	if !ok {
		o.err = newInvalidOptionError(name)
		return false
	}

	return a
}

func (o *optionLoader) loadString(name string) string {
	i := o.load(name)
	if i == nil {
		return ""
	}

	a, ok := i.(string)
	if !ok {
		o.err = newInvalidOptionError(name)
		return ""
	}

	return a
}

func (o *optionLoader) loadStringSlice(name string) []string {
	i := o.load(name)
	if i == nil {
		return nil
	}

	a, ok := i.([]string)
	if !ok {
		o.err = newInvalidOptionError(name)
		return nil
	}

	return a
}

func (o *optionLoader) loadAPIObjects() []*unstructured.Unstructured {
	i := o.load(OptionAPIObjects)
	if i == nil {
		return nil
	}

	a, ok := i.([]*unstructured.Unstructured)
	if !ok {
		o.err = newInvalidOptionError(OptionAPIObjects)
		return nil
	}

	return a
}

func (o *optionLoader) loadClientConfig() *client.Config {
	i := o.load(OptionClientConfig)
	if i == nil {
		return nil
	}

	a, ok := i.(*client.Config)
	if !ok {
		o.err = newInvalidOptionError(OptionClientConfig)
		return nil
	}

	return a
}

func (o *optionLoader) loadApp() app.App {
	i := o.load(OptionApp)
	if i == nil {
		return nil
	}

	a, ok := i.(app.App)
	if !ok {
		o.err = newInvalidOptionError(OptionApp)
		return nil
	}

	return a
}

func (ol *optionLoader) load(key string) interface{} {
	if ol.err != nil {
		return nil
	}

	i, ok := ol.m[key]
	if !ok {
		ol.err = newMissingOptionError(key)
	}

	return i
}
