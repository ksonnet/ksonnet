package cmd

import (
	"encoding/json"
	goflag "flag"
	"path/filepath"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	jsonnet "github.com/strickyak/jsonnet_cgo"
)

func init() {
	RootCmd.PersistentFlags().String("context", "", "The name of the kubeconfig context to use")
	RootCmd.PersistentFlags().StringP("jpath", "J", "", "Additional jsonnet library search path")
	RootCmd.PersistentFlags().AddGoFlagSet(goflag.CommandLine)
	RootCmd.PersistentFlags().Set("logtostderr", "true")
}

// RootCmd is the root of cobra subcommand tree
var RootCmd = &cobra.Command{
	Use:           "kubecfg",
	Short:         "Synchronise Kubernetes resources with config files",
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		goflag.CommandLine.Parse([]string{})
		glog.CopyStandardLogTo("INFO")
	},
}

// JsonnetVM constructs a new jsonnet.VM, according to command line
// flags
func JsonnetVM(cmd *cobra.Command) (*jsonnet.VM, error) {
	vm := jsonnet.Make()
	flags := cmd.Flags()

	jpath, err := flags.GetString("jpath")
	if err != nil {
		return nil, err
	}
	for _, p := range filepath.SplitList(jpath) {
		glog.V(2).Infoln("Adding jsonnet path", p)
		vm.JpathAdd(p)
	}

	return vm, nil
}

func evalFile(vm *jsonnet.VM, file string) (interface{}, error) {
	var err error
	jsonstr := ""
	if file != "" {
		jsonstr, err = vm.EvaluateFile(file)
		if err != nil {
			return nil, err
		}
	}

	glog.V(4).Infof("jsonnet result is: %s\n", jsonstr)

	var jsobj interface{}
	err = json.Unmarshal([]byte(jsonstr), &jsobj)
	if err != nil {
		return nil, err
	}

	return jsobj, nil
}
