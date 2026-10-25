package securityharness

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestCanaryScannerFindsLeaksWithoutReturningCanary(t *testing.T) {
	canary, err := NewCanary()
	if err != nil {
		t.Fatal("could not generate runtime canary")
	}
	kinds := []SourceKind{SourceFile, SourceArtifact, SourceLog, SourceTrace, SourceDatabaseDump}
	for _, kind := range kinds {
		t.Run(string(kind), func(t *testing.T) {
			data := append([]byte("synthetic="), canary...)
			findings, err := ScanSources(canary, []Source{{Kind: kind, Open: func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(data)), nil }}})
			if err != nil || len(findings) != 1 || findings[0].Code != "CANARY_LEAK" {
				t.Fatal("expected one sanitized canary finding")
			}
			if strings.Contains(strings.Join([]string{findings[0].Code, string(findings[0].Kind)}, " "), string(canary)) {
				t.Fatal("finding disclosed runtime canary")
			}
		})
	}
	leak := []byte("synthetic=" + string(canary))
	err = CheckSources(canary, []Source{{Kind: SourceLog, Open: func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(leak)), nil }}})
	if err == nil || strings.Contains(err.Error(), string(canary)) {
		t.Fatal("deliberate leak did not fail with a sanitized diagnostic")
	}
}

func TestCanaryScannerSanitizesSourceErrors(t *testing.T) {
	canary, err := NewCanary()
	if err != nil {
		t.Fatal("could not generate runtime canary")
	}
	_, err = ScanSources(canary, []Source{{Kind: SourceLog, Open: func() (io.ReadCloser, error) { return nil, errors.New(string(canary)) }}})
	if err == nil || strings.Contains(err.Error(), string(canary)) {
		t.Fatal("source failure did not return a sanitized diagnostic")
	}
	_, err = ScanSources(canary, []Source{{Kind: SourceKind(string(canary))}})
	if err == nil || strings.Contains(err.Error(), string(canary)) {
		t.Fatal("untrusted source metadata appeared in diagnostic")
	}
}

func TestCanaryScannerFindsValueAcrossReadBoundary(t *testing.T) {
	canary, err := NewCanary()
	if err != nil {
		t.Fatal("could not generate runtime canary")
	}
	boundary := canaryReadChunk + len(canary) - 1
	data := append(bytes.Repeat([]byte("x"), boundary-len(canary)/2), canary...)
	findings, err := ScanSources(canary, []Source{{Kind: SourceFile, Open: func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(data)), nil
	}}})
	if err != nil || len(findings) != 1 || findings[0].Count != 1 {
		t.Fatal("scanner missed canary spanning an input chunk boundary")
	}
}
