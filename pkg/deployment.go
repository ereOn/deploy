package deploy

import (
	"fmt"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

// A Deployment represents a complete set of manifest templates and their
// default values.
type Deployment struct {
	DefaultParameters ParametersType     `yaml:"default-parameters"`
	ManifestTemplates []ManifestTemplate `yaml:"manifest-templates"`
}

// LoadDeploymentFromRoot load all the manifest template in the specified root directory.
func LoadDeploymentFromRoot(root string) (deployment Deployment, err error) {
	deployment.DefaultParameters = make(ParametersType)

	var manifestTemplate ManifestTemplate

	if err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && root != path {
			return filepath.SkipDir
		}

		if ok, _ := filepath.Match("*.yaml", info.Name()); !ok {
			return nil
		}

		f, err := os.Open(path)

		if err != nil {
			return err
		}

		defer f.Close()

		if info.Name() == deploymentParametersFilename {
			if err = yaml.NewDecoder(f).Decode(&deployment.DefaultParameters); err != nil {
				return fmt.Errorf("parsing parameters file at `%s`: %s", info.Name(), err)
			}
		} else {
			if manifestTemplate, err = LoadManifestTemplate(info.Name(), f); err != nil {
				return err
			}

			deployment.ManifestTemplates = append(deployment.ManifestTemplates, manifestTemplate)
		}

		return nil
	}); err != nil {
		return
	}

	return
}
