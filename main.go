package main

import (
	"fmt"

	deploy "github.com/ereOn/deploy/pkg"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Short: "Deploy applications to Kubernetes.",
}

var compileCmd = &cobra.Command{
	Use:   "compile",
	Short: "Compile a list of Kubernetes manifest templates.",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := "."

		if len(args) > 0 {
			root = args[0]
		}

		data, err := deploy.CompileManifestTemplates(root)

		fmt.Printf("%s\n", data)

		return err
	},
}

func init() {
	rootCmd.AddCommand(compileCmd)
}

func main() {
	rootCmd.Execute()
}
