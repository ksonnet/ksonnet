// Run me with `../kubecfg show kubecfg_test.jsonnet`
local kubecfg = import "kubecfg.libsonnet";

assert kubecfg.parseJson("[3, 4]") == [3, 4];

local x = kubecfg.parseYaml("---
- 3
- 4
---
foo: bar
baz: xyzzy
");
assert x == [[3, 4], {foo: "bar", baz: "xyzzy"}] : "got " + x;

local i = kubecfg.resolveImage("busybox");
assert i == "busybox:latest" : "got " + i;

// Kubecfg wants to see something that looks like a k8s object
{
  apiVersion: "test",
  kind: "Result",
  result: "SUCCESS"
}
