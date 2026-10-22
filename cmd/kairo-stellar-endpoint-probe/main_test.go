package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestJSONOutputIncludesLatestLedgerCheck(t *testing.T) {
	horizon := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{
			"network_passphrase": testnetPassphrase,
		})
	}))
	defer horizon.Close()

	rpc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			ID     int    `json:"id"`
			Method string `json:"method"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Errorf("decode RPC request: %v", err)
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		var result any
		switch request.Method {
		case "getHealth":
			result = map[string]string{"status": "healthy"}
		case "getLatestLedger":
			result = map[string]uint32{"sequence": 123}
		case "getNetwork":
			result = map[string]string{"passphrase": testnetPassphrase}
		default:
			t.Errorf("unexpected RPC method %q", request.Method)
			http.Error(w, "unknown method", http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"jsonrpc": "2.0",
			"id":      request.ID,
			"result":  result,
		})
	}))
	defer rpc.Close()

	var stdout, stderr bytes.Buffer
	code := run([]string{
		"--horizon", horizon.URL,
		"--soroban-rpc", rpc.URL,
		"--network", "testnet",
		"--json",
	}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected successful JSON probe, got %d: %s", code, stderr.String())
	}

	var report struct {
		OK     bool `json:"ok"`
		Checks []struct {
			Name           string `json:"name"`
			LedgerSequence uint32 `json:"ledger_sequence"`
		} `json:"checks"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("decode CLI JSON report: %v\n%s", err, stdout.String())
	}
	if !report.OK {
		t.Fatalf("expected healthy JSON report, got %+v", report)
	}
	for _, check := range report.Checks {
		if check.Name == "Soroban RPC latest ledger" && check.LedgerSequence == 123 {
			return
		}
	}
	t.Fatalf("JSON report is missing the latest ledger sequence: %s", stdout.String())
}
