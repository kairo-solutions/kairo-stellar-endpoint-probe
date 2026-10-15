package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestHelpReturnsSuccess(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run([]string{"--help"}, &stdout, &stderr); code != 0 {
		t.Fatalf("expected help to return success, got %d", code)
	}
	if !strings.Contains(stderr.String(), "Probe Horizon and Soroban RPC") {
		t.Fatalf("expected help text, got %q", stderr.String())
	}
}

func TestMissingEndpointsReturnsUsageError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	if code := run(nil, &stdout, &stderr); code != 2 {
		t.Fatalf("expected missing endpoints to return usage error, got %d", code)
	}
	if !strings.Contains(stderr.String(), "--horizon URL") {
		t.Fatalf("expected usage instructions, got %q", stderr.String())
	}
}
