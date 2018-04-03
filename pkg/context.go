package deploy

import (
	"fmt"
	"strings"

	"github.com/tjarratt/babble"
)

// ParametersType is the type for parameters.
type ParametersType map[string]interface{}

// A Context represents a rendering context.
type Context struct {
	Release        string
	Namespace      string
	DeploymentUnit string
	Parameters     ParametersType
}

// NewContext creates a new context.
func NewContext(release string, namespace string) Context {
	if release == "" {
		babbler := babble.NewBabbler()
		babbler.Count = 2
		babbler.Separator = "-"
		release = strings.ToLower(babbler.Babble())
	}

	if namespace == "" {
		namespace = "default"
	}

	parameters := make(ParametersType)

	return Context{
		Release:    release,
		Namespace:  namespace,
		Parameters: parameters,
	}
}

// For returns a context specialized for the specified unit.
func (c Context) For(deploymentUnit string) Context {
	return Context{
		Release:        c.Release,
		Namespace:      c.Namespace,
		DeploymentUnit: deploymentUnit,
		Parameters:     c.Parameters[deploymentUnit].(ParametersType),
	}
}

// NameSuffix returns the name suffix from the associated release.
func (c Context) NameSuffix() string {
	return fmt.Sprintf("-%s", c.Release)
}
