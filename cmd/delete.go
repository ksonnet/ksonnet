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
	"sort"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

		version, err := utils.FetchVersion(disco)
		if err != nil {
			return err
		}

		sort.Sort(sort.Reverse(utils.DependencyOrder(objs)))

		deleteOpts := metav1.DeleteOptions{}
		if version.Compare(1, 6) < 0 {
			// 1.5.x option
			boolFalse := false
			deleteOpts.OrphanDependents = &boolFalse
		} else {
			// 1.6.x option (NB: Background is broken)
			fg := metav1.DeletePropagationForeground
			deleteOpts.PropagationPolicy = &fg
		}
		if gracePeriod >= 0 {
			deleteOpts.GracePeriodSeconds = &gracePeriod
		}

		for _, obj := range objs {
			desc := fmt.Sprintf("%s %s", utils.ResourceNameFor(disco, obj), utils.FqName(obj))
			log.Info("Deleting ", desc)

			c, err := utils.ClientForResource(clientpool, disco, obj, defaultNs)
			if err != nil {
				return err
			}

			err = c.Delete(obj.GetName(), &deleteOpts)
			if err != nil && !errors.IsNotFound(err) {
				return fmt.Errorf("Error deleting %s: %s", desc, err)
			}

			log.Debugf("Deleted object: ", obj)
		}

		return nil
	},
}
