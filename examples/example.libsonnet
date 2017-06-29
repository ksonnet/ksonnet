// This library is intended to support kubecfg examples, and should
// not necessarily be considered recommended jsonnet structure.

local kubecfg = import "kubecfg.libsonnet";

{
  service(name):: {
    local this = self,

    apiVersion: "v1",
    kind: "Service",
    metadata: {
      name: name,
      labels: { name: name },
    },

    targetPod_:: error "targetPod_ required in this usage",
    spec: {
      ports: [{port: p.containerPort}
              for p in this.targetPod_.spec.containers[0].ports],
      selector: this.targetPod_.metadata.labels,
    }},

  container(name, image):: {
    local this = self,

    name: name,
    image: kubecfg.resolveImage(image),

    env_:: {},  // key/value version of `env` (hidden)
    env: [{name: k, value: this.env_[k]} for k in std.objectFields(this.env_)],

    ports: [],
  },

  deployment(name):: {
    local this = self,

    apiVersion: "extensions/v1beta1",
    kind: "Deployment",
    metadata: {
      name: name,
      labels: { name: name },
    },
    spec: {
      replicas: 1,
      template: {
        metadata: { labels: this.metadata.labels },
        spec: {
          containers: []
        }}}},
}
