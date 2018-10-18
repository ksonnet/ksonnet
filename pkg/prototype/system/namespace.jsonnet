// @apiVersion 0.1
// @name io.ksonnet.pkg.namespace
// @description A simple namespace. Labels are automatically populated from the name of the namespace.
// @shortDescription Namespace with labels automatically populated from the name
// @param name string Name to give the namespace
{
  apiVersion: 'v1',
  kind: 'Namespace',
  metadata: {
    labels: {
      name: import 'param://name',
    },
    name: import 'param://name',
  },
}
