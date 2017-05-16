package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func init() {
	RootCmd.AddCommand(showCmd)
	showCmd.PersistentFlags().StringP("format", "o", "yaml", "Output format.  Supported values are: json, yaml")
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show expanded resource definitions",
	RunE: func(cmd *cobra.Command, args []string) error {
		flags := cmd.Flags()
		out := cmd.OutOrStdout()

		vm, err := JsonnetVM(cmd)
		if err != nil {
			return err
		}
		defer vm.Destroy()

		objs, err := readObjs(cmd, args)
		if err != nil {
			return err
		}

		format, err := flags.GetString("format")
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
