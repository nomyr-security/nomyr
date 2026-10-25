package opa

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/bundle"
	"github.com/open-policy-agent/opa/rego"
)

func TestEmbeddedOPAAndSignedBundleVerification(t *testing.T) {
	query, err := rego.New(
		rego.Query("data.spike.allow"),
		rego.Module("spike.rego", "package spike\nimport rego.v1\nallow if input.enabled\n"),
	).PrepareForEval(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	result, err := query.Eval(t.Context(), rego.EvalInput(map[string]any{"enabled": true}))
	if err != nil || len(result) != 1 || result[0].Expressions[0].Value != true {
		t.Fatalf("OPA evaluation result=%v error=%v", result, err)
	}

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}
	module := "package spike\nimport rego.v1\nallow if input.enabled\n"
	parsed, err := ast.ParseModule("spike.rego", module)
	if err != nil {
		t.Fatal(err)
	}
	policyBundle := bundle.Bundle{
		Manifest: bundle.Manifest{Revision: "synthetic-revision"},
		Data:     map[string]any{},
		Modules:  []bundle.ModuleFile{{URL: "spike.rego", Path: "spike.rego", Raw: []byte(module), Parsed: parsed}},
	}
	keyText := base64.RawStdEncoding.EncodeToString(key)
	if err := policyBundle.GenerateSignature(bundle.NewSigningConfig(keyText, "HS256", ""), "spike-key", false); err != nil {
		t.Fatal(err)
	}
	var archive bytes.Buffer
	if err := bundle.NewWriter(&archive).Write(policyBundle); err != nil {
		t.Fatal(err)
	}
	verify := func(data []byte, verificationKey string) error {
		configuration := bundle.NewVerificationConfig(
			map[string]*bundle.KeyConfig{"spike-key": {Key: verificationKey, Algorithm: "HS256"}},
			"spike-key", "", nil,
		)
		_, err := bundle.NewReader(bytes.NewReader(data)).WithBundleVerificationConfig(configuration).Read()
		return err
	}
	if err := verify(archive.Bytes(), keyText); err != nil {
		t.Fatalf("trusted signed bundle rejected: %v", err)
	}
	if err := verify(archive.Bytes(), strings.Repeat("untrusted-key", 4)); err == nil {
		t.Fatal("bundle signed by an untrusted key was accepted")
	}

	tamperedBundle := policyBundle
	tamperedBundle.Modules = append([]bundle.ModuleFile(nil), policyBundle.Modules...)
	tamperedBundle.Modules[0].Raw = bytes.Replace(
		append([]byte(nil), policyBundle.Modules[0].Raw...),
		[]byte("input.enabled"),
		[]byte("input.disabled"),
		1,
	)
	var tampered bytes.Buffer
	if err := bundle.NewWriter(&tampered).Write(tamperedBundle); err != nil {
		t.Fatal(err)
	}
	if err := verify(tampered.Bytes(), keyText); err == nil {
		t.Fatal("tampered bundle was accepted")
	}
}
