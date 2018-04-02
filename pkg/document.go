package deploy

import (
	"fmt"

	yaml "gopkg.in/yaml.v2"
)

// A Document represents Kubernetes document.
type Document struct {
	raw        map[interface{}]interface{}
	apiVersion string
	kind       string
	name       string
	labels     map[interface{}]interface{}
	documents  []Document
}

// Raw returns the raw document.
func (d Document) Raw() map[interface{}]interface{} { return d.raw }

// MarshalYAML implements YAML serialization.
func (d Document) MarshalYAML() (interface{}, error) {
	return yaml.Marshal(d.raw)
}

// UnmarshalYAML implements YAML deserialization.
func (d *Document) UnmarshalYAML(unmarshal func(interface{}) error) error {
	if err := unmarshal(&d.raw); err != nil {
		return err
	}

	if err := d.init(); err != nil {
		return fmt.Errorf("verifying YAML data: %s", err)
	}

	return nil
}

func (d *Document) init() error {
	var ok bool

	if d.apiVersion, ok = d.raw["apiVersion"].(string); !ok {
		return fmt.Errorf("document has no `apiVersion`")
	}

	if d.kind, ok = d.raw["kind"].(string); !ok {
		return fmt.Errorf("document has no `kind`")
	}

	if d.IsList() {
		var items []interface{}

		if items, ok = d.raw["items"].([]interface{}); !ok {
			return fmt.Errorf("document %s has no `items`", d.Type())
		}

		for i, item := range items {
			var raw map[interface{}]interface{}

			if raw, ok = item.(map[interface{}]interface{}); !ok {
				return fmt.Errorf("sub-document %d of list document %s is not a valid document", i, d.Type())
			}

			document := Document{
				raw: raw,
			}

			if err := document.init(); err != nil {
				return fmt.Errorf("in sub-document %d of list document %s: %s", i, d.Type(), err)
			}

			d.documents = append(d.documents, document)
		}
	} else {
		var metadata map[interface{}]interface{}

		if metadata, ok = d.raw["metadata"].(map[interface{}]interface{}); !ok {
			return fmt.Errorf("document %s has no `metadata`", d.Type())
		}

		if d.name, ok = metadata["name"].(string); !ok {
			return fmt.Errorf("document %s has no `name`", d.Type())
		}

		if d.labels, ok = metadata["labels"].(map[interface{}]interface{}); !ok {
			return fmt.Errorf("document %s has no `labels`", d.Type())
		}
	}

	return nil
}

// IsList indicates whether the document is a list.
func (d Document) IsList() bool { return d.apiVersion == "v1" && d.kind == "List" }

// APIVersion returns the API version of the document.
func (d Document) APIVersion() string { return d.apiVersion }

// Kind returns the kind of the document.
func (d Document) Kind() string { return d.kind }

// Type returns the type of the document.
func (d Document) Type() string { return fmt.Sprintf("%s.%s", d.apiVersion, d.kind) }

// Name returns the name version of the document.
//
// Lists don't have names.
func (d Document) Name() string { return d.name }

// Labels returns the labels.
//
// Lists don't have labels.
func (d Document) Labels() map[interface{}]interface{} { return d.labels }

// Documents returns the sub-documents.
//
// Only lists can have sub-documents.
func (d Document) Documents() []Document { return d.documents }

// AsFlatList returns a flattened list of all documents contained in the
// current document.
func (d Document) AsFlatList() (documents []Document) {
	if d.IsList() {
		for _, document := range d.Documents() {
			documents = append(documents, document.AsFlatList()...)
		}
	} else {
		documents = append(documents, d)
	}

	return
}
