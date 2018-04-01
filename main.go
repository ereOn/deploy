package main

import (
	"fmt"

	deploy "github.com/ereOn/deploy/pkg"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Short: "Deploy applications to Kubernetes.",
}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build a deployment.",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := "."

		if len(args) > 0 {
			root = args[0]
		}

		manifestTemplates, err := deploy.LoadManifestTemplatesFromRoot(root)

		for _, manifestTemplate := range manifestTemplates {
			fmt.Printf("- %s\n", manifestTemplate.Name)
		}

		return err
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)
}

func main() {
	rootCmd.Execute()
}
