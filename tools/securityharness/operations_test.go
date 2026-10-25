package securityharness

import (
	"encoding/json"
	"os"
	"testing"
)

func TestOpenAPIOperationEnumerationIsDeterministicAndSeparatesStatus(t *testing.T) {
	data, err := os.ReadFile("../../api/openapi/v1.yaml")
	if err != nil {
		t.Fatal("could not read OpenAPI contract")
	}
	document, err := LoadOpenAPI(data)
	if err != nil {
		t.Fatal("could not load OpenAPI contract")
	}
	first, err := EnumerateOperations(document)
	if err != nil {
		t.Fatal(err)
	}
	second, err := EnumerateOperations(document)
	if err != nil {
		t.Fatal(err)
	}
	firstJSON, _ := json.Marshal(first)
	secondJSON, _ := json.Marshal(second)
	if string(firstJSON) != string(secondJSON) {
		t.Fatal("operation enumeration changed between runs")
	}
	if first.Summary.Declared != 2 || first.Summary.Implemented != 1 || first.Summary.Unimplemented != 1 {
		t.Fatalf("unexpected operation accounting: %#v", first.Summary)
	}
	if first.Summary.TenantScoped != 0 || first.Summary.TenantIsolationNA != 2 {
		t.Fatalf("tenant isolation applicability was not explicit: %#v", first.Summary)
	}
	if first.Operations[0].Path != "/capabilities" || first.Operations[0].Implementation != "unimplemented" || first.Operations[0].TenantIsolation != "not_applicable" {
		t.Fatalf("capabilities status was misreported: %#v", first.Operations[0])
	}
	if first.Operations[1].Path != "/healthz" || first.Operations[1].Implementation != "implemented" || first.Operations[1].TenantIsolation != "not_applicable" {
		t.Fatalf("public health status was misreported: %#v", first.Operations[1])
	}
	if first.Operations[1].Kind != "get" {
		t.Fatalf("health operation was misclassified: %#v", first.Operations[1])
	}
}

func TestOperationClassificationAndTenantScope(t *testing.T) {
	document, err := LoadOpenAPI([]byte(`openapi: 3.1.0
info: {title: test, version: "1"}
paths:
  /tenants/{tenant_id}/records:
    get:
      operationId: listRecords
      responses: {"200": {description: ok}}
  /tenants/{tenant_id}/records/{record_id}:
    get:
      operationId: getRecord
      responses: {"200": {description: ok}}
  /tenants/{tenant_id}/record-aggregate:
    get:
      operationId: getRecordAggregate
      responses: {"200": {description: ok}}
`))
	if err != nil {
		t.Fatal("could not load synthetic OpenAPI contract")
	}
	report, err := EnumerateOperations(document)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"aggregate", "list", "get"}
	if len(report.Operations) != len(want) {
		t.Fatalf("expected %d operations, got %d", len(want), len(report.Operations))
	}
	for index, operation := range report.Operations {
		if operation.Kind != want[index] || !operation.TenantScoped || operation.TenantIsolation != "not_applicable" {
			t.Fatalf("unexpected operation classification: %#v", operation)
		}
	}
}
