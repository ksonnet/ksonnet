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

package e2e

import "sort"

func genGuestBookParams() []paramListRow {
	return []paramListRow{
		{
			Component: "guestbook-ui",
			Param:     "containerPort",
			Value:     "80",
		},
		{
			Component: "guestbook-ui",
			Param:     "image",
			Value:     "'gcr.io/heptio-images/ks-guestbook-demo:0.1'",
		},
		{
			Component: "guestbook-ui",
			Param:     "name",
			Value:     "'guestbook-ui'",
		},
		{
			Component: "guestbook-ui",
			Param:     "replicas",
			Value:     "1",
		},
		{
			Component: "guestbook-ui",
			Param:     "servicePort",
			Value:     "80",
		},
		{
			Component: "guestbook-ui",
			Param:     "type",
			Value:     "'ClusterIP'",
		},
	}
}

func setGuestBookRow(in []paramListRow, name, value string) []paramListRow {
	out := make([]paramListRow, len(in))
	copy(out, in)

	for i := range out {
		if out[i].Param == name {
			out[i].Value = value
			return out
		}
	}

	out = append(out, paramListRow{Component: "guestbook-ui", Param: name, Value: value})
	sort.Slice(out, func(i, j int) bool {
		return out[i].Param < out[j].Param
	})

	return out
}

func deleteGuestBookRow(in []paramListRow, name string) []paramListRow {
	out := make([]paramListRow, 0)
	for _, row := range in {
		if row.Param != name {
			out = append(out, row)
		}
	}

	return out
}

func genEnvList(serverVersion string) []envListRow {
	return []envListRow{
		{
			KubernetesVersion: serverVersion,
			Name:              "default",
			Namespace:         "default",
			Override:          "",
			Server:            "http://example.com",
		},
	}
}

func setEnvListRow(in []envListRow, name, k8sVersion, namespace, override, server string) []envListRow {
	out := make([]envListRow, len(in))
	copy(out, in)

	for i := range out {
		if out[i].Name == name {
			out[i].KubernetesVersion = k8sVersion
			out[i].Namespace = namespace
			out[i].Override = override
			out[i].Server = server

			return out
		}
	}

	new := envListRow{
		Name:              name,
		KubernetesVersion: k8sVersion,
		Namespace:         name,
		Override:          override,
		Server:            server,
	}

	out = append(out, new)
	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})

	return out
}

func genIncubatorPkgList() []pkgListRow {
	// TODO: is there a better way to set this?
	version := "40285d8a14f1ac5787e405e1023cf0c07f6aa28c"

	return []pkgListRow{
		{Registry: "incubator", Name: "apache", Version: version, Installed: ""},
		{Registry: "incubator", Name: "efk", Version: version, Installed: ""},
		{Registry: "incubator", Name: "mariadb", Version: version, Installed: ""},
		{Registry: "incubator", Name: "memcached", Version: version, Installed: ""},
		{Registry: "incubator", Name: "mongodb", Version: version, Installed: ""},
		{Registry: "incubator", Name: "mysql", Version: version, Installed: ""},
		{Registry: "incubator", Name: "nginx", Version: version, Installed: ""},
		{Registry: "incubator", Name: "node", Version: version, Installed: ""},
		{Registry: "incubator", Name: "postgres", Version: version, Installed: ""},
		{Registry: "incubator", Name: "redis", Version: version, Installed: ""},
		{Registry: "incubator", Name: "tomcat", Version: version, Installed: ""},
	}
}
