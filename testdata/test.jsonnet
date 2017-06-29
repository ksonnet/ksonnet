local test = import "test.libsonnet";
local aVar = std.extVar("aVar");
local anVar = std.extVar("anVar");

{
  apiVersion: "v1",
  kind: "List",
  items: [
    test {
      string: "bar",
      notAVal : aVar,
      notAnotherVal : anVar,
    }
  ],
}
