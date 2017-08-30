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
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

const (
	flagFormat = "format"
)

func init() {
	RootCmd.AddCommand(showCmd)
	showCmd.PersistentFlags().StringP(flagFormat, "o", "yaml", "Output format.  Supported values are: json, yaml")
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show expanded resource definitions",
	RunE: func(cmd *cobra.Command, args []string) error {
		flags := cmd.Flags()
		out := cmd.OutOrStdout()

		vm, err := newExpander(cmd)
		if err != nil {
			return err
		}

		objs, err := vm.Expand(args)
		if err != nil {
			return err
		}

		format, err := flags.GetString(flagFormat)
		if err != nil {
			return err
		}
		switch format {
		case "yaml":
			for _, obj := range objs {
				fmt.Fprintln(out, "---")
				// Urgh.  Go via json because we need
				// to trigger the custom scheme
				// encoding.
				buf, err := json.Marshal(obj)
				if err != nil {
					return err
				}
				o := map[string]interface{}{}
				if err := json.Unmarshal(buf, &o); err != nil {
					return err
				}
				buf, err = yaml.Marshal(o)
				if err != nil {
					return err
				}
				out.Write(buf)
			}
		case "json":
			enc := json.NewEncoder(out)
			enc.SetIndent("", "  ")
			for _, obj := range objs {
				// TODO: this is not valid framing for JSON
				if len(objs) > 1 {
					fmt.Fprintln(out, "---")
				}
				if err := enc.Encode(obj); err != nil {
					return err
				}
			}
		default:
			return fmt.Errorf("Unknown --format: %s", format)
		}

		return nil
	},
}
