// @apiVersion 0.1
// @name io.ksonnet.pkg.single-port-service
// @description long description
//   line 2
// @param name string Name of the service
// @param selector object label
// @optionalParam servicePort string 80 Port for the service to expose
// @optionalParam targetPort string 80 Port for the service target
// @optionalParam protocol string TCP Protocol to use (either TCP or UDP)
// @shortDescription short description

// our object
{
   "apiVersion": "v1",
   "kind": "Service",
   "metadata": {
      "name": params.name
   },
   "spec": {
      "ports": [
         {
            "protocol": params.protocol,
            "port": params.servicePort,
            "targetPort": params.targetPort
         }
      ],
      // our selector
      // @optionalParam serviceType string ClusterIP Type of service to expose
      "selector": params.targetLabelSelector,
      "type": params.type
   }
}
