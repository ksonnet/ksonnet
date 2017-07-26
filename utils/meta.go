package utils

import (
	"fmt"
	"strconv"

	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery"
)

// ServerVersion captures k8s major.minor version in a parsed form
type ServerVersion struct {
	Major int
	Minor int
}

// ParseVersion parses version.Info into a ServerVersion struct
func ParseVersion(v *version.Info) (ret ServerVersion, err error) {
	ret.Major, err = strconv.Atoi(v.Major)
	if err != nil {
		return
	}
	ret.Minor, err = strconv.Atoi(v.Minor)
	if err != nil {
		return
	}
	return
}

// FetchVersion fetches version information from discovery client, and parses
func FetchVersion(v discovery.ServerVersionInterface) (ret ServerVersion, err error) {
	version, err := v.ServerVersion()
	if err != nil {
		return ServerVersion{}, err
	}
	return ParseVersion(version)
}

// Compare returns -1/0/+1 iff v is less than / equal / greater than major.minor
func (v ServerVersion) Compare(major, minor int) int {
	a := v.Major
	b := major

	if a == b {
		a = v.Minor
		b = minor
	}

	var res int
	if a > b {
		res = 1
	} else if a == b {
		res = 0
	} else {
		res = -1
	}
	return res
}

func (v ServerVersion) String() string {
	return fmt.Sprintf("%d.%d", v.Major, v.Minor)
}
