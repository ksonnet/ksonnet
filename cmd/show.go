package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func init() {
	RootCmd.AddCommand(showCmd)
	showCmd.PersistentFlags().StringP("file", "f", "", "Input jsonnet file")
	showCmd.MarkFlagFilename("file", "jsonnet", "libsonnet")
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

		file, err := flags.GetString("file")
		if err != nil {
			return err
		}
		jsobj, err := evalFile(vm, file)
		if err != nil {
			return err
		}

		format, err := flags.GetString("format")
		if err != nil {
			return err
		}
		switch format {
		case "yaml":
			buf, err := yaml.Marshal(jsobj)
			if err != nil {
				return err
			}
			out.Write(buf)
		case "json":
			enc := json.NewEncoder(out)
			enc.SetIndent("", "  ")
			if err := enc.Encode(&jsobj); err != nil {
				return err
			}
		default:
			return fmt.Errorf("Unknown --format: %s", format)
		}

		return nil
	},
}
