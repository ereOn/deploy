package deploy

import yaml "gopkg.in/yaml.v2"

// Metadata represents a Kubernetes document metadata.
type Metadata struct {
	Name   string            `yaml:"name"`
	Labels map[string]string `yaml:"labels"`
}

// DocumentBase represents the document base.
type DocumentBase struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Metadata   `yaml:"metadata"`
}

// A Document represents Kubernetes document.
type Document struct {
	DocumentBase
	Raw interface{} `yaml:"-"`
}

// MarshalYAML implements YAML serialization.
func (d Document) MarshalYAML() (interface{}, error) {
	return yaml.Marshal(d.Raw)
}

// UnmarshalYAML implements YAML deserialization.
func (d *Document) UnmarshalYAML(unmarshal func(interface{}) error) (err error) {
	if err = unmarshal(&d.DocumentBase); err != nil {
		return err
	}

	return unmarshal(&d.Raw)
}
