local params = std.extVar("__ksonnet/params");
local envParams = params + {
  components +: {
    "app.project-1.deployment": {
      type: "nested",
      replicas: 1,
    },
    "deployment": {
      type: "root",
      replicas: 3,
    },
  },
};

{
  components: {
    [x]: envParams.components[x], for x in std.objectFields(envParams.components)
  },
}