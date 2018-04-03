package deploy

import (
	"encoding/json"
	"fmt"
)

// A Document represents Kubernetes document.
type Document struct {
	Context     Context
	raw         map[interface{}]interface{}
	apiVersion  string
	kind        string
	name        string
	labels      map[interface{}]interface{}
	annotations map[interface{}]interface{}
	documents   []Document
}

// Raw returns the raw document.
func (d Document) Raw() map[interface{}]interface{} { return d.raw }

// MarshalYAML implements YAML serialization.
func (d Document) MarshalYAML() (interface{}, error) {
	return d.raw, nil
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
				Context: d.Context,
				raw:     raw,
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

		var labels interface{}

		if labels, ok = metadata["labels"]; ok {
			if d.labels, ok = labels.(map[interface{}]interface{}); !ok {
				return fmt.Errorf("document %s `labels` has an unexpected format", d.Type())
			}
		} else {
			d.labels = make(map[interface{}]interface{})
		}

		var annotations interface{}

		if annotations, ok = metadata["annotations"]; ok {
			if d.annotations, ok = annotations.(map[interface{}]interface{}); !ok {
				return fmt.Errorf("document %s `annotations` has an unexpected format", d.Type())
			}
		} else {
			d.annotations = make(map[interface{}]interface{})
		}
	}

	return nil
}

func (d *Document) initForRendering() error {
	// This is guaranteed to work if init() succeeded.
	metadata := d.raw["metadata"].(map[interface{}]interface{})

	if value, ok := d.labels[releaseLabel]; ok {
		return fmt.Errorf("document %s already has a `%s` label with value `%v` which is not allowed", d.Type(), releaseLabel, value)
	}

	if value, ok := d.annotations[releaseParametersAnnotation]; ok {
		return fmt.Errorf("document %s already has a `%s` annotation with value `%v` which is not allowed", d.Type(), releaseParametersAnnotation, value)
	}

	d.name = fmt.Sprintf("%s%s", d.name, d.Context.NameSuffix())
	metadata["name"] = d.name

	// Add the reserved label. This step is mandatory.
	d.labels[releaseLabel] = d.Context.Release

	// As a debugging facility, mark the document with the set of deployment parameters.
	releaseParameters, _ := json.Marshal(d.Context.Parameters)
	d.annotations[releaseParametersAnnotation] = string(releaseParameters)
	metadata["annotations"] = d.annotations

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

// Annotations returns the annotations.
//
// Lists don't have annotations.
func (d Document) Annotations() map[interface{}]interface{} { return d.annotations }

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

// Release returns the document release.
func (d Document) Release() string {
	value, _ := d.labels[releaseLabel].(string)

	return value
}
