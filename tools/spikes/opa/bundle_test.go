package opa

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
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

	signer, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	privateKey := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: mustMarshalPKCS8(t, signer)})
	publicKey := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: mustMarshalPublicKey(t, &signer.PublicKey)})
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
	if err := policyBundle.GenerateSignature(bundle.NewSigningConfig(string(privateKey), "RS256", ""), "spike-key", false); err != nil {
		t.Fatal(err)
	}
	var archive bytes.Buffer
	if err := bundle.NewWriter(&archive).Write(policyBundle); err != nil {
		t.Fatal(err)
	}
	verify := func(data []byte, verificationKey []byte) error {
		configuration := bundle.NewVerificationConfig(
			map[string]*bundle.KeyConfig{"spike-key": {Key: string(verificationKey), Algorithm: "RS256"}},
			"spike-key", "", nil,
		)
		_, err := bundle.NewReader(bytes.NewReader(data)).WithBundleVerificationConfig(configuration).Read()
		return err
	}
	if err := verify(archive.Bytes(), publicKey); err != nil {
		t.Fatalf("trusted signed bundle rejected: %v", err)
	}
	untrustedSigner, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	untrustedPublicKey := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: mustMarshalPublicKey(t, &untrustedSigner.PublicKey)})
	if err := verify(archive.Bytes(), untrustedPublicKey); err == nil {
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
	if err := verify(tampered.Bytes(), publicKey); err == nil {
		t.Fatal("tampered bundle was accepted")
	}
}

func mustMarshalPKCS8(t *testing.T, key *rsa.PrivateKey) []byte {
	t.Helper()
	encoded, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}

func mustMarshalPublicKey(t *testing.T, key *rsa.PublicKey) []byte {
	t.Helper()
	encoded, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return encoded
}
