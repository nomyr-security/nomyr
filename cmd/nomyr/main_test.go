package main

import "testing"

func TestLoopbackAddressGuard(t *testing.T) {
	testCases := []struct {
		address string
		allowed bool
	}{
		{address: "127.0.0.1:8080", allowed: true},
		{address: "localhost:8080", allowed: true},
		{address: "[::1]:8080", allowed: true},
		{address: "0.0.0.0:8080", allowed: false},
		{address: "example.com:8080", allowed: false},
		{address: "invalid", allowed: false},
	}
	for _, testCase := range testCases {
		if actual := isLoopbackAddress(testCase.address); actual != testCase.allowed {
			t.Errorf("isLoopbackAddress(%q) = %v, want %v", testCase.address, actual, testCase.allowed)
		}
	}
}
