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
	"bytes"
	"encoding/json"
	goflag "flag"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/ksonnet/ksonnet/metadata"
	"github.com/ksonnet/ksonnet/template"
	"github.com/ksonnet/ksonnet/utils"

	// Register auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

const (
	flagVerbose    = "verbose"
	flagJpath      = "jpath"
	flagExtVar     = "ext-str"
	flagExtVarFile = "ext-str-file"
	flagTlaVar     = "tla-str"
	flagTlaVarFile = "tla-str-file"
	flagResolver   = "resolve-images"
	flagResolvFail = "resolve-images-error"
	flagAPISpec    = "api-spec"

	// For use in the commands (e.g., diff, apply, delete) that require either an
	// environment or the -f flag.
	flagComponent      = "component"
	flagComponentShort = "c"
)

var clientConfig clientcmd.ClientConfig
var overrides clientcmd.ConfigOverrides
var loadingRules clientcmd.ClientConfigLoadingRules

func init() {
	RootCmd.PersistentFlags().CountP(flagVerbose, "v", "Increase verbosity. May be given multiple times.")

	// The "usual" clientcmd/kubectl flags
	loadingRules = *clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	clientConfig = clientcmd.NewInteractiveDeferredLoadingClientConfig(&loadingRules, &overrides, os.Stdin)

	RootCmd.PersistentFlags().Set("logtostderr", "true")
}

func bindJsonnetFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringSliceP(flagJpath, "J", nil, "Additional jsonnet library search path")
	cmd.PersistentFlags().StringSliceP(flagExtVar, "V", nil, "Values of external variables")
	cmd.PersistentFlags().StringSlice(flagExtVarFile, nil, "Read external variable from a file")
	cmd.PersistentFlags().StringSliceP(flagTlaVar, "A", nil, "Values of top level arguments")
	cmd.PersistentFlags().StringSlice(flagTlaVarFile, nil, "Read top level argument from a file")
	cmd.PersistentFlags().String(flagResolver, "noop", "Change implementation of resolveImage native function. One of: noop, registry")
	cmd.PersistentFlags().String(flagResolvFail, "warn", "Action when resolveImage fails. One of ignore,warn,error")
}

func bindClientGoFlags(cmd *cobra.Command) {
	kflags := clientcmd.RecommendedConfigOverrideFlags("")
	ep := &loadingRules.ExplicitPath
	cmd.PersistentFlags().StringVar(ep, "kubeconfig", "", "Path to a kubeconfig file. Alternative to env var $KUBECONFIG.")
	clientcmd.BindOverrideFlags(&overrides, cmd.PersistentFlags(), kflags)
}

// RootCmd is the root of cobra subcommand tree
var RootCmd = &cobra.Command{
	Use:   "ks",
	Short: `Configure your application to deploy to a Kubernetes cluster`,
	Long: `
You can use the ` + "`ks`" + ` commands to write, share, and deploy your Kubernetes
application configuration to remote clusters.

----
`,
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		goflag.CommandLine.Parse([]string{})
		flags := cmd.Flags()
		out := cmd.OutOrStderr()
		log.SetOutput(out)

		logFmt := NewLogFormatter(out)
		log.SetFormatter(logFmt)

		verbosity, err := flags.GetCount(flagVerbose)
		if err != nil {
			return err
		}
		log.SetLevel(logLevel(verbosity))

		return nil
	},
}

// clientConfig.Namespace() is broken in client-go 3.0:
// namespace in config erroneously overrides explicit --namespace
func namespace() (string, error) {
	return namespaceFor(clientConfig, &overrides)
}

func namespaceFor(c clientcmd.ClientConfig, overrides *clientcmd.ConfigOverrides) (string, error) {
	if overrides.Context.Namespace != "" {
		return overrides.Context.Namespace, nil
	}
	ns, _, err := clientConfig.Namespace()
	return ns, err
}

// resolveContext returns the server and namespace of the cluster at the
// provided context. If the context string is empty, the default context is
// used.
func resolveContext(context string) (server, namespace string, err error) {
	rawConfig, err := clientConfig.RawConfig()
	if err != nil {
		return "", "", err
	}

	// use the default context where context is empty
	if context == "" {
		if rawConfig.CurrentContext == "" && len(rawConfig.Clusters) == 0 {
			// User likely does not have a kubeconfig file.
			return "", "", fmt.Errorf("No current context found. Make sure a kubeconfig file is present")
		}
		// Note: "" is a valid rawConfig.CurrentContext
		context = rawConfig.CurrentContext
	}

	ctx := rawConfig.Contexts[context]
	if ctx == nil {
		return "", "", fmt.Errorf("context '%s' does not exist in the kubeconfig file", context)
	}

	log.Infof("Using context '%s' from the kubeconfig file specified at the environment variable $KUBECONFIG", context)
	cluster, exists := rawConfig.Clusters[ctx.Cluster]
	if !exists {
		return "", "", fmt.Errorf("No cluster with name '%s' exists", ctx.Cluster)
	}

	return cluster.Server, ctx.Namespace, nil
}

func logLevel(verbosity int) log.Level {
	switch verbosity {
	case 0:
		return log.InfoLevel
	default:
		return log.DebugLevel
	}
}

type logFormatter struct {
	escapes  *terminal.EscapeCodes
	colorise bool
}

// NewLogFormatter creates a new log.Formatter customised for writer
func NewLogFormatter(out io.Writer) log.Formatter {
	var ret = logFormatter{}
	if f, ok := out.(*os.File); ok {
		ret.colorise = terminal.IsTerminal(int(f.Fd()))
		ret.escapes = terminal.NewTerminal(f, "").Escape
	}
	return &ret
}

func (f *logFormatter) levelEsc(level log.Level) []byte {
	switch level {
	case log.DebugLevel:
		return []byte{}
	case log.WarnLevel:
		return f.escapes.Yellow
	case log.ErrorLevel, log.FatalLevel, log.PanicLevel:
		return f.escapes.Red
	default:
		return f.escapes.Blue
	}
}

func (f *logFormatter) Format(e *log.Entry) ([]byte, error) {
	buf := bytes.Buffer{}
	if f.colorise {
		buf.Write(f.levelEsc(e.Level))
		fmt.Fprintf(&buf, "%-5s ", strings.ToUpper(e.Level.String()))
		buf.Write(f.escapes.Reset)
	}

	buf.WriteString(strings.TrimSpace(e.Message))
	buf.WriteString("\n")

	return buf.Bytes(), nil
}

func newExpander(cmd *cobra.Command) (*template.Expander, error) {
	flags := cmd.Flags()
	spec := template.Expander{}
	var err error

	spec.EnvJPath = filepath.SplitList(os.Getenv("KUBECFG_JPATH"))

	spec.FlagJpath, err = flags.GetStringSlice(flagJpath)
	if err != nil {
		return nil, err
	}

	spec.ExtVars, err = flags.GetStringSlice(flagExtVar)
	if err != nil {
		return nil, err
	}

	spec.ExtVarFiles, err = flags.GetStringSlice(flagExtVarFile)
	if err != nil {
		return nil, err
	}

	spec.TlaVars, err = flags.GetStringSlice(flagTlaVar)
	if err != nil {
		return nil, err
	}

	spec.TlaVarFiles, err = flags.GetStringSlice(flagTlaVarFile)
	if err != nil {
		return nil, err
	}

	spec.Resolver, err = flags.GetString(flagResolver)
	if err != nil {
		return nil, err
	}
	spec.FailAction, err = flags.GetString(flagResolvFail)
	if err != nil {
		return nil, err
	}

	return &spec, nil
}

// For debugging
func dumpJSON(v interface{}) string {
	buf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(buf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return err.Error()
	}
	return string(buf.Bytes())
}

func restClient(cmd *cobra.Command, envName *string, config clientcmd.ClientConfig, overrides *clientcmd.ConfigOverrides) (dynamic.ClientPool, discovery.DiscoveryInterface, error) {
	if envName != nil {
		err := overrideCluster(*envName, config, overrides)
		if err != nil {
			return nil, nil, err
		}
	}

	conf, err := config.ClientConfig()
	if err != nil {
		return nil, nil, err
	}

	disco, err := discovery.NewDiscoveryClientForConfig(conf)
	if err != nil {
		return nil, nil, err
	}

	discoCache := utils.NewMemcachedDiscoveryClient(disco)
	mapper := discovery.NewDeferredDiscoveryRESTMapper(discoCache, dynamic.VersionInterfaces)
	pathresolver := dynamic.LegacyAPIPathResolverFunc

	pool := dynamic.NewClientPool(conf, mapper, pathresolver)
	return pool, discoCache, nil
}

func restClientPool(cmd *cobra.Command, envName *string) (dynamic.ClientPool, discovery.DiscoveryInterface, error) {
	return restClient(cmd, envName, clientConfig, &overrides)
}

// addEnvCmdFlags adds the flags that are common to the family of commands
// whose form is `[<env>|-f <file-name>]`, e.g., `apply` and `delete`.
func addEnvCmdFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringArrayP(flagComponent, flagComponentShort, nil, "Name of a specific component (multiple -c flags accepted, allows YAML, JSON, and Jsonnet)")
}

// overrideCluster ensures that the server specified in the environment is
// associated in the user's kubeconfig file during deployment to a ksonnet
// environment. We will error out if it is not.
//
// If the environment server the user is attempting to deploy to is not the current
// kubeconfig context, we must manually override the client-go --cluster flag
// to ensure we are deploying to the correct cluster.
func overrideCluster(envName string, clientConfig clientcmd.ClientConfig, overrides *clientcmd.ConfigOverrides) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	wd := metadata.AbsPath(cwd)

	metadataManager, err := metadata.Find(wd)
	if err != nil {
		return err
	}

	rawConfig, err := clientConfig.RawConfig()
	if err != nil {
		return err
	}

	var servers = make(map[string]string)
	for name, cluster := range rawConfig.Clusters {
		server, err := utils.NormalizeURL(cluster.Server)
		if err != nil {
			return err
		}

		servers[server] = name
	}

	//
	// check to ensure that the environment we are trying to deploy to is
	// created, and that the server is located in kubeconfig.
	//

	log.Debugf("Validating deployment at '%s' with server '%v'", envName, reflect.ValueOf(servers).MapKeys())
	env, err := metadataManager.GetEnvironment(envName)
	if err != nil {
		return err
	}

	server, err := utils.NormalizeURL(env.Server)
	if err != nil {
		return err
	}

	if _, ok := servers[server]; ok {
		clusterName := servers[server]
		if overrides.Context.Cluster == "" {
			log.Debugf("Overwriting --cluster flag with '%s'", clusterName)
			overrides.Context.Cluster = clusterName
		}
		if overrides.Context.Namespace == "" {
			log.Debugf("Overwriting --namespace flag with '%s'", env.Namespace)
			overrides.Context.Namespace = env.Namespace
		}
		return nil
	}

	return fmt.Errorf("Attempting to deploy to environment '%s' at '%s', but cannot locate a server at that address", envName, env.Server)
}

// expandEnvCmdObjs finds and expands templates for the family of commands of
// the form `[<env>|-f <file-name>]`, e.g., `apply` and `delete`. That is, if
// the user passes a list of files, we will expand all templates in those files,
// while if a user passes an environment name, we will expand all component
// files using that environment.
func expandEnvCmdObjs(cmd *cobra.Command, env string, components []string, cwd metadata.AbsPath) ([]*unstructured.Unstructured, error) {
	expander, err := newExpander(cmd)
	if err != nil {
		return nil, err
	}

	//
	// Set up the template expander to be able to expand the ksonnet application.
	//

	manager, err := metadata.Find(cwd)
	if err != nil {
		return nil, err
	}

	libPath, vendorPath, envLibPath, envComponentPath, envParamsPath := manager.LibPaths(env)
	expander.FlagJpath = append([]string{string(libPath), string(vendorPath), string(envLibPath)}, expander.FlagJpath...)

	componentPaths, err := manager.ComponentPaths()
	if err != nil {
		return nil, err
	}

	baseObj, err := constructBaseObj(componentPaths, components)
	if err != nil {
		return nil, err
	}
	params := importParams(string(envParamsPath))
	expander.ExtCodes = append([]string{baseObj, params}, expander.ExtCodes...)

	//
	// Expand the ksonnet app as rendered for environment `env`.
	//

	return expander.Expand([]string{string(envComponentPath)})
}

// constructBaseObj constructs the base Jsonnet object that represents k-v
// pairs of component name -> component imports. For example,
//
//   {
//      foo: import "components/foo.jsonnet"
//      "foo-bar": import "components/foo-bar.jsonnet"
//   }
func constructBaseObj(componentPaths, componentNames []string) (string, error) {
	// IMPLEMENTATION NOTE: If one or more `componentNames` exist, it is
	// sufficient to simply omit every name that does not appear in the list. This
	// is because we know every field of the base object will contain _only_ an
	// `import` node (see example object in the function-heading comment). This
	// would not be true in cases where one field can reference another field; in
	// this case, one would need to generate the entire object, and filter that.
	//
	// Hence, a word of caution: if the base object ever becomes more complex, you
	// will need to change the way this function performs filtering, as it will
	// lead to very confusing bugs.

	shouldFilter := len(componentNames) > 0
	filter := map[string]string{}
	for _, name := range componentNames {
		filter[name] = ""
	}

	// Add every component we know about to the base object.
	var obj bytes.Buffer
	obj.WriteString("{\n")
	for _, p := range componentPaths {
		ext := path.Ext(p)
		componentName := strings.TrimSuffix(path.Base(p), ext)

		// Filter! If the filter has more than 1 element and the component name is
		// not in the filter, skip.
		if _, exists := filter[componentName]; shouldFilter && !exists {
			continue
		} else if shouldFilter && exists {
			delete(filter, componentName)
		}

		// Generate import statement.
		var importExpr string
		switch ext {
		case ".jsonnet":
			importExpr = fmt.Sprintf(`import "%s"`, p)

		// TODO: Pull in YAML and JSON when we build the base object.
		//
		// case ".yaml", ".yml":
		// 	importExpr = fmt.Sprintf(`util.parseYaml("%s")`, p)
		// case ".json":
		// 	importExpr = fmt.Sprintf(`util.parseJson("%s")`, p)
		default:
			continue
		}

		// Emit object field. Sanitize the name to guarantee we generate valid
		// Jsonnet.
		componentName = utils.QuoteNonASCII(componentName)
		fmt.Fprintf(&obj, "  %s: %s,\n", componentName, importExpr)
	}

	// Check that we found all the components the user asked for.
	if shouldFilter && len(filter) != 0 {
		names := []string{}
		for name := range filter {
			names = append(names, "'"+name+"'")
		}
		return "", fmt.Errorf("Failed to filter components; the following components don't exist: [ %s ]", strings.Join(names, ","))
	}

	// Terminate object.
	fmt.Fprintf(&obj, "}\n")

	// Emit `base.libsonnet`.
	return fmt.Sprintf("%s=%s", metadata.ComponentsExtCodeKey, obj.String()), nil
}

func importParams(path string) string {
	return fmt.Sprintf(`%s=import "%s"`, metadata.ParamsExtCodeKey, path)
}
