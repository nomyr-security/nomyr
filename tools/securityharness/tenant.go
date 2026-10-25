package securityharness

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
)

const ErrTenantIsolation = "tenant isolation assertion failed"

type TenantPrincipal struct {
	TenantID string
}

type OperationTarget struct {
	ID     string
	Method string
	Path   string
}

type ExpectedDenial struct {
	Status      int
	ContentType string
	Problem     bool
}

type TenantScenario struct {
	Operation      OperationTarget
	Principal      TenantPrincipal
	ForeignTenant  string
	ExpectedDenial ExpectedDenial
}

func AssertTenantIsolation(handler http.Handler, scenario TenantScenario) error {
	if handler == nil || scenario.Operation.ID == "" || scenario.Operation.Method == "" || scenario.Operation.Path == "" ||
		scenario.Principal.TenantID == "" || scenario.ForeignTenant == "" || scenario.Principal.TenantID == scenario.ForeignTenant ||
		!isDenialStatus(scenario.ExpectedDenial.Status) || scenario.ExpectedDenial.ContentType == "" {
		return errors.New(ErrTenantIsolation)
	}
	parsed, err := url.Parse(scenario.Operation.Path)
	if err != nil || parsed.IsAbs() || parsed.Host != "" || parsed.User != nil || parsed.Fragment != "" || parsed.Path == "" || parsed.Path[0] != '/' {
		return errors.New(ErrTenantIsolation)
	}
	query := parsed.Query()
	query.Set("tenant_id", scenario.ForeignTenant)
	parsed.RawQuery = query.Encode()
	request, err := http.NewRequest(scenario.Operation.Method, parsed.String(), nil)
	if err != nil {
		return errors.New(ErrTenantIsolation)
	}
	request.Header.Set("X-Synthetic-Tenant", scenario.Principal.TenantID)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	body, err := io.ReadAll(response.Body)
	if err != nil || response.Code != scenario.ExpectedDenial.Status || bytes.Contains(body, []byte(scenario.ForeignTenant)) {
		return errors.New(ErrTenantIsolation)
	}
	mediaType, _, err := mime.ParseMediaType(response.Header().Get("Content-Type"))
	if err != nil || mediaType != scenario.ExpectedDenial.ContentType {
		return errors.New(ErrTenantIsolation)
	}
	if scenario.ExpectedDenial.Problem && !validProblem(body, response.Code) {
		return errors.New(ErrTenantIsolation)
	}
	return nil
}

func isDenialStatus(status int) bool {
	return status == http.StatusUnauthorized || status == http.StatusForbidden || status == http.StatusNotFound
}

func validProblem(body []byte, status int) bool {
	var problem map[string]json.RawMessage
	if json.Unmarshal(body, &problem) != nil {
		return false
	}
	allowed := map[string]struct{}{
		"type": {}, "title": {}, "status": {}, "detail": {}, "instance": {},
		"code": {}, "correlation_id": {}, "retry_after_seconds": {},
	}
	for key := range problem {
		if _, exists := allowed[key]; !exists {
			return false
		}
	}
	var typeValue, title, code string
	var problemStatus int
	if json.Unmarshal(problem["type"], &typeValue) != nil || json.Unmarshal(problem["title"], &title) != nil ||
		json.Unmarshal(problem["status"], &problemStatus) != nil || json.Unmarshal(problem["code"], &code) != nil {
		return false
	}
	for _, key := range []string{"detail", "instance", "correlation_id"} {
		if raw, exists := problem[key]; exists {
			var value string
			if json.Unmarshal(raw, &value) != nil {
				return false
			}
			if key == "instance" && value != "" && !isURIReference(value) {
				return false
			}
		}
	}
	if raw, exists := problem["retry_after_seconds"]; exists {
		var retryAfter int
		if json.Unmarshal(raw, &retryAfter) != nil || retryAfter < 0 {
			return false
		}
	}
	typeValid := typeValue != "" && isURIReference(typeValue)
	codeMatches, _ := regexp.MatchString(`^NHI-[A-Z]+-[0-9]{4}$`, code)
	return typeValid && title != "" && problemStatus == status && codeMatches
}

func isURIReference(value string) bool {
	_, err := url.Parse(value)
	return err == nil && !strings.ContainsAny(value, " \t\r\n")
}
