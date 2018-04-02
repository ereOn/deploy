package deploy

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"text/template"

	"gopkg.in/yaml.v2"
)

// ManifestTemplate represents a templated Kubernetes manifest.
type ManifestTemplate struct {
	Name string `yaml:"name"`
	Data string `yaml:"data"`
}

// LoadManifestTemplate loads a manifest template from a reader.
func LoadManifestTemplate(name string, r io.Reader) (ManifestTemplate, error) {
	data, err := ioutil.ReadAll(r)

	if err != nil {
		return ManifestTemplate{}, fmt.Errorf("reading manifest template at %s: %s", name, err)
	}

	return ManifestTemplate{
		Name: name,
		Data: string(data),
	}, nil
}

// Render the manifest template.
func (m ManifestTemplate) Render(ctx Context) (manifest Manifest, err error) {
	var tmpl *template.Template
	tmpl, err = template.New(m.Name).Parse(m.Data)

	if err != nil {
		return
	}

	buf := &bytes.Buffer{}

	if err = tmpl.Execute(buf, ctx); err != nil {
		return
	}

	document := Document{Context: ctx}
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

		manifest.Documents = append(manifest.Documents, document.AsFlatList()...)
	}

	manifest.Name = m.Name

	return
}
