package deploy

import "fmt"

// A Context represents a rendering context.
type Context struct {
	Release    string
	Parameters ParametersType
}

// NameSuffix returns the name suffix from the associated release.
func (c Context) NameSuffix() string {
	return fmt.Sprintf("-%s", c.Release)
}
