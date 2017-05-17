package cmd

import (
	goflag "flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	jsonnet "github.com/strickyak/jsonnet_cgo"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/backsplice/kubecfg/utils"
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

	jpath := os.Getenv("KUBECFG_JPATH")
	for _, p := range filepath.SplitList(jpath) {
		glog.V(2).Infoln("Adding jsonnet search path", p)
		vm.JpathAdd(p)
	}

	jpath, err := flags.GetString("jpath")
	if err != nil {
		return nil, err
	}
	for _, p := range filepath.SplitList(jpath) {
		glog.V(2).Infoln("Adding jsonnet search path", p)
		vm.JpathAdd(p)
	}

	return vm, nil
}

func readObjs(cmd *cobra.Command, paths []string) ([]metav1.Object, error) {
	vm, err := JsonnetVM(cmd)
	if err != nil {
		return nil, err
	}
	defer vm.Destroy()

	res := []metav1.Object{}
	for _, path := range paths {
		objs, err := utils.Read(vm, path)
		if err != nil {
			return nil, fmt.Errorf("Error reading %s: %v", path, err)
		}
		res = append(res, utils.FlattenToV1(objs)...)
	}
	return res, nil
}
