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

package env

import (
	"bytes"
	"fmt"
)

const (
	// ComponentsExtCodeKey is the ExtCode key for component imports
	ComponentsExtCodeKey = "__ksonnet/components"
)

// DefaultOverrideData generates the contents for an environment's `main.jsonnet`.
func DefaultOverrideData() []byte {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("local base = import \"%s\";\n", "base.libsonnet"))
	buf.WriteString(fmt.Sprintf("local k = import \"%s\";\n\n", "k.libsonnet"))
	buf.WriteString("base + {\n")
	buf.WriteString("  // Insert user-specified overrides here. For example if a component is named \"nginx-deployment\", you might have something like:\n")
	buf.WriteString("  //   \"nginx-deployment\"+: k.deployment.mixin.metadata.labels({foo: \"bar\"})\n")
	buf.WriteString("}\n")
	return buf.Bytes()
}

// DefaultParamsData generates the contents for an environment's `params.libsonnet`
func DefaultParamsData() []byte {
	const (
		relComponentParamsPath = "../../components/params.libsonnet"
	)

	return []byte(`local params = import "` + relComponentParamsPath + `";
params + {
  components +: {
    // Insert component parameter overrides here. Ex:
    // guestbook +: {
    //   name: "guestbook-dev",
    //   replicas: params.global.replicas,
    // },
  },
}
`)
}

// DefaultBaseData generates environment `base.libsonnet`.
func DefaultBaseData() []byte {
	return []byte(`local components = std.extVar("` + ComponentsExtCodeKey + `");
components + {
  // Insert user-specified overrides here.
}
`)
}
