package version

import (
	"strings"

	"github.com/blang/semver"
)

// Version represents a version.
type Version struct {
	raw string
	v   semver.Version
}

// Make takes a string and converts it to a version. It supports the following:
// * 1.2.3
// * 1.2
// * 1
// * v1.2.3
// * 1.2.3-build-1
func Make(s string) (Version, error) {
	versionStr := strings.TrimPrefix(s, "v")

	parts := strings.SplitN(versionStr, ".", 3)

	switch len(parts) {
	case 3:
		// nothing to do because we have three parts of a version
	case 2:
		// assume we have major and minor
		versionStr = strings.Join(append(parts, "0"), ".")
	case 1:
		// assume we have major
		versionStr = strings.Join(append(parts, "0", "0"), ".")
	}

	v, err := semver.Make(versionStr)
	if err != nil {
		return Version{}, err
	}

	return Version{
		raw: s,
		v:   v,
	}, nil
}

func (v *Version) String() string {
	return v.raw
}
