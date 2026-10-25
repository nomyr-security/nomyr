package main

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
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

func TestSafeCommandEnvironmentRejectsInheritedPathAndConfigHome(t *testing.T) {
	environment := safeCommandEnvironment([]string{
		"HOME=/tmp/home",
		"PATH=/tmp/untrusted-bin",
		"NOMYR_CONFIG_HOME=/tmp/old-config",
		"GOPROXY=https://proxy.golang.org,direct",
	}, "/tmp/disabled-telemetry-config")
	pathCount := 0
	configCount := 0
	for _, variable := range environment {
		if strings.HasPrefix(variable, "PATH=") {
			pathCount++
			if variable != "PATH=/usr/bin:/bin" {
				t.Errorf("unexpected child PATH: %s", variable)
			}
		}
		if strings.HasPrefix(variable, "NOMYR_CONFIG_HOME=") {
			configCount++
			if variable != "NOMYR_CONFIG_HOME=/tmp/disabled-telemetry-config" {
				t.Errorf("unexpected child config home: %s", variable)
			}
		}
	}
	if pathCount != 1 || configCount != 1 {
		t.Fatalf("expected one constrained PATH and telemetry config home, got PATH=%d config=%d", pathCount, configCount)
	}
}

func TestGoExecutableUsesInstalledToolchainPath(t *testing.T) {
	path := goExecutable()
	if !filepath.IsAbs(path) {
		t.Fatalf("Go executable path is not absolute: %s", path)
	}
	if filepath.Dir(filepath.Dir(path)) != runtime.GOROOT() {
		t.Fatalf("Go executable does not resolve within GOROOT: %s", path)
	}
	if info, err := os.Stat(path); err != nil || !info.Mode().IsRegular() {
		t.Fatalf("Go executable is not present at the toolchain path: %s, %v", path, err)
	}
}
