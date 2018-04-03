package deploy

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

// Template represents a templated Kubernetes manifest.
type Template struct {
	Name string `yaml:"name"`
	Data string `yaml:"data"`
}

// LoadTemplate loads a manifest template from a reader.
func LoadTemplate(name string, r io.Reader) (Template, error) {
	data, err := ioutil.ReadAll(r)

	if err != nil {
		return Template{}, fmt.Errorf("reading manifest template at %s: %s", name, err)
	}

	return Template{
		Name: name,
		Data: string(data),
	}, nil
}

// RenderManifest generates a manifest from the template.
func (m Template) RenderManifest(ctx Context) (manifest Manifest, err error) {
	var tmpl *template.Template
	tmpl, err = template.New(m.Name).Parse(m.Data)

	if err != nil {
		err = fmt.Errorf("instanciating template from `%s`: %s", m.Name, err)
		return
	}

	buf := &bytes.Buffer{}

	if err = tmpl.Execute(buf, ctx); err != nil {
		err = fmt.Errorf("rendering `%s`: %s", m.Name, err)
		return
	}

	document := &Document{Context: ctx}
	decoder := yaml.NewDecoder(buf)

	for {
		if err = decoder.Decode(document); err != nil {
			if err == io.EOF {
				err = nil
				break
			}

			err = fmt.Errorf("decoding rendered output from `%s`: %s", m.Name, err)

			return
		}

		if err = document.initForRendering(); err != nil {
			err = fmt.Errorf("checking rendered output from `%s`: %s", m.Name, err)
		}

		manifest.Documents = append(manifest.Documents, document.AsFlatList()...)
	}

	manifest.Name = m.Name

	return
}

// A Manifest represents a collection of Kubernetes documents.
type Manifest struct {
	Name      string
	Documents []Document
}
