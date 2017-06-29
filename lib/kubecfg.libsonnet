{
  // parseJson(data): parses the `data` string as a json document, and
  // returns the resulting jsonnet object.
  parseJson:: std.native("parseJson"),

  // parseYaml(data): parse the `data` string as a YAML stream, and
  // returns an *array* of the resulting jsonnet objects.  A single
  // YAML document will still be returned as an array with one
  // element.
  parseYaml:: std.native("parseYaml"),

  // escapeStringRegex(s): Quote the regex metacharacters found in s.
  // The result is a regex that will match the original literal
  // characters.
  escapeStringRegex:: std.native("escapeStringRegex"),

  // resolveImage(image): convert the docker image string from
  // image:tag into a more specific image@digest, depending on kubecfg
  // command line flags.
  resolveImage:: std.native("resolveImage"),

  // regexMatch(regex, string): Returns true if regex is found in
  // string. Regex is as implemented in golang regexp package
  // (python-ish).
  regexMatch:: std.native("regexMatch"),

  // regexSubst(regex, src, repl): Return the result of replacing
  // regex in src with repl.  Replacement string may include $1, etc
  // to refer to submatches.  Regex is as implemented in golang regexp
  // package (python-ish).
  regexSubst:: std.native("regexSubst"),
}
