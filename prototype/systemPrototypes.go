package prototype

var defaultPrototypes = []*SpecificationSchema{
	&SpecificationSchema{
		APIVersion: "0.1",
		Name:       "io.ksonnet.pkg.yaml-namespace",
		Params: ParamSchemas{
			RequiredParam("name", "name", "Name to give the namespace."),
		},
		Template: SnippetSchema{
			Description: `A simple namespace. Labels are automatically populated from the name of the
namespace.`,
			Body: []string{
				"kind: Namespace",
				"apiVersion: v1",
				"metadata:",
				"  name: ${name}",
				"  labels:",
				"    name: ${name}",
			},
		},
	},
	&SpecificationSchema{
		APIVersion: "0.1",
		Name:       "io.ksonnet.pkg.yaml-single-port-service",
		Params: ParamSchemas{
			RequiredParam("name", "serviceName", "Name of the service"),
			RequiredParam("targetLabelSelector", "selector", "Label for the service to target (e.g., 'app: MyApp')."),
			RequiredParam("servicePort", "port", "Port for the service to expose."),
			RequiredParam("targetPort", "port", "Port for the service target."),
			OptionalParam("protocol", "protocol", "Protocol to use (either TCP or UDP).", "TCP"),
		},
		Template: SnippetSchema{
			Description: `A service that exposes 'servicePort', and directs traffic
to 'targetLabelSelector', at 'targetPort'.`,
			Body: []string{
				"kind: Service",
				"apiVersion: v1",
				"metadata:",
				"  name: ${name}",
				"spec:",
				"  selector:",
				"    ${targetLabelSelector}",
				"  ports:",
				"  - protocol: ${protocol}",
				"    port: ${servicePort}",
				"    targetPort: ${targetPort}",
			},
		},
	},
	&SpecificationSchema{
		APIVersion: "0.1",
		Name:       "io.ksonnet.pkg.yaml-empty-configMap",
		Params: ParamSchemas{
			RequiredParam("serviceName", "name", "Name to give the configMap."),
		},
		Template: SnippetSchema{
			Description: `A simple config map. Contains no data.`,
			Body: []string{
				"apiVersion: v1",
				"kind: ConfigMap",
				"metadata:",
				"  name: ${name}",
				"data:",
				"  // K/V pairs go here.",
			},
		},
	},
	&SpecificationSchema{
		APIVersion: "0.1",
		Name:       "io.ksonnet.pkg.yaml-single-port-deployment",
		Params: ParamSchemas{
			RequiredParam("name", "deploymentName", "Name of the deployment"),
			RequiredParam("image", "containerImage", "Container image to deploy"),
			OptionalParam("replicas", "replicas", "Number of replicas", "1"),
			OptionalParam("port", "containerPort", "Port to expose", "80"),
		},
		Template: SnippetSchema{
			Description: `A deployment that replicates container 'image' some number of times
(default: 1), and exposes a port (default: 80). Labels are automatically
populated from 'name'.`,
			Body: []string{
				"apiVersion: apps/v1beta1",
				"kind: Deployment",
				"metadata:",
				"  name: ${name}",
				"spec:",
				"  replicas: ${replicas:1}",
				"  template:",
				"    metadata:",
				"      labels:",
				"        app: ${name}",
				"    spec:",
				"      containers:",
				"      - name: ${name}",
				"        image: ${image}",
				"        ports:",
				"        - containerPort: ${containerPort:80}",
			},
		},
	},
}
