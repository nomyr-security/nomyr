package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestLoopbackAddressGuard(t *testing.T) {
	testCases := []struct {
		address string
		allowed bool
	}{
		{address: "127.0.0.1:8080", allowed: true},
		{address: "localhost:8080", allowed: true},
		{address: "[::1]:8080", allowed: true},
		{address: "0.0.0.0:8080", allowed: false},
		{address: "example.com:8080", allowed: false},
		{address: "invalid", allowed: false},
	}
	for _, testCase := range testCases {
		if actual := isLoopbackAddress(testCase.address); actual != testCase.allowed {
			t.Errorf("isLoopbackAddress(%q) = %v, want %v", testCase.address, actual, testCase.allowed)
		}
	}
}

func TestGlobalOutputAndCompletion(t *testing.T) {
	for _, format := range []string{"json", "yaml"} {
		command := newRootCommand()
		output := &bytes.Buffer{}
		command.SetOut(output)
		command.SetArgs([]string{"--output", format, "context"})
		if err := command.Execute(); err != nil {
			t.Fatalf("context --output %s: %v", format, err)
		}
		if format == "json" {
			var result map[string]string
			if err := json.Unmarshal(output.Bytes(), &result); err != nil || result["context"] != "local (demo)" {
				t.Errorf("invalid JSON context output %q: %v", output.String(), err)
			}
		} else {
			var result map[string]string
			if err := yaml.Unmarshal(output.Bytes(), &result); err != nil || result["context"] != "local (demo)" {
				t.Errorf("invalid YAML context output %q: %v", output.String(), err)
			}
		}
	}
	for _, shell := range []string{"bash", "zsh", "fish", "powershell"} {
		command := newRootCommand()
		output := &bytes.Buffer{}
		command.SetOut(output)
		command.SetArgs([]string{"completion", shell})
		if err := command.Execute(); err != nil {
			t.Errorf("completion %s: %v", shell, err)
		}
		if output.Len() == 0 {
			t.Errorf("completion %s generated no output", shell)
		}
	}
	noInput := newRootCommand()
	noInput.SetOut(&bytes.Buffer{})
	noInput.SetArgs([]string{"--no-input", "context"})
	if err := noInput.Execute(); err != nil {
		t.Fatalf("--no-input context: %v", err)
	}
}

func TestScaffoldCommandRejectsExistingAndUnsafeNames(t *testing.T) {
	directory := t.TempDir()
	command := newRootCommand()
	command.SetArgs([]string{"scaffold", "connector", "new", "example-connector", "--directory", directory})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "example-connector.yaml")
	first, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(first), "apiVersion: nhi.connector/v1") || !strings.Contains(string(first), "capabilities: []") || !strings.Contains(string(first), "no provider support") {
		t.Fatalf("scaffold overstates support or is missing status: %s", first)
	}
	second := newRootCommand()
	second.SetArgs([]string{"scaffold", "connector", "new", "example-connector", "--directory", directory})
	if err := second.Execute(); err == nil {
		t.Fatal("existing file was overwritten")
	}
	unchanged, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(unchanged) != string(first) {
		t.Fatal("existing scaffold contents changed")
	}
	unsafe := newRootCommand()
	unsafe.SetArgs([]string{"scaffold", "rule", "new", "../escape", "--directory", directory})
	if err := unsafe.Execute(); err == nil {
		t.Fatal("path traversal name accepted")
	}
	if _, err := os.Stat(filepath.Join(directory, "..", "escape.yaml")); !os.IsNotExist(err) {
		t.Fatalf("unsafe path created a file: %v", err)
	}
}

func TestJSONOutputIsMachineReadable(t *testing.T) {
	command := newRootCommand()
	output := &bytes.Buffer{}
	command.SetOut(output)
	command.SetArgs([]string{"--output", "json", "version"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	var result map[string]any
	if err := json.Unmarshal(output.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := result["version"]; !ok {
		t.Fatalf("version field missing: %#v", result)
	}
}

func TestUnsupportedLoginUsesStructuredMachineOutput(t *testing.T) {
	command := newRootCommand()
	command.SetArgs([]string{"--output", "json", "login"})
	_, err := command.ExecuteC()
	if err == nil {
		t.Fatal("login unexpectedly succeeded")
	}
	output := &bytes.Buffer{}
	if writeError := renderExecutionError(output, "json", err); writeError != nil {
		t.Fatal(writeError)
	}
	var result map[string]any
	if unmarshalError := json.Unmarshal(output.Bytes(), &result); unmarshalError != nil {
		t.Fatalf("invalid structured error JSON: %v", unmarshalError)
	}
	for _, field := range []string{"code", "what", "cause", "fix", "docs"} {
		if result[field] == nil || result[field] == "" {
			t.Errorf("error field %q missing: %#v", field, result)
		}
	}
}
