package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateRejectsDuplicateID(t *testing.T) {
	root := t.TempDir()
	writeTestSource(t, root)
	cat := catalog{Records: []record{
		{ID: "FR-A-01", Kind: "functional_requirement", Status: "specified", Source: "requirements/README.md#L28"},
		{ID: "FR-A-01", Kind: "functional_requirement", Status: "specified", Source: "requirements/README.md#L28"},
	}}
	err := validate(cat, root)
	if err == nil || !strings.Contains(err.Error(), "duplicate requirement ID") {
		t.Fatalf("validate() error = %v, want duplicate ID error", err)
	}
}

func TestValidateRejectsMissingCategoryEntry(t *testing.T) {
	root := t.TempDir()
	writeTestSource(t, root)
	cat := catalog{Records: []record{{ID: "FR-A-01", Kind: "functional_requirement", Status: "specified", Source: "requirements/README.md#L28"}}}
	err := validate(cat, root)
	if err == nil || !strings.Contains(err.Error(), `kind "acceptance_test" has 0 records; want 32`) {
		t.Fatalf("validate() error = %v, want missing-count error", err)
	}
}

func TestValidateRejectsUntraceableSource(t *testing.T) {
	root := t.TempDir()
	writeTestSource(t, root)
	if err := validateLink(root, "absent.md#L1"); err == nil {
		t.Fatal("validateLink() accepted missing source")
	}
	if err := validateLink(root, "source.md#L2"); err == nil {
		t.Fatal("validateLink() accepted missing line")
	}
}

func TestTransitionRequiresEvidenceAndAllowedStep(t *testing.T) {
	root := t.TempDir()
	writeTestSource(t, root)
	cat := catalog{Records: []record{{ID: "FR-A-01", Kind: "functional_requirement", Status: "specified", Source: "requirements/README.md#L28"}}}
	if err := applyTransition(&cat, "FR-A-01", "implemented", "source.md#L1", root); err == nil {
		t.Fatal("applyTransition() accepted skipped status")
	}
	if err := applyTransition(&cat, "FR-A-01", "designed", "source.md#L1", root); err != nil {
		t.Fatalf("applyTransition() error = %v", err)
	}
	if cat.Records[0].Status != "designed" || len(cat.Records[0].History) != 1 {
		t.Fatalf("transition did not retain status and evidence: %+v", cat.Records[0])
	}
}

func TestValidateRequiresTransitionHistory(t *testing.T) {
	root := t.TempDir()
	writeTestSource(t, root)
	cat := catalog{Records: []record{{ID: "FR-A-01", Kind: "functional_requirement", Status: "implemented", Source: "requirements/README.md#L28"}}}
	err := validate(cat, root)
	if err == nil || !strings.Contains(err.Error(), "does not match evidence history") {
		t.Fatalf("validate() error = %v, want evidence-history error", err)
	}
}

func TestTransitionRejectsUnavailableEvidence(t *testing.T) {
	root := t.TempDir()
	writeTestSource(t, root)
	cat := catalog{Records: []record{{ID: "D01", Kind: "detection", Status: "specified", Source: "requirements/README.md#L29"}}}
	if err := applyTransition(&cat, "D01", "unavailable", "missing.md#L1", root); err == nil {
		t.Fatal("applyTransition() accepted invalid evidence link")
	}
}

func TestValidateRejectsMissingIDEvenWhenCategoryCountMatches(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	cat, err := readCatalog(filepath.Join(root, catalogPath))
	if err != nil {
		t.Fatal(err)
	}
	cat.Records[0].ID = "FR-BOGUS-01"
	err = validate(cat, root)
	if err == nil || !strings.Contains(err.Error(), "missing requirement ID") {
		t.Fatalf("validate() error = %v, want missing ID error", err)
	}
}

func TestValidateRejectsUnknownConnectorRow(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	cat, err := readCatalog(filepath.Join(root, catalogPath))
	if err != nil {
		t.Fatal(err)
	}
	for index := range cat.Records {
		if cat.Records[index].Kind == "connector" {
			cat.Records[index].ID = "unknown connector"
			break
		}
	}
	err = validate(cat, root)
	if err == nil || !strings.Contains(err.Error(), "missing connector-family row") {
		t.Fatalf("validate() error = %v, want missing connector-row error", err)
	}
}

func writeTestSource(t *testing.T, root string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, "source.md"), []byte("evidence\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "requirements"), 0700); err != nil {
		t.Fatal(err)
	}
	readme := strings.Repeat("\n", 27) + "- Functional requirements (`FR-*`): product specification §5.\n" +
		"- Detections (`D*`): product specification §8.2.\n"
	if err := os.WriteFile(filepath.Join(root, "requirements", "README.md"), []byte(readme), 0600); err != nil {
		t.Fatal(err)
	}
}
