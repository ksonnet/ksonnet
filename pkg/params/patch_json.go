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

package params

import "github.com/ksonnet/ksonnet/pkg/util/jsonnet"

// PatchJSON patches components.
func PatchJSON(jsonObject, patch, patchName string) (string, error) {
	vm := jsonnet.NewVM()
	vm.TLACode("target", jsonObject)
	vm.TLACode("patch", patch)
	vm.TLAVar("patchName", patchName)

	return vm.EvaluateSnippet("patchJSON", snippetMergeComponentPatch)
}

var snippetMergeComponentPatch = `
function(target, patch, patchName)
	if std.objectHas(patch.components, patchName) then
		std.mergePatch(target, patch.components[patchName])
	else
		target
`
