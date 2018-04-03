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
	verbose    = false
	outputFile = "-"
	release    = ""
	namespace  = ""
)

var rootCmd = &cobra.Command{
	Short: "Deploy applications to Kubernetes.",
}

var buildCmd = &cobra.Command{
	Use:   "build [path]...",
	Short: "Build a deployment",
	RunE: func(cmd *cobra.Command, args []string) error {
		paths := args

		if len(paths) == 0 {
			paths = []string{"."}
		}

		cmd.SilenceUsage = true

		deployment, err := deploy.LoadDeploymentFromPaths(paths)

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

var installCmd = &cobra.Command{
	Use:   "install [deployment]",
	Short: "Install a deployment",
	Long:  "Install the specified deployment. If no deployment is specified, it will be expected from the standard input.",
	Args:  cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		deploymentFile := ""

		if len(args) > 0 {
			deploymentFile = args[0]
		}

		cmd.SilenceUsage = true

		var input io.ReadCloser

		if deploymentFile == "" {
			input = os.Stdin
		} else {
			if input, err = os.Open(deploymentFile); err != nil {
				return err
			}
		}

		defer input.Close()

		decoder := yaml.NewDecoder(input)

		var deployment deploy.Deployment

		if err := decoder.Decode(&deployment); err != nil {
			if err != io.EOF {
				return fmt.Errorf("parsing deployment: %s", err)
			}
		}

		ctx := deployment.NewContext(release, namespace)

		var output []byte

		if output, err = deploy.Install(deployment, ctx); err != nil {
			return err
		}

		if verbose {
			fmt.Fprintf(cmd.OutOrStderr(), "%s", output)
		}

		fmt.Fprintf(cmd.OutOrStderr(), "deployment units installed as \"%s\" in namespace \"%s\"\n", ctx.Release, ctx.Namespace)

		return nil
	},
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall <release>...",
	Short: "Uninstall a deployment",
	Long:  "Uninstall the deployment with the specified release.",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		releases := args

		cmd.SilenceUsage = true

		for _, release := range releases {
			if release == "" {
				continue
			}

			ctx := deploy.NewContext(release, namespace)

			var output []byte

			if output, err = deploy.Uninstall(ctx); err != nil {
				return err
			}

			if verbose {
				fmt.Fprintf(cmd.OutOrStderr(), "%s", output)
			}

			fmt.Fprintf(cmd.OutOrStderr(), "deployment units \"%s\" uninstalled from namespace \"%s\"\n", ctx.Release, ctx.Namespace)
		}

		return nil
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all releases in the namespace",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		releases, err := deploy.List(namespace)

		if err != nil {
			return err
		}

		for _, release := range releases {
			fmt.Fprintf(cmd.OutOrStdout(), "%s\n", release)
		}

		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", verbose, "Enable verbose output.")

	buildCmd.Flags().StringVarP(&outputFile, "output-file", "o", outputFile, "The file to write the deployment to. Specify `-` to write to the standard output.")
	buildCmd.MarkFlagFilename("output-file")

	rootCmd.AddCommand(buildCmd)

	installCmd.Flags().StringVarP(&release, "release", "r", release, "The release name to use. If not specified, a random one will be generated.")
	installCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "The namespace name to install into. If not specified, `default` will be assumed.")
	rootCmd.AddCommand(installCmd)

	uninstallCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "The namespace name to uninstall from. If not specified, `default` will be assumed.")
	rootCmd.AddCommand(uninstallCmd)

	listCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "The namespace name to list releases from. If not specified, `default` will be assumed.")
	rootCmd.AddCommand(listCmd)
}

func main() {
	rootCmd.Execute()
}
