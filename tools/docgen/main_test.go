package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestPublicSourceAllowlistExcludesPrivateAndPlanningMaterial(t *testing.T) {
	allowed := map[string]bool{}
	for _, path := range publicInputs {
		allowed[path] = true
		for _, forbiddenRoot := range []string{"docs/", "openwiki/", "deploy/telemetry-worker/"} {
			if strings.HasPrefix(path, forbiddenRoot) {
				t.Fatalf("private or planning tree entered public allowlist: %s", path)
			}
		}
	}
	for _, path := range []string{"docs/spec/", "docs/plan/", "docs/phases/", "docs/adr/", "openwiki/", "deploy/telemetry-worker/", ".env"} {
		if allowed[path] {
			t.Fatalf("private or planning source entered public allowlist: %s", path)
		}
	}
	if len(allowed) != len(publicInputs) {
		t.Fatal("public source allowlist contains duplicate paths")
	}
}

func TestGeneratedPagesAreDeterministic(t *testing.T) {
	first, err := generate("../..")
	if err != nil {
		t.Fatal(err)
	}
	second, err := generate("../..")
	if err != nil {
		t.Fatal(err)
	}
	for path, content := range first {
		if !bytes.Equal(content, second[path]) {
			t.Errorf("generation differs for %s", path)
		}
	}
}

func TestOutputAllowlistStaysInsideDocsSite(t *testing.T) {
	for _, path := range outputFiles {
		if len(path) < len("docs-site/docs/") || path[:len("docs-site/docs/")] != "docs-site/docs/" {
			t.Errorf("generated path escapes public docs site: %s", path)
		}
	}
	for _, unsafe := range []string{"docs/spec/generated.md", "openwiki/index.md", "deploy/telemetry-worker/config.md"} {
		for _, path := range outputFiles {
			if path == unsafe {
				t.Errorf("unsafe generated destination allowed: %s", unsafe)
			}
		}
	}
}

func TestReadPublicRejectsPrivatePlanningAndTelemetrySources(t *testing.T) {
	for _, path := range []string{"docs/spec/product.md", "docs/plan/internal.md", "openwiki/index.md", "deploy/telemetry-worker/config.env", ".env"} {
		if _, err := readPublic(".", path); err == nil {
			t.Errorf("private or unapproved source was readable: %s", path)
		}
	}
	if _, err := readPublic(".", "api/openapi/v1.yaml"); err != nil && !os.IsNotExist(err) {
		t.Fatalf("allowlisted public source rejected unexpectedly: %v", err)
	}
}
