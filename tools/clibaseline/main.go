package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"time"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: clibaseline BINARY")
		os.Exit(2)
	}
	binary := os.Args[1]
	configHome, err := os.MkdirTemp("", "nomyr-cli-baseline-")
	if err != nil {
		fail(err)
	}
	defer func() { _ = os.RemoveAll(configHome) }()
	if err := os.WriteFile(filepath.Join(configHome, "telemetry.json"), []byte(`{"enabled":false}`), 0o600); err != nil {
		fail(err)
	}
	latencies := make([]float64, 0, 25)
	for index := 0; index < cap(latencies); index++ {
		started := time.Now()
		command := exec.Command(binary, "version")
		command.Env = append(os.Environ(), "NOMYR_CONFIG_HOME="+configHome)
		output, runErr := command.CombinedOutput()
		latencies = append(latencies, float64(time.Since(started).Microseconds())/1000)
		if runErr != nil {
			_, _ = os.Stderr.Write(output)
			fail(runErr)
		}
	}
	file, err := os.Stat(binary)
	if err != nil {
		fail(err)
	}
	sort.Float64s(latencies)
	report := struct {
		Command         string  `json:"command"`
		Runs            int     `json:"runs"`
		ColdStartFirst  float64 `json:"cold_start_first_ms"`
		MedianStart     float64 `json:"median_start_ms"`
		BinarySizeBytes int64   `json:"binary_size_bytes"`
	}{"nomyr version", len(latencies), latencies[0], latencies[len(latencies)/2], file.Size()}
	if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
