package deploy

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CompileManifestTemplates compiles manifest template files in a given folder.
func CompileManifestTemplates(root string) ([]byte, error) {
	result := &bytes.Buffer{}

	if err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		if ok, _ := filepath.Match("*.yaml", info.Name()); !ok {
			return nil
		}

		f, err := os.Open(path)

		if err != nil {
			return err
		}

		defer f.Close()

		if result.Len() > 0 {
			fmt.Fprintf(result, "---\n")
		}

		if _, err := io.Copy(result, f); err != nil {
			return fmt.Errorf("reading %s: %s", path, err)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return result.Bytes(), nil
}
