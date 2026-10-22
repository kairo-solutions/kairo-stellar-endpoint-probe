package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	endpointprobe "github.com/kairo-solutions/kairo-stellar-endpoint-probe"
)

const (
	publicPassphrase  = "Public Global Stellar Network ; September 2015"
	testnetPassphrase = "Test SDF Network ; September 2015"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("kairo-stellar-endpoint-probe", flag.ContinueOnError)
	flags.SetOutput(stderr)
	horizonURL := flags.String("horizon", "", "Horizon base URL (required)")
	rpcURL := flags.String("soroban-rpc", "", "Soroban RPC base URL (required)")
	network := flags.String("network", "testnet", "Expected network: testnet, public, or custom")
	customPassphrase := flags.String("network-passphrase", "", "Expected passphrase when --network=custom")
	timeout := flags.Duration("timeout", 5*time.Second, "Timeout for each HTTP request")
	jsonOutput := flags.Bool("json", false, "Print the report as JSON")
	flags.Usage = func() {
		fmt.Fprintln(stderr, "Probe Horizon and Soroban RPC endpoint health and network identity.")
		fmt.Fprintln(stderr, "\nUsage:")
		fmt.Fprintln(stderr, "  kairo-stellar-endpoint-probe --horizon URL --soroban-rpc URL [options]")
		fmt.Fprintln(stderr, "\nOptions:")
		flags.PrintDefaults()
	}
	if err := flags.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}
	if *horizonURL == "" || *rpcURL == "" {
		flags.Usage()
		return 2
	}

	expectedPassphrase := ""
	switch *network {
	case "testnet":
		expectedPassphrase = testnetPassphrase
	case "public":
		expectedPassphrase = publicPassphrase
	case "custom":
		if *customPassphrase == "" {
			fmt.Fprintln(stderr, "--network-passphrase is required with --network=custom")
			return 2
		}
		expectedPassphrase = *customPassphrase
	default:
		fmt.Fprintln(stderr, "--network must be testnet, public, or custom")
		return 2
	}

	report, err := endpointprobe.Probe(context.Background(), endpointprobe.Config{
		HorizonURL:        *horizonURL,
		SorobanRPCURL:     *rpcURL,
		Network:           *network,
		NetworkPassphrase: expectedPassphrase,
		Timeout:           *timeout,
	})
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}

	if *jsonOutput {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			fmt.Fprintln(stderr, "encode report:", err)
			return 2
		}
	} else {
		status := "OK"
		if !report.OK {
			status = "FAILED"
		}
		fmt.Fprintf(stdout, "%s: %s network\n", status, report.Network)
		for _, check := range report.Checks {
			if check.Endpoint != "" {
				fmt.Fprintf(stdout, "- %-42s %-7s %s (%d ms) - %s\n",
					check.Name, check.Status, check.Endpoint, check.LatencyMS, check.Message)
			} else {
				fmt.Fprintf(stdout, "- %-42s %-7s %s\n", check.Name, check.Status, check.Message)
			}
		}
	}
	if !report.OK {
		return 1
	}
	return 0
}
