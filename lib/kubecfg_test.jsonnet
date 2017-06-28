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

assert kubecfg.regexMatch("o$", "foo");

local r1 = kubecfg.escapeStringRegex("f[o");
assert r1 == "f\\[o" : "got " + r1;

local r2 = kubecfg.regexSubst("e", "tree", "oll");
assert r2 == "trolloll" : "got " + r2;

// Kubecfg wants to see something that looks like a k8s object
{
  apiVersion: "test",
  kind: "Result",
  result: "SUCCESS"
}
