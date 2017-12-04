// Copyright 2017 The kubecfg authors
//
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/pflag"

	"github.com/ksonnet/ksonnet/metadata"
	"github.com/ksonnet/ksonnet/prototype"
	"github.com/ksonnet/ksonnet/prototype/snippet"
	"github.com/ksonnet/ksonnet/prototype/snippet/jsonnet"
	"github.com/ksonnet/ksonnet/utils"
	"github.com/spf13/cobra"
)

var protoShortDesc = map[string]string{
	"list":     "List all locally available ksonnet prototypes",
	"describe": "See more info about a prototype's output and usage",
	"preview":  "Preview a prototype's output without creating a component (stdout)",
	"search":   "Search for a prototype",
	"use":      "Use the specified prototype to generate a component manifest",
}

func init() {
	RootCmd.AddCommand(prototypeCmd)
	RootCmd.AddCommand(generateCmd)
	prototypeCmd.AddCommand(prototypeListCmd)
	prototypeCmd.AddCommand(prototypeDescribeCmd)
	prototypeCmd.AddCommand(prototypeSearchCmd)
	prototypeCmd.AddCommand(prototypeUseCmd)
	prototypeCmd.AddCommand(prototypePreviewCmd)
}

var prototypeCmd = &cobra.Command{
	Use:   "prototype",
	Short: `Instantiate, inspect, and get examples for ksonnet prototypes`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return fmt.Errorf("%s is not a valid subcommand\n\n%s", strings.Join(args, " "), cmd.UsageString())
		}
		return fmt.Errorf("Command 'prototype' requires a subcommand\n\n%s", cmd.UsageString())
	},
	Long: `
Use the` + " `prototype` " + `subcommands to manage, inspect, instantiate, and get
examples for ksonnet prototypes.

Prototypes are pre-written but incomplete Kubernetes manifests, with "holes"
(parameters) that can be filled in with the ksonnet CLI or manually. For example,
the prototype` + " `io.ksonnet.pkg.single-port-deployment` " + `requires a name and image,
and the ksonnet CLI can expand this into a fully-formed 'Deployment' object.

These complete manifests are output into your ` + "`components/`" + ` directory. In other
words, prototypes provide the basis for the **components** of your app. You can
use prototypes to autogenerate boilerplate code and focus on customizing them
for your use case.

----
`,
}

var prototypeListCmd = &cobra.Command{
	Use:   "list",
	Short: protoShortDesc["list"],
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return fmt.Errorf("Command 'prototype list' does not take any arguments")
		}

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		wd := metadata.AbsPath(cwd)

		manager, err := metadata.Find(wd)
		if err != nil {
			return err
		}

		extProtos, err := manager.GetAllPrototypes()
		if err != nil {
			return err
		}

		index := prototype.NewIndex(extProtos)
		protos, err := index.List()
		if err != nil {
			return err
		} else if len(protos) == 0 {
			return fmt.Errorf("No prototypes found")
		}

		fmt.Print(protos)

		return nil
	},
	Long: `
The ` + "`list`" + ` command displays all prototypes that are available locally, as
well as brief descriptions of what they generate.

ksonnet comes with a set of system prototypes that you can use out-of-the-box
(e.g.` + " `io.ksonnet.pkg.configMap`" + `). However, you can use more advanced
prototypes like ` + "`io.ksonnet.pkg.redis-stateless`" + ` by downloading extra packages
from the *incubator* registry.

### Related Commands

* ` + "`ks prototype describe` " + `— ` + protoShortDesc["describe"] + `
* ` + "`ks prototype preview` " + `— ` + protoShortDesc["preview"] + `
* ` + "`ks prototype use` " + `— ` + protoShortDesc["use"] + `
* ` + "`ks pkg install` " + pkgShortDesc["install"] + `

### Syntax
`,
}

var prototypeDescribeCmd = &cobra.Command{
	Use:   "describe <prototype-name>",
	Short: protoShortDesc["describe"],
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("Command 'prototype describe' requires a prototype name\n\n%s", cmd.UsageString())
		}

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		wd := metadata.AbsPath(cwd)

		extProtos := prototype.SpecificationSchemas{}
		manager, err := metadata.Find(wd)
		if err == nil {
			extProtos, err = manager.GetAllPrototypes()
			if err != nil {
				return err
			}
		}

		query := args[0]

		proto, err := fundUniquePrototype(query, extProtos)
		if err != nil {
			return err
		}

		fmt.Println(`PROTOTYPE NAME:`)
		fmt.Println(proto.Name)
		fmt.Println()
		fmt.Println(`DESCRIPTION:`)
		fmt.Println(proto.Template.Description)
		fmt.Println()
		fmt.Println(`REQUIRED PARAMETERS:`)
		fmt.Println(proto.RequiredParams().PrettyString("  "))
		fmt.Println()
		fmt.Println(`OPTIONAL PARAMETERS:`)
		fmt.Println(proto.OptionalParams().PrettyString("  "))
		fmt.Println()
		fmt.Println(`TEMPLATE TYPES AVAILABLE:`)
		fmt.Println(fmt.Sprintf("  %s", proto.Template.AvailableTemplates()))

		return nil
	},
	Long: `
This command outputs documentation, examples, and other information for
the specified prototype (identified by name). Specifically, this describes:

  1. What sort of component is generated
  2. Which parameters (required and optional) can be passed in via CLI flags
     to customize the component
  3. The file format of the generated component manifest (currently, Jsonnet only)

### Related Commands

* ` + "`ks prototype preview` " + `— ` + protoShortDesc["preview"] + `
* ` + "`ks prototype use` " + `— ` + protoShortDesc["use"] + `

### Syntax
`,

	Example: `
# Display documentation about the prototype 'io.ksonnet.pkg.single-port-deployment'
ks prototype describe deployment`,
}

var prototypeSearchCmd = &cobra.Command{
	Use:   "search <name-substring>",
	Short: protoShortDesc["search"],
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("Command 'prototype search' requires a prototype name\n\n%s", cmd.UsageString())
		}

		query := args[0]

		index := prototype.NewIndex([]*prototype.SpecificationSchema{})
		protos, err := index.SearchNames(query, prototype.Substring)
		if err != nil {
			return err
		} else if len(protos) == 0 {
			return fmt.Errorf("Failed to find any search results for query '%s'", query)
		}

		fmt.Print(protos)

		return nil
	},
	Long: `
The ` + "`prototype search`" + ` command allows you to search for specific prototypes by name.
Specifically, it matches any prototypes with names that contain the string <name-substring>.

### Related Commands

* ` + "`ks prototype describe` " + `— ` + protoShortDesc["describe"] + `
* ` + "`ks prototype list` " + `— ` + protoShortDesc["list"] + `

### Syntax
`,
	Example: `
# Search for prototypes with names that contain the string 'service'.
ks prototype search service`,
}

var prototypePreviewCmd = &cobra.Command{
	Use:                "preview <prototype-name> [parameter-flags]",
	Short:              protoShortDesc["preview"],
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, rawArgs []string) error {
		if len(rawArgs) == 1 && (rawArgs[0] == "--help" || rawArgs[0] == "-h") {
			return cmd.Help()
		}

		if len(rawArgs) < 1 {
			return fmt.Errorf("Command 'prototype preview' requires a prototype name\n\n%s", cmd.UsageString())
		}

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		wd := metadata.AbsPath(cwd)

		extProtos := prototype.SpecificationSchemas{}
		manager, err := metadata.Find(wd)
		if err == nil {
			extProtos, err = manager.GetAllPrototypes()
			if err != nil {
				return err
			}
		}

		query := rawArgs[0]

		proto, err := fundUniquePrototype(query, extProtos)
		if err != nil {
			return err
		}

		bindPrototypeFlags(cmd, proto)

		cmd.DisableFlagParsing = false
		err = cmd.ParseFlags(rawArgs)
		if err != nil {
			return err
		}
		flags := cmd.Flags()

		// Try to find the template type (if it is supplied) after the args are
		// parsed. Note that the case that `len(args) == 0` is handled at the
		// beginning of this command.
		var templateType prototype.TemplateType
		if args := flags.Args(); len(args) == 1 {
			templateType = prototype.Jsonnet
		} else if len(args) == 2 {
			templateType, err = prototype.ParseTemplateType(args[1])
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("Incorrect number of arguments supplied to 'prototype preview'\n\n%s", cmd.UsageString())
		}

		params, err := getParameters(proto, flags)
		if err != nil {
			return err
		}

		text, err := expandPrototype(proto, templateType, params, "preview")
		if err != nil {
			return err
		}

		fmt.Println(text)
		return nil
	},
	Long: `
This ` + "`preview`" + ` command expands a prototype with CLI flag parameters, and
emits the resulting manifest to stdout. This allows you to see the potential
output of a ` + "`ks generate`" + ` command without actually creating a new component file.

The output is formatted in Jsonnet. To see YAML or JSON equivalents, first create
a component with ` + "`ks generate`" + ` and then use ` + "`ks show`" + `.

### Related Commands

* ` + "`ks generate` " + `— ` + protoShortDesc["use"] + `

### Syntax
`,
	Example: `
# Preview prototype 'io.ksonnet.pkg.single-port-deployment', using the
# 'nginx' image, and port 80 exposed.
ks prototype preview single-port-deployment \
  --name=nginx                              \
  --image=nginx                             \
  --port=80`,
}

// generateCmd acts as an alias for `prototype use`
var generateCmd = &cobra.Command{
	Use:                "generate <prototype-name> <component-name> [type] [parameter-flags]",
	Short:              prototypeUseCmd.Short,
	DisableFlagParsing: prototypeUseCmd.DisableFlagParsing,
	RunE:               prototypeUseCmd.RunE,
	Long:               prototypeUseCmd.Long,
	Example:            prototypeUseCmd.Example,
}

var prototypeUseCmd = &cobra.Command{
	Use:                "use <prototype-name> <componentName> [type] [parameter-flags]",
	Short:              protoShortDesc["use"],
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, rawArgs []string) error {
		if len(rawArgs) == 1 && (rawArgs[0] == "--help" || rawArgs[0] == "-h") {
			return cmd.Help()
		}

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		manager, err := metadata.Find(metadata.AbsPath(cwd))
		if err != nil {
			return fmt.Errorf("Command can only be run in a ksonnet application directory:\n\n%v", err)
		}

		extProtos, err := manager.GetAllPrototypes()
		if err != nil {
			return err
		}

		if len(rawArgs) < 1 {
			return fmt.Errorf("Command requires a prototype name\n\n%s", cmd.UsageString())
		}

		query := rawArgs[0]

		proto, err := fundUniquePrototype(query, extProtos)
		if err != nil {
			return err
		}

		bindPrototypeFlags(cmd, proto)

		cmd.DisableFlagParsing = false
		err = cmd.ParseFlags(rawArgs)
		if err != nil {
			return err
		}
		flags := cmd.Flags()

		// Try to find the template type (if it is supplied) after the args are
		// parsed. Note that the case that `len(args) == 0` is handled at the
		// beginning of this command.
		var componentName string
		var templateType prototype.TemplateType
		if args := flags.Args(); len(args) == 1 {
			return fmt.Errorf("Command is missing argument 'componentName'\n\n%s", cmd.UsageString())
		} else if len(args) == 2 {
			componentName = args[1]
			templateType = prototype.Jsonnet
		} else if len(args) == 3 {
			componentName = args[1]
			templateType, err = prototype.ParseTemplateType(args[1])
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("Command has too many arguments (takes a prototype name and a component name)\n\n%s", cmd.UsageString())
		}

		name, err := flags.GetString("name")
		if err != nil {
			return err
		}

		if name == "" {
			flags.Set("name", componentName)
		}

		params, err := getParameters(proto, flags)
		if err != nil {
			return err
		}

		text, err := expandPrototype(proto, templateType, params, componentName)
		if err != nil {
			return err
		}

		return manager.CreateComponent(componentName, text, params, templateType)
	},
	Long: `
The ` + "`generate`" + ` command (aliased from ` + "`prototype use`" + `) generates Kubernetes-
compatible, Jsonnet ` + `manifests for components in your ksonnet app. Each component
corresponds to a single manifest in the` + " `components/` " + `directory. This manifest
can define one or more Kubernetes resources, and is generated from a ksonnet
*prototype* (a customizable, reusable Kubernetes configuration snippet).

1. The first argument, the **prototype name**, can either be fully qualified
(e.g.` + " `io.ksonnet.pkg.single-port-service`" + `) or a partial match (e.g.` +
		" `service`" + `).
If using a partial match, note that any ambiguity in resolving the name will
result in an error.

2. The second argument, the **component name**, determines the filename for the
generated component manifest. For example, the following command will expand
template` + " `io.ksonnet.pkg.single-port-deployment` " + `and place it in the
file` + " `components/nginx-depl.jsonnet` " + `. Note that by default ksonnet will
expand prototypes into Jsonnet files.

       ks prototype use io.ksonnet.pkg.single-port-deployment nginx-depl \
         --image=nginx

  If the optional ` + "`--name`"  + ` tag is not specified, all Kubernetes API resources
  declared by this prototype use this argument as their own ` + "`metadata.name`" + `

3. Prototypes can be further customized by passing in **parameters** via additional
command line flags, such as ` + " `--image` " + `in the example above. Note that
different prototypes support their own unique flags.

### Related Commands

* ` + "`ks show` " + `— ` + showShortDesc + `
* ` + "`ks apply` " + `— ` + applyShortDesc + `
* ` + "`ks param set` " + paramShortDesc["set"] + `

### Syntax
`,
	Example: `
# Instantiate prototype 'io.ksonnet.pkg.single-port-deployment', using the
# 'nginx' image. The expanded prototype is placed in
# 'components/nginx-depl.jsonnet'.
# The associated Deployment has metadata.name 'nginx-depl'.
ks prototype use io.ksonnet.pkg.single-port-deployment nginx-depl \
  --image=nginx

# Instantiate prototype 'io.ksonnet.pkg.single-port-deployment' using the
# suffix, 'deployment'. (This works unless there is an ambiguity, e.g. another
# prototype with 'deployment' in its name.) The expanded prototype is again
# placed in 'components/nginx-depl.jsonnet'.
# The associated Deployment has metadata.name 'nginx' instead of 'nginx-depl'
# (due to --name).
ks prototype use deployment nginx-depl \
  --name=nginx                         \
  --image=nginx`,
}

func bindPrototypeFlags(cmd *cobra.Command, proto *prototype.SpecificationSchema) {
	for _, param := range proto.RequiredParams() {
		cmd.PersistentFlags().String(param.Name, "", param.Description)
	}

	for _, param := range proto.OptionalParams() {
		cmd.PersistentFlags().String(param.Name, *param.Default, param.Description)
	}
}

func expandPrototype(proto *prototype.SpecificationSchema, templateType prototype.TemplateType, params map[string]string, componentName string) (string, error) {
	template, err := proto.Template.Body(templateType)
	if err != nil {
		return "", err
	}
	if templateType == prototype.Jsonnet {
		componentsText := "components." + componentName
		if !utils.IsASCIIIdentifier(componentName) {
			componentsText = fmt.Sprintf(`components["%s"]`, componentName)
		}
		template = append([]string{`local params = std.extVar("` + metadata.ParamsExtCodeKey + `").` + componentsText + ";"}, template...)
		return jsonnet.Parse(componentName, strings.Join(template, "\n"))
	}

	tm := snippet.Parse(strings.Join(template, "\n"))
	return tm.Evaluate(params)
}

func getParameters(proto *prototype.SpecificationSchema, flags *pflag.FlagSet) (map[string]string, error) {
	missingReqd := prototype.ParamSchemas{}
	values := map[string]string{}
	for _, param := range proto.RequiredParams() {
		val, err := flags.GetString(param.Name)
		if err != nil {
			return nil, err
		} else if val == "" {
			missingReqd = append(missingReqd, param)
		} else if _, ok := values[param.Name]; ok {
			return nil, fmt.Errorf("Prototype '%s' has multiple parameters with name '%s'", proto.Name, param.Name)
		}

		quoted, err := param.Quote(val)
		if err != nil {
			return nil, err
		}
		values[param.Name] = quoted
	}

	if len(missingReqd) > 0 {
		return nil, fmt.Errorf("Failed to instantiate prototype '%s'. The following required parameters are missing:\n%s", proto.Name, missingReqd.PrettyString(""))
	}

	for _, param := range proto.OptionalParams() {
		val, err := flags.GetString(param.Name)
		if err != nil {
			return nil, err
		} else if _, ok := values[param.Name]; ok {
			return nil, fmt.Errorf("Prototype '%s' has multiple parameters with name '%s'", proto.Name, param.Name)
		}

		quoted, err := param.Quote(val)
		if err != nil {
			return nil, err
		}
		values[param.Name] = quoted
	}

	return values, nil
}

func fundUniquePrototype(query string, extProtos prototype.SpecificationSchemas) (*prototype.SpecificationSchema, error) {
	index := prototype.NewIndex(extProtos)

	suffixProtos, err := index.SearchNames(query, prototype.Suffix)
	if err != nil {
		return nil, err
	}

	if len(suffixProtos) == 1 {
		// Success.
		return suffixProtos[0], nil
	} else if len(suffixProtos) > 1 {
		// Ambiguous match.
		names := specNames(suffixProtos)
		return nil, fmt.Errorf("Ambiguous match for '%s':\n%s", query, strings.Join(names, "\n"))
	} else {
		// No matches.
		substrProtos, err := index.SearchNames(query, prototype.Substring)
		if err != nil || len(substrProtos) == 0 {
			return nil, fmt.Errorf("No prototype names matched '%s'", query)
		}

		partialMatches := specNames(substrProtos)
		partials := strings.Join(partialMatches, "\n")
		return nil, fmt.Errorf("No prototype names matched '%s'; a list of partial matches:\n%s", query, partials)
	}
}

func specNames(protos []*prototype.SpecificationSchema) []string {
	partialMatches := []string{}
	for _, proto := range protos {
		partialMatches = append(partialMatches, proto.Name)
	}

	return partialMatches
}
