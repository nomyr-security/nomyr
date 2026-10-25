package securityharness

import (
	"net/http"
	"testing"
)

func TestTenantIsolationAcceptsDenialAndRejectsForeignTenantResponse(t *testing.T) {
	principal := TenantPrincipal{TenantID: "synthetic-requester"}
	denied := http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusForbidden)
	})
	if err := AssertTenantIsolation(denied, principal, "synthetic-foreign"); err != nil {
		t.Fatalf("expected synthetic denial to pass: %v", err)
	}
	leakingDenial := http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusForbidden)
		_, _ = response.Write([]byte("synthetic-foreign"))
	})
	if err := AssertTenantIsolation(leakingDenial, principal, "synthetic-foreign"); err == nil {
		t.Fatal("foreign tenant data in a denial response was accepted")
	}
	unsafe := http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"tenant_id":"synthetic-foreign","id":"synthetic-record"}`))
	})
	if err := AssertTenantIsolation(unsafe, principal, "synthetic-foreign"); err == nil || err.Error() != ErrTenantIsolation {
		t.Fatal("unsafe foreign-tenant response was not rejected")
	}
}

func TestTenantIsolationRejectsUnexpectedServerFailure(t *testing.T) {
	broken := http.HandlerFunc(func(response http.ResponseWriter, _ *http.Request) {
		http.Error(response, "synthetic error", http.StatusInternalServerError)
	})
	if err := AssertTenantIsolation(broken, TenantPrincipal{TenantID: "synthetic-requester"}, "synthetic-foreign"); err == nil {
		t.Fatal("unexpected server failure was accepted as an isolation pass")
	}
}
