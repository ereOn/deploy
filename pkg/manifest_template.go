package deploy

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// ManifestTemplate represents a templated Kubernetes manifest.
type ManifestTemplate struct {
	Name string
	Data []byte
}

// LoadManifestTemplate loads a manifest template from a reader.
func LoadManifestTemplate(name string, r io.Reader) (ManifestTemplate, error) {
	data, err := ioutil.ReadAll(r)

	if err != nil {
		return ManifestTemplate{}, fmt.Errorf("reading manifest template at %s: %s", name, err)
	}

	return ManifestTemplate{
		Name: name,
		Data: data,
	}, nil
}

// LoadManifestTemplatesFromRoot load all the manifest template in the specified root directory.
func LoadManifestTemplatesFromRoot(root string) (templates []ManifestTemplate, err error) {
	var manifestTemplate ManifestTemplate

	if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
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

		if manifestTemplate, err = LoadManifestTemplate(info.Name(), f); err != nil {
			return err
		}

		templates = append(templates, manifestTemplate)

		return nil
	}); err != nil {
		return nil, err
	}

	return
}
