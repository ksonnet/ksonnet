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

package snippet

// AppendComponent takes the following params
//
//   component: the name of the new component to be added.
//   snippet: a jsonnet snippet resembling the current component parameters.
//   params: the parameters for the new component.
//
// and returns the jsonnet snippet with the appended component and parameters.
func AppendComponent(component, snippet string, params map[string]string) (string, error) {
	return appendComponent(component, snippet, params)
}

// GetComponentParams takes
//
//  component: the name of the component to retrieve params for.
//  snippet: the jsonnet snippet containing the component parameters.
//
// and returns a map of key-value param pairs corresponding to that component.
//
// An error will be returned if the component is not found in the snippet.
func GetComponentParams(component, snippet string) (map[string]string, error) {
	params, _, err := getComponentParams(component, snippet)
	return params, err
}

// SetComponentParams takes
//
//   component: the name of the new component to be modified.
//   snippet: a jsonnet snippet resembling the current component parameters.
//   params: the parameters to be set for 'component'.
//
// and returns the jsonnet snippet with the modified set of component parameters.
func SetComponentParams(component, snippet string, params map[string]string) (string, error) {
	return setComponentParams(component, snippet, params)
}

// SetEnvironmentParams takes
//
//   component: the name of the new component to be modified.
//   snippet: a jsonnet snippet resembling the current environment parameters (not expanded).
//   params: the parameters to be set for 'component'.
//
// and returns the jsonnet snippet with the modified set of environment parameters.
func SetEnvironmentParams(component, snippet string, params map[string]string) (string, error) {
	return setEnvironmentParams(component, snippet, params)
}
