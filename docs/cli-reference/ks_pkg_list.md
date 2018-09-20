## ks pkg list

List all packages known (downloaded or not) for the current ksonnet app

### Synopsis


The `list` command outputs a table that describes all *known* packages (not
necessarily downloaded, but available from existing registries). This includes
the following info:

1. Library name
2. Registry name
3. Installed status — an asterisk indicates 'installed'

### Related Commands

* `ks pkg install` — Install a package (e.g. extra prototypes) for the current ksonnet app
* `ks pkg describe` — Describe a ksonnet package and its contents
* `ks registry describe` — Describe a ksonnet registry and the packages it contains

### Syntax


```
ks pkg list [flags]
```

### Options

```
  -h, --help            help for list
      --installed       Only list installed packages
  -o, --output string   Output format. Valid options: table|json
```

### Options inherited from parent commands

```
      --dir string        Ksonnet application root to use; Defaults to CWD
      --tls-skip-verify   Skip verification of TLS server certificates
  -v, --verbose count     Increase verbosity. May be given multiple times.
```

### SEE ALSO

* [ks pkg](ks_pkg.md)	 - Manage packages and dependencies for the current ksonnet application

