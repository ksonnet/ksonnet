// @apiVersion 0.1
// @name io.ksonnet.pkg.single-port-service
// @description A service that exposes 'servicePort', and directs traffic
//   to 'targetLabelSelector', at 'targetPort'. Since 'targetLabelSelector' is an
//   object literal that specifies which labels the service is meant to target, this
//   will typically look something like:
//
//     ks prototype use service --targetLabelSelector "{app: 'nginx'}" [...]
// @shortDescription Service that exposes a single port
// @param name string Name of the service
// @param targetLabelSelector object Label for the service to target (e.g., "{app: 'MyApp'}"").
// @optionalParam servicePort number 80 Port for the service to expose
// @optionalParam targetPort number 80 Port for the service target
// @optionalParam protocol string TCP Protocol to use (either TCP or UDP)
// @optionalParam serviceType string ClusterIP Type of service to expose
{
  apiVersion: 'v1',
  kind: 'Service',
  metadata: {
    name: import 'param://name',
  },
  spec: {
    ports: [
      {
        protocol: import 'param://protocol',
        port: import 'param://servicePort',
        targetPort: import 'param://targetPort',
      },
    ],
    selector: import 'param://targetLabelSelector',
    type: import 'param://serviceType',
  },
}
