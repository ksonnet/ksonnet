package cmd

import (
	"fmt"
	"sort"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"k8s.io/client-go/pkg/api/errors"
	"k8s.io/client-go/pkg/api/v1"

	"github.com/ksonnet/kubecfg/utils"
)

const (
	flagGracePeriod = "grace-period"
)

func init() {
	RootCmd.AddCommand(deleteCmd)
	deleteCmd.PersistentFlags().Int64(flagGracePeriod, -1, "Number of seconds given to resources to terminate gracefully. A negative value is ignored")
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete Kubernetes resources described in local config",
	RunE: func(cmd *cobra.Command, args []string) error {
		flags := cmd.Flags()
		boolFalse := false

		gracePeriod, err := flags.GetInt64(flagGracePeriod)
		if err != nil {
			return err
		}

		objs, err := readObjs(cmd, args)
		if err != nil {
			return err
		}

		clientpool, disco, err := restClientPool(cmd)
		if err != nil {
			return err
		}

		defaultNs, _, err := clientConfig.Namespace()
		if err != nil {
			return err
		}

		sort.Sort(sort.Reverse(utils.DependencyOrder(objs)))

		deleteOpts := v1.DeleteOptions{OrphanDependents: &boolFalse}
		if gracePeriod >= 0 {
			deleteOpts.GracePeriodSeconds = &gracePeriod
		}

		for _, obj := range objs {
			desc := fmt.Sprintf("%s/%s", obj.GetKind(), fqName(obj))
			glog.Info("Deleting ", desc)

			c, err := clientForResource(clientpool, disco, obj, defaultNs)
			if err != nil {
				return err
			}

			err = c.Delete(obj.GetName(), &deleteOpts)
			if err != nil && !errors.IsNotFound(err) {
				return fmt.Errorf("Error deleting %s: %s", desc, err)
			}

			glog.V(2).Info("Deleted object: ", obj)
		}

		return nil
	},
}
