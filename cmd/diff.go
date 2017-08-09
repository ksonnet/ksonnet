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
	"io"
	"os"
	"sort"

	"github.com/mattn/go-isatty"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/yudai/gojsondiff"
	"github.com/yudai/gojsondiff/formatter"
	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/ksonnet/kubecfg/utils"
)

const flagDiffStrategy = "diff-strategy"

var ErrDiffFound = fmt.Errorf("Differences found.")

func init() {
	diffCmd.PersistentFlags().String(flagDiffStrategy, "all", "Diff strategy, all or subset.")
	RootCmd.AddCommand(diffCmd)
}

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Display differences between server and local config",
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

		flags := cmd.Flags()
		diffStrategy, err := flags.GetString(flagDiffStrategy)
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

		sort.Sort(utils.AlphabeticalOrder(objs))

		diffFound := false
		for _, obj := range objs {
			desc := fmt.Sprintf("%s %s", utils.ResourceNameFor(disco, obj), utils.FqName(obj))
			log.Debugf("Fetching ", desc)

			c, err := utils.ClientForResource(clientpool, disco, obj, defaultNs)
			if err != nil {
				return err
			}

			liveObj, err := c.Get(obj.GetName())
			if err != nil && errors.IsNotFound(err) {
				log.Debugf("%s doesn't exist on the server", desc)
				liveObj = nil
			} else if err != nil {
				return fmt.Errorf("Error fetching %s: %v", desc, err)
			}

			fmt.Fprintln(out, "---")
			fmt.Fprintf(out, "- live %s\n+ config %s\n", desc, desc)
			if liveObj == nil {
				fmt.Fprintf(out, "%s doesn't exist on server\n", desc)
				diffFound = true
				continue
			}

			liveObjObject := liveObj.Object
			if diffStrategy == "subset" {
				liveObjObject = removeMapFields(obj.Object, liveObjObject)
			}
			diff := gojsondiff.New().CompareObjects(liveObjObject, obj.Object)

			if diff.Modified() {
				diffFound = true
				fcfg := formatter.AsciiFormatterConfig{
					Coloring: istty(out),
				}
				formatter := formatter.NewAsciiFormatter(liveObjObject, fcfg)
				text, err := formatter.Format(diff)
				if err != nil {
					return err
				}
				fmt.Fprintf(out, "%s", text)
			} else {
				fmt.Fprintf(out, "%s unchanged\n", desc)
			}
		}

		if diffFound {
			return ErrDiffFound
		}
		return nil
	},
}

func removeFields(config, live interface{}) interface{} {
	switch c := config.(type) {
	case map[string]interface{}:
		return removeMapFields(c, live.(map[string]interface{}))
	case []interface{}:
		return removeListFields(c, live.([]interface{}))
	default:
		return live
	}
}

func removeMapFields(config, live map[string]interface{}) map[string]interface{} {
	result := map[string]interface{}{}
	for k, v1 := range config {
		v2, ok := live[k]
		if !ok {
			continue
		}
		result[k] = removeFields(v1, v2)
	}
	return result
}

func removeListFields(config, live []interface{}) []interface{} {
	// If live is longer than config, then the extra elements at the end of the
	// list will be returned as is so they appear in the diff.
	result := make([]interface{}, 0, len(live))
	for i, v2 := range live {
		if len(config) > i {
			result = append(result, removeFields(config[i], v2))
		} else {
			result = append(result, v2)
		}
	}
	return result
}

func istty(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return isatty.IsTerminal(f.Fd())
	}
	return false
}
