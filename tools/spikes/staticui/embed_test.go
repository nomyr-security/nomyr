package staticui

import (
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestEmbeddedIdentityRouteAndColdStart(t *testing.T) {
	binaryPath := filepath.Join("..", "..", "..", "bin", "nomyr")
	if _, err := os.Stat(binaryPath); err != nil {
		t.Skip("build bin/nomyr first with make build")
	}
	configHome := t.TempDir()
	environment := append(os.Environ(), "NOMYR_CONFIG_HOME="+configHome)
	disableTelemetry := exec.Command(binaryPath, "telemetry", "disable")
	disableTelemetry.Env = environment
	if output, err := disableTelemetry.CombinedOutput(); err != nil {
		t.Fatalf("disable telemetry in temporary config home: %v: %s", err, output)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	address := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}

	command := exec.Command(binaryPath, "demo", "--listen", address)
	command.Env = environment
	stdout, err := command.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	stderr, err := command.StderrPipe()
	if err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = command.Process.Kill()
		_ = command.Wait()
		_, _ = io.Copy(io.Discard, stdout)
		_, _ = io.Copy(io.Discard, stderr)
	})

	client := &http.Client{Timeout: 200 * time.Millisecond}
	baseURL := "http://" + address
	var healthReady bool
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		response, requestErr := client.Get(baseURL + "/api/v1/healthz")
		if requestErr == nil {
			_, _ = io.Copy(io.Discard, response.Body)
			response.Body.Close()
			if response.StatusCode == http.StatusOK {
				healthReady = true
				break
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !healthReady {
		_ = command.Process.Kill()
		_ = command.Wait()
		out, _ := io.ReadAll(stdout)
		errOutput, _ := io.ReadAll(stderr)
		t.Fatalf("demo health endpoint did not become ready: stdout=%s stderr=%s", out, errOutput)
	}
	coldStart := time.Since(started)
	t.Logf("cold_start_to_healthz=%s platform=%s/%s", coldStart, runtime.GOOS, runtime.GOARCH)

	response, err := client.Get(baseURL + "/identities/demo-service-account/?as_of=2026-01-01T00:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("identity static route status=%d, want 200", response.StatusCode)
	}
	for _, expected := range []string{"demo-service-account", "Static route spike", "As of:"} {
		if !strings.Contains(string(body), expected) {
			t.Fatalf("identity route response did not contain %q", expected)
		}
	}
}
