local params = std.extVar('__ksonnet/params');

params + {
  components+: {
    "app.project-1.ds"+: {
      replicas: 3,
    },
  },
}
