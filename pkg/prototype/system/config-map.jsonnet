// @apiVersion 0.1
// @name io.ksonnet.pkg.configMap
// @description A simple config map with optional user-specified data.
// @shortDescription A simple config map with optional user-specified data
// @param name string Name to give the configMap.
// @optionalParam data object {} Data for the configMap.
{
   "apiVersion": "v1",
   "data": import 'param://data',
   "kind": "ConfigMap",
   "metadata": {
    "name": import 'param://name'
  }
}