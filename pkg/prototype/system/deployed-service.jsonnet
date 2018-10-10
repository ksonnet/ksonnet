// @apiVersion 0.1
// @name io.ksonnet.pkg.deployed-service
// @description A service that exposes 'servicePort', and directs traffic to 'targetLabelSelector', at 'targetPort'.
// @shortDescription A deployment exposed with a service
// @param name string Name of the service and deployment resources
// @param image string Container image to deploy
// @optionalParam servicePort number 80 Port for the service to expose.
// @optionalParam containerPort number 80 Container port for service to target.
// @optionalParam replicas number 1 Number of replicas
// @optionalParam type string ClusterIP Type of service to expose
[
  {
    apiVersion: 'v1',
    kind: 'Service',
    metadata: {
      name: import 'param://name',
    },
    spec: {
      ports: [
        {
          port: import 'param://servicePort',
          targetPort: import 'param://containerPort',
        },
      ],
      selector: {
        app: import 'param://name',
      },
      type: import 'param://type',
    },
  },
  {
    apiVersion: 'apps/v1beta2',
    kind: 'Deployment',
    metadata: {
      name: import 'param://name',
    },
    spec: {
      replicas: import 'param://replicas',
      selector: {
        matchLabels: {
          app: import 'param://name',
        },
      },
      template: {
        metadata: {
          labels: {
            app: import 'param://name',
          },
        },
        spec: {
          containers: [
            {
              image: import 'param://image',
              name: import 'param://name',
              ports: [
                {
                  containerPort: import 'param://containerPort',
                },
              ],
            },
          ],
        },
      },
    },
  },
]
