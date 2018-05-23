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

package prototype

import "fmt"

// ParamType represents a type constraint for a prototype parameter (e.g., it
// must be a number).
type ParamType string

const (
	// Number represents a prototype parameter that must be a number.
	Number ParamType = "number"

	// String represents a prototype parameter that must be a string.
	String ParamType = "string"

	// NumberOrString represents a prototype parameter that must be either a
	// number or a string.
	NumberOrString ParamType = "numberOrString"

	// Object represents a prototype parameter that must be an object.
	Object ParamType = "object"

	// Array represents a prototype parameter that must be a array.
	Array ParamType = "array"
)

func parseParamType(t string) (ParamType, error) {
	switch t {
	case "number":
		return Number, nil
	case "string":
		return String, nil
	case "numberOrString":
		return NumberOrString, nil
	case "object":
		return Object, nil
	case "array":
		return Array, nil
	default:
		return "", fmt.Errorf("unknown param type '%s'", t)
	}
}

func (pt ParamType) String() string {
	switch pt {
	case Number:
		return "number"
	case String:
		return "string"
	case NumberOrString:
		return "numberOrString"
	case Object:
		return "object"
	case Array:
		return "array"
	default:
		return "unknown"
	}
}
