package usagemetrics

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDefaultMetricsRecordOnlyAnonymousContractFields(t *testing.T) {
	t.Setenv("NOMYR_CONFIG_HOME", t.TempDir())
	received := make(chan []Event, 1)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		var payload struct {
			Events []Event `json:"events"`
		}
		if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
			t.Errorf("decode metrics: %v", err)
		}
		received <- payload.Events
		writer.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()
	t.Setenv("NOMYR_TELEMETRY_ENDPOINT", server.URL)

	RecordCommand(context.Background(), "nomyr doctor --json /private/path", 25*time.Millisecond, true)
	events := <-received
	if len(events) != 2 || events[0].Event != "installation_created" || events[1].Event != "command_completed" {
		t.Fatalf("events = %#v", events)
	}
	if events[1].Command != "doctor --json" {
		t.Fatalf("command = %q", events[1].Command)
	}
	if events[1].Success == nil || !*events[1].Success {
		t.Fatal("successful command was not recorded")
	}

	stored, err := Status()
	if err != nil {
		t.Fatal(err)
	}
	if !stored.Enabled {
		t.Fatal("usage metrics should be enabled by default")
	}
	if len(stored.InstallationID) != 32 {
		t.Fatalf("installation ID length = %d, want 32", len(stored.InstallationID))
	}
	if !stored.InstallationReported {
		t.Fatal("installation event was not marked as reported")
	}
}

func TestOptedOutMetricsDoNotSend(t *testing.T) {
	t.Setenv("NOMYR_CONFIG_HOME", t.TempDir())
	if err := Disable(); err != nil {
		t.Fatal(err)
	}
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		requests++
	}))
	defer server.Close()
	t.Setenv("NOMYR_TELEMETRY_ENDPOINT", server.URL)

	RecordCommand(context.Background(), "nomyr version", time.Millisecond, true)
	if requests != 0 {
		t.Fatalf("requests = %d, want 0", requests)
	}
}
