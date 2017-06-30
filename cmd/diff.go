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
	"k8s.io/client-go/pkg/api/errors"

	"github.com/ksonnet/kubecfg/utils"
)

func init() {
	RootCmd.AddCommand(diffCmd)
}

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Display differences between server and local config",
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()

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

		for _, obj := range objs {
			desc := fmt.Sprintf("%s/%s", obj.GetKind(), fqName(obj))
			log.Debugf("Fetching ", desc)

			c, err := clientForResource(clientpool, disco, obj, defaultNs)
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
			fmt.Fprintf(out, "- live %s\n+ config %s", desc, desc)
			if liveObj == nil {
				fmt.Fprintf(out, "%s doesn't exist on server\n", desc)
				continue
			}

			diff := gojsondiff.New().CompareObjects(liveObj.Object, obj.Object)

			if diff.Modified() {
				fcfg := formatter.AsciiFormatterConfig{
					Coloring: istty(out),
				}
				formatter := formatter.NewAsciiFormatter(liveObj.Object, fcfg)
				text, err := formatter.Format(diff)
				if err != nil {
					return err
				}
				fmt.Fprintf(out, "%s", text)
			} else {
				fmt.Fprintf(out, "%s unchanged\n", desc)
			}
		}

		return nil
	},
}

func istty(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return isatty.IsTerminal(f.Fd())
	}
	return false
}
