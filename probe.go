package endpointprobe

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const maxResponseSize = 1 << 20

// Config describes the network endpoints to inspect.
type Config struct {
	HorizonURL        string
	SorobanRPCURL     string
	Network           string
	NetworkPassphrase string
	Timeout           time.Duration
}

// Check is one observable result from an endpoint probe.
type Check struct {
	Name           string `json:"name"`
	Status         string `json:"status"`
	Endpoint       string `json:"endpoint,omitempty"`
	LatencyMS      int64  `json:"latency_ms,omitempty"`
	LedgerSequence uint32 `json:"ledger_sequence,omitempty"`
	Message        string `json:"message,omitempty"`
}

// Report contains the outcome of all requested checks.
type Report struct {
	Network string  `json:"network"`
	OK      bool    `json:"ok"`
	Checks  []Check `json:"checks"`
}

type endpointResult struct {
	check      Check
	passphrase string
}

// Probe checks Horizon's root endpoint and Soroban RPC health/network methods.
func Probe(ctx context.Context, cfg Config) (Report, error) {
	if err := validateEndpoint("Horizon", cfg.HorizonURL); err != nil {
		return Report{}, err
	}
	if err := validateEndpoint("Soroban RPC", cfg.SorobanRPCURL); err != nil {
		return Report{}, err
	}
	if cfg.Timeout <= 0 {
		return Report{}, fmt.Errorf("timeout must be greater than zero")
	}

	client := &http.Client{Timeout: cfg.Timeout}
	var horizon endpointResult
	var health Check
	var latestLedger Check
	var network endpointResult
	var checks sync.WaitGroup
	checks.Add(4)
	go func() {
		defer checks.Done()
		horizon = checkHorizon(ctx, client, cfg.HorizonURL)
	}()
	go func() {
		defer checks.Done()
		health = checkRPCHealth(ctx, client, cfg.SorobanRPCURL)
	}()
	go func() {
		defer checks.Done()
		latestLedger = checkRPCLatestLedger(ctx, client, cfg.SorobanRPCURL)
	}()
	go func() {
		defer checks.Done()
		network = checkRPCNetwork(ctx, client, cfg.SorobanRPCURL)
	}()
	checks.Wait()

	report := Report{
		Network: cfg.Network,
		Checks:  []Check{horizon.check, health, latestLedger, network.check},
	}

	if cfg.NetworkPassphrase != "" {
		if horizon.passphrase != "" {
			report.Checks = append(report.Checks, comparePassphrase(
				"Horizon network", cfg.HorizonURL, horizon.passphrase, cfg.NetworkPassphrase,
			))
		}
		if network.passphrase != "" {
			report.Checks = append(report.Checks, comparePassphrase(
				"Soroban RPC network", cfg.SorobanRPCURL, network.passphrase, cfg.NetworkPassphrase,
			))
		}
	}

	if horizon.passphrase != "" && network.passphrase != "" {
		report.Checks = append(report.Checks, comparePassphrase(
			"Horizon and Soroban network consistency",
			"",
			network.passphrase,
			horizon.passphrase,
		))
	} else {
		report.Checks = append(report.Checks, Check{
			Name:   "Horizon and Soroban network consistency",
			Status: "skipped",
			Message: "Could not compare network passphrases because one or both " +
				"network checks did not return a passphrase",
		})
	}

	report.OK = true
	for _, check := range report.Checks {
		if check.Status == "fail" {
			report.OK = false
			break
		}
	}
	return report, nil
}

func validateEndpoint(name, rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return fmt.Errorf("%s endpoint must be an absolute HTTP or HTTPS URL", name)
	}
	if parsed.User != nil {
		return fmt.Errorf("%s endpoint URL must not contain user credentials", name)
	}
	return nil
}

func checkHorizon(ctx context.Context, client *http.Client, endpoint string) endpointResult {
	started := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, horizonRootURL(endpoint), nil)
	if err != nil {
		return endpointResult{check: failedCheck("Horizon root", endpoint, started, err.Error())}
	}
	resp, err := client.Do(req)
	if err != nil {
		return endpointResult{check: failedCheck(
			"Horizon root", endpoint, started, transportError(err),
		)}
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return endpointResult{check: failedCheck(
			"Horizon root", endpoint, started, fmt.Sprintf("HTTP status %d", resp.StatusCode),
		)}
	}

	var body struct {
		NetworkPassphrase string `json:"network_passphrase"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxResponseSize)).Decode(&body); err != nil {
		return endpointResult{check: failedCheck(
			"Horizon root", endpoint, started, "could not decode Horizon root response: "+err.Error(),
		)}
	}
	if body.NetworkPassphrase == "" {
		return endpointResult{check: failedCheck(
			"Horizon root", endpoint, started, "response is missing network_passphrase",
		)}
	}

	return endpointResult{
		check: Check{
			Name:      "Horizon root",
			Status:    "ok",
			Endpoint:  displayEndpoint(endpoint),
			LatencyMS: time.Since(started).Milliseconds(),
			Message:   "Horizon returned its network passphrase",
		},
		passphrase: body.NetworkPassphrase,
	}
}

func checkRPCHealth(ctx context.Context, client *http.Client, endpoint string) Check {
	started := time.Now()
	result, err := callRPC(ctx, client, endpoint, "getHealth")
	if err != nil {
		return failedCheck("Soroban RPC health", endpoint, started, err.Error())
	}

	var health struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(result, &health); err != nil {
		return failedCheck("Soroban RPC health", endpoint, started, "invalid getHealth result: "+err.Error())
	}
	if health.Status != "healthy" {
		message := "getHealth did not report a healthy status"
		if health.Status != "" {
			message = fmt.Sprintf("getHealth reported status %q", health.Status)
		}
		return failedCheck("Soroban RPC health", endpoint, started, message)
	}

	return Check{
		Name:      "Soroban RPC health",
		Status:    "ok",
		Endpoint:  displayEndpoint(endpoint),
		LatencyMS: time.Since(started).Milliseconds(),
		Message:   "getHealth reported healthy",
	}
}

func checkRPCLatestLedger(ctx context.Context, client *http.Client, endpoint string) Check {
	started := time.Now()
	result, err := callRPC(ctx, client, endpoint, "getLatestLedger")
	if err != nil {
		return failedCheck("Soroban RPC latest ledger", endpoint, started, err.Error())
	}

	var latestLedger struct {
		Sequence uint32 `json:"sequence"`
	}
	if err := json.Unmarshal(result, &latestLedger); err != nil {
		return failedCheck(
			"Soroban RPC latest ledger",
			endpoint,
			started,
			"invalid getLatestLedger result: "+err.Error(),
		)
	}
	if latestLedger.Sequence == 0 {
		return failedCheck(
			"Soroban RPC latest ledger",
			endpoint,
			started,
			"getLatestLedger result is missing a positive sequence",
		)
	}

	return Check{
		Name:           "Soroban RPC latest ledger",
		Status:         "ok",
		Endpoint:       displayEndpoint(endpoint),
		LatencyMS:      time.Since(started).Milliseconds(),
		LedgerSequence: latestLedger.Sequence,
		Message:        fmt.Sprintf("getLatestLedger returned sequence %d", latestLedger.Sequence),
	}
}

func checkRPCNetwork(ctx context.Context, client *http.Client, endpoint string) endpointResult {
	started := time.Now()
	result, err := callRPC(ctx, client, endpoint, "getNetwork")
	if err != nil {
		return endpointResult{check: failedCheck("Soroban RPC network", endpoint, started, err.Error())}
	}

	var network struct {
		Passphrase string `json:"passphrase"`
	}
	if err := json.Unmarshal(result, &network); err != nil {
		return endpointResult{check: failedCheck(
			"Soroban RPC network", endpoint, started, "invalid getNetwork result: "+err.Error(),
		)}
	}
	if network.Passphrase == "" {
		return endpointResult{check: failedCheck(
			"Soroban RPC network", endpoint, started, "getNetwork result is missing passphrase",
		)}
	}

	return endpointResult{
		check: Check{
			Name:      "Soroban RPC network",
			Status:    "ok",
			Endpoint:  displayEndpoint(endpoint),
			LatencyMS: time.Since(started).Milliseconds(),
			Message:   "getNetwork returned its network passphrase",
		},
		passphrase: network.Passphrase,
	}
}

func callRPC(ctx context.Context, client *http.Client, endpoint, method string) (json.RawMessage, error) {
	requestBody, err := json.Marshal(struct {
		JSONRPC string `json:"jsonrpc"`
		ID      int    `json:"id"`
		Method  string `json:"method"`
	}{
		JSONRPC: "2.0",
		ID:      1,
		Method:  method,
	})
	if err != nil {
		return nil, fmt.Errorf("encode %s request: %w", method, err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		strings.TrimRight(endpoint, "/"),
		bytes.NewReader(requestBody),
	)
	if err != nil {
		return nil, fmt.Errorf("create %s request: %w", method, err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s request failed: %s", method, transportError(err))
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("%s returned HTTP status %d", method, resp.StatusCode)
	}

	var response struct {
		JSONRPC string          `json:"jsonrpc"`
		ID      int             `json:"id"`
		Result  json.RawMessage `json:"result"`
		Error   *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxResponseSize)).Decode(&response); err != nil {
		return nil, fmt.Errorf("decode %s response: %w", method, err)
	}
	if response.JSONRPC != "2.0" || response.ID != 1 {
		return nil, fmt.Errorf("%s response has an invalid JSON-RPC version or id", method)
	}
	if response.Error != nil {
		return nil, fmt.Errorf("%s RPC error %d: %s", method, response.Error.Code, response.Error.Message)
	}
	if len(response.Result) == 0 || string(response.Result) == "null" {
		return nil, fmt.Errorf("%s response is missing result", method)
	}
	return response.Result, nil
}

func comparePassphrase(name, endpoint, actual, expected string) Check {
	if actual != expected {
		return Check{
			Name:     name,
			Status:   "fail",
			Endpoint: displayEndpoint(endpoint),
			Message:  fmt.Sprintf("network passphrase mismatch (expected %q, got %q)", expected, actual),
		}
	}
	return Check{
		Name:     name,
		Status:   "ok",
		Endpoint: displayEndpoint(endpoint),
		Message:  "network passphrase matches",
	}
}

func failedCheck(name, endpoint string, started time.Time, message string) Check {
	return Check{
		Name:      name,
		Status:    "fail",
		Endpoint:  displayEndpoint(endpoint),
		LatencyMS: time.Since(started).Milliseconds(),
		Message:   message,
	}
}

func displayEndpoint(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	parsed.RawQuery = ""
	parsed.ForceQuery = false
	parsed.Fragment = ""
	parsed.RawFragment = ""
	return parsed.String()
}

func horizonRootURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/") + "/"
	parsed.RawPath = ""
	return parsed.String()
}

func transportError(err error) string {
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return urlErr.Err.Error()
	}
	return err.Error()
}
