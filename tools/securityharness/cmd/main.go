package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"nomyr/tools/securityharness"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, output io.Writer) error {
	if len(args) < 1 || args[0] != "verify" {
		return errors.New("usage: go run ./tools/securityharness/cmd verify [openapi-file]")
	}
	openAPIPath := "api/openapi/v1.yaml"
	if len(args) == 2 {
		openAPIPath = args[1]
	} else if len(args) > 2 {
		return errors.New("usage: go run ./tools/securityharness/cmd verify [openapi-file]")
	}
	data, err := os.ReadFile(openAPIPath)
	if err != nil {
		return fmt.Errorf("cannot read OpenAPI contract")
	}
	document, err := securityharness.LoadOpenAPI(data)
	if err != nil {
		return err
	}
	operations, err := securityharness.EnumerateOperations(document)
	if err != nil {
		return err
	}
	canary, err := securityharness.NewCanary()
	if err != nil {
		return err
	}
	if err := verifyCanary(canary); err != nil {
		return err
	}
	if err := verifyTenantIsolation(); err != nil {
		return err
	}
	return json.NewEncoder(output).Encode(struct {
		Operations securityharness.OperationReport `json:"operations"`
		Assertions map[string]string               `json:"assertions"`
	}{Operations: operations, Assertions: map[string]string{
		"safe_canary_inputs":          "pass",
		"deliberate_canary_leak":      "detected_sanitized",
		"safe_tenant_fixture":         "pass",
		"unsafe_tenant_fixture":       "detected_sanitized",
		"live_database_dump_coverage": "not_exercised",
	}})
}

func verifyCanary(canary []byte) error {
	safeKinds := []securityharness.SourceKind{securityharness.SourceFile, securityharness.SourceArtifact, securityharness.SourceLog, securityharness.SourceTrace, securityharness.SourceDatabaseDump}
	sources := make([]securityharness.Source, 0, len(safeKinds))
	for _, kind := range safeKinds {
		sources = append(sources, bytesSource(kind, []byte("synthetic safe content")))
	}
	findings, err := securityharness.ScanSources(canary, sources)
	if err != nil || len(findings) != 0 {
		return errors.New("safe synthetic canary scan failed")
	}
	leakSource := []securityharness.Source{bytesSource(securityharness.SourceLog, append([]byte("synthetic="), canary...))}
	findings, err = securityharness.ScanSources(canary, leakSource)
	if err != nil {
		return errors.New("synthetic leak scan failed")
	}
	if len(findings) != 1 || findings[0].Code != "CANARY_LEAK" || findings[0].Count != 1 {
		return errors.New("synthetic canary leak was not detected")
	}
	encoded, _ := json.Marshal(findings)
	if bytes.Contains(encoded, canary) {
		return errors.New("canary diagnostic was not sanitized")
	}
	if err := securityharness.CheckSources(canary, leakSource); err == nil || bytes.Contains([]byte(err.Error()), canary) {
		return errors.New("deliberate canary leak did not fail with a sanitized diagnostic")
	}
	return nil
}

func verifyTenantIsolation() error {
	safe := http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		http.Error(response, `{"status":404,"code":"NOT_FOUND"}`, http.StatusNotFound)
	})
	principal := securityharness.TenantPrincipal{TenantID: "synthetic-requester"}
	if err := securityharness.AssertTenantIsolation(safe, principal, "synthetic-foreign"); err != nil {
		return errors.New("safe synthetic tenant fixture failed")
	}
	unsafe := http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(response, `{"tenant_id":"synthetic-foreign","id":"synthetic-record"}`)
	})
	if err := securityharness.AssertTenantIsolation(unsafe, principal, "synthetic-foreign"); err == nil || err.Error() != securityharness.ErrTenantIsolation {
		return errors.New("unsafe synthetic tenant fixture was not detected")
	}
	return nil
}

func bytesSource(kind securityharness.SourceKind, data []byte) securityharness.Source {
	return securityharness.Source{Kind: kind, Open: func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(data)), nil
	}}
}
