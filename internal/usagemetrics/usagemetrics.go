package usagemetrics

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"nomyr/internal/buildinfo"
)

const defaultEndpoint = "https://telemetry.nomyr.io/v1/events"

type Config struct {
	Enabled              bool   `json:"enabled"`
	InstallationID       string `json:"installation_id,omitempty"`
	InstallationReported bool   `json:"installation_reported,omitempty"`
}

type Event struct {
	SchemaVersion  string `json:"schema_version"`
	Event          string `json:"event"`
	InstallationID string `json:"installation_id"`
	OccurredAt     string `json:"occurred_at"`
	Version        string `json:"version"`
	OS             string `json:"os"`
	Arch           string `json:"arch"`
	Command        string `json:"command,omitempty"`
	Success        *bool  `json:"success,omitempty"`
	DurationMS     *int64 `json:"duration_ms,omitempty"`
}

func Enable() (Config, error) {
	config, err := load()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, err
	}
	if config.InstallationID == "" {
		config.InstallationID, err = newInstallationID()
		if err != nil {
			return Config{}, err
		}
	}
	config.Enabled = true
	return config, save(config)
}

func Disable() error {
	return save(Config{Enabled: false})
}

func Status() (Config, error) {
	config, err := load()
	if errors.Is(err, os.ErrNotExist) {
		return Config{Enabled: true}, nil
	}
	return config, err
}

func RecordCommand(ctx context.Context, command string, elapsed time.Duration, succeeded bool) {
	config, err := Status()
	if err != nil || !config.Enabled {
		return
	}
	if config.InstallationID == "" {
		config.InstallationID, err = newInstallationID()
		if err != nil || save(config) != nil {
			return
		}
	}

	version := buildinfo.Current().Version
	now := time.Now().UTC().Format(time.RFC3339)
	events := make([]Event, 0, 2)
	if !config.InstallationReported {
		events = append(events, Event{
			SchemaVersion: "1", Event: "installation_created", InstallationID: config.InstallationID,
			OccurredAt: now, Version: version, OS: runtime.GOOS, Arch: runtime.GOARCH,
		})
	}
	durationMS := elapsed.Milliseconds()
	if durationMS < 0 {
		durationMS = 0
	}
	events = append(events, Event{
		SchemaVersion: "1", Event: "command_completed", InstallationID: config.InstallationID,
		OccurredAt: now, Version: version, OS: runtime.GOOS, Arch: runtime.GOARCH,
		Command: sanitizeCommand(command), Success: &succeeded, DurationMS: &durationMS,
	})

	if send(ctx, events) == nil && !config.InstallationReported {
		config.InstallationReported = true
		_ = save(config)
	}
}

func sanitizeCommand(command string) string {
	command = strings.TrimSpace(strings.TrimPrefix(command, "nomyr"))
	if command == "" {
		return "root"
	}
	parts := strings.Fields(command)
	if len(parts) > 2 {
		parts = parts[:2]
	}
	return strings.Join(parts, " ")
}

func send(ctx context.Context, events []Event) error {
	body, err := json.Marshal(struct {
		Events []Event `json:"events"`
	}{Events: events})
	if err != nil {
		return err
	}
	endpoint := os.Getenv("NOMYR_TELEMETRY_ENDPOINT")
	if endpoint == "" {
		endpoint = defaultEndpoint
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("usage metrics endpoint returned %s", response.Status)
	}
	return nil
}

func configPath() (string, error) {
	if override := os.Getenv("NOMYR_CONFIG_HOME"); override != "" {
		return filepath.Join(override, "telemetry.json"), nil
	}
	root, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "nomyr", "telemetry.json"), nil
}

func load() (Config, error) {
	path, err := configPath()
	if err != nil {
		return Config{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("read usage metrics configuration: %w", err)
	}
	return config, nil
}

func save(config Config) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o600)
}

func newInstallationID() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return hex.EncodeToString(value), nil
}
