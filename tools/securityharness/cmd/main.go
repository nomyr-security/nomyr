package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"nomyr/tools/securityharness"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, output io.Writer) error {
	if len(args) == 0 {
		return errors.New("usage: securityharness verify [openapi-file] | scan [options]")
	}
	switch args[0] {
	case "verify":
		return runVerify(args[1:], output)
	case "scan":
		return runScan(args[1:], output)
	default:
		return errors.New("usage: securityharness verify [openapi-file] | scan [options]")
	}
}

func runVerify(args []string, output io.Writer) error {
	openAPIPath := "api/openapi/v1.yaml"
	if len(args) == 1 {
		openAPIPath = args[0]
	} else if len(args) > 1 {
		return errors.New("usage: securityharness verify [openapi-file]")
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
	if err := securityharness.ValidatePhase0Inventory(operations); err != nil {
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

type inputPaths []string

func (paths *inputPaths) String() string {
	return ""
}

func (paths *inputPaths) Set(value string) error {
	*paths = append(*paths, value)
	return nil
}

func runScan(args []string, output io.Writer) error {
	flags := flag.NewFlagSet("scan", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	var canaryPath string
	var files, artifacts, logs, traces, databaseDumps inputPaths
	flags.StringVar(&canaryPath, "canary-file", "", "private file containing the runtime canary")
	flags.Var(&files, "file", "file input to scan; repeatable")
	flags.Var(&artifacts, "artifact", "artifact input to scan; repeatable")
	flags.Var(&logs, "log", "log input to scan; repeatable")
	flags.Var(&traces, "trace", "trace input to scan; repeatable")
	flags.Var(&databaseDumps, "database-dump", "database dump input to scan; repeatable")
	if flags.Parse(args) != nil || flags.NArg() != 0 {
		return errors.New("invalid scan arguments")
	}
	if canaryPath == "" {
		return errors.New("canary file is required")
	}
	info, err := os.Stat(canaryPath)
	if err != nil || info.Mode().Perm() != 0o600 {
		return errors.New("canary file must have mode 0600")
	}
	canary, err := os.ReadFile(canaryPath)
	if err != nil {
		return errors.New("canary file cannot be read")
	}
	canary = bytes.TrimSpace(canary)
	if len(canary) == 0 {
		return errors.New("canary file is empty")
	}
	sources := make([]securityharness.Source, 0, len(files)+len(artifacts)+len(logs)+len(traces)+len(databaseDumps))
	appendSources := func(kind securityharness.SourceKind, paths []string) {
		for _, path := range paths {
			sources = append(sources, fileSource(kind, path))
		}
	}
	appendSources(securityharness.SourceFile, files)
	appendSources(securityharness.SourceArtifact, artifacts)
	appendSources(securityharness.SourceLog, logs)
	appendSources(securityharness.SourceTrace, traces)
	appendSources(securityharness.SourceDatabaseDump, databaseDumps)
	if len(sources) == 0 {
		return errors.New("at least one input is required")
	}
	if err := securityharness.CheckSources(canary, sources); err != nil {
		return err
	}
	return json.NewEncoder(output).Encode(map[string]any{"status": "clean", "sources_scanned": len(sources)})
}

func verifyCanary(canary []byte) (result error) {
	fixtureDirectory, err := os.MkdirTemp("", "nomyr-securityharness-")
	if err != nil {
		return errors.New("cannot create temporary harness fixtures")
	}
	defer func() {
		if err := os.RemoveAll(fixtureDirectory); err != nil && result == nil {
			result = errors.New("cannot remove temporary harness fixtures")
		}
	}()
	canaryPath := filepath.Join(fixtureDirectory, "canary")
	if err := os.WriteFile(canaryPath, canary, 0o600); err != nil {
		return errors.New("cannot write private runtime canary")
	}
	canaryInfo, err := os.Stat(canaryPath)
	if err != nil || canaryInfo.Mode().Perm() != 0o600 {
		return errors.New("private runtime canary mode is invalid")
	}
	safeKinds := []securityharness.SourceKind{securityharness.SourceFile, securityharness.SourceArtifact, securityharness.SourceLog, securityharness.SourceTrace, securityharness.SourceDatabaseDump}
	sources := make([]securityharness.Source, 0, len(safeKinds))
	for index, kind := range safeKinds {
		inputPath := filepath.Join(fixtureDirectory, fmt.Sprintf("input-%d", index))
		if err := os.WriteFile(inputPath, []byte("synthetic safe content"), 0o600); err != nil {
			return errors.New("cannot write synthetic safe input")
		}
		sources = append(sources, fileSource(kind, inputPath))
	}
	if err := securityharness.CheckSources(canary, sources); err != nil {
		return errors.New("safe synthetic canary scan failed")
	}
	leakPath := filepath.Join(fixtureDirectory, "leaking-log")
	if err := os.WriteFile(leakPath, append([]byte("synthetic="), canary...), 0o600); err != nil {
		return errors.New("cannot write deliberate leak fixture")
	}
	leakSource := []securityharness.Source{fileSource(securityharness.SourceLog, leakPath)}
	if err := securityharness.CheckSources(canary, leakSource); err == nil || bytes.Contains([]byte(err.Error()), canary) {
		return errors.New("deliberate canary leak did not fail with a sanitized diagnostic")
	}
	return nil
}

func verifyTenantIsolation() error {
	scenario := securityharness.TenantScenario{
		Operation:      securityharness.OperationTarget{ID: "listRecords", Method: http.MethodGet, Path: "/synthetic/records"},
		Principal:      securityharness.TenantPrincipal{TenantID: "synthetic-requester"},
		ForeignTenant:  "synthetic-foreign",
		ExpectedDenial: securityharness.ExpectedDenial{Status: http.StatusNotFound, ContentType: "application/problem+json", Problem: true},
	}
	safe := http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.Header().Set("Content-Type", "application/problem+json")
		response.WriteHeader(http.StatusNotFound)
		_, _ = io.WriteString(response, `{"type":"about:blank","title":"Not found","status":404,"code":"NHI-CORE-0001"}`)
	})
	if err := securityharness.AssertTenantIsolation(safe, scenario); err != nil {
		return errors.New("safe synthetic tenant fixture failed")
	}
	unsafe := http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(response, `{"tenant_id":"synthetic-foreign","id":"synthetic-record"}`)
	})
	if err := securityharness.AssertTenantIsolation(unsafe, scenario); err == nil || err.Error() != securityharness.ErrTenantIsolation {
		return errors.New("unsafe synthetic tenant fixture was not detected")
	}
	return nil
}

func fileSource(kind securityharness.SourceKind, path string) securityharness.Source {
	return securityharness.Source{Kind: kind, Open: func() (io.ReadCloser, error) {
		return os.Open(path)
	}}
}
