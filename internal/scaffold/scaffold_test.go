package scaffold

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestCreateProducesValidScaffoldYAML(t *testing.T) {
	for _, kind := range []string{"connector", "rule"} {
		t.Run(kind, func(t *testing.T) {
			directory := t.TempDir()
			path, err := Create(directory, kind, "starter")
			if err != nil {
				t.Fatal(err)
			}
			if path != filepath.Join(directory, "starter.yaml") {
				t.Fatalf("unexpected path: %s", path)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			var document map[string]any
			if err := yaml.Unmarshal(data, &document); err != nil {
				t.Fatalf("invalid YAML: %v", err)
			}
			if kind == "connector" {
				metadata, ok := document["metadata"].(map[string]any)
				if !ok || document["apiVersion"] != "nhi.connector/v1" || document["kind"] != "CapabilityManifest" || metadata["name"] != "starter" || len(document["capabilities"].([]any)) != 0 {
					t.Fatalf("unexpected connector manifest scaffold: %#v", document)
				}
			} else if document["status"] != "scaffold-only" || document["name"] != "starter" {
				t.Fatalf("unexpected rule scaffold: %#v", document)
			}
		})
	}
}

func TestCreateRejectsTraversalAndExistingFile(t *testing.T) {
	directory := t.TempDir()
	if _, err := Create(directory, "connector", "../outside"); err == nil {
		t.Fatal("path traversal name was accepted")
	}
	if _, err := Create(directory, "connector", "already-here"); err != nil {
		t.Fatal(err)
	}
	if _, err := Create(directory, "connector", "already-here"); err == nil {
		t.Fatal("existing target was accepted")
	}
	if _, err := Create(filepath.Join(directory, "missing"), "rule", "valid-name"); err == nil {
		t.Fatal("missing destination was accepted")
	}
}
