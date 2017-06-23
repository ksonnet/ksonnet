{
  // parseJson(data): parses the `data` string as a json document, and
  // returns the resulting jsonnet object.
  parseJson:: std.native("parseJson"),

  // parseYaml(data): parse the `data` string as a YAML stream, and
  // returns an *array* of the resulting jsonnet objects.  A single
  // YAML document will still be returned as an array with one
  // element.
  parseYaml:: std.native("parseYaml"),

  // resolveImage(image): convert the docker image string from
  // image:tag into a more specific image@digest, depending on kubecfg
  // command line flags.
  resolveImage:: std.native("resolveImage")
}
