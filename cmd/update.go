package cmd

import (
	"encoding/json"
	"fmt"
	"sort"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/api/errors"
	"k8s.io/client-go/pkg/runtime"
	"k8s.io/client-go/pkg/util/diff"

	"github.com/ksonnet/kubecfg/utils"
)

const (
	flagCreate = "create"
)

func init() {
	RootCmd.AddCommand(updateCmd)
	updateCmd.PersistentFlags().Bool(flagCreate, true, "Create missing resources")
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Kubernetes resources with local config",
	RunE: func(cmd *cobra.Command, args []string) error {
		flags := cmd.Flags()

		create, err := flags.GetBool(flagCreate)
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

		sort.Sort(utils.DependencyOrder(objs))

		for _, obj := range objs {
			desc := fmt.Sprintf("%s/%s", obj.GetKind(), fqName(obj))
			log.Info("Updating ", desc)

			c, err := clientForResource(clientpool, disco, obj, defaultNs)
			if err != nil {
				return err
			}

			asPatch, err := json.Marshal(obj)
			if err != nil {
				return err
			}
			newobj, err := c.Patch(obj.GetName(), api.MergePatchType, asPatch)
			if create && errors.IsNotFound(err) {
				log.Info(" Creating non-existent ", desc)
				newobj, err = c.Create(obj)
			}
			if err != nil {
				// TODO: retry
				return fmt.Errorf("Error updating %s: %s", desc, err)
			}

			log.Debug("Updated object: ", diff.ObjectDiff(obj, newobj))
		}

		return nil
	},
}

func fqName(o *runtime.Unstructured) string {
	if o.GetNamespace() == "" {
		return o.GetName()
	}
	return fmt.Sprintf("%s.%s", o.GetNamespace(), o.GetName())
}
