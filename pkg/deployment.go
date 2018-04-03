package deploy

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

// Deployment represents a deployment.
type Deployment struct {
	Units []DeploymentUnit `yaml:"units"`
}

// LoadDeploymentFromPaths load a deployment from a list of paths.
func LoadDeploymentFromPaths(paths []string) (deployment Deployment, err error) {
	pathsSet := make(map[string]bool)

	for _, path := range paths {
		pathsSet[filepath.Clean(path)] = true
	}

	var deploymentUnit DeploymentUnit

	for path := range pathsSet {
		if deploymentUnit, err = LoadDeploymentUnitFromPath(path); err != nil {
			return
		}

		for _, existingDeploymentUnit := range deployment.Units {
			if existingDeploymentUnit.Name == deploymentUnit.Name {
				err = fmt.Errorf("loading deployment unit at `%s`: another deployment unit with the name `%s` was already loaded", path, deploymentUnit.Name)
				return
			}
		}

		deployment.Units = append(deployment.Units, deploymentUnit)
	}

	return
}

// NewContext creates a new context from the deployment default parameters.
func (d Deployment) NewContext(release string, namespace string) Context {
	ctx := NewContext(release, namespace)

	for _, unit := range d.Units {
		unitParameters := make(ParametersType)

		for key, value := range unit.Parameters {
			unitParameters[key] = value
		}

		ctx.Parameters[unit.Name] = unitParameters
	}

	return ctx
}

// RenderManifests produces the manifests for all deployment units in the
// deployment.
func (d Deployment) RenderManifests(ctx Context) (manifests []Manifest, err error) {
	var unitManifests []Manifest

	for _, unit := range d.Units {
		if unitManifests, err = unit.RenderManifests(ctx.For(unit.Name)); err != nil {
			err = fmt.Errorf("rendering deployment: %s", err)
			return
		}

		manifests = append(manifests, unitManifests...)
	}

	return
}

// Render the deployment.
func (d Deployment) Render(ctx Context) ([]byte, error) {
	manifests, err := d.RenderManifests(ctx)

	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	encoder := yaml.NewEncoder(buf)

	for _, manifest := range manifests {
		for _, document := range manifest.Documents {
			encoder.Encode(document)
		}
	}

	return buf.Bytes(), nil
}

// DeploymentUnitAttributes contains the attributes of a deployment.
type DeploymentUnitAttributes struct {
	Name       string         `yaml:"name"`
	Parameters ParametersType `yaml:"parameters,omitempty"`
}

// A DeploymentUnit represents a complete set of manifest templates and their
// default values.
type DeploymentUnit struct {
	DeploymentUnitAttributes `yaml:",inline"`
	Templates                []Template `yaml:"manifest-templates"`
}

// LoadDeploymentUnitFromPath load all the manifest template in the specified
// path directory.
func LoadDeploymentUnitFromPath(path string) (deploymentUnit DeploymentUnit, err error) {
	deploymentUnit.DeploymentUnitAttributes.Parameters = make(ParametersType)

	var manifestTemplate Template

	if err = filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && path != p {
			return filepath.SkipDir
		}

		if ok, _ := filepath.Match("*.yaml", info.Name()); !ok {
			return nil
		}

		f, err := os.Open(p)

		if err != nil {
			return err
		}

		defer f.Close()

		if info.Name() == deploymentUnitAttributesFilename {
			decoder := yaml.NewDecoder(f)
			decoder.SetStrict(true)

			if err = decoder.Decode(&deploymentUnit.DeploymentUnitAttributes); err != nil {
				return fmt.Errorf("parsing deployment attributes file at `%s`: %s", info.Name(), err)
			}
		} else {
			if manifestTemplate, err = LoadTemplate(info.Name(), f); err != nil {
				return err
			}

			deploymentUnit.Templates = append(deploymentUnit.Templates, manifestTemplate)
		}

		return nil
	}); err != nil {
		return
	}

	if deploymentUnit.Name == "" {
		err = fmt.Errorf("loading deployment unit at `%s`: no name was defined", path)
	}

	return
}

// RenderManifests produces the manifests for the deployment unit.
func (u DeploymentUnit) RenderManifests(ctx Context) (manifests []Manifest, err error) {
	var manifest Manifest

	for _, template := range u.Templates {
		if manifest, err = template.RenderManifest(ctx); err != nil {
			err = fmt.Errorf("rendering deployment unit `%s`: %s", u.Name, err)
			return
		}

		manifests = append(manifests, manifest)
	}

	return
}
