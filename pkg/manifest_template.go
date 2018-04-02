package deploy

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"gopkg.in/yaml.v2"
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

// Render the manifest template.
func (m ManifestTemplate) Render(ctx Context) (manifest Manifest, err error) {
	var tmpl *template.Template
	tmpl, err = template.New(m.Name).Parse(string(m.Data))

	if err != nil {
		return
	}

	buf := &bytes.Buffer{}

	if err = tmpl.Execute(buf, ctx); err != nil {
		return
	}

	var document Document

	decoder := yaml.NewDecoder(buf)

	for {
		if err = decoder.Decode(&document); err != nil {
			if err == io.EOF {
				err = nil
				break
			}

			err = fmt.Errorf("checking rendered output from `%s`: %s", m.Name, err)

			return
		}

		manifest.Documents = append(manifest.Documents, document)
	}

	manifest.Name = m.Name

	return
}
