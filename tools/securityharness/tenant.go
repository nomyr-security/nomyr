package securityharness

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
)

const ErrTenantIsolation = "tenant isolation assertion failed"

type TenantPrincipal struct {
	TenantID string
}

func AssertTenantIsolation(handler http.Handler, principal TenantPrincipal, foreignTenantID string) error {
	if principal.TenantID == "" || foreignTenantID == "" || principal.TenantID == foreignTenantID {
		return errors.New(ErrTenantIsolation)
	}
	request := httptest.NewRequest(http.MethodGet, "/synthetic/records?tenant_id="+foreignTenantID, nil)
	request.Header.Set("X-Synthetic-Tenant", principal.TenantID)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	body, err := io.ReadAll(response.Body)
	if err != nil || bytes.Contains(body, []byte(foreignTenantID)) {
		return errors.New(ErrTenantIsolation)
	}
	if response.Code == http.StatusNotFound || response.Code == http.StatusForbidden || response.Code == http.StatusUnauthorized {
		return nil
	}
	return errors.New(ErrTenantIsolation)
}
