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
	Implementation      string `json:"implementation"`
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
	} `json:"summary"`
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
			implemented := isImplemented(method, path)
			tenantScoped := isTenantScoped(path, operation.Extensions)
			record := Operation{
				Method:         method,
				Path:           path,
				ID:             id,
				Kind:           kind,
				Implementation: "unimplemented",
				TenantScoped:   tenantScoped,
			}
			if implemented {
				record.Implementation = "implemented"
			}
			if tenantScoped && implemented {
				record.TenantIsolation = "not_tested"
				record.TenantIsolationNote = "operation is tenant-scoped; no product authorization assertion is wired"
			} else {
				record.TenantIsolation = "not_applicable"
				switch {
				case !implemented:
					record.TenantIsolationNote = "operation has no implemented handler"
				case !tenantScoped:
					record.TenantIsolationNote = "operation is not declared tenant-scoped"
				}
			}
			report.Operations = append(report.Operations, record)
		}
	}
	for _, operation := range report.Operations {
		report.Summary.Declared++
		if operation.Implementation == "implemented" {
			report.Summary.Implemented++
		} else {
			report.Summary.Unimplemented++
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

func isImplemented(method, path string) bool {
	return method == "GET" && (path == "/healthz" || path == "/demo-estate")
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
