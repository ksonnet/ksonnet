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
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/ksonnet/kubecfg/metadata"
	"github.com/ksonnet/kubecfg/pkg/kubecfg"
	"github.com/ksonnet/kubecfg/template"
	"github.com/ksonnet/kubecfg/utils"

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

	// For use in the commands (e.g., diff, update, delete) that require either an
	// environment or the -f flag.
	flagFile      = "file"
	flagFileShort = "f"
)

var clientConfig clientcmd.ClientConfig

func init() {
	RootCmd.PersistentFlags().CountP(flagVerbose, "v", "Increase verbosity. May be given multiple times.")
	RootCmd.PersistentFlags().StringP(flagJpath, "J", "", "Additional jsonnet library search path")
	RootCmd.PersistentFlags().StringSliceP(flagExtVar, "V", nil, "Values of external variables")
	RootCmd.PersistentFlags().StringSlice(flagExtVarFile, nil, "Read external variable from a file")
	RootCmd.PersistentFlags().StringSliceP(flagTlaVar, "A", nil, "Values of top level arguments")
	RootCmd.PersistentFlags().StringSlice(flagTlaVarFile, nil, "Read top level argument from a file")
	RootCmd.PersistentFlags().String(flagResolver, "noop", "Change implementation of resolveImage native function. One of: noop, registry")
	RootCmd.PersistentFlags().String(flagResolvFail, "warn", "Action when resolveImage fails. One of ignore,warn,error")

	// The "usual" clientcmd/kubectl flags
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	overrides := clientcmd.ConfigOverrides{}
	kflags := clientcmd.RecommendedConfigOverrideFlags("")
	RootCmd.PersistentFlags().StringVar(&loadingRules.ExplicitPath, "kubeconfig", "", "Path to a kube config. Only required if out-of-cluster")
	clientcmd.BindOverrideFlags(&overrides, RootCmd.PersistentFlags(), kflags)
	clientConfig = clientcmd.NewInteractiveDeferredLoadingClientConfig(loadingRules, &overrides, os.Stdin)

	RootCmd.PersistentFlags().Set("logtostderr", "true")
}

// RootCmd is the root of cobra subcommand tree
var RootCmd = &cobra.Command{
	Use:           "kubecfg",
	Short:         "Synchronise Kubernetes resources with config files",
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

func logLevel(verbosity int) log.Level {
	switch verbosity {
	case 0:
		return log.WarnLevel
	case 1:
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

	jpath, err := flags.GetString(flagJpath)
	if err != nil {
		return nil, err
	}
	spec.FlagJpath = filepath.SplitList(jpath)

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

func restClientPool(cmd *cobra.Command) (dynamic.ClientPool, discovery.DiscoveryInterface, error) {
	conf, err := clientConfig.ClientConfig()
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

func addEnvCmdFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringArrayP(flagFile, flagFileShort, nil, "Filename or directory that contains the configuration to apply (accepts YAML, JSON, and Jsonnet)")
}

func parseEnvCmd(cmd *cobra.Command, args []string) (*string, []string, error) {
	flags := cmd.Flags()

	files, err := flags.GetStringArray(flagFile)
	if err != nil {
		return nil, nil, err
	}

	var env *string
	if len(args) == 1 {
		env = &args[0]
	}

	return env, files, nil
}

// TODO: Remove this and use `kubecfg.GetFiles` when we move commands into
// `pkg`.
func getFiles(cmd *cobra.Command, args []string) ([]string, error) {
	env, files, err := parseEnvCmd(cmd, args)
	if err != nil {
		return nil, err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	return kubecfg.GetFiles(metadata.AbsPath(cwd), env, files)
}
