package prototype

import (
	"fmt"
	"strings"
)

const (
	delimiter = "\x00"
)

type index struct {
	prototypes map[string]*SpecificationSchema
}

func (idx *index) SearchNames(query string, opts SearchOptions) ([]*SpecificationSchema, error) {
	// TODO(hausdorff): This is the world's worst search algorithm. Improve it at
	// some point.

	prototypes := []*SpecificationSchema{}

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
