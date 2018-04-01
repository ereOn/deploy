package deploy

// A Manifest represents a collection of Kubernetes documents.
type Manifest struct {
	Name      string
	Documents []Document
}
