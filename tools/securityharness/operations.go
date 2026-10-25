package securityharness

import (
	"fmt"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

type Operation struct {
	Method              string `json:"method"`
	Path                string `json:"path"`
	ID                  string `json:"operation_id,omitempty"`
	Kind                string `json:"kind"`
	Review              string `json:"review"`
	Implementation      string `json:"implementation"`
	Access              string `json:"access"`
	TenantScoped        bool   `json:"tenant_scoped"`
	TenantIsolation     string `json:"tenant_isolation"`
	TenantIsolationNote string `json:"tenant_isolation_note,omitempty"`
}

type OperationReport struct {
	Operations []Operation `json:"operations"`
	Summary    struct {
		Declared                int `json:"declared"`
		Implemented             int `json:"implemented"`
		Unimplemented           int `json:"unimplemented"`
		TenantScoped            int `json:"tenant_scoped"`
		TenantIsolationPassing  int `json:"tenant_isolation_passing"`
		TenantIsolationUntested int `json:"tenant_isolation_not_tested"`
		TenantIsolationNA       int `json:"tenant_isolation_not_applicable"`
		NotVerified             int `json:"not_verified"`
	} `json:"summary"`
}

type inventoryEntry struct {
	operationID     string
	implementation  string
	access          string
	tenantScoped    bool
	tenantIsolation string
	note            string
}

var phaseZeroInventory = map[string]inventoryEntry{
	"GET /healthz": {
		operationID: "getHealth", implementation: "implemented", access: "public",
		tenantIsolation: "not_applicable", note: "public process health endpoint",
	},
	"GET /capabilities": {
		operationID: "listCapabilities", implementation: "known_unimplemented", access: "authenticated",
		tenantIsolation: "not_applicable", note: "declared contract has no implemented handler",
	},
}

func EnumerateOperations(document *openapi3.T) (OperationReport, error) {
	var report OperationReport
	if document == nil || document.Paths == nil {
		return report, fmt.Errorf("OpenAPI document has no paths")
	}
	paths := document.Paths.Keys()
	sort.Strings(paths)
	for _, path := range paths {
		item := document.Paths.Value(path)
		if item == nil {
			report.Summary.NotVerified++
			continue
		}
		methods := make([]string, 0)
		operations := make(map[string]*openapi3.Operation)
		for method, operation := range item.Operations() {
			if operation != nil {
				method = strings.ToUpper(method)
				methods = append(methods, method)
				operations[method] = operation
			}
		}
		sort.Strings(methods)
		for _, method := range methods {
			operation := operations[method]
			id := operation.OperationID
			kind := classifyOperation(method, path, id)
			key := method + " " + path
			entry, reviewed := phaseZeroInventory[key]
			access := effectiveAccess(document, operation)
			record := Operation{
				Method:          method,
				Path:            path,
				ID:              id,
				Kind:            kind,
				Review:          "not_verified",
				Implementation:  "not_verified",
				Access:          "not_verified",
				TenantScoped:    isTenantScoped(path, operation.Extensions),
				TenantIsolation: "not_verified",
			}
			if reviewed && entry.operationID == id && entry.access == access {
				record.Review = "reviewed"
				record.Implementation = entry.implementation
				record.Access = entry.access
				record.TenantScoped = entry.tenantScoped
				record.TenantIsolation = entry.tenantIsolation
				record.TenantIsolationNote = entry.note
			}
			report.Operations = append(report.Operations, record)
		}
	}
	for key := range phaseZeroInventory {
		found := false
		for _, operation := range report.Operations {
			if operation.Method+" "+operation.Path == key && operation.Review == "reviewed" {
				found = true
				break
			}
		}
		if !found {
			report.Summary.NotVerified++
		}
	}
	for _, operation := range report.Operations {
		report.Summary.Declared++
		if operation.Implementation == "implemented" {
			report.Summary.Implemented++
		}
		if operation.Implementation == "known_unimplemented" {
			report.Summary.Unimplemented++
		}
		if operation.Review != "reviewed" {
			report.Summary.NotVerified++
		}
		if operation.TenantScoped {
			report.Summary.TenantScoped++
			switch operation.TenantIsolation {
			case "pass":
				report.Summary.TenantIsolationPassing++
			case "not_tested":
				report.Summary.TenantIsolationUntested++
			}
		}
		if operation.TenantIsolation == "not_applicable" {
			report.Summary.TenantIsolationNA++
		}
	}
	return report, nil
}

func ValidatePhase0Inventory(report OperationReport) error {
	if report.Summary.NotVerified > 0 {
		return fmt.Errorf("OpenAPI operation inventory requires review")
	}
	return nil
}

func effectiveAccess(document *openapi3.T, operation *openapi3.Operation) string {
	security := document.Security
	if operation.Security != nil {
		security = *operation.Security
	}
	if len(security) == 0 {
		return "public"
	}
	for _, requirement := range security {
		if len(requirement) == 0 {
			return "not_verified"
		}
		for name := range requirement {
			if name != "bearerAuth" {
				return "not_verified"
			}
		}
	}
	if document.Components == nil {
		return "not_verified"
	}
	scheme := document.Components.SecuritySchemes["bearerAuth"]
	if scheme == nil || scheme.Value == nil || scheme.Value.Type != "http" || !strings.EqualFold(scheme.Value.Scheme, "bearer") {
		return "not_verified"
	}
	return "authenticated"
}

func classifyOperation(method, path, id string) string {
	value := strings.ToLower(id + " " + path)
	for _, marker := range []string{"aggregate", "summary", "statistics", "stats"} {
		if strings.Contains(value, marker) {
			return "aggregate"
		}
	}
	if strings.Contains(strings.ToLower(id), "list") {
		return "list"
	}
	if strings.HasPrefix(strings.ToLower(id), "get") {
		return "get"
	}
	if method == "GET" && strings.Contains(path, "{") {
		return "get"
	}
	return "other"
}

func isTenantScoped(path string, extensions map[string]any) bool {
	if extensions["x-nomyr-tenant-scoped"] == true {
		return true
	}
	for _, segment := range strings.Split(path, "/") {
		if strings.EqualFold(segment, "tenants") || strings.EqualFold(segment, "tenant") {
			return true
		}
		if strings.HasPrefix(segment, "{") && strings.Contains(strings.ToLower(segment), "tenant") {
			return true
		}
	}
	return false
}

func LoadOpenAPI(data []byte) (*openapi3.T, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = false
	document, err := loader.LoadFromData(data)
	if err != nil {
		return nil, fmt.Errorf("load OpenAPI document")
	}
	return document, nil
}
