package main

import (
	"fmt"
	"io"
	"os"

	deploy "github.com/ereOn/deploy/pkg"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

var (
	outputFile = "-"
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

		deployment, err := deploy.LoadDeploymentFromRoot(root)

		if err != nil {
			return err
		}

		var output io.WriteCloser

		if outputFile == "-" {
			output = os.Stdout
		} else {
			output, err = os.Create(outputFile)
		}

		defer output.Close()

		if err := yaml.NewEncoder(output).Encode(deployment); err != nil {
			return fmt.Errorf("failed to write deployment: %s", err)
		}

		return nil
	},
}

func init() {
	buildCmd.Flags().StringVarP(&outputFile, "output-file", "o", outputFile, "The file to write the deployment to. Specify `-` to write to the standard output.")
	buildCmd.MarkFlagFilename("output-file")
	rootCmd.AddCommand(buildCmd)
}

func main() {
	rootCmd.Execute()
}
