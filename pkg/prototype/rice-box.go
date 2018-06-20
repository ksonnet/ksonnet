package prototype

import (
	"github.com/GeertJohan/go.rice/embedded"
	"time"
)

func init() {

	// define files
	file2 := &embedded.EmbeddedFile{
		Filename:    "config-map.jsonnet",
		FileModTime: time.Unix(1526918148, 0),
		Content:     string("// @apiVersion 0.1\n// @name io.ksonnet.pkg.configMap\n// @description A simple config map with optional user-specified data.\n// @shortDescription A simple config map with optional user-specified data\n// @param name string Name to give the configMap.\n// @optionalParam data object {} Data for the configMap.\n{\n   \"apiVersion\": \"v1\",\n   \"data\": import 'param://data',\n   \"kind\": \"ConfigMap\",\n   \"metadata\": {\n    \"name\": import 'param://name'\n  }\n}"),
	}
	file3 := &embedded.EmbeddedFile{
		Filename:    "deployed-service.jsonnet",
		FileModTime: time.Unix(1526918148, 0),
		Content:     string("// @apiVersion 0.1\n// @name io.ksonnet.pkg.deployed-service\n// @description A service that exposes 'servicePort', and directs traffic to 'targetLabelSelector', at 'targetPort'.\n// @shortDescription A deployment exposed with a service\n// @param name string Name of the service and deployment resources\n// @param image string Container image to deploy\n// @optionalParam servicePort number 80 Port for the service to expose.\n// @optionalParam containerPort number 80 Container port for service to target.\n// @optionalParam replicas number 1 Number of replicas\n// @optionalParam type string ClusterIP Type of service to expose\n[\n   {\n      \"apiVersion\": \"v1\",\n      \"kind\": \"Service\",\n      \"metadata\": {\n         \"name\": import 'param://name'\n      },\n      \"spec\": {\n         \"ports\": [\n            {\n               \"port\": import 'param://servicePort',\n               \"targetPort\": import 'param://containerPort'\n            }\n         ],\n         \"selector\": {\n            \"app\": import 'param://name'\n         },\n         \"type\": import 'param://type'\n      }\n   },\n   {\n      \"apiVersion\": \"apps/v1beta2\",\n      \"kind\": \"Deployment\",\n      \"metadata\": {\n         \"name\": import 'param://name'\n      },\n      \"spec\": {\n         \"replicas\": import 'param://replicas',\n         \"selector\": {\n            \"matchLabels\": {\n               \"app\": import 'param://name'\n            },\n         },\n         \"template\": {\n            \"metadata\": {\n               \"labels\": {\n                  \"app\": import 'param://name'\n               }\n            },\n            \"spec\": {\n               \"containers\": [\n                  {\n                     \"image\": import 'param://image',\n                     \"name\": import 'param://name',\n                     \"ports\": [\n                     {\n                        \"containerPort\": import 'param://containerPort'\n                     }\n                     ]\n                  }\n               ]\n            }\n         }\n      }\n   }\n]\n"),
	}
	file4 := &embedded.EmbeddedFile{
		Filename:    "namespace.jsonnet",
		FileModTime: time.Unix(1526918148, 0),
		Content:     string("// @apiVersion 0.1\n// @name io.ksonnet.pkg.namespace\n// @description A simple namespace. Labels are automatically populated from the name of the namespace.\n// @shortDescription Namespace with labels automatically populated from the name\n// @param name string Name to give the namespace\n{\n   \"apiVersion\": \"v1\",\n   \"kind\": \"Namespace\",\n   \"metadata\": {\n      \"labels\": {\n         \"name\": import 'param://name'\n      },\n      \"name\": import 'param://name'\n   }\n}\n"),
	}
	file5 := &embedded.EmbeddedFile{
		Filename:    "single-port-deployment.jsonnet",
		FileModTime: time.Unix(1526918148, 0),
		Content:     string("// @apiVersion 0.1\n// @name io.ksonnet.pkg.single-port-deployment\n// @description A deployment that replicates container 'image' some number of times (default: 1), and exposes a port (default: 80). Labels are automatically populated from 'name'.\n// @shortDescription Replicates a container n times, exposes a single port\n// @param name string Name of the deployment\n// @param image string Container image to deploy\n// @optionalParam replicas number 1 Number of replicas\n// @optionalParam containerPort number 80 Port to expose\n{\n   \"apiVersion\": \"apps/v1beta1\",\n   \"kind\": \"Deployment\",\n   \"metadata\": {\n      \"name\": import 'param://name'\n   },\n   \"spec\": {\n      \"replicas\": import 'param://replicas',\n      \"template\": {\n         \"metadata\": {\n            \"labels\": {\n               \"app\": import 'param://name'\n            }\n         },\n         \"spec\": {\n            \"containers\": [\n               {\n                  \"image\": import 'param://image',\n                  \"name\": import 'param://name',\n                  \"ports\": [\n                     {\n                        \"containerPort\": import 'param://containerPort'\n                     }\n                  ]\n               }\n            ]\n         }\n      }\n   }\n}"),
	}
	file6 := &embedded.EmbeddedFile{
		Filename:    "single-port-service.jsonnet",
		FileModTime: time.Unix(1526918148, 0),
		Content:     string("// @apiVersion 0.1\n// @name io.ksonnet.pkg.single-port-service\n// @description A service that exposes 'servicePort', and directs traffic\n//   to 'targetLabelSelector', at 'targetPort'. Since 'targetLabelSelector' is an\n//   object literal that specifies which labels the service is meant to target, this\n//   will typically look something like:\n//\n//     ks prototype use service --targetLabelSelector \"{app: 'nginx'}\" [...]\n// @shortDescription Service that exposes a single port\n// @param name string Name of the service\n// @param targetLabelSelector object Label for the service to target (e.g., \"{app: 'MyApp'}\"\").\n// @optionalParam servicePort number 80 Port for the service to expose\n// @optionalParam targetPort number 80 Port for the service target\n// @optionalParam protocol string TCP Protocol to use (either TCP or UDP)\n// @optionalParam serviceType string ClusterIP Type of service to expose\n{\n   \"apiVersion\": \"v1\",\n   \"kind\": \"Service\",\n   \"metadata\": {\n      \"name\": import 'param://name'\n   },\n   \"spec\": {\n      \"ports\": [\n         {\n            \"protocol\": import 'param://protocol',\n            \"port\": import 'param://servicePort',\n            \"targetPort\": import 'param://targetPort'\n         }\n      ],\n      \"selector\": import 'param://targetLabelSelector',\n      \"type\": import 'param://serviceType'\n   }\n}"),
	}

	// define dirs
	dir1 := &embedded.EmbeddedDir{
		Filename:   "",
		DirModTime: time.Unix(1526918148, 0),
		ChildFiles: []*embedded.EmbeddedFile{
			file2, // "config-map.jsonnet"
			file3, // "deployed-service.jsonnet"
			file4, // "namespace.jsonnet"
			file5, // "single-port-deployment.jsonnet"
			file6, // "single-port-service.jsonnet"

		},
	}

	// link ChildDirs
	dir1.ChildDirs = []*embedded.EmbeddedDir{}

	// register embeddedBox
	embedded.RegisterEmbeddedBox(`system`, &embedded.EmbeddedBox{
		Name: `system`,
		Time: time.Unix(1526918148, 0),
		Dirs: map[string]*embedded.EmbeddedDir{
			"": dir1,
		},
		Files: map[string]*embedded.EmbeddedFile{
			"config-map.jsonnet":             file2,
			"deployed-service.jsonnet":       file3,
			"namespace.jsonnet":              file4,
			"single-port-deployment.jsonnet": file5,
			"single-port-service.jsonnet":    file6,
		},
	})
}
