package endpointprobe

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const testnet = "Test SDF Network ; September 2015"

func TestProbeHealthyMatchingEndpoints(t *testing.T) {
	horizon := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/" {
			t.Errorf("unexpected Horizon request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"network_passphrase": testnet})
	}))
	defer horizon.Close()

	rpc := httptest.NewServer(rpcHandler(t, testnet, "healthy"))
	defer rpc.Close()

	report, err := Probe(context.Background(), Config{
		HorizonURL:        horizon.URL,
		SorobanRPCURL:     rpc.URL,
		Network:           "testnet",
		NetworkPassphrase: testnet,
		Timeout:           time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !report.OK {
		t.Fatalf("expected successful report, got %+v", report)
	}
	if len(report.Checks) != 6 {
		t.Fatalf("expected six checks, got %d: %+v", len(report.Checks), report.Checks)
	}
}

func TestProbeDetectsNetworkMismatch(t *testing.T) {
	horizon := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"network_passphrase": testnet})
	}))
	defer horizon.Close()

	rpc := httptest.NewServer(rpcHandler(t, "Public Global Stellar Network ; September 2015", "healthy"))
	defer rpc.Close()

	report, err := Probe(context.Background(), Config{
		HorizonURL:        horizon.URL,
		SorobanRPCURL:     rpc.URL,
		Network:           "testnet",
		NetworkPassphrase: testnet,
		Timeout:           time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	if report.OK {
		t.Fatalf("expected mismatch to fail, got %+v", report)
	}
	if report.Checks[4].Status != "fail" {
		t.Fatalf("expected Soroban network check to fail, got %+v", report.Checks[4])
	}
	if report.Checks[5].Status != "fail" {
		t.Fatalf("expected consistency check to fail, got %+v", report.Checks[5])
	}
}

func TestProbeRedactsEndpointQueryParameters(t *testing.T) {
	horizon := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("api_key") != "sample-token" {
			t.Errorf("expected API key query parameter to reach Horizon request")
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"network_passphrase": testnet})
	}))
	defer horizon.Close()

	rpc := httptest.NewServer(rpcHandler(t, testnet, "healthy"))
	defer rpc.Close()

	report, err := Probe(context.Background(), Config{
		HorizonURL:        horizon.URL + "?api_key=sample-token",
		SorobanRPCURL:     rpc.URL + "?api_key=sample-token",
		Network:           "testnet",
		NetworkPassphrase: testnet,
		Timeout:           time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(report)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "sample-token") {
		t.Fatalf("report unexpectedly contains endpoint query credentials: %s", encoded)
	}
}

func TestProbeRejectsCredentialBearingEndpoint(t *testing.T) {
	_, err := Probe(context.Background(), Config{
		HorizonURL:    "https://user:secret@example.test",
		SorobanRPCURL: "https://rpc.example.test",
		Timeout:       time.Second,
	})
	if err == nil {
		t.Fatal("expected endpoint with credentials to be rejected")
	}
}

func rpcHandler(t *testing.T, passphrase, healthStatus string) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
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
			result = map[string]string{"status": healthStatus}
		case "getNetwork":
			result = map[string]string{"passphrase": passphrase}
		default:
			t.Errorf("unexpected RPC method %q", request.Method)
			http.Error(w, "unknown method", http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"jsonrpc": "2.0",
			"id":      1,
			"result":  result,
		})
	})
}
