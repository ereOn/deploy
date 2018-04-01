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

		ctx := deploy.Context{}

		for _, manifestTemplate := range manifestTemplates {
			manifest, err := manifestTemplate.Render(ctx)

			if err != nil {
				return err
			}

			fmt.Printf("- %s\n", manifest.Name)

			for _, document := range manifest.Documents {
				fmt.Printf("  - %s.%s:%s\n", document.APIVersion, document.Kind, document.Name)
				fmt.Printf("    %v\n", document.Labels)
			}
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
