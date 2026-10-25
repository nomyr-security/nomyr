package securityharness

import (
	"fmt"
	"net/http"
	"testing"
)

func TestTenantIsolationUsesConfiguredListGetAndAggregateDenials(t *testing.T) {
	scenarios := []TenantScenario{
		{
			Operation:      OperationTarget{ID: "listRecords", Method: http.MethodGet, Path: "/synthetic/records"},
			Principal:      TenantPrincipal{TenantID: "synthetic-requester"},
			ForeignTenant:  "synthetic-foreign",
			ExpectedDenial: ExpectedDenial{Status: http.StatusNotFound, ContentType: "application/problem+json", Problem: true},
		},
		{
			Operation:      OperationTarget{ID: "getRecord", Method: http.MethodGet, Path: "/synthetic/records/synthetic-record"},
			Principal:      TenantPrincipal{TenantID: "synthetic-requester"},
			ForeignTenant:  "synthetic-foreign",
			ExpectedDenial: ExpectedDenial{Status: http.StatusNotFound, ContentType: "application/problem+json", Problem: true},
		},
		{
			Operation:      OperationTarget{ID: "aggregateRecords", Method: http.MethodGet, Path: "/synthetic/records/aggregate"},
			Principal:      TenantPrincipal{TenantID: "synthetic-requester"},
			ForeignTenant:  "synthetic-foreign",
			ExpectedDenial: ExpectedDenial{Status: http.StatusForbidden, ContentType: "application/problem+json", Problem: true},
		},
	}
	for _, scenario := range scenarios {
		t.Run(scenario.Operation.ID, func(t *testing.T) {
			handler := http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				if request.Method != scenario.Operation.Method || request.URL.Path != scenario.Operation.Path || request.URL.Query().Get("tenant_id") != scenario.ForeignTenant || request.Header.Get("X-Synthetic-Tenant") != scenario.Principal.TenantID {
					t.Errorf("fixture received unexpected operation or principal")
				}
				response.Header().Set("Content-Type", "application/problem+json")
				response.WriteHeader(scenario.ExpectedDenial.Status)
				_, _ = fmt.Fprintf(response, `{"type":"about:blank","title":"Access denied","status":%d,"code":"NHI-CORE-0001"}`, scenario.ExpectedDenial.Status)
			})
			if err := AssertTenantIsolation(handler, scenario); err != nil {
				t.Fatalf("valid synthetic problem denial failed: %v", err)
			}
		})
	}
}

func TestTenantIsolationRejectsMalformedOrLeakingDenials(t *testing.T) {
	scenario := TenantScenario{
		Operation:      OperationTarget{ID: "getRecord", Method: http.MethodGet, Path: "/synthetic/records/synthetic-record"},
		Principal:      TenantPrincipal{TenantID: "synthetic-requester"},
		ForeignTenant:  "synthetic-foreign",
		ExpectedDenial: ExpectedDenial{Status: http.StatusNotFound, ContentType: "application/problem+json", Problem: true},
	}
	tests := []struct {
		name    string
		handler http.Handler
	}{
		{
			name:    "missing RFC 9457 required fields",
			handler: problemHandler(http.StatusNotFound, "application/problem+json", `{"status":404}`),
		},
		{
			name:    "wrong problem content type",
			handler: problemHandler(http.StatusNotFound, "text/plain", `{"type":"about:blank","title":"Missing","status":404,"code":"NHI-CORE-0001"}`),
		},
		{
			name:    "problem status mismatch",
			handler: problemHandler(http.StatusNotFound, "application/problem+json", `{"type":"about:blank","title":"Missing","status":403,"code":"NHI-CORE-0001"}`),
		},
		{
			name:    "malformed error code",
			handler: problemHandler(http.StatusNotFound, "application/problem+json", `{"type":"about:blank","title":"Missing","status":404,"code":"bad"}`),
		},
		{
			name:    "denial leaks foreign tenant data",
			handler: problemHandler(http.StatusNotFound, "application/problem+json", `{"type":"about:blank","title":"Missing","status":404,"code":"NHI-CORE-0001","detail":"synthetic-foreign"}`),
		},
		{
			name:    "wrong denial status",
			handler: problemHandler(http.StatusForbidden, "application/problem+json", `{"type":"about:blank","title":"Denied","status":403,"code":"NHI-CORE-0001"}`),
		},
		{
			name:    "unexpected server error",
			handler: problemHandler(http.StatusInternalServerError, "application/problem+json", `{"type":"about:blank","title":"Error","status":500,"code":"NHI-CORE-0001"}`),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := AssertTenantIsolation(test.handler, scenario); err == nil || err.Error() != ErrTenantIsolation {
				t.Fatal("invalid or leaking denial response was accepted")
			}
		})
	}
}

func TestTenantIsolationRejectsInvalidScenario(t *testing.T) {
	scenario := TenantScenario{
		Operation:      OperationTarget{ID: "getRecord", Method: http.MethodGet, Path: "/synthetic/records"},
		Principal:      TenantPrincipal{TenantID: "synthetic-requester"},
		ForeignTenant:  "synthetic-foreign",
		ExpectedDenial: ExpectedDenial{Status: http.StatusOK, ContentType: "application/json"},
	}
	if err := AssertTenantIsolation(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), scenario); err == nil {
		t.Fatal("invalid expected denial status was accepted")
	}
}

func problemHandler(status int, contentType, body string) http.Handler {
	return http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.Header().Set("Content-Type", contentType)
		response.WriteHeader(status)
		_, _ = fmt.Fprint(response, body)
	})
}
