local test = import "test.libsonnet";
local aVar = std.extVar("aVar");

{
  apiVersion: "v1",
  kind: "List",
  items: [
    test {
      string: "bar",
      notAVal : aVar,
    }
  ],
}
