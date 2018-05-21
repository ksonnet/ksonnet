// @apiVersion 0.1
// @name io.ksonnet.pkg.single-port-deployment
// @description A deployment that replicates container 'image' some number of times (default: 1), and exposes a port (default: 80). Labels are automatically populated from 'name'.
// @shortDescription Replicates a container n times, exposes a single port
// @param name string Name of the deployment
// @param image string Container image to deploy
// @optionalParam replicas number 1 Number of replicas
// @optionalParam containerPort number 80 Port to expose
{
   "apiVersion": "apps/v1beta1",
   "kind": "Deployment",
   "metadata": {
      "name": import 'param://name'
   },
   "spec": {
      "replicas": import 'param://replicas',
      "template": {
         "metadata": {
            "labels": {
               "app": import 'param://name'
            }
         },
         "spec": {
            "containers": [
               {
                  "image": import 'param://image',
                  "name": import 'param://name',
                  "ports": [
                     {
                        "containerPort": import 'param://containerPort'
                     }
                  ]
               }
            ]
         }
      }
   }
}