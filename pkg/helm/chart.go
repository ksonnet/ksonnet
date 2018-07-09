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

package helm

import (
	"path/filepath"
<<<<<<< HEAD
	"strings"
=======
>>>>>>> 4417ebd4... update version sorting

	"github.com/blang/semver"
	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/util/version"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

// LatestChartVersion finds the latest vendored version of a chart.
func LatestChartVersion(a app.App, repoName, chartName string) (string, error) {
	path := filepath.Join(a.Root(), "vendor", repoName, chartName, "helm")

	fis, err := afero.ReadDir(a.Fs(), path)
	if err != nil {
		return "", errors.Wrapf(err, "reading dir %q", path)
	}

	var versions []version.Version

	for _, fi := range fis {
		if !fi.IsDir() {
			continue
		}

		v, err := version.Make(fi.Name())
		if err != nil {
			return "", err
		}

		versions = append(versions, v)

	}

	if len(versions) == 0 {
		return "", errors.Errorf("chart %s/%s doesn't have any releases", repoName, chartName)
	}

	version.Sort(versions)
	lastestVersion := versions[len(versions)-1]
	return lastestVersion.String(), nil
}
