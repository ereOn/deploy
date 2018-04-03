package deploy

import (
	"bytes"
	"fmt"
	"os/exec"

	yaml "gopkg.in/yaml.v2"
)

// Install the specified deployment.
func Install(deployment Deployment, ctx Context) ([]byte, error) {
	data, err := deployment.Render(ctx)

	if err != nil {
		return nil, err
	}

	input := bytes.NewBuffer(data)
	args := []string{"apply", "--prune", "-f", "-", fmt.Sprintf("--namespace=%s", ctx.Namespace), fmt.Sprintf("-l %s=%s", releaseLabel, ctx.Release)}

	cmd := exec.Command("kubectl", args...)
	cmd.Stdin = input

	output, err := cmd.Output()

	if err != nil {
		if err, ok := err.(*exec.ExitError); ok {
			return output, fmt.Errorf("kubectl failed: %s", err.Stderr)
		}
	}

	return output, err
}

// Uninstall the deployment matching the specified context.
func Uninstall(ctx Context) ([]byte, error) {
	args := []string{"delete", "all", fmt.Sprintf("--namespace=%s", ctx.Namespace), fmt.Sprintf("-l %s=%s", releaseLabel, ctx.Release)}
	cmd := exec.Command("kubectl", args...)

	output, err := cmd.Output()

	if err != nil {
		if err, ok := err.(*exec.ExitError); ok {
			return output, fmt.Errorf("kubectl failed: %s", err.Stderr)
		}
	}

	return output, err
}

// List the deployment releases in the specified namespace.
func List(namespace string) (releases map[string]ParametersType, err error) {
	args := []string{"get", "all", fmt.Sprintf("--namespace=%s", namespace), "-o", "yaml"}
	cmd := exec.Command("kubectl", args...)

	output, err := cmd.Output()

	if err != nil {
		if err, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("kubectl failed: %s", err.Stderr)
		}

		return nil, err
	}

	document := Document{Context: NewContext("", namespace)}

	if err = yaml.NewDecoder(bytes.NewBuffer(output)).Decode(&document); err != nil {
		return nil, fmt.Errorf("failed to decode YAML output of kubectl: %s", err)
	}

	releases = make(map[string]ParametersType)

	for _, document := range document.AsFlatList() {
		if release := document.Release(); release != "" {
			// TODO: Read parameters for deployment unit
			releases[release] = nil
		}
	}

	return
}
