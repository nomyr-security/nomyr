package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"nomyr/tools/securityharness"
)

func TestCLIProcessBoundaries(t *testing.T) {
	temporaryDirectory := t.TempDir()
	t.Cleanup(func() {
		if err := os.RemoveAll(temporaryDirectory); err != nil {
			t.Error("canary-bearing test fixtures could not be removed")
		}
	})
	binaryPath := filepath.Join(temporaryDirectory, "securityharness")
	build := exec.Command("go", "build", "-o", binaryPath, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("could not build harness CLI: %s", output)
	}

	canary, err := securityharness.NewCanary()
	if err != nil {
		t.Fatal("could not generate test runtime canary")
	}
	canaryPath := filepath.Join(temporaryDirectory, "runtime-canary")
	if err := os.WriteFile(canaryPath, canary, 0o600); err != nil {
		t.Fatal("could not write private runtime canary")
	}
	canaryInfo, err := os.Stat(canaryPath)
	if err != nil || canaryInfo.Mode().Perm() != 0o600 {
		t.Fatal("runtime canary file is not mode 0600")
	}

	arguments := []string{"scan", "--canary-file", canaryPath}
	inputKinds := []struct {
		name string
		flag string
	}{
		{name: "file", flag: "--file"},
		{name: "artifact", flag: "--artifact"},
		{name: "log", flag: "--log"},
		{name: "trace", flag: "--trace"},
		{name: "database-dump", flag: "--database-dump"},
	}
	inputPaths := make(map[string]string, len(inputKinds))
	for _, kind := range inputKinds {
		inputPath := filepath.Join(temporaryDirectory, kind.name)
		if err := os.WriteFile(inputPath, []byte("synthetic safe content"), 0o600); err != nil {
			t.Fatal("could not create synthetic scan input")
		}
		inputPaths[kind.name] = inputPath
		arguments = append(arguments, kind.flag, inputPath)
	}
	stdout, stderr, err := runCLI(binaryPath, arguments...)
	if err != nil || len(stderr) != 0 || bytes.Contains(stdout, canary) {
		t.Fatal("safe CLI scan did not complete with sanitized output")
	}
	var safeReport struct {
		Status         string `json:"status"`
		SourcesScanned int    `json:"sources_scanned"`
	}
	if json.Unmarshal(stdout, &safeReport) != nil || safeReport.Status != "clean" || safeReport.SourcesScanned != len(inputKinds) {
		t.Fatal("safe CLI scan did not report all real file inputs")
	}

	leakingLog := inputPaths["log"]
	if err := os.WriteFile(leakingLog, append([]byte("synthetic="), canary...), 0o600); err != nil {
		t.Fatal("could not create deliberate real-file leak")
	}
	stdout, stderr, err = runCLI(binaryPath, arguments...)
	var exitError *exec.ExitError
	if !errors.As(err, &exitError) || exitError.ExitCode() != 1 {
		t.Fatal("deliberate canary leak did not return CLI exit status 1")
	}
	if len(stdout) != 0 || string(stderr) != "canary leak detected\n" || bytes.Contains(stderr, canary) {
		t.Fatal("leaking CLI scan did not keep stdout and stderr sanitized")
	}

	unknownContractPath := filepath.Join(temporaryDirectory, "drift.yaml")
	unknownContract := []byte(`openapi: 3.1.0
info: {title: test, version: "1"}
security: [{bearerAuth: []}]
components:
  securitySchemes:
    bearerAuth: {type: http, scheme: bearer}
paths:
  /healthz:
    get:
      operationId: getHealth
      security: []
      responses: {"200": {description: ok}}
  /capabilities:
    get:
      operationId: listCapabilities
      responses: {"200": {description: ok}}
  /tenants/{tenant_id}/records:
    get:
      operationId: listTenantRecords
      responses: {"200": {description: ok}}
`)
	if err := os.WriteFile(unknownContractPath, unknownContract, 0o600); err != nil {
		t.Fatal("could not write contract drift fixture")
	}
	stdout, stderr, err = runCLI(binaryPath, "verify", unknownContractPath)
	if !errors.As(err, &exitError) || exitError.ExitCode() != 1 || len(stdout) != 0 || string(stderr) != "OpenAPI operation inventory requires review\n" || bytes.Contains(stderr, canary) {
		t.Fatal("unknown tenant operation did not fail closed with sanitized output")
	}
	if strings.Contains(string(stdout)+string(stderr), string(canary)) {
		t.Fatal("CLI output disclosed the runtime canary")
	}
}

func runCLI(binaryPath string, arguments ...string) ([]byte, []byte, error) {
	command := exec.Command(binaryPath, arguments...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}
