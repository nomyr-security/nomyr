package main

import "testing"

func TestLatencyStatsPreservesFirstRun(t *testing.T) {
	firstRun, medianRun := latencyStats([]float64{9, 1, 5, 3, 7})
	if firstRun != 9 || medianRun != 5 {
		t.Fatalf("got first=%v median=%v; want first=9 median=5", firstRun, medianRun)
	}
}
