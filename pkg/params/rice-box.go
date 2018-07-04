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

import (
	"github.com/GeertJohan/go.rice/embedded"
	"time"
)

func init() {

	// define files
	file2 := &embedded.EmbeddedFile{
		Filename:    "modularize_params.libsonnet",
		FileModTime: time.Unix(1530453707, 0),
		Content:     string("function(moduleName, params)\n    local prefix = if (moduleName == \"/\" || moduleName == \"\") then \"\" else \"%s.\" % moduleName;\n\n    local baseObject = if std.objectHas(params, \"global\")\n        then {global: params.global}\n        else {};\n\n    baseObject + {\n        components: {\n            [\"%s%s\" % [prefix, key]]: params.components[key]\n            for key in std.objectFieldsAll(params.components)\n        },\n    }"),
	}
	file3 := &embedded.EmbeddedFile{
		Filename:    "params_for_module.libsonnet",
		FileModTime: time.Unix(1530407391, 0),
		Content:     string("function(moduleName, input)\n   local isModule(key) = std.length(std.split(key, \".\")) > 1;\n\n   local localizeKey(key) =\n      if isModule(key)\n      then\n         local parts = std.split(key, \".\");\n         parts[std.length(parts)-1]\n      else key;\n\n   local findInRoot(key, value) =\n      if isModule(key)\n      then {[key]:null}\n      else {[key]:value};\n\n   local findInModule(moduleName, key, value) =\n      if std.startsWith(key, moduleName)\n      then {[localizeKey(key)]: value}\n      else {[localizeKey(key)]: null};\n\n   local findValue(moduleName, key, value) =\n      if moduleName == \"/\"\n      then findInRoot(key, value)\n      else findInModule(moduleName, key, value);\n\n   local fn(moduleName, params) = [\n         findValue(moduleName, key, params.components[key])\n         for key in std.objectFields(params.components)\n   ];\n\n   {\n      local mash(ag, obj) = ag + obj,\n      components+: std.foldl(mash, std.prune(fn(moduleName, input)), {})\n   }\n"),
	}

	// define dirs
	dir1 := &embedded.EmbeddedDir{
		Filename:   "",
		DirModTime: time.Unix(1530366356, 0),
		ChildFiles: []*embedded.EmbeddedFile{
			file2, // "modularize_params.libsonnet"
			file3, // "params_for_module.libsonnet"

		},
	}

	// link ChildDirs
	dir1.ChildDirs = []*embedded.EmbeddedDir{}

	// register embeddedBox
	embedded.RegisterEmbeddedBox(`scripts`, &embedded.EmbeddedBox{
		Name: `scripts`,
		Time: time.Unix(1530366356, 0),
		Dirs: map[string]*embedded.EmbeddedDir{
			"": dir1,
		},
		Files: map[string]*embedded.EmbeddedFile{
			"modularize_params.libsonnet": file2,
			"params_for_module.libsonnet": file3,
		},
	})
}
