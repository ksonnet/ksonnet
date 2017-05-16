local test = import "test.libsonnet";

{
  apiVersion: "v1",
  kind: "List",
  items: [
    test {
      string: "bar",
    }
  ],
}
