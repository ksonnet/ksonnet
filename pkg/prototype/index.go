// Copyright 2018 The kubecfg authors
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

import (
	"fmt"
	"strings"
)

type index struct {
	prototypes map[string]*Prototype
}

func (idx *index) List() (Prototypes, error) {
	prototypes := []*Prototype{}
	for _, prototype := range idx.prototypes {
		prototypes = append(prototypes, prototype)
	}
	return prototypes, nil
}

func (idx *index) SearchNames(query string, opts SearchOptions) (Prototypes, error) {
	// TODO(hausdorff): This is the world's worst search algorithm. Improve it at
	// some point.

	prototypes := []*Prototype{}

	for name, prototype := range idx.prototypes {
		isSearchResult := false
		switch opts {
		case Prefix:
			isSearchResult = strings.HasPrefix(name, query)
		case Suffix:
			isSearchResult = strings.HasSuffix(name, query)
		case Substring:
			isSearchResult = strings.Contains(name, query)
		default:
			return nil, fmt.Errorf("Unrecognized search option '%d'", opts)
		}

		if isSearchResult {
			prototypes = append(prototypes, prototype)
		}
	}

	return prototypes, nil
}
